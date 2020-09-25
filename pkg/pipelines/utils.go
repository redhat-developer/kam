package pipelines

import (
	"fmt"
	"net/url"
)

func repoURL(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("failed to parse %q: %w", u, err)
	}
	parsed.Path = ""
	parsed.User = nil
	return parsed.String(), nil
}
