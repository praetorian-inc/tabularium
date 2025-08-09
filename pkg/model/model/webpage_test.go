package model

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
)

const (
	testBaseURL     = "https://example.com"
	testPath        = "/path"
	testQuery       = "?query=value"
	testFragment    = "#fragment"
	testUserAgent   = "Test/1.0"
	testContentType = "application/json"
)

type webpageTestCase struct {
	name          string
	url           string
	request       WebpageRequest
	expectedURL   string
	expectedState string
}

type stateTestCase struct {
	name          string
	initialState1 string
	initialState2 string
	expectedState string
}

type metadataTestCase struct {
	name             string
	initialMetadata1 map[string]any
	initialMetadata2 map[string]any
	expectedMetadata map[string]any
}

var basicWebpageTestCases = []webpageTestCase{
	{
		name:          "basic URL",
		url:           testBaseURL,
		request:       WebpageRequest{RawURL: testBaseURL + "/"},
		expectedURL:   testBaseURL + "/",
		expectedState: Unanalyzed,
	},
	{
		name:          "URL with path",
		url:           testBaseURL + testPath,
		request:       WebpageRequest{RawURL: testBaseURL + testPath},
		expectedURL:   testBaseURL + testPath,
		expectedState: Unanalyzed,
	},
	{
		name:          "URL with query parameters",
		url:           testBaseURL + testPath + testQuery,
		request:       WebpageRequest{RawURL: testBaseURL + testPath + testQuery},
		expectedURL:   testBaseURL + testPath,
		expectedState: Interesting,
	},
	{
		name:          "URL with query and fragment",
		url:           testBaseURL + testPath + testQuery + testFragment,
		request:       WebpageRequest{RawURL: testBaseURL + testPath + testQuery + testFragment},
		expectedURL:   testBaseURL + testPath,
		expectedState: Interesting,
	},
}

func TestWebpageConstructors(t *testing.T) {
	parent := createTestParent()

	for _, tc := range basicWebpageTestCases {
		t.Run(tc.name+" (from string)", func(t *testing.T) {
			webpage := NewWebpageFromString(tc.url, parent, WithRequests(tc.request))
			assertWebpage(t, webpage, tc.expectedURL, tc.expectedState)
		})

		t.Run(tc.name+" (from URL)", func(t *testing.T) {
			parsedURL, err := url.Parse(tc.url)
			assert.NoError(t, err)
			webpage := NewWebpage(*parsedURL, parent, WithRequests(tc.request))
			assertWebpage(t, webpage, tc.expectedURL, tc.expectedState)
		})
	}
}

func TestWebpage_IsPrivate(t *testing.T) {
	publicAsset := NewAsset("contoso.com", "18.1.2.4")
	privateAsset := NewAsset("contoso.com", "10.0.0.1")

	privateHTTPS := NewAttribute("https", "443", &privateAsset)
	publicHTTPS := NewAttribute("https", "443", &publicAsset)

	privatePort := NewAttribute("port", "443", &privateAsset)
	publicPort := NewAttribute("port", "443", &publicAsset)

	url, _ := url.Parse("https://contoso.com:443")

	w := NewWebpage(*url, &privateHTTPS)
	assert.True(t, w.IsPrivate(), "private https should be private")

	w = NewWebpage(*url, &publicHTTPS)
	assert.False(t, w.IsPrivate(), "public https should not be private")

	w = NewWebpage(*url, &privatePort)
	assert.True(t, w.IsPrivate(), "private port should be private")

	w = NewWebpage(*url, &publicPort)
	assert.False(t, w.IsPrivate(), "public port should not be private")

	w.Parent.Model = nil
	marshalled, err := json.Marshal(w)
	assert.NoError(t, err)

	var unmarshalled Webpage
	err = json.Unmarshal(marshalled, &unmarshalled)
	assert.NoError(t, err)
	assert.False(t, unmarshalled.IsPrivate(), "no parent should not be private")
}

func TestWebpageConstructorEdgeCases(t *testing.T) {
	parent := createTestParent()

	testCases := map[string]func(*testing.T){
		"empty path gets default": func(t *testing.T) {
			u, _ := url.Parse(testBaseURL)
			u.Path = ""
			webpage := NewWebpage(*u, parent)
			assertWebpageURL(t, webpage, testBaseURL+"/")
		},
		"parent assignment": func(t *testing.T) {
			webpage := createTestWebpage(testBaseURL + testPath)
			assert.Equal(t, parent, webpage.Parent.Model)
		},
		"valid webpage creation": func(t *testing.T) {
			webpage := createTestWebpage(testBaseURL + testPath)
			assert.True(t, webpage.Valid())
		},
	}

	for name, testFunc := range testCases {
		t.Run(name, testFunc)
	}
}

