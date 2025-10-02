package model

import (
	"encoding/base64"
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

	entry := burpData.Data["14"]
	assert.Equal(t, "GET", entry.OriginalRequest.Method)
	assert.Equal(t, "http://example.com:8888/", entry.OriginalRequest.RawURL)
	assert.Equal(t, "burp.test", entry.ToolSource)
	assert.Equal(t, 0, entry.RequestMessageID)
	assert.False(t, entry.RequestInScope)
}

func TestBurpHTTPData_ExtractBaseURLs(t *testing.T) {
	burpData := &BurpHTTPData{
		Data: map[string]BurpHTTPEntry{
			"1": {
				OriginalRequest: &WebpageRequest{
					RawURL: "http://api.example.com:8888/api/users",
					Method: "GET",
				},
			},
			"2": {
				OriginalRequest: &WebpageRequest{
					RawURL: "http://api.example.com:8888/login",
					Method: "GET",
				},
			},
			"3": {
				OriginalRequest: &WebpageRequest{
					RawURL: "https://secure.example.com:443/",
					Method: "GET",
				},
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
		Data: map[string]BurpHTTPEntry{
			"1": {
				OriginalRequest: &WebpageRequest{
					RawURL: "http://api.example.com:8888/api/users",
					Method: "GET",
					Body:   "",
					Headers: map[string][]string{
						"Host":       {"api.example.com:8888"},
						"User-Agent": {"Mozilla/5.0"},
					},
				},
				OriginalResponse: &WebpageResponse{
					Body: `{"users": []}`,
					Headers: map[string][]string{
						"Content-Type": {"application/json"},
					},
					StatusCode: 200,
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
		Data: map[string]BurpHTTPEntry{
			"1": {ToolSource: "burp.proxy"},
			"2": {ToolSource: "burp.repeater"},
			"3": {ToolSource: "burp.proxy"},
			"4": {ToolSource: ""},
		},
	}

	sources := burpData.ExtractToolSources()
	assert.Equal(t, 2, len(sources))
	assert.Contains(t, sources, "burp.proxy")
	assert.Contains(t, sources, "burp.repeater")
}

func TestBurpHTTPEntry_GetBurpMetadata(t *testing.T) {
	entry := BurpHTTPEntry{
		RequestMessageID:             123,
		RequestInScope:               true,
		ResponseMessageID:            124,
		ResponseInScope:              false,
		ToolSource:                   "burp.repeater",
		WasRequestIntercepted:        true,
		WasResponseIntercepted:       false,
		WasRequestModified:           true,
		WasResponseModified:          false,
		WasRequestBodyBase64Encoded:  false,
		WasResponseBodyBase64Encoded: true,
	}

	metadata := entry.GetBurpMetadata()

	assert.Equal(t, 123, metadata["request_message_id"])
	assert.Equal(t, true, metadata["request_in_scope"])
	assert.Equal(t, 124, metadata["response_message_id"])
	assert.Equal(t, false, metadata["response_in_scope"])
	assert.Equal(t, "burp.repeater", metadata["tool_source"])
	assert.Equal(t, true, metadata["was_request_intercepted"])
	assert.Equal(t, false, metadata["was_response_intercepted"])
	assert.Equal(t, true, metadata["was_request_modified"])
	assert.Equal(t, false, metadata["was_response_modified"])
	assert.Equal(t, false, metadata["was_request_body_base64_encoded"])
	assert.Equal(t, true, metadata["was_response_body_base64_encoded"])
}

func TestBurpHTTPEntry_HelperMethods(t *testing.T) {
	entry := BurpHTTPEntry{
		OriginalRequest: &WebpageRequest{
			Method: "GET",
		},
		OriginalResponse: &WebpageResponse{
			StatusCode: 200,
		},
		ModifiedRequest: &WebpageRequest{
			Method: "POST",
		},
		RequestMessageID:  123,
		RequestInScope:    true,
		ResponseMessageID: 124,
	}

	assert.True(t, entry.HasModifiedRequest())
	assert.False(t, entry.HasModifiedResponse())
	assert.Equal(t, 123, entry.GetOriginalRequestMessageID())
	assert.Equal(t, 124, entry.GetOriginalResponseMessageID())
	assert.True(t, entry.IsInScope())
}

func TestBurpHTTPEntry_HelperMethods_NilRequests(t *testing.T) {
	entry := BurpHTTPEntry{}

	assert.False(t, entry.HasModifiedRequest())
	assert.False(t, entry.HasModifiedResponse())
	assert.Equal(t, 0, entry.GetOriginalRequestMessageID())
	assert.Equal(t, 0, entry.GetOriginalResponseMessageID())
	assert.False(t, entry.IsInScope())
}

func TestBurpHTTPEntry_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"originalRequest": {
			"body": "test body",
			"messageId": 456,
			"inScope": true,
			"method": "POST",
			"path": "/test",
			"url": "https://test.example.com:443/test",
			"headers": [
				{"Host": "test.example.com:443"},
				{"Content-Type": "application/json"}
			]
		},
		"originalResponse": {
			"body": "response body",
			"messageId": 457,
			"inScope": false,
			"headers": [
				{"Content-Type": "text/html"}
			]
		},
		"toolSource": "burp.proxy",
		"wasRequestIntercepted": true
	}`

	var entry BurpHTTPEntry
	err := entry.UnmarshalJSON([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "POST", entry.OriginalRequest.Method)
	assert.Equal(t, "https://test.example.com:443/test", entry.OriginalRequest.RawURL)
	assert.Equal(t, "test body", entry.OriginalRequest.Body)
	assert.Equal(t, "response body", entry.OriginalResponse.Body)
	assert.Equal(t, "burp.proxy", entry.ToolSource)
	assert.Equal(t, 456, entry.RequestMessageID)
	assert.True(t, entry.RequestInScope)
	assert.Equal(t, 457, entry.ResponseMessageID)
	assert.False(t, entry.ResponseInScope)
	assert.True(t, entry.WasRequestIntercepted)
}

func TestBurpHTTPEntry_UnmarshalJSON_ComplexWithModified(t *testing.T) {
	jsonData := `{
		"originalRequest": {
			"body": "original request body",
			"messageId": 100,
			"inScope": true,
			"method": "POST",
			"path": "/api/login",
			"url": "https://example.com:443/api/login",
			"headers": [
				{"Host": "example.com:443"},
				{"Content-Type": "application/json"},
				{"Authorization": "Bearer original-token"}
			]
		},
		"originalResponse": {
			"body": "{\"status\": \"success\", \"token\": \"abc123\"}",
			"messageId": 101,
			"inScope": true,
			"headers": [
				{"Content-Type": "application/json"},
				{"Set-Cookie": "session=xyz789"}
			]
		},
		"modifiedRequest": {
			"body": "modified request body with injection",
			"messageId": 102,
			"inScope": true,
			"method": "POST",
			"path": "/api/login",
			"url": "https://example.com:443/api/login",
			"headers": [
				{"Host": "example.com:443"},
				{"Content-Type": "application/json"},
				{"Authorization": "Bearer modified-token"},
				{"X-Injection": "malicious-payload"}
			]
		},
		"modifiedResponse": {
			"body": "{\"status\": \"error\", \"message\": \"access denied\"}",
			"messageId": 103,
			"inScope": false,
			"headers": [
				{"Content-Type": "application/json"},
				{"X-Error": "auth-failed"}
			]
		},
		"toolSource": "burp.repeater",
		"wasRequestIntercepted": true,
		"wasResponseIntercepted": true,
		"wasRequestModified": true,
		"wasResponseModified": true,
		"wasRequestBodyBase64Encoded": false,
		"wasResponseBodyBase64Encoded": false,
		"wasModifiedRequestBodyBase64Encoded": false,
		"wasModifiedResponseBodyBase64Encoded": false
	}`

	var entry BurpHTTPEntry
	err := entry.UnmarshalJSON([]byte(jsonData))
	require.NoError(t, err)

	// Test original request
	require.NotNil(t, entry.OriginalRequest)
	assert.Equal(t, "POST", entry.OriginalRequest.Method)
	assert.Equal(t, "https://example.com:443/api/login", entry.OriginalRequest.RawURL)
	assert.Equal(t, "original request body", entry.OriginalRequest.Body)
	assert.Equal(t, []string{"example.com:443"}, entry.OriginalRequest.Headers["Host"])
	assert.Equal(t, []string{"application/json"}, entry.OriginalRequest.Headers["Content-Type"])
	assert.Equal(t, []string{"Bearer original-token"}, entry.OriginalRequest.Headers["Authorization"])

	// Test original response
	require.NotNil(t, entry.OriginalResponse)
	assert.Equal(t, `{"status": "success", "token": "abc123"}`, entry.OriginalResponse.Body)
	assert.Equal(t, []string{"application/json"}, entry.OriginalResponse.Headers["Content-Type"])
	assert.Equal(t, []string{"session=xyz789"}, entry.OriginalResponse.Headers["Set-Cookie"])

	// Test modified request
	require.NotNil(t, entry.ModifiedRequest)
	assert.Equal(t, "POST", entry.ModifiedRequest.Method)
	assert.Equal(t, "https://example.com:443/api/login", entry.ModifiedRequest.RawURL)
	assert.Equal(t, "modified request body with injection", entry.ModifiedRequest.Body)
	assert.Equal(t, []string{"example.com:443"}, entry.ModifiedRequest.Headers["Host"])
	assert.Equal(t, []string{"application/json"}, entry.ModifiedRequest.Headers["Content-Type"])
	assert.Equal(t, []string{"Bearer modified-token"}, entry.ModifiedRequest.Headers["Authorization"])
	assert.Equal(t, []string{"malicious-payload"}, entry.ModifiedRequest.Headers["X-Injection"])

	// Test modified response
	require.NotNil(t, entry.ModifiedResponse)
	assert.Equal(t, `{"status": "error", "message": "access denied"}`, entry.ModifiedResponse.Body)
	assert.Equal(t, []string{"application/json"}, entry.ModifiedResponse.Headers["Content-Type"])
	assert.Equal(t, []string{"auth-failed"}, entry.ModifiedResponse.Headers["X-Error"])

	// Test Burp metadata
	assert.Equal(t, "burp.repeater", entry.ToolSource)
	assert.Equal(t, 100, entry.RequestMessageID)
	assert.True(t, entry.RequestInScope)
	assert.Equal(t, 101, entry.ResponseMessageID)
	assert.True(t, entry.ResponseInScope)

	// Test interception and modification flags
	assert.True(t, entry.WasRequestIntercepted)
	assert.True(t, entry.WasResponseIntercepted)
	assert.True(t, entry.WasRequestModified)
	assert.True(t, entry.WasResponseModified)
	assert.False(t, entry.WasRequestBodyBase64Encoded)
	assert.False(t, entry.WasResponseBodyBase64Encoded)
	assert.False(t, entry.WasModifiedRequestBodyBase64Encoded)
	assert.False(t, entry.WasModifiedResponseBodyBase64Encoded)

	// Test helper methods
	assert.True(t, entry.HasModifiedRequest())
	assert.True(t, entry.HasModifiedResponse())
	assert.Equal(t, 100, entry.GetOriginalRequestMessageID())
	assert.Equal(t, 101, entry.GetOriginalResponseMessageID())
	assert.True(t, entry.IsInScope())

	// Test metadata extraction
	metadata := entry.GetBurpMetadata()
	assert.Equal(t, 100, metadata["request_message_id"])
	assert.Equal(t, true, metadata["request_in_scope"])
	assert.Equal(t, 101, metadata["response_message_id"])
	assert.Equal(t, true, metadata["response_in_scope"])
	assert.Equal(t, "burp.repeater", metadata["tool_source"])
	assert.Equal(t, true, metadata["was_request_intercepted"])
	assert.Equal(t, true, metadata["was_response_intercepted"])
	assert.Equal(t, true, metadata["was_request_modified"])
	assert.Equal(t, true, metadata["was_response_modified"])
	assert.Equal(t, false, metadata["was_request_body_base64_encoded"])
	assert.Equal(t, false, metadata["was_response_body_base64_encoded"])
}

func TestBurpHTTPEntry_UnmarshalJSON_Base64EncodedModifiedRequest(t *testing.T) {
	// Base64 encode the bytes 0xdeadcafebabe
	binaryData := []byte{0xde, 0xad, 0xca, 0xfe, 0xba, 0xbe}
	base64EncodedData := base64.StdEncoding.EncodeToString(binaryData)

	jsonData := `{
		"originalRequest": {
			"body": "normal plain text request body",
			"messageId": 200,
			"inScope": true,
			"method": "POST",
			"path": "/api/upload",
			"url": "https://api.example.com/api/upload",
			"headers": [
				{"Host": "api.example.com"},
				{"Content-Type": "text/plain"}
			]
		},
		"originalResponse": {
			"body": "{\"status\": \"received\"}",
			"messageId": 201,
			"inScope": true,
			"headers": [
				{"Content-Type": "application/json"}
			]
		},
		"modifiedRequest": {
			"body": "` + base64EncodedData + `",
			"messageId": 202,
			"inScope": true,
			"method": "POST",
			"path": "/api/upload",
			"url": "https://api.example.com/api/upload",
			"headers": [
				{"Host": "api.example.com"},
				{"Content-Type": "application/octet-stream"},
				{"Content-Encoding": "base64"}
			]
		},
		"modifiedResponse": {
			"body": "{\"status\": \"error\", \"message\": \"malicious content detected\"}",
			"messageId": 203,
			"inScope": false,
			"headers": [
				{"Content-Type": "application/json"},
				{"X-Security-Alert": "true"}
			]
		},
		"toolSource": "burp.intruder",
		"wasRequestIntercepted": true,
		"wasResponseIntercepted": true,
		"wasRequestModified": true,
		"wasResponseModified": true,
		"wasRequestBodyBase64Encoded": false,
		"wasResponseBodyBase64Encoded": false,
		"wasModifiedRequestBodyBase64Encoded": true,
		"wasModifiedResponseBodyBase64Encoded": false
	}`

	var entry BurpHTTPEntry
	err := entry.UnmarshalJSON([]byte(jsonData))
	require.NoError(t, err)

	// Test original request - should contain plain text
	require.NotNil(t, entry.OriginalRequest)
	assert.Equal(t, "normal plain text request body", entry.OriginalRequest.Body)
	assert.Equal(t, "POST", entry.OriginalRequest.Method)
	assert.Equal(t, "https://api.example.com/api/upload", entry.OriginalRequest.RawURL)
	assert.Equal(t, []string{"text/plain"}, entry.OriginalRequest.Headers["Content-Type"])

	// Test modified request - should contain base64 encoded data
	require.NotNil(t, entry.ModifiedRequest)
	assert.Equal(t, base64EncodedData, entry.ModifiedRequest.Body)
	assert.Equal(t, "POST", entry.ModifiedRequest.Method)
	assert.Equal(t, "https://api.example.com/api/upload", entry.ModifiedRequest.RawURL)
	assert.Equal(t, []string{"application/octet-stream"}, entry.ModifiedRequest.Headers["Content-Type"])
	assert.Equal(t, []string{"base64"}, entry.ModifiedRequest.Headers["Content-Encoding"])

	// Test original response
	require.NotNil(t, entry.OriginalResponse)
	assert.Equal(t, `{"status": "received"}`, entry.OriginalResponse.Body)
	assert.Equal(t, []string{"application/json"}, entry.OriginalResponse.Headers["Content-Type"])

	// Test modified response
	require.NotNil(t, entry.ModifiedResponse)
	assert.Equal(t, `{"status": "error", "message": "malicious content detected"}`, entry.ModifiedResponse.Body)
	assert.Equal(t, []string{"application/json"}, entry.ModifiedResponse.Headers["Content-Type"])
	assert.Equal(t, []string{"true"}, entry.ModifiedResponse.Headers["X-Security-Alert"])

	// Test Burp metadata
	assert.Equal(t, "burp.intruder", entry.ToolSource)
	assert.Equal(t, 200, entry.RequestMessageID)
	assert.True(t, entry.RequestInScope)
	assert.Equal(t, 201, entry.ResponseMessageID)
	assert.True(t, entry.ResponseInScope)

	// Test Base64 encoding flags - this is the key test
	assert.False(t, entry.WasRequestBodyBase64Encoded)        // Original request is NOT base64 encoded
	assert.True(t, entry.WasModifiedRequestBodyBase64Encoded) // Modified request IS base64 encoded
	assert.False(t, entry.WasResponseBodyBase64Encoded)
	assert.False(t, entry.WasModifiedResponseBodyBase64Encoded)

	// Test other flags
	assert.True(t, entry.WasRequestIntercepted)
	assert.True(t, entry.WasResponseIntercepted)
	assert.True(t, entry.WasRequestModified)
	assert.True(t, entry.WasResponseModified)

	// Test helper methods
	assert.True(t, entry.HasModifiedRequest())
	assert.True(t, entry.HasModifiedResponse())

	// Verify the base64 data decodes correctly to our expected bytes
	// This validates that the test data is correct
	decodedBytes, err := base64.StdEncoding.DecodeString(entry.ModifiedRequest.Body)
	require.NoError(t, err)
	assert.Equal(t, binaryData, decodedBytes)
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

	issuesByURL := burpData.GetIssuesByBaseURL()
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
