package rules

import (
	"testing"

	"github.com/flavio/kube-image-bouncer/rules"
)

func TestIsFromWhitelistedRegistry(t *testing.T) {
	whitelist := []string{
		"localhost:5000",
		"registry.com",
	}

	testData := map[string]bool{
		"busybox":                   false,
		"registry.com/busybox":      true,
		"registry.com/team/busybox": true,
		"localhost:5000/redis":      true,
	}

	for image, expected := range testData {
		actual, err := rules.IsFromWhiteListedRegistry(image, whitelist)
		if err != nil {
			t.Fatalf("Unexpected error while processing image %s: %v", image, err)
		}

		if actual != expected {
			if actual {
				t.Fatalf("Expected '%s' to not be accepted", image)
			} else {
				t.Fatalf("Expected '%s' to be accepted", image)
			}
		}
	}
}