var stateMergingTestCases = []stateTestCase{
	{"Interesting overwrites Unanalyzed", Unanalyzed, Interesting, Interesting},
	{"Uninteresting overwrites Unanalyzed", Unanalyzed, Uninteresting, Uninteresting},
	{"Interesting overwrites Uninteresting", Uninteresting, Interesting, Interesting},
	{"Uninteresting does not overwrite Interesting", Interesting, Uninteresting, Interesting},
	{"Unanalyzed does not overwrite Interesting", Interesting, Unanalyzed, Interesting},
}

var metadataMergingTestCases = []metadataTestCase{
	{
		name:             "Add new metadata",
		initialMetadata1: map[string]any{"key1": "value1"},
		initialMetadata2: map[string]any{"key2": "value2"},
		expectedMetadata: map[string]any{"key1": "value1", "key2": "value2"},
	},
	{
		name:             "Overwrite existing single values",
		initialMetadata1: map[string]any{"key1": "value1"},
		initialMetadata2: map[string]any{"key1": "value2", "key2": "value3"},
		expectedMetadata: map[string]any{"key1": "value2", "key2": "value3"},
	},
	{
		name:             "Append to array of any values",
		initialMetadata1: map[string]any{"key1": []any{"value1"}},
		initialMetadata2: map[string]any{"key1": []any{"value2", "value3"}},
		expectedMetadata: map[string]any{"key1": []any{"value1", "value2", "value3"}},
	},
	{
		name:             "Dont append duplicate values",
		initialMetadata1: map[string]any{"key1": []any{"value1"}},
		initialMetadata2: map[string]any{"key1": []any{"value1", "value2"}},
		expectedMetadata: map[string]any{"key1": []any{"value1", "value2"}},
	},
	{
		name:             "Append to array of string values",
		initialMetadata1: map[string]any{"key1": []string{"value1"}},
		initialMetadata2: map[string]any{"key1": []string{"value2", "value3"}},
		expectedMetadata: map[string]any{"key1": []string{"value1", "value2", "value3"}},
	},
	{
		name:             "Merge with empty initial metadata",
		initialMetadata1: map[string]any{},
		initialMetadata2: map[string]any{"key1": "value1"},
		expectedMetadata: map[string]any{"key1": "value1"},
	},
	{
		name:             "Merge empty metadata into existing",
		initialMetadata1: map[string]any{"key1": "value1"},
		initialMetadata2: map[string]any{},
		expectedMetadata: map[string]any{"key1": "value1"},
	},
	{
		name:             "Merge multiple metadata",
		initialMetadata1: map[string]any{"key1": "value1", "key2": []string{"value2"}},
		initialMetadata2: map[string]any{"key2": []string{"value3"}, "key3": "value4"},
		expectedMetadata: map[string]any{"key1": "value1", "key2": []string{"value2", "value3"}, "key3": "value4"},
	},
}

