package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBurpHTTPData(t *testing.T) {
	sampleJSON := `{
  "metadata": {
    "excluded_extensions": [
      "jpg",
      "css",
      "js"
    ],
    "scheduledInterval": 10,
    "mapType": "http",
    "sizeThreshold": 1024,
    "ai_enabled": true,
    "scope_enabled": false,
    "timeUnit": "SECONDS"
  },
  "data": {
    "14": {
      "originalRequest": {
        "body": "",
        "messageId": 0,
        "inScope": false,
        "method": "GET",
        "path": "/",
        "url": "http://example.com:8888/",
        "headers": [
          {
            "Host": "example.com:8888"
          },
          {
            "User-Agent": "Mozilla/5.0"
          }
        ]
      },
      "originalResponse": {
        "body": "<html><body>Test</body></html>",
        "messageId": 0,
        "inScope": false,
        "method": "",
        "path": "",
        "statusCode": 200,
        "headers": [
          {
            "Content-Type": "text/html"
          }
        ]
      },
      "toolSource": "burp.test"
    }
  }
}`

	burpData, err := ParseBurpHTTPData([]byte(sampleJSON))
	require.NoError(t, err)

	assert.Equal(t, "http", burpData.Metadata.MapType)
	assert.True(t, burpData.Metadata.AIEnabled)
	assert.Equal(t, 1, len(burpData.Data))

	webpage := burpData.Data["14"]
	assert.Equal(t, "http://example.com:8888/", webpage.URL)
	assert.Equal(t, []string{BURP_COURIER_SOURCE}, webpage.Source)
	assert.Equal(t, "burp.test", webpage.Metadata["tool_source"])

	// Check that the webpage has requests
	assert.Equal(t, 1, len(webpage.Requests))
	request := webpage.Requests[0]
	assert.Equal(t, "GET", request.Method)
	assert.Equal(t, "http://example.com:8888/", request.RawURL)
	assert.Equal(t, []string{"example.com:8888"}, request.Headers["Host"])
	assert.Equal(t, []string{"Mozilla/5.0"}, request.Headers["User-Agent"])

	// Check response
	require.NotNil(t, request.Response)
	assert.Equal(t, 200, request.Response.StatusCode)
	assert.Equal(t, "<html><body>Test</body></html>", request.Response.Body)
	assert.Equal(t, []string{"text/html"}, request.Response.Headers["Content-Type"])
}

func TestBurpHTTPData_ExtractBaseURLs(t *testing.T) {
	burpData := &BurpHTTPData{
		Data: map[string]Webpage{
			"1": {
				URL: "http://api.example.com:8888/api/users",
			},
			"2": {
				URL: "http://api.example.com:8888/login",
			},
			"3": {
				URL: "https://secure.example.com:443/",
			},
		},
	}

	urls := burpData.ExtractBaseURLs()
	assert.Equal(t, 2, len(urls))
	assert.Contains(t, urls, "http://api.example.com:8888")
	assert.Contains(t, urls, "https://secure.example.com:443")
}

func TestBurpHTTPData_ToWebpageRequests(t *testing.T) {
	burpData := &BurpHTTPData{
		Data: map[string]Webpage{
			"1": {
				URL: "http://api.example.com:8888/api/users",
				WebpageDetails: WebpageDetails{
					Requests: []WebpageRequest{
						{
							RawURL: "http://api.example.com:8888/api/users",
							Method: "GET",
							Body:   "",
							Headers: map[string][]string{
								"Host":       {"api.example.com:8888"},
								"User-Agent": {"Mozilla/5.0"},
							},
							Response: &WebpageResponse{
								Body: `{"users": []}`,
								Headers: map[string][]string{
									"Content-Type": {"application/json"},
								},
								StatusCode: 200,
							},
						},
					},
				},
			},
		},
	}

	requests, err := burpData.ToWebpageRequests()
	require.NoError(t, err)
	assert.Equal(t, 1, len(requests))

	req := requests[0]
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "http://api.example.com:8888/api/users", req.RawURL)
	assert.Equal(t, "", req.Body)
	assert.NotNil(t, req.Response)
	assert.Equal(t, `{"users": []}`, req.Response.Body)
	assert.Equal(t, 200, req.Response.StatusCode)
}

