package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
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

	go func() {
		version.GetOnlineVersion()
		if version.UpdateAvailable() {
			daulog.Info("*** NEW VERSION AVAILABLE ***")
			daulog.Infof("You are currently on version %s, but version %s is available\n", version.CurrentVersion, version.LatestVersionInfo.TagName)
			daulog.Info("----------- Release Info -----------")
			daulog.Info(version.LatestVersionInfo.Body)
			daulog.Info("------------------------------------")
			daulog.Info("Upgrade at https://github.com/tardisx/discord-auto-upload/releases/latest")
		}
	}()

	// create the watchers, restart them if config changes
	// blocks forever
	startWatchers(config, up, configChanged)

}

func startWatchers(config *config.ConfigService, up *upload.Uploader, configChange chan bool) {
	for {
		daulog.Debug("Creating watchers")
		ctx, cancel := context.WithCancel(context.Background())
		for _, c := range config.Config.Watchers {
			daulog.Infof("Creating watcher for %s with interval %d", c.Path, config.Config.WatchInterval)
			watcher := watch{uploader: up, lastCheck: time.Now(), newLastCheck: time.Now(), config: c}
			go watcher.Watch(config.Config.WatchInterval, ctx)
		}
		// wait for single that the config changed
		<-configChange
		cancel()
		daulog.Info("starting new watchers due to config change")
	}

}

func (w *watch) Watch(interval int, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			daulog.Info("Killing old watcher")
			return
		default:
			newFiles := w.ProcessNewFiles()
			for _, f := range newFiles {
				w.uploader.AddFile(f, w.config)
			}
			// upload them
			w.uploader.Upload()
			daulog.Debugf("sleeping for %ds before next check of %s", interval, w.config.Path)
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
		daulog.Errorf("Problem with path '%s': %s", w.config.Path, err)
		return false
	}
	if !src.IsDir() {
		daulog.Errorf("Problem with path '%s': is not a directory", w.config.Path)
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
