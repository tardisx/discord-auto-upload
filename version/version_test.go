package version

import (
	"testing"
)

func TestVersioningUpdate(t *testing.T) {
	// pretend there is a new version
	LatestVersion = "v0.13.9"
	if !UpdateAvailable() {
		t.Error("should be a version newer than " + CurrentVersion)
	}
}

func TestVersioningNoUpdate(t *testing.T) {
	// pretend there is a new version
	LatestVersion = "v0.12.1"
	if UpdateAvailable() {
		t.Error("should NOT be a version newer than " + CurrentVersion)
	}
}