func TestBurpHTTPData_ExtractToolSources(t *testing.T) {
	burpData := &BurpHTTPData{
		Data: map[string]Webpage{
			"1": {
				Metadata: map[string]any{"tool_source": "burp.proxy"},
			},
			"2": {
				Metadata: map[string]any{"tool_source": "burp.repeater"},
			},
			"3": {
				Metadata: map[string]any{"tool_source": "burp.proxy"},
			},
			"4": {
				Metadata: map[string]any{"tool_source": ""},
			},
		},
	}

	sources := burpData.ExtractToolSources()
	assert.Equal(t, 2, len(sources))
	assert.Contains(t, sources, "burp.proxy")
	assert.Contains(t, sources, "burp.repeater")
}

// Tests for Burp Issues structures

func TestParseBurpIssuesData(t *testing.T) {
	sampleJSON := `{
  "metadata": {
    "excluded_extensions": [
      "jpg",
      "css",
      "js"
    ],
    "scheduledInterval": 10,
    "mapType": "issues",
    "sizeThreshold": 1024,
    "ai_enabled": true,
    "scope_enabled": false,
    "timeUnit": "SECONDS"
  },
  "data": {
    "-2101509812": {
      "baseUrl": "http://ecs5.gladiator.systems:8888/robots.txt",
      "collaboratorInteractions": [],
      "confidence": "CERTAIN",
      "severity": "INFO",
      "requests": [
        {
          "body": "",
          "messageId": 0,
          "inScope": false,
          "method": "GET",
          "path": "/robots.txt",
          "url": "http://ecs5.gladiator.systems:8888/robots.txt",
          "headers": [
            {
              "Host": "ecs5.gladiator.systems:8888"
            },
            {
              "User-Agent": "Mozilla/5.0"
            }
          ]
        }
      ],
      "responses": [
        {
          "body": "User-agent: *\\nDisallow: /",
          "messageId": 0,
          "inScope": false,
          "method": "",
          "path": "",
          "headers": [
            {
              "Content-Type": "text/plain"
            }
          ]
        }
      ],
      "name": "Robots.txt file",
      "detail": "The web server contains a robots.txt file."
    }
  }
}`

	burpData, err := ParseBurpIssuesData([]byte(sampleJSON))
	require.NoError(t, err)

	assert.Equal(t, "issues", burpData.Metadata.MapType)
	assert.True(t, burpData.Metadata.AIEnabled)
	assert.Equal(t, 1, len(burpData.Data))

	issue := burpData.Data["-2101509812"]
	assert.Equal(t, "http://ecs5.gladiator.systems:8888/robots.txt", issue.BaseURL)
	assert.Equal(t, "CERTAIN", issue.Confidence)
	assert.Equal(t, "INFO", issue.Severity)
	assert.Equal(t, "Robots.txt file", issue.Name)
	assert.Equal(t, "The web server contains a robots.txt file.", issue.Detail)
	assert.Equal(t, 0, len(issue.CollaboratorInteractions))
	assert.Equal(t, 1, len(issue.ProcessedRequests))
	assert.Equal(t, 1, len(issue.ProcessedResponses))

	// Test processed request
	req := issue.ProcessedRequests[0]
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "http://ecs5.gladiator.systems:8888/robots.txt", req.RawURL)
	assert.Equal(t, []string{"ecs5.gladiator.systems:8888"}, req.Headers["Host"])
	assert.Equal(t, []string{"Mozilla/5.0"}, req.Headers["User-Agent"])

	// Test processed response
	resp := issue.ProcessedResponses[0]
	assert.Equal(t, 200, resp.StatusCode) // Default status code
	assert.Equal(t, "User-agent: *\\nDisallow: /", resp.Body)
	assert.Equal(t, []string{"text/plain"}, resp.Headers["Content-Type"])
}

func TestBurpIssuesData_ExtractBaseURLs(t *testing.T) {
	burpData := &BurpIssuesData{
		Data: map[string]BurpIssueEntry{
			"1": {
				BaseURL: "http://api.example.com:8888/api/users",
			},
			"2": {
				BaseURL: "http://api.example.com:8888/login",
			},
			"3": {
				BaseURL: "https://secure.example.com:443/",
			},
		},
	}

	urls := burpData.ExtractBaseURLs()
	assert.Equal(t, 2, len(urls))
	assert.Contains(t, urls, "http://api.example.com:8888")
	assert.Contains(t, urls, "https://secure.example.com:443")
}

