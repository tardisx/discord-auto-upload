package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
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

	// "github.com/skratchdot/open-golang/open"

	"github.com/tardisx/discord-auto-upload/config"
	daulog "github.com/tardisx/discord-auto-upload/log"
	"github.com/tardisx/discord-auto-upload/upload"

	// "github.com/tardisx/discord-auto-upload/upload"
	"github.com/tardisx/discord-auto-upload/version"
	"github.com/tardisx/discord-auto-upload/web"
)

type watch struct {
	lastCheck    time.Time
	newLastCheck time.Time
	config       config.Watcher
	uploader     *upload.Uploader
}

func main() {

	parseOptions()

	// grab the config, register to notice changes
	config := config.DefaultConfigService()
	configChanged := make(chan bool)
	config.Changed = configChanged
	config.LoadOrInit()

	// create the uploader
	up := upload.NewUploader()

	// log.Print("Opening web browser")
	// open.Start("http://localhost:9090")
	web := web.WebService{Config: config, Uploader: up}
	web.StartWebServer()

	go func() { checkUpdates() }()

	// create the watchers, restart them if config changes
	// blocks forever
	startWatchers(config, up, configChanged)

}

func startWatchers(config *config.ConfigService, up *upload.Uploader, configChange chan bool) {
	for {
		log.Printf("Creating watchers")
		ctx, cancel := context.WithCancel(context.Background())
		for _, c := range config.Config.Watchers {
			log.Printf("Creating watcher for %s interval %d", c.Path, config.Config.WatchInterval)
			watcher := watch{uploader: up, lastCheck: time.Now(), newLastCheck: time.Now(), config: c}
			go watcher.Watch(config.Config.WatchInterval, ctx)
		}
		// wait for single that the config changed
		<-configChange
		cancel()
		log.Printf("starting new watchers due to config change")
	}

}

func (w *watch) Watch(interval int, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Killing old watcher")
			return
		default:
			newFiles := w.ProcessNewFiles()
			for _, f := range newFiles {
				w.uploader.AddFile(f, w.config)
			}
			// upload them
			w.uploader.Upload()
			daulog.SendLog(fmt.Sprintf("sleeping for %ds before next check of %s", interval, w.config.Path), daulog.LogTypeDebug)
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
}

// ProcessNewFiles returns an array of new files that have appeared since
// the last time ProcessNewFiles was run.
func (w *watch) ProcessNewFiles() []string {
	var newFiles []string
	// check the path each time around, in case it goes away or something
	if w.checkPath() {
		// walk the path
		err := filepath.WalkDir(w.config.Path,
			func(path string, d fs.DirEntry, err error) error {
				return w.checkFile(path, &newFiles, w.config.Exclude)
			})

		if err != nil {
			log.Fatal("could not watch path", err)
		}
		w.lastCheck = w.newLastCheck
	}
	return newFiles
}

// checkPath makes sure the path exists, and is a directory.
// It logs errors if there are problems, and returns false
func (w *watch) checkPath() bool {
	src, err := os.Stat(w.config.Path)
	if err != nil {
		log.Printf("Problem with path '%s': %s", w.config.Path, err)
		return false
	}
	if !src.IsDir() {
		log.Printf("Problem with path '%s': is not a directory", w.config.Path)
		return false
	}
	return true
}

// checkFile checks if a file is eligible, first looking at extension (to
// avoid statting files uselessly) then modification times.
// If the file is eligible, not excluded and new enough to care we add it
// to the passed in array of files
func (w *watch) checkFile(path string, found *[]string, exclusions []string) error {

	extension := strings.ToLower(filepath.Ext(path))

	if !(extension == ".png" || extension == ".jpg" || extension == ".gif") {
		return nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fi.ModTime().After(w.lastCheck) && fi.Mode().IsRegular() {
		excluded := false
		for _, exclusion := range exclusions {
			if strings.Contains(path, exclusion) {
				excluded = true
			}
		}
		if !excluded {
			*found = append(*found, path)
		}
	}

	if w.newLastCheck.Before(fi.ModTime()) {
		w.newLastCheck = fi.ModTime()
	}

	return nil
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
	var versionFlag bool
	flag.BoolVar(&versionFlag, "version", false, "show version")
	flag.Parse()

	if versionFlag {
		fmt.Println("dau - https://github.com/tardisx/discord-auto-upload")
		fmt.Printf("Version: %s\n", version.CurrentVersion)
		os.Exit(0)
	}

}
