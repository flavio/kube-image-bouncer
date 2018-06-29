package rules

import (
	"fmt"
	"strings"

	"github.com/containers/image/docker/reference"
)

func IsFromWhiteListedRegistry(image string, whitelist []string) (bool, error) {
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return false, err
	}
	res := strings.SplitN(named.Name(), "/", 2)
	if len(res) != 2 {
		return false, fmt.Errorf("Error while identifying the registry of %s", image)
	}

	for _, allowed := range whitelist {
		if res[0] == allowed {
			return true, nil
		}
	}

	return false, nil
}
