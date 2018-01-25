package rules

import (
	"strings"

	"github.com/containers/image/docker/reference"
)

func IsUsingLatestTag(image string) (bool, error) {
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return false, err
	}

	return strings.HasSuffix(reference.TagNameOnly(named).String(), ":latest"), nil
}
