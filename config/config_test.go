package config_test

import (
	"testing"

	"github.com/tardisx/discord-auto-upload/config"
)

func TestVersioning(t *testing.T) {
	if !config.NewVersionAvailable("v0.1.0") {
		t.Error("should be a version newer than v0.1.0")
	}
}
