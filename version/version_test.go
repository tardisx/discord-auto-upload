package version

import (
	"testing"
)

func TestVersioning(t *testing.T) {
	if !NewVersionAvailable("v1.0.0") {
		t.Error("should be a version newer than v1.0.0")
	}
}
