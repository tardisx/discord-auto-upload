package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/tardisx/discord-auto-upload/config"
	"github.com/tardisx/discord-auto-upload/upload"
)

func TestWatchNewFiles(t *testing.T) {
	dir := createFileTree()
	defer os.RemoveAll(dir)
	time.Sleep(time.Second)

	w := watch{
		config:       config.Watcher{Path: dir},
		uploader:     upload.Uploader{},
		lastCheck:    time.Now(),
		newLastCheck: time.Now(),
	}
	files := w.ProcessNewFiles()
	if len(files) != 0 {
		t.Errorf("was not zero files (%d): %v", len(files), files)
	}

	// create a new file
	time.Sleep(time.Second)
	os.Create(fmt.Sprintf("%s%c%s", dir, os.PathSeparator, "b.gif"))
	files = w.ProcessNewFiles()
	if len(files) != 1 {
		t.Errorf("was not one file - got: %v", files)
	}
	if files[0] != fmt.Sprintf("%s%c%s", dir, os.PathSeparator, "b.gif") {
		t.Error("wrong file")
	}
}

func TestExclsion(t *testing.T) {
	dir := createFileTree()
	defer os.RemoveAll(dir)
	time.Sleep(time.Second)

	w := watch{
		config:       config.Watcher{Path: dir, Exclude: []string{"thumb", "tiny"}},
		uploader:     upload.Uploader{},
		lastCheck:    time.Now(),
		newLastCheck: time.Now(),
	}
	files := w.ProcessNewFiles()
	if len(files) != 0 {
		t.Errorf("was not zero files (%d): %v", len(files), files)
	}
	// create a new file that would not hit exclusion, and two that would
	os.Create(fmt.Sprintf("%s%c%s", dir, os.PathSeparator, "b.gif"))
	os.Create(fmt.Sprintf("%s%c%s", dir, os.PathSeparator, "b_thumb.gif"))
	os.Create(fmt.Sprintf("%s%c%s", dir, os.PathSeparator, "tiny_b.jpg"))
	files = w.ProcessNewFiles()
	if len(files) != 1 {
		t.Error("was not one new file")
	}

}

func TestCheckPath(t *testing.T) {
	dir := createFileTree()
	defer os.RemoveAll(dir)

	w := watch{
		config:       config.Watcher{Path: dir},
		uploader:     upload.Uploader{},
		lastCheck:    time.Now(),
		newLastCheck: time.Now(),
	}
	if !w.checkPath() {
		t.Error("checkPath failed?")
	}
	os.RemoveAll(dir)
	if w.checkPath() {
		t.Error("checkPath succeeded when shouldn't?")
	}
}

func createFileTree() string {
	dir, err := ioutil.TempDir("", "dau-test")
	if err != nil {
		log.Fatal(err)
	}
	os.Create(fmt.Sprintf("%s%c%s", dir, os.PathSeparator, "a.gif"))
	os.Create(fmt.Sprintf("%s%c%s", dir, os.PathSeparator, "a.jpg"))
	os.Create(fmt.Sprintf("%s%c%s", dir, os.PathSeparator, "a.png"))

	return dir
}
