package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	daulog "github.com/tardisx/discord-auto-upload/log"

	"github.com/mitchellh/go-homedir"
)

// Config for the application
type ConfigV1 struct {
	WebHookURL  string
	Path        string
	Watch       int
	Username    string
	NoWatermark bool
	Exclude     string
}

type ConfigV2Watcher struct {
	WebHookURL  string
	Path        string
	Username    string
	NoWatermark bool
	Exclude     string
}

type ConfigV2 struct {
	WatchInterval int
	Version       int
	Watchers      []ConfigV2Watcher
}

var Config ConfigV2
var configPath string

func Init() {
	configPath = defaultConfigPath()
}

// LoadOrInit loads the current configuration from the config file, or creates
// a new config file if none exists.
func LoadOrInit() error {
	daulog.SendLog(fmt.Sprintf("Trying to load config from %s", configPath), daulog.LogTypeDebug)
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		daulog.SendLog("NOTE: No config file, writing out sample configuration", daulog.LogTypeInfo)
		daulog.SendLog("You need to set the configuration via the web interface", daulog.LogTypeInfo)
		Config.Version = 2
		Config.WatchInterval = 10
		return SaveConfig()
	} else {
		return LoadConfig()
	}
}

// LoadConfig will load the configuration from a known-to-exist config file.
func LoadConfig() error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("cannot read config file %s: %s", configPath, err.Error())
	}
	err = json.Unmarshal([]byte(data), &Config)
	if err != nil {
		return fmt.Errorf("cannot decode config file %s: %s", configPath, err.Error())
	}
	return nil
}

func SaveConfig() error {
	daulog.SendLog("saving configuration", daulog.LogTypeInfo)
	jsonString, _ := json.Marshal(Config)
	err := ioutil.WriteFile(configPath, jsonString, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot save config %s: %s", configPath, err.Error())
	}
	return nil
}

func homeDir() string {
	dir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return dir
}

func defaultConfigPath() string {
	homeDir := homeDir()
	return homeDir + string(os.PathSeparator) + ".dau.json"
}
