package rules

import (
	"testing"

	"github.com/flavio/kube-image-bouncer/rules"
)

func TestIsUsingLatest(t *testing.T) {
	testData := map[string]bool{
		"busybox":                     true,
		"registry.com/busybox":        true,
		"busybox:latest":              true,
		"registry.com/busybox:latest": true,
		"busybox:1.2":                 false,
		"registry.com/busybox:1.2":    false,
	}

	for image, expected := range testData {
		actual, err := rules.IsUsingLatestTag(image)
		if err != nil {
			t.Fatalf("Unexpected error while processing image %s: %v", image, err)
		}

		if actual != expected {
			if actual {
				t.Fatalf("Expected '%s' to not be flagged as latest", image)
			} else {
				t.Fatalf("Expected '%s' to be flagged as latest", image)
			}
		}
	}
}
