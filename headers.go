package rodder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	requestMirrorURL          = "https://requestmirror.dev/api/v1"
	requestMirrorDataSelector = "body > pre"
)

// GetHeaders returns a list of selected headers the browser add to each requests
// by requesting a mirror request on https://requestmirror.dev. The list is filtered (and
// modified for some sec-fetch-* headers) to help you make Golang http requests that mimick
// what a real browser will do with javascript API requests. Don't forget to add your own
// "accept", "accept-encoding" and "content-type" if necessary. Example:
// map[
//
//	Accept-Language:[fr-FR] Priority:[u=0, i] Sec-Ch-Ua:["Not;A=Brand";v="24", "Chromium";v="128"]
//	Sec-Ch-Ua-Mobile:[?0] Sec-Ch-Ua-Platform:["Windows"] Sec-Fetch-Dest:[empty] Sec-Fetch-Mode:[cors]
//	Sec-Fetch-Site:[same-origin] Upgrade-Insecure-Requests:[1]
//	User-Agent:[Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36]
//
// ]
func (b *Browser) GetHeaders() (headers http.Header, err error) {
	// Get a new page
	page, err := b.NewPage()
	if err != nil {
		err = fmt.Errorf("failed to get a new page from remote browser: %w", err)
		return
	}
	defer func() {
		if errClose := page.Close(); errClose != nil && err == nil {
			err = errClose
		}
	}()
	// Navigate to the request mirror URL
	if err = page.Navigate(requestMirrorURL); err != nil {
		err = fmt.Errorf("failed to navigate to request mirror URL: %w", err)
		return
	}
	if err = page.WaitStable(time.Second); err != nil {
		err = fmt.Errorf("failed to wait for page to stabilize: %w", err)
		return
	}
	// Pinpoint content
	present, contentElement, err := page.Has(requestMirrorDataSelector)
	if err != nil {
		err = fmt.Errorf("failed to extract HTML content: %w", err)
		return
	}
	if !present {
		err = fmt.Errorf("failed to retreive data from page: selector %q has not been found", requestMirrorDataSelector)
		return
	}
	// Extract content
	content, err := contentElement.HTML()
	if err != nil {
		err = fmt.Errorf("failed to extract HTML data: %w", err)
		return
	}
	content = strings.TrimSuffix(strings.TrimPrefix(content, "<pre>"), "</pre>")
	// Parse content
	var data mirrorPayload
	if err = json.Unmarshal([]byte(content), &data); err != nil {
		err = fmt.Errorf("failed to unmarshall data as JSON: %w", err)
		return
	}
	// Extract headers
	headers = make(http.Header, len(data.Headers))
	for key, value := range data.Headers {
		switch key {
		case "accept-language", "priority", "sec-ch-ua", "sec-ch-ua-mobile",
			"sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent":
			headers.Set(key, value)
		case "sec-fetch-dest":
			headers.Set(key, "empty")
		case "sec-fetch-mode":
			headers.Set(key, "cors")
		case "sec-fetch-site":
			headers.Set(key, "same-origin")
		case "sec-fetch-user":
			// ignore
		}
	}
	return
}

type mirrorPayload struct {
	Timestamp time.Time         `json:"timestamp"`
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Path      string            `json:"path"`
	Query     map[string]string `json:"query"`
	Headers   map[string]string `json:"headers"`
	Cookies   map[string]any    `json:"cookies"`
	Body      struct {
		Type        string `json:"type"`
		ContentType string `json:"contentType"`
		Size        int    `json:"size"`
		Note        string `json:"note"`
	} `json:"body"`
	IP struct {
		V4       string `json:"v4"`
		V6       string `json:"v6"`
		Original string `json:"original"`
	} `json:"ip"`
	Protocol string `json:"protocol"`
}
