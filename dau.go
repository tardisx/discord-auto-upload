package main

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
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/image/font/inconsolata"

	"discord-auto-upload/web"
)

const currentVersion = "0.7"

var lastCheck = time.Now()
var newLastCheck = time.Now()

// Config for the application
type Config struct {
	webhookURL  string
	path        string
	watch       int
	username    string
	noWatermark bool
	exclude     string
}

func main() {

	config := parseOptions()
	checkPath(config.path)
	wconfig := web.Init()
	go processWebChanges(wconfig)

	log.Print("Opening web browser")
	open.Start("http://localhost:9090")

	checkUpdates()

	log.Print("Waiting for images to appear in ", config.path)
	// wander the path, forever
	for {
		err := filepath.Walk(config.path,
			func(path string, f os.FileInfo, err error) error { return checkFile(path, f, err, config) })
		if err != nil {
			log.Fatal("could not watch path", err)
		}
		lastCheck = newLastCheck
		time.Sleep(time.Duration(config.watch) * time.Second)
	}
}

func processWebChanges(wc web.DAUWebServer) {
	for {
		change := <-wc.ConfigChange
		log.Print(change)
		log.Print("Got a change!")
	}
}

func checkPath(path string) {
	src, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	if !src.IsDir() {
		log.Fatal(path, " is not a directory")
		os.Exit(1)
	}
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

	if currentVersion < latest.TagName {
		fmt.Printf("You are currently on version %s, but version %s is available\n", currentVersion, latest.TagName)
		fmt.Println("----------- Release Info -----------")
		fmt.Println(latest.Body)
		fmt.Println("------------------------------------")
		fmt.Println("Upgrade at https://github.com/tardisx/discord-auto-upload/releases/latest")
	}

}

func parseOptions() Config {

	var newConfig Config
	// Declare the flags to be used
	webhookFlag := getopt.StringLong("webhook", 'w', "", "discord webhook URL")
	pathFlag := getopt.StringLong("directory", 'd', "", "directory to scan, optional, defaults to current directory")
	watchFlag := getopt.Int16Long("watch", 's', 10, "time between scans")
	usernameFlag := getopt.StringLong("username", 'u', "", "username for the bot upload")
	excludeFlag := getopt.StringLong("exclude", 'x', "", "exclude files containing this string")
	noWatermarkFlag := getopt.BoolLong("no-watermark", 'n', "do not put a watermark on images before uploading")
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
		fmt.Printf("Version: %s\n", currentVersion)
		os.Exit(0)
	}

	if !getopt.IsSet("directory") {
		*pathFlag = "./"
		log.Println("Defaulting to current directory")
	}

	if !getopt.IsSet("webhook") {
		log.Fatal("ERROR: You must specify a --webhook URL")
	}

	newConfig.path = *pathFlag
	newConfig.webhookURL = *webhookFlag
	newConfig.watch = int(*watchFlag)
	newConfig.username = *usernameFlag
	newConfig.noWatermark = *noWatermarkFlag
	newConfig.exclude = *excludeFlag

	return newConfig
}

func checkFile(path string, f os.FileInfo, err error, config Config) error {

	if f.ModTime().After(lastCheck) && f.Mode().IsRegular() {

		if fileEligible(config, path) {
			// process file
			processFile(config, path)
		}

		if newLastCheck.Before(f.ModTime()) {
			newLastCheck = f.ModTime()
		}
	}

	return nil
}

func fileEligible(config Config, file string) bool {

	if config.exclude != "" && strings.Contains(file, config.exclude) {
		return false
	}

	extension := strings.ToLower(filepath.Ext(file))
	if extension == ".png" || extension == ".jpg" || extension == ".gif" {
		return true
	}

	return false
}

func processFile(config Config, file string) {

	if !config.noWatermark {
		log.Print("Copying to temp location and watermarking ", file)
		file = mungeFile(file)
	}

	log.Print("Uploading ", file)

	extraParams := map[string]string{}

	if config.username != "" {
		extraParams["username"] = config.username
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
		request, err := newfileUploadRequest(config.webhookURL, extraParams, "file", file)
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

	if !config.noWatermark {
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
