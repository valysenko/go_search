package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Headers map[string]string

type HttpClient struct {
	http    *http.Client
	baseUrl string
}

func NewHttpClient(timeoutSeconds int, baseUrl string) *HttpClient {
	return &HttpClient{
		http: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		baseUrl: baseUrl,
	}
}

func (c *HttpClient) Get(ctx context.Context, path string, headers Headers, out any) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseUrl+path,
		nil,
	)
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

// example: methods with receivers can't use generics
func notUsedGet[T any](ctx context.Context, c *HttpClient, path string, headers Headers) (*T, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseUrl+path,
		nil,
	)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result T
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
