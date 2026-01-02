package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	FieldOne string
	FieldTwo int
}

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "not allowed"}`))
			return
		}

		if r.URL.Path != "/endpoint/1" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "path not found"}`))
			return
		}

		if r.Header.Get("User-Agent") != "Browser" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "missing header"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testStruct{FieldOne: "Apple", FieldTwo: 42})
	}))
	defer server.Close()

	t.Run("check successful response", func(t *testing.T) {
		client := NewHttpClient(5, server.URL)

		var result testStruct
		err := client.Get(context.Background(), "/endpoint/1", Headers{"User-Agent": "Browser"}, &result)
		assert.Nil(t, err)
		assert.Equal(t, "Apple", result.FieldOne)
		assert.Equal(t, 42, result.FieldTwo)
	})

	t.Run("missing header returns 500 with error", func(t *testing.T) {
		client := NewHttpClient(5, server.URL)

		var result testStruct
		err := client.Get(context.Background(), "/endpoint/1", nil, &result)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrServerError))
		assert.Contains(t, err.Error(), "500")
		assert.Contains(t, err.Error(), "server error")
		assert.Contains(t, err.Error(), "missing header")
	})

	t.Run("wrong path returns 404 path not found", func(t *testing.T) {
		client := NewHttpClient(5, server.URL)

		var result testStruct
		err := client.Get(context.Background(), "/wrong/path", Headers{"User-Agent": "Browser"}, &result)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrClientError))
		assert.Contains(t, err.Error(), "404")
		assert.Contains(t, err.Error(), "client error")
		assert.Contains(t, err.Error(), "path not found")
	})

	t.Run("connection refused error", func(t *testing.T) {
		client := NewHttpClient(1, "http://127.0.0.1:1")

		var result testStruct
		err := client.Get(context.Background(), "/endpoint/1", Headers{"User-Agent": "Browser"}, &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute request")
		assert.Contains(t, err.Error(), "connection refused")
		assert.False(t, errors.Is(err, ErrServerError))
		assert.False(t, errors.Is(err, ErrClientError))
	})

	t.Run("context cancelled before request", func(t *testing.T) {
		client := NewHttpClient(5, server.URL)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var result testStruct
		err := client.Get(ctx, "/endpoint/1", Headers{"User-Agent": "Browser"}, &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute request")
		assert.True(t, errors.Is(err, context.Canceled))
	})

	t.Run("http client timeout exceeded", func(t *testing.T) {
		slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer slowServer.Close()

		client := NewHttpClient(1, slowServer.URL)

		var result testStruct
		err := client.Get(context.Background(), "/endpoint/1", Headers{"User-Agent": "Browser"}, &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute request")
		assert.Contains(t, err.Error(), "Client.Timeout")
		assert.False(t, errors.Is(err, ErrServerError))
		assert.False(t, errors.Is(err, ErrClientError))
	})

	t.Run("context timeout exceeded", func(t *testing.T) {
		slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer slowServer.Close()

		client := NewHttpClient(5, slowServer.URL)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		var result testStruct
		err := client.Get(ctx, "/endpoint/1", Headers{"User-Agent": "Browser"}, &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}
