package normalize

import (
	"fmt"
	"net/url"
	"strings"
)

func Normalize(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("empty URL")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Scheme == "" {
		return "", fmt.Errorf("URL missing scheme")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("URL missing host")
	}

	*parsed = RemoveDefaultPorts(*parsed)
	*parsed = NormalizePath(*parsed)
	*parsed = RemoveQueryAndFragment(*parsed)
	*parsed = NormalizeCasing(*parsed)

	return parsed.String(), nil
}

func RemoveDefaultPorts(u url.URL) url.URL {
	p := u.Port()
	if p == "80" || p == "443" {
		u.Host = u.Hostname()
	}
	return u
}

func NormalizePath(u url.URL) url.URL {
	if u.Path == "" {
		u.Path = "/"
	}
	return u
}

func RemoveQueryAndFragment(u url.URL) url.URL {
	u.RawQuery = ""
	u.Fragment = ""
	return u
}

func NormalizeCasing(u url.URL) url.URL {
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	return u
}
