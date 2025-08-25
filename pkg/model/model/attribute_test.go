package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttribute_Preseed(t *testing.T) {
	tests := []struct {
		name string
		attr Attribute
		want Preseed
	}{
		{
			name: "basic preseed",
			attr: NewAttribute("preseed", "#preseed#whois+company#company#Chariot Systems", &Asset{BaseAsset: BaseAsset{Key: "#asset#gladiator.systems#gladiator.systems"}}),
			want: NewPreseed("whois+company", "company", "Chariot Systems"),
		},
		{
			name: "not a preseed",
			attr: NewAttribute("http", "80", &Asset{BaseAsset: BaseAsset{Key: "#asset#gladiator.systems#gladiator.systems"}}),
			want: Preseed{},
		},
		{
			name: "long truncated preseed",
			attr: NewAttribute("preseed", "#preseed#tlscert#530046309f9a6e3424d4ae66953b3671#-----BEGIN CERTIFICATE-----\nMIIEIzCCA6mgAwIBAgISBKY4tUq22IC5By7JQHN/o3ulMAoGCCqGSM49BAMDMDIx\nCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MQswCQYDVQQDEwJF\nNjAeFw0yNDExMTEyMDMwNDNaFw0yNTAyMDkyMDMwNDJaMBwxGjAYBgNVBAMTEWds\nYWRpYXRvci5zeXN0ZW1zMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEWgBr/u1z\nKSzmPSlG00wHyoIsDJKcdiEQWj9i/Al2dm2DZUgY6BG3ThJL2mjgeL59HIxoGMpr\ntbux5XhAwi0bLaOCArMwggKvMA4GA1UdDwEB/wQEAwIHgDAdBgNVHSUEFjAUBggr\nBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQU9djs2bmY\nbsY6YFH8KVA85/ZV0SAwHwYDVR0jBBgwFoAUkydGmAOpUWiOmNbEQkjbI79YlNIw\nVQYIKwYBBQUHAQEESTBHMCEGCCsGAQUFBzABhhVodHRwOi8vZTYuby5sZW5jci5v\ncmcwIgYIKwYBBQUHMAKGFmh0dHA6Ly9lNi5pLmxlbmNyLm9yZy8wgbwGA1UdEQSB\ntDCBsYIYY2ljZXJvLmdsYWRpYXRvci5zeXN0ZW1zghFnbGFkaWF0b3Iuc3lzdGVt\nc4IZZ3JhY2N1cy5nbGFkaWF0b3Iuc3lzdGVtc4IYbWFyY3VzLmdsYWRpYXRvci5z\neXN0ZW1zghltYXhpbXVzLmdsYWRpYXRvci5zeXN0ZW1zghZuZXJvLmdsYWRpYXRv\nci5zeXN0ZW1zghpzdWIubmVyby5nbGFkaWF0b3Iuc3lzdGVtczATBgNVHSAEDDAK\nMAgGBmeBDAECATCCAQMGCisGAQQB1nkCBAIEgfQEgfEA7wB2AObSMWNAd4zBEEEG\n13G5zsHSQPaWhIb7uocyHf0eN45QAAABkx0hfAsAAAQDAEcwRQIgRFQZz2woupNO\nnACvzN+VA6hdFPtURqHSQ515DjPcRHgCIQDV9ESS2H+5goazEVEvJl5mHJ0b3nAs\nWp5ZvHUCGiUxKwB1AOCSs/wMHcjnaDYf3mG5lk0KUngZinLWcsSwTaVtb1QEAAAB\nkx0hfBIAAAQDAEYwRAIgDLwbI0v7nax4VIqninXbEUHfnU8TYc4084cWKXsdbQ0C\nID+PJFgxIPAHEQ5T+WMGtMnMk9VbFW3oqKS6iVXy4kRTMAoGCCqGSM49BAMDA2gA\nMGUCMQCEw6wNjTvzYPUbCy0uVg5ZBfFQO51S8broOZV+y2xSPrfuwe8jpWqhjceF\nO97uffgCMHtc96t9h+UxWbuGwAugUEHvog9vsZmOq3UrNX5PntVznRg1Jvx+htje\nWPvnk2BEMQ==\n-----END CERTIFICATE-----\n", &Asset{BaseAsset: BaseAsset{Key: "#asset#gladiator.systems#52.169.142.100"}}),
			want: NewPreseed("tlscert", "530046309f9a6e3424d4ae66953b3671", "-----BEGIN CERTIFICATE-----\nMIIEIzCCA6mgAwIBAgISBKY4tUq22IC5By7JQHN/o3ulMAoGCCqGSM49BAMDMDIx\nCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MQswCQYDVQQDEwJF\nNjAeFw0yNDExMTEyMDMwNDNaFw0yNTAyMDkyMDMwNDJaMBwxGjAYBgNVBAMTEWds\nYWRpYXRvci5zeXN0ZW1zMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEWgBr/u1z\nKSzmPSlG00wHyoIsDJKcdiEQWj9i/Al2dm2DZUgY6BG3ThJL2mjgeL59HIxoGMpr\ntbux5XhAwi0bLaOCArMwggKvMA4GA1UdDwEB/wQEAwIHgDAdBgNVHSUEFjAUBggr\nBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQU9djs2bmY\nbsY6YFH8KVA85/ZV0SAwHwYDVR0jBBgwFoAUkydGmAOpUWiOmNbEQkjbI79YlNIw\nVQYIKwYBBQUHAQEESTBHMCEGCCsGAQUFBzABhhVodHRwOi8vZTYuby5sZW5jci5v\ncmcwIgYIKwYBBQUHMAKGFmh0dHA6Ly9lNi5pLmxlbmNyLm9yZy8wgbwGA1UdEQSB\ntDCBsYIYY2ljZXJvLmdsYWRpYXRvci5zeXN0ZW1zghFnbGFkaWF0b3Iuc3lzdGVt\nc4IZZ3JhY2N1cy5nbGFkaWF0b3Iuc3lzdGVtc4IYbWFyY3VzLmdsYWRpYXRvci5z\neXN0ZW1zghltYXhpbXVzLmdsYWRpYXRvci5zeXN0ZW1zghZuZXJvLmdsYWRpYXRv\nci5zeXN0ZW1zghpzdWIubmVyby5nbGFkaWF0b3Iuc3lzdGVtczATBgNVHSAEDDAK\nMAgGBmeBDAECATCCAQMGCisGAQQB1nkCBAIEgfQEgfEA7wB2AObSMWNAd4zBEEEG\n13G5zsHSQPaWhIb7uocyHf0eN45QAAABkx0hfAsAAAQDAEcwRQIgRFQZz2woupNO\nnACvzN+VA6hdFPtURqHSQ515DjPcRHgCIQDV9ESS2H+5goazEVEvJl5mHJ0b3nAs\nWp5ZvHUCGiUxKwB1AOCSs/wMHcjnaDYf3mG5lk0KUngZinLWcsSwTaVtb1QEAAAB\nkx0hfBIAAAQDAEYwRAIgDLwbI0v7nax4VIqninXbEUHfnU8TYc4084cWKXsdbQ0C\nID+PJFgxIPAHEQ5T+WMGtMnMk9VbFW3oqKS6iVXy4kRTMAoGCCqGSM49BAMDA2gA\nMGUCMQCEw6wNjTvzYPUbCy0uVg5ZBfFQO51S8broOZV+y2xSPrfuwe8jpWqhjceF\nO97uffgCMHtc96t9h+UxWbuGwAugUEHvog9vsZmOq3UrNX5PntVznRg1Jvx+htje\nWPvnk2BEMQ==\n-----END CERTIFICATE-----\n"),
		},
	}

	for _, test := range tests {
		actual := test.attr.Preseed()
		assert.Equal(t, test.want, actual, "test case %s failed: expected %v, got %v", test.name, test.want, actual)
	}
}