func TestWebpageMerge(t *testing.T) {
	t.Run("state merging", func(t *testing.T) {
		for _, tc := range stateMergingTestCases {
			t.Run(tc.name, func(t *testing.T) {
				webpage1 := createTestWebpage(testBaseURL+testPath, WithState(tc.initialState1))
				webpage2 := createTestWebpage(testBaseURL+testPath, WithState(tc.initialState2))
				webpage1.Merge(webpage2)
				assertWebpageState(t, webpage1, tc.expectedState)
			})
		}
	})

	t.Run("metadata merging", func(t *testing.T) {
		for _, tc := range metadataMergingTestCases {
			t.Run(tc.name, func(t *testing.T) {
				webpage1, webpage2 := createTestWebpagePair(testBaseURL + testPath)
				webpage1.Metadata = tc.initialMetadata1
				webpage2.Metadata = tc.initialMetadata2
				webpage1.Merge(webpage2)
				assert.Equal(t, tc.expectedMetadata, webpage1.Metadata)
			})
		}
	})

	t.Run("metadata merging edge cases", func(t *testing.T) {
		edgeCases := []metadataTestCase{
			{
				name:             "merge nil metadata",
				initialMetadata1: nil,
				initialMetadata2: map[string]any{"key": "value"},
				expectedMetadata: map[string]any{"key": "value"},
			},
			{
				name:             "merge into nil metadata",
				initialMetadata1: map[string]any{"key": "value"},
				initialMetadata2: nil,
				expectedMetadata: map[string]any{"key": "value"},
			},
			{
				name:             "merge different slice types",
				initialMetadata1: map[string]any{"key": []int{1, 2}},
				initialMetadata2: map[string]any{"key": []string{"a", "b"}},
				expectedMetadata: map[string]any{"key": []string{"a", "b"}},
			},
			{
				name:             "merge non-slice types",
				initialMetadata1: map[string]any{"key": "original"},
				initialMetadata2: map[string]any{"key": "replacement"},
				expectedMetadata: map[string]any{"key": "replacement"},
			},
		}

		for _, tc := range edgeCases {
			t.Run(tc.name, func(t *testing.T) {
				webpage1, webpage2 := createTestWebpages()
				webpage1.Metadata = tc.initialMetadata1
				webpage2.Metadata = tc.initialMetadata2
				webpage1.Merge(webpage2)
				assert.Equal(t, tc.expectedMetadata, webpage1.Metadata)
			})
		}
	})

	t.Run("source merging", func(t *testing.T) {
		sourceTestCases := []struct {
			name            string
			initialSources1 []string
			initialSources2 []string
			expectedSources []string
		}{
			{"merge different sources", []string{"crawl"}, []string{"login"}, []string{"crawl", "login"}},
			{"avoid duplicate sources", []string{"crawl"}, []string{"crawl"}, []string{"crawl"}},
			{"preserve order", []string{"first", "second"}, []string{"third", "fourth"}, []string{"first", "second", "third", "fourth"}},
			{"merge with empty sources", []string{"existing"}, []string{}, []string{"existing"}},
		}

		parent := createTestParent()
		for _, tc := range sourceTestCases {
			t.Run(tc.name, func(t *testing.T) {
				webpage1 := NewWebpageFromString(testBaseURL+"/", parent)
				webpage1.Source = tc.initialSources1
				webpage2 := NewWebpageFromString(testBaseURL+"/", parent)
				webpage2.Source = tc.initialSources2
				webpage1.Visit(webpage2)
				assert.Equal(t, tc.expectedSources, webpage1.Source)
			})
		}
	})
}

func TestWebpageRequestManagement(t *testing.T) {
	t.Run("add requests", func(t *testing.T) {
		req1 := createTestRequest(testBaseURL+"/page1", "GET", "body1")
		req2 := createTestRequest(testBaseURL+"/page2", "POST", "body2")
		webpage := createTestWebpage(testBaseURL+"/", WithRequests(req1, req2))

		assertWebpageRequestCount(t, webpage, 2)
		assert.Equal(t, req1, webpage.Requests[0])
		assert.Equal(t, req2, webpage.Requests[1])
	})

	t.Run("merge requests", func(t *testing.T) {
		testCases := []struct {
			name             string
			requests1        []WebpageRequest
			requests2        []WebpageRequest
			expectedLength   int
			shouldHaveUpdate bool
		}{
			{
				name: "no duplicates",
				requests1: []WebpageRequest{
					{RawURL: testBaseURL + "/1", Method: "GET", Body: "body1"},
					{RawURL: testBaseURL + "/2", Method: "GET", Body: "body2"},
				},
				requests2: []WebpageRequest{
					{RawURL: testBaseURL + "/3", Method: "GET", Body: "body3"},
					{RawURL: testBaseURL + "/4", Method: "GET", Body: "body4"},
				},
				expectedLength: 4,
			},
			{
				name: "with duplicates",
				requests1: []WebpageRequest{
					{RawURL: testBaseURL + "/1", Method: "GET", Body: "body1"},
					{RawURL: testBaseURL + "/2", Method: "GET", Body: "body2"},
				},
				requests2: []WebpageRequest{
					{RawURL: testBaseURL + "/1", Method: "GET", Body: "body1_updated"},
					{RawURL: testBaseURL + "/3", Method: "GET", Body: "body3"},
				},
				expectedLength:   4,
				shouldHaveUpdate: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				webpage1, webpage2 := createTestWebpagePair(testBaseURL + "/")
				webpage1.Requests = tc.requests1
				webpage2.Requests = tc.requests2
				webpage1.Merge(webpage2)
				assertWebpageRequestCount(t, webpage1, tc.expectedLength)

				if tc.shouldHaveUpdate {
					found := false
					for _, req := range webpage1.Requests {
						if req.RawURL == testBaseURL+"/1" && req.Body == "body1_updated" {
							found = true
							break
						}
					}
					assert.True(t, found, "Updated duplicate request should be present")
				}
			})
		}
	})

	t.Run("max request limit", func(t *testing.T) {
		webpage1, webpage2 := createTestWebpagePair(testBaseURL + "/")

		for i := 0; i < DefaultMaxRequestsPerWebpage+10; i++ {
			req := createTestRequest(fmt.Sprintf("%s/%d", testBaseURL, i), "GET", fmt.Sprintf("body%d", i))
			if i < DefaultMaxRequestsPerWebpage/2 {
				webpage1.AddRequest(req)
			} else {
				webpage2.AddRequest(req)
			}
		}

		webpage1.Merge(webpage2)
		assert.Equal(t, DefaultMaxRequestsPerWebpage, len(webpage1.Requests))
	})
}