func TestBurpIssuesData_GetIssuesByBaseURL(t *testing.T) {
	burpData := &BurpIssuesData{
		Data: map[string]BurpIssueEntry{
			"1": {
				BaseURL: "http://api.example.com:8888/api/users",
				Name:    "Issue 1",
			},
			"2": {
				BaseURL: "http://api.example.com:8888/login",
				Name:    "Issue 2",
			},
			"3": {
				BaseURL: "https://secure.example.com:443/",
				Name:    "Issue 3",
			},
		},
	}

	issuesByURL := burpData.GetIssuesByURL()
	assert.Equal(t, 2, len(issuesByURL))

	// Check that issues for same base URL are grouped together
	apiIssues := issuesByURL["http://api.example.com:8888"]
	assert.Equal(t, 2, len(apiIssues))
	assert.Contains(t, []string{"Issue 1", "Issue 2"}, apiIssues[0].Name)
	assert.Contains(t, []string{"Issue 1", "Issue 2"}, apiIssues[1].Name)

	secureIssues := issuesByURL["https://secure.example.com:443"]
	assert.Equal(t, 1, len(secureIssues))
	assert.Equal(t, "Issue 3", secureIssues[0].Name)
}

func TestBurpIssueEntry_GetRiskSeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		expected string
	}{
		// Test severity-based mapping
		{"High severity", "HIGH", "H"},
		{"Medium severity", "MEDIUM", "M"},
		{"Low severity", "LOW", "L"},
		{"Info severity", "INFO", "I"},
		{"Information severity", "INFORMATION", "I"},
		{"Informational severity", "INFORMATIONAL", "I"},
		{"Case insensitive severity", "high", "H"},

		// Test default behavior for unknown/empty severity
		{"Unknown severity", "UNKNOWN_SEVERITY", "I"},
		{"Empty severity", "", "I"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issue := BurpIssueEntry{
				Severity: test.severity,
			}
			assert.Equal(t, test.expected, issue.GetRiskSeverity(),
				"Expected %s for severity %s", test.expected, test.severity)
		})
	}
}

func TestBurpIssueEntry_HasCollaboratorInteractions(t *testing.T) {
	// Test with no interactions
	issue := BurpIssueEntry{}
	assert.False(t, issue.HasCollaboratorInteractions())

	// Test with interactions
	issue.CollaboratorInteractions = []CollaboratorInteraction{
		{
			Type:        "DNS",
			Protocol:    "UDP",
			LookupType:  "A",
			Interaction: "test.collaborator.net",
		},
	}
	assert.True(t, issue.HasCollaboratorInteractions())
}

func TestBurpIssueEntry_UnmarshalJSON_WithCollaboratorInteractions(t *testing.T) {
	jsonData := `{
		"baseUrl": "https://test.example.com/",
		"collaboratorInteractions": [
			{
				"type": "DNS",
				"protocol": "UDP",
				"lookupType": "A",
				"interaction": "test.collaborator.net",
				"rawDetail": "DNS lookup details"
			}
		],
		"confidence": "FIRM",
		"severity": "HIGH",
		"requests": [
			{
				"body": "test=payload",
				"messageId": 100,
				"inScope": true,
				"method": "POST",
				"path": "/submit",
				"url": "https://test.example.com/submit",
				"headers": [
					{"Content-Type": "application/x-www-form-urlencoded"}
				]
			}
		],
		"responses": [
			{
				"body": "Success",
				"messageId": 101,
				"inScope": true,
				"headers": [
					{"Content-Type": "text/plain"}
				]
			}
		],
		"name": "SSRF Vulnerability",
		"detail": "Server-Side Request Forgery vulnerability detected."
	}`

	var issue BurpIssueEntry
	err := issue.UnmarshalJSON([]byte(jsonData))
	require.NoError(t, err)

	// Test basic fields
	assert.Equal(t, "https://test.example.com/", issue.BaseURL)
	assert.Equal(t, "FIRM", issue.Confidence)
	assert.Equal(t, "HIGH", issue.Severity)
	assert.Equal(t, "SSRF Vulnerability", issue.Name)
	assert.Equal(t, "Server-Side Request Forgery vulnerability detected.", issue.Detail)

	// Test collaborator interactions
	assert.True(t, issue.HasCollaboratorInteractions())
	assert.Equal(t, 1, len(issue.CollaboratorInteractions))
	interaction := issue.CollaboratorInteractions[0]
	assert.Equal(t, "DNS", interaction.Type)
	assert.Equal(t, "UDP", interaction.Protocol)
	assert.Equal(t, "A", interaction.LookupType)
	assert.Equal(t, "test.collaborator.net", interaction.Interaction)
	assert.Equal(t, "DNS lookup details", interaction.RawDetail)

	// Test processed requests
	assert.Equal(t, 1, len(issue.ProcessedRequests))
	req := issue.ProcessedRequests[0]
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "https://test.example.com/submit", req.RawURL)
	assert.Equal(t, "test=payload", req.Body)
	assert.Equal(t, []string{"application/x-www-form-urlencoded"}, req.Headers["Content-Type"])

	// Test processed responses
	assert.Equal(t, 1, len(issue.ProcessedResponses))
	resp := issue.ProcessedResponses[0]
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Success", resp.Body)
	assert.Equal(t, []string{"text/plain"}, resp.Headers["Content-Type"])
}

