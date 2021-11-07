package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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

type Watcher struct {
	WebHookURL  string
	Path        string
	Username    string
	NoWatermark bool
	HoldUploads bool
	Exclude     []string
}

type ConfigV2 struct {
	WatchInterval int
	Version       int
	Port          int
	Watchers      []Watcher
}

type ConfigService struct {
	Config         *ConfigV2
	Changed        chan bool
	ConfigFilename string
}

func DefaultConfigService() *ConfigService {
	c := ConfigService{
		ConfigFilename: defaultConfigPath(),
	}
	return &c
}

// LoadOrInit loads the current configuration from the config file, or creates
// a new config file if none exists.
func (c *ConfigService) LoadOrInit() error {
	daulog.SendLog(fmt.Sprintf("Trying to load config from %s\n", c.ConfigFilename), daulog.LogTypeDebug)
	_, err := os.Stat(c.ConfigFilename)
	if os.IsNotExist(err) {
		daulog.SendLog("NOTE: No config file, writing out sample configuration", daulog.LogTypeInfo)
		daulog.SendLog("You need to set the configuration via the web interface", daulog.LogTypeInfo)
		c.Config = DefaultConfig()
		return c.Save()
	} else {
		return c.Load()
	}
}

func DefaultConfig() *ConfigV2 {
	c := ConfigV2{}
	c.Version = 2
	c.WatchInterval = 10
	c.Port = 9090
	w := Watcher{
		WebHookURL:  "https://webhook.url.here",
		Path:        "/your/screenshot/dir/here",
		Username:    "",
		NoWatermark: false,
		Exclude:     []string{},
	}
	c.Watchers = []Watcher{w}
	return &c
}

// Load will load the configuration from a known-to-exist config file.
func (c *ConfigService) Load() error {
	fmt.Printf("Loading from %s\n\n", c.ConfigFilename)

	data, err := ioutil.ReadFile(c.ConfigFilename)
	if err != nil {
		return fmt.Errorf("cannot read config file %s: %s", c.ConfigFilename, err.Error())
	}
	err = json.Unmarshal([]byte(data), &c.Config)
	if err != nil {
		return fmt.Errorf("cannot decode config file %s: %s", c.ConfigFilename, err.Error())
	}

	// Version 0 predates config migrations
	if c.Config.Version == 0 {
		// need to migrate this
		daulog.SendLog("Migrating config to V2", daulog.LogTypeInfo)

		configV1 := ConfigV1{}
		err = json.Unmarshal([]byte(data), &configV1)
		if err != nil {
			return fmt.Errorf("cannot decode legacy config file as v1 %s: %s", c.ConfigFilename, err.Error())
		}

		// copy stuff across
		c.Config.Version = 2
		c.Config.WatchInterval = configV1.Watch
		c.Config.Port = 9090 // this never used to be configurable

		onlyWatcher := Watcher{
			WebHookURL:  configV1.WebHookURL,
			Path:        configV1.Path,
			Username:    configV1.Username,
			NoWatermark: configV1.NoWatermark,
			Exclude:     strings.Split(configV1.Exclude, " "),
		}

		c.Config.Watchers = []Watcher{onlyWatcher}
	}

	return nil
}

func (c *ConfigService) Save() error {
	daulog.SendLog("saving configuration", daulog.LogTypeInfo)
	// sanity checks
	for _, watcher := range c.Config.Watchers {

		// give the sample one a pass? this is kinda gross...
		if watcher.Path == "/your/screenshot/dir/here" {
			continue
		}
		info, err := os.Stat(watcher.Path)
		if os.IsNotExist(err) {
			return fmt.Errorf("path '%s' does not exist", watcher.Path)
		}
		if !info.IsDir() {
			return fmt.Errorf("path '%s' is not a directory", watcher.Path)
		}
	}

	for _, watcher := range c.Config.Watchers {
		if strings.Index(watcher.WebHookURL, "https://") != 0 {
			return fmt.Errorf("webhook URL '%s' does not look valid", watcher.WebHookURL)
		}
	}

	if c.Config.WatchInterval < 1 {
		return fmt.Errorf("watch interval should be greater than 0 - '%d' invalid", c.Config.WatchInterval)
	}

	jsonString, _ := json.Marshal(c.Config)
	err := ioutil.WriteFile(c.ConfigFilename, jsonString, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot save config %s: %s", c.ConfigFilename, err.Error())
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