func TestWebpageBasicAnalysis(t *testing.T) {
	testCases := []struct {
		name           string
		rawURL         string
		initialState   string
		expectedState  string
		expectedParams bool
	}{
		{"no parameters", testBaseURL + "/page", "", Unanalyzed, false},
		{"with parameters", testBaseURL + "/page?param=value", "", Interesting, true},
		{"only version parameter", testBaseURL + "/page?ver=1.0", "", Unanalyzed, false},
		{"already interesting", testBaseURL + "/page?param=value", Interesting, Interesting, true},
		{"uninteresting to interesting", testBaseURL + "/page?param=value", Uninteresting, Interesting, true},
		{"invalid URL", "invalid-url", "", Unanalyzed, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var options []WebpageOption
			if tc.initialState != "" {
				options = append(options, WithState(tc.initialState))
			}
			options = append(options, WithRequests(WebpageRequest{RawURL: tc.rawURL}))

			webpage := createTestWebpage(testBaseURL+"/page", options...)

			assertWebpageState(t, webpage, tc.expectedState)
			if tc.expectedParams {
				assertWebpageMetadata(t, webpage, PARAMETERS_IDENTIFIED, true)
			}
		})
	}
}

func TestWebpageURLParsing(t *testing.T) {
	testCases := []struct {
		name             string
		url              string
		expectedProtocol string
		expectedHostname string
		expectedPath     string
		expectedPort     int
	}{
		{"basic URL with port", "https://example.com:8080/path/to/resource", "https", "example.com", "/path/to/resource", 8080},
		{"HTTP default port", "http://example.com/", "http", "example.com", "/", 80},
		{"HTTPS default port", "https://example.com/", "https", "example.com", "/", 443},
		{"default path", "https://example.com", "https", "example.com", "/", 443},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			webpage := createTestWebpage(tc.url)
			assert.Equal(t, tc.expectedProtocol, webpage.Protocol())
			assert.Equal(t, tc.expectedHostname, webpage.Hostname())
			assert.Equal(t, tc.expectedPath, webpage.UrlPath())
			assert.Equal(t, tc.expectedPort, webpage.Port())
		})
	}
}

func TestWebpageValidation(t *testing.T) {
	parent := createTestParent()

	testCases := []struct {
		name      string
		setupFunc func() Webpage
		isValid   bool
	}{
		{
			name:      "valid webpage",
			setupFunc: func() Webpage { return NewWebpageFromString(testBaseURL+testPath, parent) },
			isValid:   true,
		},
		{
			name:      "invalid key",
			setupFunc: func() Webpage { return Webpage{Key: "invalid-key"} },
			isValid:   false,
		},
		{
			name: "long URL key truncation",
			setupFunc: func() Webpage {
				longPath := strings.Repeat("a", 2100)
				return NewWebpageFromString(testBaseURL+"/"+longPath, parent)
			},
			isValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			webpage := tc.setupFunc()
			assert.Equal(t, tc.isValid, webpage.Valid())

			if tc.name == "valid webpage" {
				assert.Contains(t, webpage.Key, "#webpage#"+testBaseURL+testPath)
			}
			if tc.name == "long URL key truncation" {
				assert.LessOrEqual(t, len(webpage.Key), 2048)
			}
		})
	}
}

func TestWebpageParent(t *testing.T) {
	expectedParent := createTestParent()
	webpage1, _ := createTestWebpages()
	assert.Equal(t, expectedParent, webpage1.Parent.Model)
}