func TestBurpIssueEntry_UnmarshalJSON_EmptyCollaboratorInteractions(t *testing.T) {
	jsonData := `{
		"baseUrl": "https://test.example.com/",
		"collaboratorInteractions": [],
		"confidence": "TENTATIVE",
		"severity": "LOW",
		"requests": [],
		"responses": [],
		"name": "Information Disclosure",
		"detail": "Potential information disclosure."
	}`

	var issue BurpIssueEntry
	err := issue.UnmarshalJSON([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "https://test.example.com/", issue.BaseURL)
	assert.Equal(t, "TENTATIVE", issue.Confidence)
	assert.Equal(t, "LOW", issue.Severity)
	assert.Equal(t, "Information Disclosure", issue.Name)
	assert.Equal(t, "Potential information disclosure.", issue.Detail)
	assert.False(t, issue.HasCollaboratorInteractions())
	assert.Equal(t, 0, len(issue.ProcessedRequests))
	assert.Equal(t, 0, len(issue.ProcessedResponses))
}

func TestBurpIssueEntry_UnmarshalJSON_ComplexExample(t *testing.T) {
	jsonData := `{
		"baseUrl": "http://ecs5.gladiator.systems:8888/",
		"collaboratorInteractions": [],
		"confidence": "FIRM",
		"severity": "MEDIUM",
		"requests": [
			{
				"body": "",
				"messageId": 0,
				"inScope": false,
				"method": "GET",
				"path": "/",
				"url": "http://ecs5.gladiator.systems:8888/",
				"headers": [
					{"Host": "ecs5.gladiator.systems:8888"},
					{"Cache-Control": "max-age=0"},
					{"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}
				]
			},
			{
				"body": "",
				"messageId": 0,
				"inScope": false,
				"method": "GET",
				"path": "/index.php/grda9p/",
				"url": "http://ecs5.gladiator.systems:8888/index.php/grda9p/",
				"headers": [
					{"Host": "ecs5.gladiator.systems:8888"},
					{"Accept": "text/html,application/xhtml+xml"}
				]
			}
		],
		"responses": [
			{
				"body": "<!DOCTYPE html>\\n<html>...</html>",
				"messageId": 0,
				"inScope": false,
				"method": "",
				"path": "",
				"headers": [
					{"Date": "Tue, 09 Sep 2025 18:22:48 GMT"},
					{"Server": "Apache/2.4.10 (Debian)"},
					{"Content-Type": "text/html;charset=utf-8"}
				]
			},
			{
				"body": "<!DOCTYPE html>\\n<html>...</html>",
				"messageId": 0,
				"inScope": false,
				"method": "",
				"path": "",
				"headers": [
					{"Date": "Tue, 09 Sep 2025 18:22:52 GMT"},
					{"Server": "Apache/2.4.10 (Debian)"},
					{"Content-Type": "text/html;charset=utf-8"}
				]
			}
		],
		"name": "Path-relative style sheet import",
		"detail": "The application may be vulnerable to path-relative style sheet import (PRSSI) attacks."
	}`

	var issue BurpIssueEntry
	err := issue.UnmarshalJSON([]byte(jsonData))
	require.NoError(t, err)

	// Test basic fields
	assert.Equal(t, "http://ecs5.gladiator.systems:8888/", issue.BaseURL)
	assert.Equal(t, "FIRM", issue.Confidence)
	assert.Equal(t, "MEDIUM", issue.Severity)
	assert.Equal(t, "Path-relative style sheet import", issue.Name)
	assert.Equal(t, "The application may be vulnerable to path-relative style sheet import (PRSSI) attacks.", issue.Detail)
	assert.False(t, issue.HasCollaboratorInteractions())

	// Test multiple requests
	assert.Equal(t, 2, len(issue.ProcessedRequests))

	req1 := issue.ProcessedRequests[0]
	assert.Equal(t, "GET", req1.Method)
	assert.Equal(t, "http://ecs5.gladiator.systems:8888/", req1.RawURL)
	assert.Equal(t, "", req1.Body)
	assert.Equal(t, []string{"ecs5.gladiator.systems:8888"}, req1.Headers["Host"])
	assert.Equal(t, []string{"max-age=0"}, req1.Headers["Cache-Control"])
	assert.Equal(t, []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}, req1.Headers["User-Agent"])

	req2 := issue.ProcessedRequests[1]
	assert.Equal(t, "GET", req2.Method)
	assert.Equal(t, "http://ecs5.gladiator.systems:8888/index.php/grda9p/", req2.RawURL)
	assert.Equal(t, []string{"text/html,application/xhtml+xml"}, req2.Headers["Accept"])

	// Test multiple responses
	assert.Equal(t, 2, len(issue.ProcessedResponses))

	resp1 := issue.ProcessedResponses[0]
	assert.Equal(t, 200, resp1.StatusCode)
	assert.Equal(t, "<!DOCTYPE html>\\n<html>...</html>", resp1.Body)
	assert.Equal(t, []string{"Tue, 09 Sep 2025 18:22:48 GMT"}, resp1.Headers["Date"])
	assert.Equal(t, []string{"Apache/2.4.10 (Debian)"}, resp1.Headers["Server"])
	assert.Equal(t, []string{"text/html;charset=utf-8"}, resp1.Headers["Content-Type"])

	resp2 := issue.ProcessedResponses[1]
	assert.Equal(t, 200, resp2.StatusCode)
	assert.Equal(t, "<!DOCTYPE html>\\n<html>...</html>", resp2.Body)
	assert.Equal(t, []string{"Tue, 09 Sep 2025 18:22:52 GMT"}, resp2.Headers["Date"])
}