func TestAttribute_Target(t *testing.T) {
	tests := []struct {
		name     string
		attr     Attribute
		expected string
	}{
		{
			name:     "http target",
			attr:     NewAttribute("http", "80", &Asset{BaseAsset: BaseAsset{Key: "#asset#gladiator.systems#52.169.142.100"}}),
			expected: "http://gladiator.systems:80",
		},
		{
			name:     "ssh target",
			attr:     NewAttribute("ssh", "22", &Asset{BaseAsset: BaseAsset{Key: "#asset#gladiator.systems#52.169.142.100"}}),
			expected: "ssh://gladiator.systems:22",
		},
		{
			name:     "port target",
			attr:     NewAttribute("port", "80", &Asset{BaseAsset: BaseAsset{Key: "#asset#gladiator.systems#52.169.142.100"}}),
			expected: "gladiator.systems:80",
		},
		{
			name:     "protocol target",
			attr:     NewAttribute("protocol", "http", &Asset{BaseAsset: BaseAsset{Key: "#asset#gladiator.systems#52.169.142.100"}}),
			expected: "http://gladiator.systems",
		},
	}

	for _, test := range tests {
		actual := test.attr.Target()
		if actual != test.expected {
			t.Errorf("Attribute.Target() = %v, want %v", actual, test.expected)
		}
	}
}

func TestAttribute_TargetImplementation(t *testing.T) {
	tests := []struct {
		name               string
		attr               Attribute
		expectedGroup      string
		expectedIdentifier string
	}{
		{
			name:               "http target",
			attr:               NewAttribute("http", "80", &Asset{BaseAsset: BaseAsset{Key: "#asset#gladiator.systems#52.169.142.100"}}),
			expectedGroup:      "gladiator.systems",
			expectedIdentifier: "http://gladiator.systems:80",
		},
	}

	for _, test := range tests {
		actual := test.attr.Group()
		if actual != test.expectedGroup {
			t.Errorf("Attribute.Target() = %v, want %v", actual, test.expectedGroup)
		}
		actual = test.attr.Identifier()
		if actual != test.expectedIdentifier {
			t.Errorf("Attribute.Target() = %v, want %v", actual, test.expectedIdentifier)
		}
	}
}

func TestAttribute_IsPrivate(t *testing.T) {
	publicAsset := NewAsset("contoso.com", "18.1.2.4")
	privateAsset := NewAsset("contoso.local", "10.0.0.1")

	tests := []struct {
		name string
		attr Attribute
		want bool
	}{
		{
			name: "public https",
			attr: NewAttribute("https", "443", &publicAsset),
			want: false,
		},
		{
			name: "private https",
			attr: NewAttribute("https", "443", &privateAsset),
			want: true,
		},
		{
			name: "public port",
			attr: NewAttribute("port", "443", &publicAsset),
			want: false,
		},
		{
			name: "private port",
			attr: NewAttribute("port", "443", &privateAsset),
			want: true,
		},
	}

	for _, tc := range tests {
		actual := tc.attr.IsPrivate()
		assert.Equal(t, tc.want, actual, "test case %s failed: expected %t, got %t", tc.name, tc.want, actual)
	}
}