func TestWebpageHydrationAndDehydration(t *testing.T) {
	t.Run("basic hydration and dehydration", func(t *testing.T) {
		webpage := NewWebpageFromString(testBaseURL+testPath, createTestParent())
		expectedDetails := createTestWebpageDetails("0")
		webpage.WebpageDetails = expectedDetails

		detailsFile, dehydratedWebpage := webpage.Dehydrate()
		detailsPath, hydrate := dehydratedWebpage.Hydrate()

		assert.True(t, strings.HasPrefix(detailsPath, "webpage/example.com/443/"+RemoveReservedCharacters(testBaseURL+testPath)+"/details"))

		err := hydrate(detailsFile.Bytes)
		assert.NoError(t, err)
		assert.Equal(t, expectedDetails, dehydratedWebpage.(*Webpage).WebpageDetails)

		var fileDetails WebpageDetails
		err = json.Unmarshal(detailsFile.Bytes, &fileDetails)
		assert.NoError(t, err)
		assert.Equal(t, expectedDetails, fileDetails)
	})

	t.Run("dehydration edge cases", func(t *testing.T) {
		parent := createTestParent()
		testCases := []struct {
			name        string
			setupFunc   func() Webpage
			expectEmpty bool
		}{
			{
				name: "with max requests",
				setupFunc: func() Webpage {
					webpage := NewWebpageFromString(testBaseURL+"/", parent)
					for i := 0; i < DefaultMaxRequestsPerWebpage+10; i++ {
						webpage.AddRequest(createTestRequest(fmt.Sprintf("%s/%d", testBaseURL, i), "GET", fmt.Sprintf("body%d", i)))
					}
					return webpage
				},
			},
			{
				name:        "with empty requests",
				setupFunc:   func() Webpage { return NewWebpageFromString(testBaseURL+"/", parent) },
				expectEmpty: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				webpage := tc.setupFunc()
				file, dehydrated := webpage.Dehydrate()

				assert.NotEmpty(t, file.Bytes)
				assert.Equal(t, 0, len(dehydrated.(*Webpage).Requests))
				assert.NotEmpty(t, dehydrated.(*Webpage).DetailsFilepath)

				var details WebpageDetails
				err := json.Unmarshal(file.Bytes, &details)
				assert.NoError(t, err)

				if tc.expectEmpty {
					assert.Empty(t, details.Requests)
				} else {
					assert.Equal(t, DefaultMaxRequestsPerWebpage, len(details.Requests))
				}
			})
		}
	})
}

func TestWebpageFileGeneration(t *testing.T) {
	parent := createTestParent()
	webpage := NewWebpageFromString(testBaseURL+testPath, parent)

	t.Run("details file generation", func(t *testing.T) {
		details := createTestWebpageDetails("test1")
		file := webpage.GetDetailsFile(details)

		assert.NotEmpty(t, file.Name)
		assert.Contains(t, file.Name, "webpage/example.com/443/")
		assert.Contains(t, file.Name, "details")

		var unmarshalled WebpageDetails
		err := json.Unmarshal(file.Bytes, &unmarshalled)
		assert.NoError(t, err)
		assert.Equal(t, details, unmarshalled)
	})

	t.Run("response body file generation", func(t *testing.T) {
		request := createTestRequest(testBaseURL+"/test", "GET", "test body")
		request.Response.Body = "response content"

		file := webpage.GetResponseBodyAsFile(request)
		assert.NotEmpty(t, file.Name)
		assert.Contains(t, file.Name, "webpage/example.com/443/")
		assert.Equal(t, SmartBytes([]byte("response content")), file.Bytes)
	})
}

func TestWebpageGetLabels(t *testing.T) {
	webpage := &Webpage{}
	labels := webpage.GetLabels()
	assert.Contains(t, labels, "Webpage")
	assert.Contains(t, labels, TTLLabel.String())
}

func TestWebpageGetDescription(t *testing.T) {
	webpage := &Webpage{}
	description := webpage.GetDescription()
	assert.NotEmpty(t, description)
	assert.Contains(t, description, "webpage")
	assert.Contains(t, description, "URL")
}