// Test functions for the new BurpHTTPData UnmarshalJSON functionality

func TestBurpHTTPData_UnmarshalJSON_CreatesWebpages(t *testing.T) {
	sampleJSON := `{
  "metadata": {
    "mapType": "http",
    "ai_enabled": true,
    "scope_enabled": false
  },
  "data": {
    "100": {
      "originalRequest": {
        "body": "test=data",
        "messageId": 100,
        "inScope": true,
        "method": "POST",
        "path": "/api/login",
        "url": "https://example.com/api/login",
        "headers": [
          {"Host": "example.com"},
          {"Content-Type": "application/x-www-form-urlencoded"}
        ]
      },
      "originalResponse": {
        "body": "{\"success\": true}",
        "messageId": 101,
        "inScope": true,
        "statusCode": 200,
        "headers": [
          {"Content-Type": "application/json"}
        ]
      },
      "toolSource": "burp.proxy",
      "wasRequestIntercepted": true,
      "wasResponseIntercepted": false
    },
    "200": {
      "originalRequest": {
        "body": "",
        "messageId": 200,
        "inScope": false,
        "method": "GET",
        "path": "/robots.txt",
        "url": "https://example.com/robots.txt",
        "headers": [
          {"Host": "example.com"},
          {"User-Agent": "BurpSuite/1.0"}
        ]
      },
      "originalResponse": {
        "body": "User-agent: *\nDisallow: /",
        "messageId": 201,
        "inScope": false,
        "statusCode": 200,
        "headers": [
          {"Content-Type": "text/plain"}
        ]
      },
      "toolSource": "burp.spider"
    }
  }
}`

	var burpData BurpHTTPData
	err := burpData.UnmarshalJSON([]byte(sampleJSON))
	require.NoError(t, err)

	// Test metadata
	assert.Equal(t, "http", burpData.Metadata.MapType)
	assert.True(t, burpData.Metadata.AIEnabled)
	assert.False(t, burpData.Metadata.ScopeEnabled)

	// Test that we have 2 webpages
	assert.Equal(t, 2, len(burpData.Data))

	// Test first webpage (messageID "100")
	webpage1 := burpData.Data["100"]
	assert.Equal(t, "https://example.com/api/login", webpage1.URL)
	assert.Equal(t, []string{BURP_COURIER_SOURCE}, webpage1.Source)
	assert.Equal(t, "burp.proxy", webpage1.Metadata["tool_source"])
	assert.Equal(t, true, webpage1.Metadata["was_request_intercepted"])
	assert.Equal(t, false, webpage1.Metadata["was_response_intercepted"])

	// Test first webpage requests
	assert.Equal(t, 1, len(webpage1.Requests))
	request1 := webpage1.Requests[0]
	assert.Equal(t, "POST", request1.Method)
	assert.Equal(t, "https://example.com/api/login", request1.RawURL)
	assert.Equal(t, "test=data", request1.Body)
	assert.Equal(t, []string{"example.com"}, request1.Headers["Host"])
	assert.Equal(t, []string{"application/x-www-form-urlencoded"}, request1.Headers["Content-Type"])
	assert.True(t, request1.WasIntercepted)

	// Test first webpage response
	require.NotNil(t, request1.Response)
	assert.Equal(t, 200, request1.Response.StatusCode)
	assert.Equal(t, "{\"success\": true}", request1.Response.Body)
	assert.Equal(t, []string{"application/json"}, request1.Response.Headers["Content-Type"])

	// Test second webpage (messageID "200")
	webpage2 := burpData.Data["200"]
	assert.Equal(t, "https://example.com/robots.txt", webpage2.URL)
	assert.Equal(t, []string{BURP_COURIER_SOURCE}, webpage2.Source)
	assert.Equal(t, "burp.spider", webpage2.Metadata["tool_source"])

	// Test second webpage requests
	assert.Equal(t, 1, len(webpage2.Requests))
	request2 := webpage2.Requests[0]
	assert.Equal(t, "GET", request2.Method)
	assert.Equal(t, "https://example.com/robots.txt", request2.RawURL)
	assert.Equal(t, "", request2.Body)
	assert.Equal(t, []string{"example.com"}, request2.Headers["Host"])
	assert.Equal(t, []string{"BurpSuite/1.0"}, request2.Headers["User-Agent"])

	// Test second webpage response
	require.NotNil(t, request2.Response)
	assert.Equal(t, 200, request2.Response.StatusCode)
	assert.Equal(t, "User-agent: *\nDisallow: /", request2.Response.Body)
	assert.Equal(t, []string{"text/plain"}, request2.Response.Headers["Content-Type"])
}

