package config

// Config for the application
var Config struct {
	WebHookURL  string
	Path        string
	Watch       int
	Username    string
	NoWatermark bool
	Exclude     string
}

const CurrentVersion string = "0.6"
