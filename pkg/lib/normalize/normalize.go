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

	*parsed = FixSchemePortMismatch(*parsed)
	*parsed = RemoveDefaultPorts(*parsed)
	*parsed = NormalizePath(*parsed)
	*parsed = RemoveQueryAndFragment(*parsed)
	*parsed = NormalizeCasing(*parsed)

	return parsed.String(), nil
}

func RemoveDefaultPorts(u url.URL) url.URL {
	if u.Scheme == "http" && strings.HasSuffix(u.Host, ":80") {
		u.Host = strings.TrimSuffix(u.Host, ":80")
	} else if u.Scheme == "https" && strings.HasSuffix(u.Host, ":443") {
		u.Host = strings.TrimSuffix(u.Host, ":443")
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
	u.Path = strings.ToLower(u.Path)
	return u
}

// FingerprintX sometimes returns HTTP on a 443 port
func FixSchemePortMismatch(u url.URL) url.URL {
	if u.Port() == "443" && u.Scheme == "http" {
		// HTTP scheme with HTTPS port -> change to HTTPS
		u.Scheme = "https"
	} else if u.Port() == "80" && u.Scheme == "https" {
		// HTTPS scheme with HTTP port -> change to HTTP
		u.Scheme = "http"
	}
	return u
}