func TestBurpHTTPData_UnmarshalJSON_WithModifiedRequests(t *testing.T) {
	sampleJSON := `{
  "metadata": {
    "mapType": "http"
  },
  "data": {
    "300": {
      "originalRequest": {
        "body": "username=admin&password=test",
        "messageId": 300,
        "inScope": true,
        "method": "POST",
        "path": "/login",
        "url": "https://test.com/login",
        "headers": [
          {"Host": "test.com"},
          {"Content-Type": "application/x-www-form-urlencoded"}
        ]
      },
      "originalResponse": {
        "body": "Login failed",
        "messageId": 301,
        "inScope": true,
        "statusCode": 401,
        "headers": [
          {"Content-Type": "text/plain"}
        ]
      },
      "modifiedRequest": {
        "body": "username=admin&password=admin123",
        "messageId": 302,
        "inScope": true,
        "method": "POST",
        "path": "/login",
        "url": "https://test.com/login",
        "headers": [
          {"Host": "test.com"},
          {"Content-Type": "application/x-www-form-urlencoded"},
          {"X-Injection": "test"}
        ]
      },
      "modifiedResponse": {
        "body": "Login successful",
        "messageId": 303,
        "inScope": true,
        "statusCode": 200,
        "headers": [
          {"Content-Type": "text/plain"},
          {"Set-Cookie": "session=abc123"}
        ]
      },
      "toolSource": "burp.repeater",
      "wasRequestIntercepted": true,
      "wasResponseIntercepted": true,
      "wasRequestModified": true,
      "wasResponseModified": true
    }
  }
}`

	var burpData BurpHTTPData
	err := burpData.UnmarshalJSON([]byte(sampleJSON))
	require.NoError(t, err)

	assert.Equal(t, 1, len(burpData.Data))

	webpage := burpData.Data["300"]
	assert.Equal(t, "https://test.com/login", webpage.URL)
	assert.Equal(t, []string{BURP_COURIER_SOURCE}, webpage.Source)
	assert.Equal(t, "burp.repeater", webpage.Metadata["tool_source"])
	assert.Equal(t, true, webpage.Metadata["was_request_intercepted"])
	assert.Equal(t, true, webpage.Metadata["was_response_intercepted"])
	assert.Equal(t, true, webpage.Metadata["was_request_modified"])
	assert.Equal(t, true, webpage.Metadata["was_response_modified"])

	// Should have both original and modified requests
	assert.Equal(t, 2, len(webpage.Requests))

	// Test original request
	originalRequest := webpage.Requests[0]
	assert.Equal(t, "POST", originalRequest.Method)
	assert.Equal(t, "https://test.com/login", originalRequest.RawURL)
	assert.Equal(t, "username=admin&password=test", originalRequest.Body)
	assert.True(t, originalRequest.WasIntercepted)
	assert.False(t, originalRequest.WasModified) // Original request is never modified by definition

	// Test original response
	require.NotNil(t, originalRequest.Response)
	assert.Equal(t, 401, originalRequest.Response.StatusCode)
	assert.Equal(t, "Login failed", originalRequest.Response.Body)
	assert.True(t, originalRequest.Response.WasIntercepted)
	assert.False(t, originalRequest.Response.WasModified) // Original response is never modified by definition

	// Test modified request
	modifiedRequest := webpage.Requests[1]
	assert.Equal(t, "POST", modifiedRequest.Method)
	assert.Equal(t, "https://test.com/login", modifiedRequest.RawURL)
	assert.Equal(t, "username=admin&password=admin123", modifiedRequest.Body)
	assert.True(t, modifiedRequest.WasIntercepted)
	assert.True(t, modifiedRequest.WasModified)
	assert.Equal(t, []string{"test"}, modifiedRequest.Headers["X-Injection"])

	// Test modified response
	require.NotNil(t, modifiedRequest.Response)
	assert.Equal(t, 200, modifiedRequest.Response.StatusCode)
	assert.Equal(t, "Login successful", modifiedRequest.Response.Body)
	assert.Equal(t, []string{"session=abc123"}, modifiedRequest.Response.Headers["Set-Cookie"])
	assert.True(t, modifiedRequest.Response.WasIntercepted)
	assert.True(t, modifiedRequest.Response.WasModified)
}

