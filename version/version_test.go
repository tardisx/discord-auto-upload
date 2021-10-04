package version

import (
	"testing"
)

func TestVersioning(t *testing.T) {
	if !NewVersionAvailable("v0.1.0") {
		t.Error("should be a version newer than v0.1.0")
	}
}
