package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	daulog "github.com/tardisx/discord-auto-upload/log"

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

const CurrentVersion string = "0.10"

// Load the current config or initialise with defaults
func LoadOrInit() {
	configPath := configPath()
	daulog.SendLog(fmt.Sprintf("Trying to load config from %s", configPath), daulog.LogTypeDebug)
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		daulog.SendLog("NOTE: No config file, writing out sample configuration", daulog.LogTypeInfo)
		daulog.SendLog("You need to set the configuration via the web interface", daulog.LogTypeInfo)

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
	daulog.SendLog("saving configuration", daulog.LogTypeInfo)
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