func TestCreateWebpage(t *testing.T) {
	urlString := "https://example.com/test"
	burpEntry := &BurpHTTPEntryRaw{
		OriginalRequest: &BurpRawRequestData{
			Body:      "test body",
			MessageID: 100,
			InScope:   true,
			Method:    "POST",
			Path:      "/test",
			URL:       urlString,
			Headers: []map[string]string{
				{"Host": "example.com"},
				{"Content-Type": "application/json"},
			},
		},
		OriginalResponse: &BurpRawResponseData{
			Body:       "response body",
			MessageID:  101,
			InScope:    true,
			StatusCode: 200,
			Headers: []map[string]string{
				{"Content-Type": "application/json"},
			},
		},
		ToolSource:             "burp.test",
		WasRequestIntercepted:  true,
		WasRequestModified:     false,
		WasResponseIntercepted: false,
		WasResponseModified:    false,
	}

	webpage, err := createWebpage(burpEntry)
	require.NoError(t, err)
	require.NotNil(t, webpage)

	// Test webpage properties
	assert.Equal(t, urlString, webpage.URL)
	assert.Equal(t, []string{BURP_COURIER_SOURCE}, webpage.Source)
	assert.Equal(t, "burp.test", webpage.Metadata["tool_source"])
	assert.Equal(t, true, webpage.Metadata["was_request_intercepted"])
	assert.Equal(t, false, webpage.Metadata["was_request_modified"])

	// Test requests
	assert.Equal(t, 1, len(webpage.Requests))
	request := webpage.Requests[0]
	assert.Equal(t, "POST", request.Method)
	assert.Equal(t, urlString, request.RawURL)
	assert.Equal(t, "test body", request.Body)
	assert.Equal(t, []string{"example.com"}, request.Headers["Host"])
	assert.Equal(t, []string{"application/json"}, request.Headers["Content-Type"])
	assert.True(t, request.WasIntercepted)
	assert.False(t, request.WasModified)

	// Test response
	require.NotNil(t, request.Response)
	assert.Equal(t, 200, request.Response.StatusCode)
	assert.Equal(t, "response body", request.Response.Body)
	assert.Equal(t, []string{"application/json"}, request.Response.Headers["Content-Type"])
	assert.False(t, request.Response.WasIntercepted)
	assert.False(t, request.Response.WasModified)
}

