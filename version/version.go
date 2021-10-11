package version

import (
	"fmt"
	"log"

	"golang.org/x/mod/semver"
)

const CurrentVersion string = "v0.12.0"

func NewVersionAvailable(v string) bool {
	if !semver.IsValid(CurrentVersion) {
		panic(fmt.Sprintf("my current version '%s' is not valid", CurrentVersion))
	}
	if !semver.IsValid(v) {
		// maybe this should just be a warning
		log.Printf("passed in version '%s' is not valid - assuming no new version", v)
		return false
	}
	comp := semver.Compare(v, CurrentVersion)
	if comp == 0 {
		return false
	}
	if comp == 1 {
		return true
	}
	return false // they are using a newer one than exists?
}
