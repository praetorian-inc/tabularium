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
	h := u.Hostname()
	p := u.Port()
	if u.Scheme == "http" && p == "80" {
		u.Host = h
	} else if u.Scheme == "https" && p == "443" {
		u.Host = h
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

// FingerprintX sometimes returns HTTP on a 443 port
func FixSchemePortMismatch(u url.URL) url.URL {
	s := u.Scheme
	p := u.Port()
	if p == "443" && s == "http" {
		// HTTP scheme with HTTPS port -> change to HTTPS
		u.Scheme = "https"
	} else if p == "80" && s == "https" {
		// HTTPS scheme with HTTP port -> change to HTTP
		u.Scheme = "http"
	}
	return u
}