func TestCreateWebpage_InvalidURL(t *testing.T) {
	invalidURL := "://invalid-url"
	burpEntry := &BurpHTTPEntryRaw{
		OriginalRequest: &BurpRawRequestData{
			URL: invalidURL,
		},
	}

	webpage, err := createWebpage(burpEntry)
	assert.Error(t, err)
	assert.Nil(t, webpage)
}

func TestConvertRawRequestToWebpageRequest_WithResponse(t *testing.T) {
	rawRequest := &BurpRawRequestData{
		Body:      "request body",
		MessageID: 100,
		InScope:   true,
		Method:    "PUT",
		Path:      "/api/update",
		URL:       "https://api.example.com/api/update",
		Headers: []map[string]string{
			{"Host": "api.example.com"},
			{"Content-Type": "application/json"},
			{"Authorization": "Bearer token123"},
		},
	}

	rawResponse := &BurpRawResponseData{
		Body:       "response body",
		MessageID:  101,
		InScope:    true,
		StatusCode: 201,
		Headers: []map[string]string{
			{"Content-Type": "application/json"},
			{"Location": "/api/resource/123"},
		},
	}

	webpageRequest, err := convertRawRequestToWebpageRequest(rawRequest, rawResponse)
	require.NoError(t, err)

	// Test request fields
	assert.Equal(t, "https://api.example.com/api/update", webpageRequest.RawURL)
	assert.Equal(t, "PUT", webpageRequest.Method)
	assert.Equal(t, "request body", webpageRequest.Body)
	assert.Equal(t, []string{"api.example.com"}, webpageRequest.Headers["Host"])
	assert.Equal(t, []string{"application/json"}, webpageRequest.Headers["Content-Type"])
	assert.Equal(t, []string{"Bearer token123"}, webpageRequest.Headers["Authorization"])

	// Test response fields
	require.NotNil(t, webpageRequest.Response)
	assert.Equal(t, 201, webpageRequest.Response.StatusCode)
	assert.Equal(t, "response body", webpageRequest.Response.Body)
	assert.Equal(t, []string{"application/json"}, webpageRequest.Response.Headers["Content-Type"])
	assert.Equal(t, []string{"/api/resource/123"}, webpageRequest.Response.Headers["Location"])
}

func TestConvertRawRequestToWebpageRequest_WithoutResponse(t *testing.T) {
	rawRequest := &BurpRawRequestData{
		Body:      "request body",
		MessageID: 100,
		InScope:   true,
		Method:    "DELETE",
		Path:      "/api/delete",
		URL:       "https://api.example.com/api/delete",
		Headers: []map[string]string{
			{"Host": "api.example.com"},
		},
	}

	webpageRequest, err := convertRawRequestToWebpageRequest(rawRequest, nil)
	require.NoError(t, err)

	// Test request fields
	assert.Equal(t, "https://api.example.com/api/delete", webpageRequest.RawURL)
	assert.Equal(t, "DELETE", webpageRequest.Method)
	assert.Equal(t, "request body", webpageRequest.Body)
	assert.Equal(t, []string{"api.example.com"}, webpageRequest.Headers["Host"])

	// Test that response is nil
	assert.Nil(t, webpageRequest.Response)
}

func TestConvertRawRequestToWebpageRequest_EmptyURL(t *testing.T) {
	rawRequest := &BurpRawRequestData{
		URL: "",
	}

	webpageRequest, err := convertRawRequestToWebpageRequest(rawRequest, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "URL field is empty")
	assert.Equal(t, WebpageRequest{}, webpageRequest)
}
