package main

//go:generate go-bindata -pkg assets -o assets/static.go -prefix data/ data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"image"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/fogleman/gg"
	"github.com/pborman/getopt"
	// "github.com/skratchdot/open-golang/open"
	"golang.org/x/image/font/inconsolata"

	"github.com/tardisx/discord-auto-upload/config"
	"github.com/tardisx/discord-auto-upload/web"
)

var lastCheck = time.Now()
var newLastCheck = time.Now()

func main() {

	parseOptions()

	// log.Print("Opening web browser")
	// open.Start("http://localhost:9090")
	go web.StartWebServer()

	checkUpdates()

	log.Print("Waiting for images to appear in ", config.Config.Path)
	// wander the path, forever
	for {
		if checkPath(config.Config.Path) {
			err := filepath.Walk(config.Config.Path,
				func(path string, f os.FileInfo, err error) error { return checkFile(path, f, err) })
			if err != nil {
				log.Fatal("could not watch path", err)
			}
			lastCheck = newLastCheck
		}
		log.Print("sleeping before next check")
		time.Sleep(time.Duration(config.Config.Watch) * time.Second)
	}
}

func checkPath(path string) bool {
	src, err := os.Stat(path)
	if err != nil {
		log.Println("path problem: ", err)
		return false
	}
	if !src.IsDir() {
		log.Println(path, " is not a directory")
		return false
	}
	return true
}

func checkUpdates() {

	type GithubRelease struct {
		HTMLURL string
		TagName string
		Name    string
		Body    string
	}

	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Get("https://api.github.com/repos/tardisx/discord-auto-upload/releases/latest")
	if err != nil {
		log.Print("WARNING: Update check failed: ", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("could not check read update response")
	}

	var latest GithubRelease
	err = json.Unmarshal(body, &latest)

	if err != nil {
		log.Fatal("could not parse JSON: ", err)
	}

	if config.CurrentVersion < latest.TagName {
		fmt.Printf("You are currently on version %s, but version %s is available\n", config.CurrentVersion, latest.TagName)
		fmt.Println("----------- Release Info -----------")
		fmt.Println(latest.Body)
		fmt.Println("------------------------------------")
		fmt.Println("Upgrade at https://github.com/tardisx/discord-auto-upload/releases/latest")
	}

}

func parseOptions() {

	// Declare the flags to be used
	excludeFlag := getopt.StringLong("exclude", 'x', "", "exclude files containing this string")
	helpFlag := getopt.BoolLong("help", 'h', "help")
	versionFlag := getopt.BoolLong("version", 'v', "show version")
	getopt.SetParameters("")

	getopt.Parse()

	if *helpFlag {
		getopt.PrintUsage(os.Stderr)
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Println("dau - https://github.com/tardisx/discord-auto-upload")
		fmt.Printf("Version: %s\n", config.CurrentVersion)
		os.Exit(0)
	}

	// if !getopt.IsSet("webhook") {
	// 	log.Fatal("ERROR: You must specify a --webhook URL")
	// }

	// grab the config
	config.LoadOrInit()

	// overrides from command line
	config.Config.Exclude = *excludeFlag
}

func checkFile(path string, f os.FileInfo, err error) error {

	if f.ModTime().After(lastCheck) && f.Mode().IsRegular() {

		if fileEligible(path) {
			// process file
			processFile(path)
		}

		if newLastCheck.Before(f.ModTime()) {
			newLastCheck = f.ModTime()
		}
	}

	return nil
}

func fileEligible(file string) bool {

	if config.Config.Exclude != "" && strings.Contains(file, config.Config.Exclude) {
		return false
	}

	extension := strings.ToLower(filepath.Ext(file))
	if extension == ".png" || extension == ".jpg" || extension == ".gif" {
		return true
	}

	return false
}

func processFile(file string) {

	if !config.Config.NoWatermark {
		log.Print("Copying to temp location and watermarking ", file)
		file = mungeFile(file)
	}

	if config.Config.WebHookURL == "" {
		log.Print("WebHookURL is not configured - cannot upload!")
		return
	}

	log.Print("Uploading ", file)

	extraParams := map[string]string{}

	if config.Config.Username != "" {
		log.Print("Overriding username with " + config.Config.Username)
		extraParams["username"] = config.Config.Username
	}

	type DiscordAPIResponseAttachment struct {
		URL      string
		ProxyURL string
		Size     int
		Width    int
		Height   int
		Filename string
	}

	type DiscordAPIResponse struct {
		Attachments []DiscordAPIResponseAttachment
		ID          int64 `json:",string"`
	}

	var retriesRemaining = 5
	for retriesRemaining > 0 {

		request, err := newfileUploadRequest(config.Config.WebHookURL, extraParams, "file", file)
		if err != nil {
			log.Fatal(err)
		}
		start := time.Now()
		client := &http.Client{Timeout: time.Second * 30}
		resp, err := client.Do(request)
		if err != nil {
			log.Print("Error performing request:", err)
			retriesRemaining--
			sleepForRetries(retriesRemaining)
			continue
		} else {

			if resp.StatusCode != 200 {
				log.Print("Bad response from server:", resp.StatusCode)
				if b, err := ioutil.ReadAll(resp.Body); err == nil {
					log.Print("Body:", string(b))
				}
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}

			resBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Print("could not deal with body: ", err)
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}
			resp.Body.Close()

			var res DiscordAPIResponse
			err = json.Unmarshal(resBody, &res)

			if err != nil {
				log.Print("could not parse JSON: ", err)
				fmt.Println("Response was:", string(resBody[:]))
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}
			if len(res.Attachments) < 1 {
				log.Print("bad response - no attachments?")
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}
			var a = res.Attachments[0]
			elapsed := time.Since(start)
			rate := float64(a.Size) / elapsed.Seconds() / 1024.0

			log.Printf("Uploaded to %s %dx%d", a.URL, a.Width, a.Height)
			log.Printf("id: %d, %d bytes transferred in %.2f seconds (%.2f KiB/s)", res.ID, a.Size, elapsed.Seconds(), rate)
			break
		}
	}

	if !config.Config.NoWatermark {
		log.Print("Removing temporary file ", file)
		os.Remove(file)
	}

	if retriesRemaining == 0 {
		log.Fatal("Failed to upload, even after retries")
	}
}

