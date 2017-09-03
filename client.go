package main

import (
	"errors"
	"fmt"
	"net/http"

	"searchinform/provider"
)

// HTTPClient - custom http client
type HTTPClient struct {
	client http.Client
}

// NewHTTPClient - constuctor for HTTPClient struct
func NewHTTPClient(client *http.Client, providers *provider.Iterator) *HTTPClient {
	return &HTTPClient{
		client: *client,
	}
}

// Resolve returns country of this addr
func (c *HTTPClient) Resolve(provider *provider.Provider, addr string) (country string, err error) {
	url := fmt.Sprintf(provider.URLPattern, addr)
	req, err := http.NewRequest(provider.Method, url, nil)
	if err != nil {
		return "", err
	}
	for key, value := range provider.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || 300 <= resp.StatusCode {
		return "", errors.New("Invalid status code:" + resp.Status)
	}

	return provider.ParseBody(resp.Body)
}