func TestWebpageMergeRequests(t *testing.T) {
	parent := createTestParent()

	t.Run("merge new requests", func(t *testing.T) {
		webpage := NewWebpageFromString(testBaseURL+"/", parent)
		req1 := createTestRequest(testBaseURL+"/page1", "GET", "body1")
		req2 := createTestRequest(testBaseURL+"/page2", "POST", "body2")
		webpage.AddRequest(req1)

		req3 := createTestRequest(testBaseURL+"/page3", "GET", "body3")
		req4 := createTestRequest(testBaseURL+"/page4", "PUT", "body4")
		webpage.MergeRequests(req2, req3, req4)

		assert.Len(t, webpage.Requests, 4)
		assert.Equal(t, req1.RawURL, webpage.Requests[0].RawURL)
	})

	t.Run("merge with duplicate requests - update existing", func(t *testing.T) {
		webpage := NewWebpageFromString(testBaseURL+"/", parent)
		req1 := createTestRequest(testBaseURL+"/page1", "GET", "original_body")
		req1.Response = &WebpageResponse{StatusCode: 500}
		req2 := createTestRequest(testBaseURL+"/page2", "POST", "body2")
		webpage.AddRequest(req1)
		webpage.AddRequest(req2)

		req1Updated := createTestRequest(testBaseURL+"/page1", "GET", "original_body")
		req1Updated.Response = &WebpageResponse{StatusCode: 200}
		req3 := createTestRequest(testBaseURL+"/page3", "GET", "body3")
		webpage.MergeRequests(req1Updated, req3)

		assert.Len(t, webpage.Requests, 3)
		foundUpdated := false
		for _, req := range webpage.Requests {
			if req.RawURL == testBaseURL+"/page1" && req.Response != nil && req.Response.StatusCode == 200 {
				foundUpdated = true
				break
			}
		}
		assert.True(t, foundUpdated, "Updated request should be present with new response")
	})

	t.Run("merge respects max requests limit", func(t *testing.T) {
		webpage := NewWebpageFromString(testBaseURL+"/", parent)
		for i := 0; i < DefaultMaxRequestsPerWebpage; i++ {
			req := createTestRequest(fmt.Sprintf("%s/page%d", testBaseURL, i), "GET", fmt.Sprintf("body%d", i))
			webpage.AddRequest(req)
		}

		extraReqs := make([]WebpageRequest, 10)
		for i := 0; i < 10; i++ {
			extraReqs[i] = createTestRequest(fmt.Sprintf("%s/extra%d", testBaseURL, i), "GET", fmt.Sprintf("extrabody%d", i))
		}
		webpage.MergeRequests(extraReqs...)
		assert.Equal(t, DefaultMaxRequestsPerWebpage, len(webpage.Requests))
	})
}

func TestWebpagePopulateResponse(t *testing.T) {
	webpage := createTestWebpage(testBaseURL + testPath)

	t.Run("populate response with mock server", func(t *testing.T) {
		server := createJSONServer()
		defer server.Close()

		request := WebpageRequest{RawURL: server.URL + "/test", Method: "GET", Headers: map[string][]string{"User-Agent": {testUserAgent}}}
		err := webpage.PopulateResponse(&request)

		assert.NoError(t, err)
		assertRequestResponse(t, request, http.StatusOK, "Test Response")
		assert.Contains(t, request.Response.Headers, "Content-Type")
		assert.Contains(t, request.Response.Headers, "X-Custom-Header")
	})

	t.Run("populate response with POST data", func(t *testing.T) {
		server := createEchoServer()
		defer server.Close()

		request := WebpageRequest{
			RawURL:  server.URL + "/api",
			Method:  "POST",
			Headers: map[string][]string{"Content-Type": {testContentType}},
			Body:    "testBody",
		}
		err := webpage.PopulateResponse(&request)

		assert.NoError(t, err)
		assertRequestResponse(t, request, http.StatusOK, `"received": "testBody"`)
	})

	errorTests := []struct {
		name string
		url  string
	}{
		{"populate response with invalid URL", "invalid-url"},
		{"populate response with connection error", "http://non-existent-domain-12345.com"},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			request := WebpageRequest{RawURL: tt.url, Method: "GET"}
			err := webpage.PopulateResponse(&request)
			assert.Error(t, err)
			assert.Nil(t, request.Response)
		})
	}
}