func sleepForRetries(retry int) {
	if retry == 0 {
		return
	}
	retryTime := (6-retry)*(6-retry) + 6
	log.Printf("Will retry in %d seconds (%d remaining attempts)", retryTime, retry)
	// time.Sleep(time.Duration(retryTime) * time.Second)
}

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatal("Could not copy: ", err)
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func mungeFile(path string) string {

	reader, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	im, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	bounds := im.Bounds()
	// var S float64 = float64(bounds.Max.X)

	dc := gg.NewContext(bounds.Max.X, bounds.Max.Y)
	dc.Clear()
	dc.SetRGB(0, 0, 0)

	dc.SetFontFace(inconsolata.Regular8x16)

	dc.DrawImage(im, 0, 0)

	dc.DrawRoundedRectangle(0, float64(bounds.Max.Y-18.0), 320, float64(bounds.Max.Y), 0)
	dc.SetRGB(0, 0, 0)
	dc.Fill()

	dc.SetRGB(1, 1, 1)

	dc.DrawString("github.com/tardisx/discord-auto-upload", 5.0, float64(bounds.Max.Y)-5.0)

	tempfile, err := ioutil.TempFile("", "dau")
	if err != nil {
		log.Fatal(err)
	}
	tempfile.Close()
	os.Remove(tempfile.Name())
	actualName := tempfile.Name() + ".png"

	dc.SavePNG(actualName)
	return actualName
}
