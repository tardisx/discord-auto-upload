package version

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	daulog "github.com/tardisx/discord-auto-upload/log"

	"golang.org/x/mod/semver"
)

const CurrentVersion string = "v0.12.4"

type GithubRelease struct {
	HTMLURL string `json:"html_url"`
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
}

var LatestVersion string
var LatestVersionInfo GithubRelease

// UpdateAvailable returns true or false, depending on whether not a new version is available.
// It always returns false if the OnlineVersion has not yet been fetched.
func UpdateAvailable() bool {
	if !semver.IsValid(CurrentVersion) {
		panic(fmt.Sprintf("my current version '%s' is not valid", CurrentVersion))
	}

	if LatestVersion == "" {
		return false
	}

	if !semver.IsValid(LatestVersion) {
		// maybe this should just be a warning
		daulog.Errorf("online version '%s' is not valid - assuming no new version", LatestVersion)
		return false
	}
	comp := semver.Compare(LatestVersion, CurrentVersion)
	if comp == 0 {
		return false
	}
	if comp == 1 {
		return true
	}
	return false // they are using a newer one than exists?
}

func GetOnlineVersion() {

	daulog.Info("checking for new version")

	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Get("https://api.github.com/repos/tardisx/discord-auto-upload/releases/latest")
	if err != nil {
		daulog.Errorf("WARNING: Update check failed: %s", err)
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

	LatestVersion = latest.TagName
	LatestVersionInfo = latest
	daulog.Debugf("Latest version: %s", LatestVersion)

}
