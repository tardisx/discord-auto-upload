package config

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestNoConfig(t *testing.T) {
	if Config.Version != 0 {
		t.Error("not 0 empty config")
	}

	configPath = emptyTempFile()
	os.Remove(configPath)

	err := LoadOrInit()
	if err != nil {
		t.Errorf("unexpected failure from load: %s", err)
	}

	if Config.Version != 2 {
		t.Error("not version 2 starting config")
	}

	if fileSize(configPath) < 40 {
		t.Errorf("File is too small %d bytes", fileSize(configPath))
	}

	os.Remove(configPath)
}

func TestEmptyFileConfig(t *testing.T) {

	configPath = emptyTempFile()

	err := LoadOrInit()
	if err == nil {
		t.Error("unexpected success from LoadOrInit()")
	}

	os.Remove(configPath)
}

func emptyTempFile() string {
	f, err := ioutil.TempFile("", "dautest-*")
	if err != nil {
		panic(err)
	}
	return f.Name()
}

func fileSize(file string) int {
	fi, err := os.Stat(file)
	if err != nil {
		panic(err)
	}
	return int(fi.Size())

}
