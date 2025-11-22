package client

import (
	"fmt"
	"net/http"
	"time"
)

// HTTPClient implements the IHTTPClient interface
type HTTPClient struct {
	client  *http.Client
	baseURL string
	token   string
}

// NewHTTPClient creates a new HTTP client with the specified timeout
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Do executes an HTTP request and returns the response
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.setHeaders(req)
	return c.client.Do(req)
}

// Get performs an HTTP GET request
func (c *HTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}
	return c.Do(req)
}

// Post performs an HTTP POST request
func (c *HTTPClient) Post(url, contentType string, body interface{}) (*http.Response, error) {
	// For now, we'll create a simple implementation
	// In a full implementation, we'd handle JSON marshaling and body creation
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return c.Do(req)
}

// SetTimeout sets the request timeout
func (c *HTTPClient) SetTimeout(timeout time.Duration) {
	c.client.Timeout = timeout
}

// SetAuthHeader sets the authentication header
func (c *HTTPClient) SetAuthHeader(token string) {
	c.token = token
}

// Close closes the HTTP client and cleans up resources
func (c *HTTPClient) Close() error {
	// HTTP client doesn't need explicit closing in most cases
	return nil
}

// setHeaders sets common headers for all requests
func (c *HTTPClient) setHeaders(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("User-Agent", "supervisorctl/1.0.0")
}