package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
)

// Config for the application
var Config struct {
	WebHookURL  string
	Path        string
	Watch       int
	Username    string
	NoWatermark bool
	Exclude     string
}

const CurrentVersion string = "0.8"

// Load the current config or initialise with defaults
func LoadOrInit() {
	configPath := configPath()
	log.Printf("Trying to load from %s\n", configPath)
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		log.Printf("NOTE: No config file, writing out sample configuration")
		log.Printf("You need to set the configuration via the web interface")

		Config.WebHookURL = ""
		Config.Path = homeDir() + string(os.PathSeparator) + "screenshots"
		Config.Watch = 10
		SaveConfig()
	} else {
		LoadConfig()
	}
}

func LoadConfig() {
	path := configPath()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("cannot read config file %s: %s", path, err.Error())
	}
	err = json.Unmarshal([]byte(data), &Config)
	if err != nil {
		log.Fatalf("cannot decode config file %s: %s", path, err.Error())
	}
}

func SaveConfig() {
	log.Print("saving configuration")
	path := configPath()
	jsonString, _ := json.Marshal(Config)
	err := ioutil.WriteFile(path, jsonString, os.ModePerm)
	if err != nil {
		log.Fatalf("Cannot save config %s: %s", path, err.Error())
	}
}

func homeDir() string {
	dir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return dir
}

func configPath() string {
	homeDir := homeDir()
	return homeDir + string(os.PathSeparator) + ".dau.json"
}
