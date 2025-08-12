package model

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

func (w *Webpage) basicAnalysis() {
	if w.State == Interesting && w.Metadata[PARAMETERS_IDENTIFIED] != nil {
		return
	}

	for _, req := range w.Requests {
		url, err := url.Parse(req.RawURL)
		if err != nil {
			continue
		}
		params := url.Query()
		hasParameters := len(params) > 0
		onlyJSVer := len(params) == 1 && len(params["ver"]) == 1

		if hasParameters && !onlyJSVer {
			w.State = Interesting
			w.Metadata[PARAMETERS_IDENTIFIED] = true
			break
		}
	}
}

func (w *Webpage) PopulateResponse(request *WebpageRequest) error {
	resp, err := w.doRequest(request.Method, request.RawURL, request.Headers, request.Body)
	if err != nil {
		return err
	}
	request.Response = &resp
	return nil
}

func (w *Webpage) MergeMetadata(other Webpage) {
	// Initialize metadata map if nil
	if w.Metadata == nil {
		w.Metadata = make(map[string]any)
	}

	// We append slices and arrays otherwise overwrite
	for key, value := range other.Metadata {
		if _, ok := w.Metadata[key]; !ok {
			w.Metadata[key] = value
		} else {
			w.Metadata[key] = mergeMapOfSlice(w.Metadata[key], value)
		}
	}
}

func mergeMapOfSlice(original any, other any) any {
	switch orig := original.(type) {
	case []string:
		if oth, ok := other.([]string); ok {
			return mergeSlices(orig, oth)
		}
	case []any:
		if oth, ok := other.([]any); ok {
			return mergeSlices(orig, oth)
		}
	default:
		return other
	}
	return original
}

func mergeSlices[T comparable](original, other []T) []T {
	for _, item := range other {
		if !slices.Contains(original, item) {
			original = append(original, item)
		}
	}
	return original
}

func (w *Webpage) MergeSource(other Webpage) {
	w.Source = mergeSlices(w.Source, other.Source)
}

// We merge requests preferring existing, updating in the duplicate case, then append new from other webpage
func (w *Webpage) MergeRequests(others ...WebpageRequest) {
	type reqKey struct {
		RawURL string
		Method string
		Body   string
	}
	max := DefaultMaxRequestsPerWebpage

	newReqMap := make(map[reqKey]WebpageRequest, len(others))
	for _, req := range others {
		k := reqKey{req.RawURL, req.Method, req.Body}
		newReqMap[k] = req
	}

	merged := make([]WebpageRequest, 0, max)
	used := make(map[reqKey]struct{}, max)
	for _, req := range w.Requests {
		k := reqKey{req.RawURL, req.Method, req.Body}
		if updated, ok := newReqMap[k]; ok {
			merged = append(merged, updated)
			used[k] = struct{}{}
		} else {
			merged = append(merged, req)
			used[k] = struct{}{}
		}
		if len(merged) == max {
			w.Requests = merged
			return
		}
	}

	for _, req := range others {
		k := reqKey{req.RawURL, req.Method, req.Body}
		if _, ok := used[k]; !ok {
			merged = append(merged, req)
			used[k] = struct{}{}
			if len(merged) == max {
				break
			}
		}
	}

	w.Requests = merged
}

func (w *Webpage) BasePath() string {
	return fmt.Sprintf("webpage/%s/%d/%s", w.Hostname(), w.Port(), RemoveReservedCharacters(w.URL))
}

func (w *Webpage) GetDetailsFile(details WebpageDetails) File {
	bytes, err := json.Marshal(details)
	if err != nil {
		slog.Warn("Failed to marshal webpage details", "error", err)
		return File{}
	}
	filename := fmt.Sprintf("%s/details.json", w.BasePath())
	file := NewFile(filename)
	file.Bytes = bytes
	return file
}

func (w *Webpage) GetDisplayResponseFile() File {
	if len(w.Requests) == 0 {
		slog.Warn("no requests to get display response file for", "webpage", w.Key)
		return File{}
	}
	file := w.GetResponseBodyAsFile(w.Requests[0])
	w.Metadata[DISPLAY_RESPONSE_FILE_PATH] = file.Name
	return file
}

func (w *Webpage) GetResponseBodyAsFile(request WebpageRequest) File {
	if request.Response == nil {
		w.PopulateResponse(&request)
	}
	body := request.Response.Body
	filename := fmt.Sprintf("%s/%s", w.BasePath(), RemoveReservedCharacters(request.RawURL))
	file := NewFile(filename)
	file.Bytes = []byte(body)
	return file
}

func (w *Webpage) doRequest(method, url string, headers map[string][]string, body string) (WebpageResponse, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return WebpageResponse{}, err
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return WebpageResponse{}, err
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return WebpageResponse{}, err
	}
	respHeaders := make(map[string][]string)
	for k, vs := range res.Header {
		respHeaders[k] = vs
	}
	return WebpageResponse{
		Body:       string(respBody),
		StatusCode: res.StatusCode,
		Headers:    respHeaders,
	}, nil
}

func (w *Webpage) parseURL() url.URL {
	parsed, err := url.Parse(w.URL)
	if err != nil {
		return url.URL{}
	}
	return *parsed
}

func (w *Webpage) Protocol() string {
	return w.parseURL().Scheme
}

func (w *Webpage) Host() string {
	return w.parseURL().Host
}

func (w *Webpage) Hostname() string {
	parsed := w.parseURL()
	return parsed.Hostname()
}

func (w *Webpage) UrlPath() string {
	parsed := w.parseURL()
	if parsed.Path == "" {
		return DEFAULT_URL_PATH
	}
	return parsed.Path
}

func (w *Webpage) Port() int {
	parsed := w.parseURL()
	port := parsed.Port()
	if port == "" {
		switch w.Protocol() {
		case "http":
			return 80
		case "https":
			return 443
		default:
			return ERR_PORT
		}
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return ERR_PORT
	}
	return portInt
}

func RemoveReservedCharacters(s string) string {
	invalidChars := "<>:\"/\\|?*"
	result := s
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, string(char), "_")
	}
	return result
}

func WithState(state string) WebpageOption {
	return func(w *Webpage) error {
		w.State = state
		return nil
	}
}

func WithRequests(requests ...WebpageRequest) WebpageOption {
	return func(w *Webpage) error {
		for _, req := range requests {
			w.AddRequest(req)
		}
		return nil
	}
}