func TestWebpagePopulateResponses(t *testing.T) {
	t.Run("populate responses for requests without responses", func(t *testing.T) {
		server := createEchoServer()
		defer server.Close()

		req1 := WebpageRequest{RawURL: server.URL + "/page1", Method: "GET"}
		req2 := createTestRequest(server.URL+"/page2", "GET", "")
		req2.Response = &WebpageResponse{StatusCode: 200, Body: "Existing response"}
		req3 := WebpageRequest{RawURL: server.URL + "/page3", Method: "GET"}

		webpage := createTestWebpage(testBaseURL+"/", WithRequests(req1, req2, req3))
		webpage.PopulateResponses(false)

		assertRequestResponse(t, webpage.Requests[0], 200, "page1")
		assertRequestResponse(t, webpage.Requests[1], 200, "Existing response")
		assertRequestResponse(t, webpage.Requests[2], 200, "page3")
	})

	t.Run("populate responses with refresh flag", func(t *testing.T) {
		server := createEchoServer()
		defer server.Close()

		req := createTestRequest(server.URL+"/page", "GET", "")
		req.Response = &WebpageResponse{StatusCode: 200, Body: "Old response"}

		webpage := createTestWebpage(testBaseURL+"/", WithRequests(req))
		webpage.PopulateResponses(true)

		assertRequestResponse(t, webpage.Requests[0], 200, "Response for /page")
		assert.NotContains(t, webpage.Requests[0].Response.Body, "Old response")
	})

	t.Run("populate responses with no requests", func(t *testing.T) {
		webpage := createTestWebpage(testBaseURL + "/")
		webpage.PopulateResponses(false)
		assertWebpageRequestCount(t, webpage, 0)
		webpage.PopulateResponses(true)
		assertWebpageRequestCount(t, webpage, 0)
	})
}

func TestWebpageConstructorOptions(t *testing.T) {
	customStatusOption := func(w *Webpage) error { w.Status = "Custom"; return nil }
	customMetadataOption := func(w *Webpage) error {
		w.Metadata["custom"] = "value"
		w.Metadata["option"] = "applied"
		return nil
	}
	customStateOption := func(w *Webpage) error { w.State = Interesting; return nil }

	webpage := createTestWebpage(testBaseURL+"/test", customStatusOption, customMetadataOption, customStateOption)

	assert.Equal(t, "Custom", webpage.Status)
	assertWebpageMetadata(t, webpage, "custom", "value")
	assertWebpageMetadata(t, webpage, "option", "applied")
	assertWebpageState(t, webpage, Interesting)
}

func TestWebpageHooks(t *testing.T) {
	t.Run("construction hook - key generation", func(t *testing.T) {
		webpage := Webpage{URL: testBaseURL + "/very/long/path/with/parameters"}
		err := registry.CallHooks(&webpage)
		assert.NoError(t, err)
		assert.Equal(t, "#webpage#"+testBaseURL+"/very/long/path/with/parameters", webpage.Key)
	})

	t.Run("construction hook - key truncation", func(t *testing.T) {
		longPath := strings.Repeat("a", 2100)
		webpage := Webpage{URL: testBaseURL + "/" + longPath}
		err := registry.CallHooks(&webpage)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(webpage.Key), 2048)
	})

	t.Run("construction hook - basic analysis", func(t *testing.T) {
		req := createTestRequest(testBaseURL+"/test?param=value&id=123", "GET", "")
		webpage := createTestWebpage(testBaseURL+"/test", WithRequests(req))

		err := registry.CallHooks(&webpage)
		assert.NoError(t, err)
		assertWebpageState(t, webpage, Interesting)
		assertWebpageMetadata(t, webpage, PARAMETERS_IDENTIFIED, true)
	})
}

func TestWebpageGobEncoding(t *testing.T) {
	t.Run("gob encoding loses WebpageDetails", func(t *testing.T) {
		webpage := createTestWebpage(testBaseURL + "/test")
		expectedDetails := createTestWebpageDetails("0")
		webpage.Source = []string{"test"}
		webpage.WebpageDetails = expectedDetails

		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)
		err := encoder.Encode(webpage)
		assert.NoError(t, err)

		fmt.Println("Encoded Webpage:", base64.StdEncoding.EncodeToString(buf.Bytes()))

		var decodedWebpage Webpage
		decoder := gob.NewDecoder(&buf)
		err = decoder.Decode(&decodedWebpage)
		assert.NoError(t, err)

		assert.Empty(t, decodedWebpage.WebpageDetails.Requests, "WebpageDetails should be empty after gob encoding")
		assert.Equal(t, expectedDetails, webpage.WebpageDetails, "Original webpage should still have WebpageDetails")
		assert.NotEmpty(t, decodedWebpage.Source, "Source should not be empty")
	})
}

