package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/pborman/getopt"

	// "github.com/skratchdot/open-golang/open"

	"github.com/tardisx/discord-auto-upload/config"
	daulog "github.com/tardisx/discord-auto-upload/log"
	"github.com/tardisx/discord-auto-upload/uploads"
	"github.com/tardisx/discord-auto-upload/version"
	"github.com/tardisx/discord-auto-upload/web"
)

var lastCheck = time.Now()
var newLastCheck = time.Now()

func main() {

	parseOptions()

	// log.Print("Opening web browser")
	// open.Start("http://localhost:9090")
	web.StartWebServer()

	checkUpdates()

	daulog.SendLog(fmt.Sprintf("Waiting for images to appear in %s", config.Config.Path), daulog.LogTypeInfo)
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
		daulog.SendLog(fmt.Sprintf("sleeping for %ds before next check of %s", config.Config.Watch, config.Config.Path), daulog.LogTypeDebug)
		time.Sleep(time.Duration(config.Config.Watch) * time.Second)
	}
}

func checkPath(path string) bool {
	src, err := os.Stat(path)
	if err != nil {
		log.Printf("Problem with path '%s': %s", path, err)
		return false
	}
	if !src.IsDir() {
		log.Printf("Problem with path '%s': is not a directory", path)
		return false
	}
	return true
}

func checkUpdates() {

	type GithubRelease struct {
		HTMLURL string `json:"html_url"`
		TagName string `json:"tag_name"`
		Name    string `json:"name"`
		Body    string `json:"body"`
	}

	daulog.SendLog("checking for new version", daulog.LogTypeInfo)

	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Get("https://api.github.com/repos/tardisx/discord-auto-upload/releases/latest")
	if err != nil {
		daulog.SendLog(fmt.Sprintf("WARNING: Update check failed: %v", err), daulog.LogTypeError)
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

	// pre v0.11.0 version (ie before semver) did a simple string comparison,
	// but since "0.10.0" < "v0.11.0" they should still get prompted to upgrade
	// ok
	if version.NewVersionAvailable(latest.TagName) {
		fmt.Printf("You are currently on version %s, but version %s is available\n", version.CurrentVersion, latest.TagName)
		fmt.Println("----------- Release Info -----------")
		fmt.Println(latest.Body)
		fmt.Println("------------------------------------")
		fmt.Println("Upgrade at https://github.com/tardisx/discord-auto-upload/releases/latest")
		daulog.SendLog(fmt.Sprintf("New version available: %s - download at https://github.com/tardisx/discord-auto-upload/releases/latest", latest.TagName), daulog.LogTypeInfo)
	}

	daulog.SendLog("already running latest version", daulog.LogTypeInfo)

}

func parseOptions() {

	// Declare the flags to be used
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
		fmt.Printf("Version: %s\n", version.CurrentVersion)
		os.Exit(0)
	}

	// grab the config
	config.LoadOrInit()
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

	daulog.SendLog("Sending to uploader", daulog.LogTypeInfo)
	uploads.AddFile(file)
}
