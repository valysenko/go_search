package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ErrorType int

func (t ErrorType) String() string {
	return [...]string{"Unknown", "Client", "Server", "Network", "WrongData"}[t]
}

const (
	ErrorTypeUnknown   ErrorType = iota
	ErrorTypeClient              // 4xx
	ErrorTypeServer              // 5xx
	ErrorTypeNetwork             // dns, timeouts
	ErrorTypeWrongData           // json decoding errors
)

type RequestError struct {
	Type            ErrorType
	Message         string
	StatusCode      int
	Url             string
	Err             error
	RawResponseBody []byte
}

func (e *RequestError) Error() string {
	msg := fmt.Sprintf("request failed: %s (type=%s, status=%d)",
		e.Url, e.Type.String(), e.StatusCode,
	)
	if e.Err != nil {
		msg += fmt.Sprintf(", cause: %v", e.Err)
	}

	if len(e.RawResponseBody) > 0 {
		snippet := string(e.RawResponseBody)
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}
		msg += fmt.Sprintf(", body: [%s]", snippet)
	}

	return msg
}

func (e *RequestError) Unwrap() error {
	return e.Err
}

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

func (c *HttpClient) Get(ctx context.Context, path string, headers Headers, out any) (retErr error) {
	rawPath, query, _ := strings.Cut(path, "?")
	baseUrl := strings.TrimSuffix(c.baseUrl, "/")
	cleanPath := strings.TrimPrefix(rawPath, "/")
	fullUrl := baseUrl + "/" + cleanPath
	if query != "" {
		fullUrl += "?" + query
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fullUrl,
		nil,
	)
	if err != nil {
		return &RequestError{Type: ErrorTypeUnknown, Message: "failed to create request", Url: fullUrl, Err: err}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		errType := ErrorTypeNetwork
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return &RequestError{
				Type:    errType,
				Message: "request timed out or cancelled",
				Url:     fullUrl,
				Err:     err,
			}
		}
		return &RequestError{
			Type:    errType,
			Message: "transport level error",
			Url:     fullUrl,
			Err:     err,
		}
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && retErr == nil {
			retErr = &RequestError{
				Type:       ErrorTypeNetwork,
				Message:    "failed to close response body",
				StatusCode: resp.StatusCode,
				Url:        fullUrl,
				Err:        closeErr,
			}
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.handleErrorStatus(resp, fullUrl)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return &RequestError{Type: ErrorTypeWrongData, Message: "json decode fail", Url: fullUrl, StatusCode: resp.StatusCode, Err: err}
	}

	return nil
}

func (c *HttpClient) handleErrorStatus(resp *http.Response, fullUrl string) error {
	errType := ErrorTypeUnknown
	message := "unknown error"
	switch {
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		errType = ErrorTypeClient
		message = "client error"
	case resp.StatusCode >= 500:
		errType = ErrorTypeServer
		message = "server error"
	}

	// prevent OOM in case body is huge
	limitReader := io.LimitReader(resp.Body, 4096)
	body, err := io.ReadAll(limitReader)
	if err != nil {
		return &RequestError{
			Type:       ErrorTypeWrongData, // Or ErrorTypeNetwork depending on preference
			Message:    "failed to read error response body",
			StatusCode: resp.StatusCode,
			Url:        fullUrl,
			Err:        err, // The underlying read error (e.g. connection reset)
		}
	}

	return &RequestError{
		Type:            errType,
		Message:         message,
		StatusCode:      resp.StatusCode,
		Url:             fullUrl,
		RawResponseBody: body,
	}
}
