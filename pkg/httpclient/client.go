package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrClientError      = errors.New("client error")
	ErrServerError      = errors.New("server error")
	ErrUnexpectedStatus = errors.New("unexpected status code")
)

type Headers map[string]string

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HttpClient struct {
	http    HTTPClient
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
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.handleErrorStatus(resp, path)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *HttpClient) handleErrorStatus(resp *http.Response, path string) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response body for %s: %w", path, err)
	}

	var errResp struct {
		Message string `json:"message"`
	}
	var errMsg string
	if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
		errMsg = errResp.Message
	}

	switch {
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		if errMsg != "" {
			return fmt.Errorf("client error %d for %s (%s): %w", resp.StatusCode, path, errMsg, ErrClientError)
		}
		return fmt.Errorf("client error %d for %s: %w", resp.StatusCode, path, ErrClientError)
	case resp.StatusCode >= 500:
		if errMsg != "" {
			return fmt.Errorf("server error %d for %s (%s): %w", resp.StatusCode, path, errMsg, ErrServerError)
		}
		return fmt.Errorf("server error %d for %s: %w", resp.StatusCode, path, ErrServerError)
	default:
		return fmt.Errorf("unexpected status code %d for %s: %w", resp.StatusCode, path, ErrUnexpectedStatus)
	}
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
