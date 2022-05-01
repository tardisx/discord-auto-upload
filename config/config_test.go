package config

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestNoConfig(t *testing.T) {
	c := ConfigService{}

	c.ConfigFilename = emptyTempFile()
	err := os.Remove(c.ConfigFilename)
	if err != nil {
		t.Fatalf("could not remove file: %v", err)
	}

	defer os.Remove(c.ConfigFilename) // because we are about to create it

	err = c.LoadOrInit()
	if err != nil {
		t.Errorf("unexpected failure from load: %s", err)
	}

	if c.Config.Version != 3 {
		t.Error("not version 3 starting config")
	}

	if fileSize(c.ConfigFilename) < 40 {
		t.Errorf("File is too small %d bytes", fileSize(c.ConfigFilename))
	}

}

func TestEmptyFileConfig(t *testing.T) {
	c := ConfigService{}

	c.ConfigFilename = emptyTempFile()
	defer os.Remove(c.ConfigFilename)

	err := c.LoadOrInit()
	if err == nil {
		t.Error("unexpected success from LoadOrInit()")
	}

}

func TestMigrateFromV1toV3(t *testing.T) {
	c := ConfigService{}

	c.ConfigFilename = v1Config()
	err := c.LoadOrInit()
	if err != nil {
		t.Error("unexpected error from LoadOrInit()")
	}
	if c.Config.Version != 3 {
		t.Errorf("Version %d not 3", c.Config.Version)
	}

	if c.Config.OpenBrowserOnStart != true {
		t.Errorf("Open browser on start not true")
	}

	if len(c.Config.Watchers) != 1 {
		t.Error("wrong amount of watchers")
	}

	if c.Config.Watchers[0].Path != "/private/tmp" {
		t.Error("Wrong path")
	}
	if c.Config.WatchInterval != 69 {
		t.Error("Wrong watch interval")
	}
	if c.Config.Port != 9090 {
		t.Error("Wrong port")
	}
}

func v1Config() string {
	f, err := ioutil.TempFile("", "dautest-*")
	if err != nil {
		panic(err)
	}
	config := `{"WebHookURL":"https://discord.com/api/webhooks/abc123","Path":"/private/tmp","Watch":69,"Username":"abcdedf","NoWatermark":true,"Exclude":"ab cd ef"}`
	f.Write([]byte(config))
	defer f.Close()
	return f.Name()
}

func emptyTempFile() string {
	f, err := ioutil.TempFile("", "dautest-*")
	if err != nil {
		panic(err)
	}
	f.Close()
	return f.Name()
}

func fileSize(file string) int {
	fi, err := os.Stat(file)
	if err != nil {
		panic(err)
	}
	return int(fi.Size())

}
