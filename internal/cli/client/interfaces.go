package client

import (
	"net/http"
	"time"
)

// IHTTPClient defines the interface for HTTP client operations
type IHTTPClient interface {
	// Do executes an HTTP request and returns the response
	Do(req *http.Request) (*http.Response, error)

	// Get performs an HTTP GET request
	Get(url string) (*http.Response, error)

	// Post performs an HTTP POST request
	Post(url, contentType string, body interface{}) (*http.Response, error)

	// SetTimeout sets the request timeout
	SetTimeout(timeout time.Duration)

	// SetAuthHeader sets the authentication header
	SetAuthHeader(token string)

	// Close closes the HTTP client and cleans up resources
	Close() error
}