// Helper functions
func assertWebpage(t *testing.T, webpage Webpage, expectedURL, expectedState string) {
	t.Helper()
	assertWebpageURL(t, webpage, expectedURL)
	assertWebpageState(t, webpage, expectedState)
}

func assertWebpageState(t *testing.T, webpage Webpage, expectedState string) {
	t.Helper()
	assert.Equal(t, expectedState, webpage.State, "webpage state should match expected")
}

func assertWebpageMetadata(t *testing.T, webpage Webpage, key string, expectedValue any) {
	t.Helper()
	assert.Equal(t, expectedValue, webpage.Metadata[key], "webpage metadata[%s] should match expected", key)
}

func assertWebpageRequestCount(t *testing.T, webpage Webpage, expectedCount int) {
	t.Helper()
	assert.Len(t, webpage.Requests, expectedCount, "webpage should have expected number of requests")
}

func assertRequestResponse(t *testing.T, request WebpageRequest, expectedStatusCode int, expectedBodyContains string) {
	t.Helper()
	assert.NotNil(t, request.Response, "request should have a response")
	assert.Equal(t, expectedStatusCode, request.Response.StatusCode, "response status code should match")
	if expectedBodyContains != "" {
		assert.Contains(t, request.Response.Body, expectedBodyContains, "response body should contain expected text")
	}
}

func assertWebpageURL(t *testing.T, webpage Webpage, expectedURL string) {
	t.Helper()
	assert.Equal(t, expectedURL, webpage.URL, "webpage URL should match expected")
}

func createTestParent() GraphModel {
	parentAsset := NewAsset("gladiator", "systems")
	parentAttribute := NewAttribute("https", "443", &parentAsset)
	return &parentAttribute
}

func createTestWebpage(rawURL string, options ...WebpageOption) Webpage {
	parsedURL, _ := url.Parse(rawURL)
	return NewWebpage(*parsedURL, createTestParent(), options...)
}

func createTestWebpages() (Webpage, Webpage) {
	parent := createTestParent()
	webpage1 := NewWebpageFromString(testBaseURL+testPath, parent)
	webpage2 := NewWebpageFromString(testBaseURL+testPath, parent)
	return webpage1, webpage2
}

func createTestWebpagePair(url string) (Webpage, Webpage) {
	parent := createTestParent()
	webpage1 := NewWebpageFromString(url, parent)
	webpage2 := NewWebpageFromString(url, parent)
	return webpage1, webpage2
}

func createTestWebpageDetails(identifier string) WebpageDetails {
	return WebpageDetails{
		Requests: []WebpageRequest{createTestRequest(
			fmt.Sprintf("%s%s/?x=%s", testBaseURL, testPath, identifier),
			"GET",
			fmt.Sprintf("{\"index\": %s}", identifier),
		)},
	}
}

func createTestRequest(rawURL, method, body string) WebpageRequest {
	return WebpageRequest{
		RawURL:  rawURL,
		Method:  method,
		Headers: map[string][]string{"User-Agent": {testUserAgent}},
		Body:    body,
		Response: &WebpageResponse{
			StatusCode: 200,
			Headers:    map[string][]string{"Content-Type": {"text/html"}},
			Body:       fmt.Sprintf("{\"response\": %s}", body),
		},
	}
}

func createNumberedRequests(baseURL string, count int, method string) []WebpageRequest {
	requests := make([]WebpageRequest, count)
	for i := 0; i < count; i++ {
		requests[i] = WebpageRequest{
			RawURL:  fmt.Sprintf("%s/%d", baseURL, i),
			Method:  method,
			Headers: map[string][]string{"User-Agent": {testUserAgent}},
			Body:    fmt.Sprintf("body%d", i),
			Response: &WebpageResponse{
				StatusCode: 200,
				Headers:    map[string][]string{"Content-Type": {"text/html"}},
				Body:       fmt.Sprintf("response%d", i),
			},
		}
	}
	return requests
}

func setupMockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func createEchoServer() *httptest.Server {
	return setupMockServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", testContentType)
		w.WriteHeader(http.StatusOK)
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		if r.ContentLength != 0 {
			w.Write([]byte(fmt.Sprintf(`{"received": "%s"}`, string(body))))
		}
		w.Write([]byte("Response for " + r.URL.Path))
	}))
}

func createJSONServer() *httptest.Server {
	return setupMockServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", testContentType)
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Test Response"}`))
	}))
}
