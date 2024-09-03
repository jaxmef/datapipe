package engine

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jaxmef/datapipe/config"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPHandler(t *testing.T) {
	cfg := config.HTTPHandler{
		URL:                  "http://example.com",
		Method:               "POST",
		Timeout:              10 * time.Second,
		ExpectedResponseCode: 201,
	}

	h := newHTTPHandler("test-handler", cfg)

	assert.Equal(t, "test-handler", h.name)
	assert.Equal(t, cfg, h.cfg)
	assert.Equal(t, 10*time.Second, h.httpClient.Timeout)
}

func TestHandler_Name(t *testing.T) {
	h := newHTTPHandler("test-handler", config.HTTPHandler{})

	assert.Equal(t, "test-handler", h.Name())
}

func TestHandler_Handle(t *testing.T) {
	t.Run("Successful request", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/test-url", r.URL.Path)
			reqBody, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, "test-body", string(reqBody))

			w.WriteHeader(http.StatusOK)
			_, err = w.Write([]byte(`{"results":[{"output":"test-output"}]}`))
			assert.NoError(t, err)
		}))
		defer mockServer.Close()

		cfg := config.HTTPHandler{
			Method:               "POST",
			URL:                  mockServer.URL + "/test-url",
			Body:                 "test-body",
			Headers:              map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Timeout:              10 * time.Second,
			ExpectedResponseCode: http.StatusOK,
		}
		h := newHTTPHandler("test-handler", cfg)

		data := map[string]string{"body": "test-body"}

		result, err := h.Handle(context.Background(), data)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, `"test-output"`, string(result[0]["output"]))
	})

	t.Run("Failed to replace placeholders in URL", func(t *testing.T) {
		cfg := config.HTTPHandler{
			Method: "GET",
			URL:    "{{ missing-placeholder }}",
		}
		h := newHTTPHandler("test-handler", cfg)

		data := map[string]string{"body": "test-body"}

		result, err := h.Handle(context.Background(), data)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to replace placeholders in URL")
	})

	t.Run("Failed to replace placeholders in body", func(t *testing.T) {
		cfg := config.HTTPHandler{
			Method: "GET",
			URL:    "http://example.com",
			Body:   "{{ missing-placeholder }}",
		}
		h := newHTTPHandler("test-handler", cfg)

		data := map[string]string{"body": "test-body"}

		result, err := h.Handle(context.Background(), data)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to replace placeholders in body")
	})

	t.Run("Failed HTTP request", func(t *testing.T) {
		cfg := config.HTTPHandler{
			Method: "GET",
			URL:    "http://invalid-url",
		}
		h := newHTTPHandler("test-handler", cfg)

		data := map[string]string{"body": "test-body"}

		result, err := h.Handle(context.Background(), data)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to send HTTP request")
	})

	t.Run("Unexpected response code", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer mockServer.Close()

		cfg := config.HTTPHandler{
			Method:               "GET",
			URL:                  mockServer.URL + "/test-url",
			ExpectedResponseCode: http.StatusOK,
		}
		h := newHTTPHandler("test-handler", cfg)

		data := map[string]string{"body": "test-body"}

		result, err := h.Handle(context.Background(), data)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unexpected response code")
	})

	t.Run("Failed to decode response body", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json`))
		}))
		defer mockServer.Close()

		cfg := config.HTTPHandler{
			Method:               "GET",
			URL:                  mockServer.URL + "/test-url",
			ExpectedResponseCode: http.StatusOK,
		}
		h := newHTTPHandler("test-handler", cfg)

		data := map[string]string{"body": "test-body"}

		result, err := h.Handle(context.Background(), data)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to decode response body")
	})

	t.Run("Successful request with object result", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/test-url", r.URL.Path)
			reqBody, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, "test-body", string(reqBody))

			w.WriteHeader(http.StatusOK)
			_, err = w.Write([]byte(`{"results":[{"output": {"a":"b","c":"d"}}]}`))
			assert.NoError(t, err)
		}))
		defer mockServer.Close()

		cfg := config.HTTPHandler{
			Method:               "POST",
			URL:                  mockServer.URL + "/test-url",
			Body:                 "test-body",
			Headers:              map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Timeout:              10 * time.Second,
			ExpectedResponseCode: http.StatusOK,
		}
		h := newHTTPHandler("test-handler", cfg)

		data := map[string]string{"body": "test-body"}

		result, err := h.Handle(context.Background(), data)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, `{"a":"b","c":"d"}`, string(result[0]["output"]))
	})

	t.Run("Retry on unexpected response code", func(t *testing.T) {
		serverCalls := 0
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch serverCalls {
			case 0:
				serverCalls++
				w.WriteHeader(http.StatusInternalServerError)
				return
			default:
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"results":[{"output":"test-output"}]}`))
				assert.NoError(t, err)
			}
		}))
		defer mockServer.Close()

		cfg := config.HTTPHandler{
			Method:               "GET",
			URL:                  mockServer.URL + "/test-url",
			ExpectedResponseCode: http.StatusOK,
			Retries:              2,
			RetryInterval:        0,
		}
		h := newHTTPHandler("test-handler", cfg)

		result, err := h.Handle(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, `"test-output"`, string(result[0]["output"]))
	})

	t.Run("Fail after all retries", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}))
		defer mockServer.Close()

		cfg := config.HTTPHandler{
			Method:               "GET",
			URL:                  mockServer.URL + "/test-url",
			ExpectedResponseCode: http.StatusOK,
			Retries:              2,
			RetryInterval:        0,
		}
		h := newHTTPHandler("test-handler", cfg)

		result, err := h.Handle(context.Background(), nil)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "failed to execute HTTP request after 2 attempts"))
		assert.Nil(t, result)
	})
}

func TestReplacePlaceholders(t *testing.T) {
	t.Run("Successful replacement", func(t *testing.T) {
		s := "{'input':{{ data-source.description }}}"
		data := map[string]string{"data-source.description": "test1"}

		result, success := replacePlaceholders(s, data)

		assert.True(t, success)
		assert.Equal(t, "{'input':test1}", result)
	})

	t.Run("Multiple replacements", func(t *testing.T) {
		s := "{{ key1 }} and {{key2}}"
		data := map[string]string{"key1": "value1", "key2": "value2"}

		result, success := replacePlaceholders(s, data)

		assert.True(t, success)
		assert.Equal(t, "value1 and value2", result)
	})

	t.Run("Missing key in the map", func(t *testing.T) {
		s := "{'input':{{ data-source.description }}, 'output':{{ data-source.title }}}"
		data := map[string]string{"data-source.description": "test1"}

		result, success := replacePlaceholders(s, data)

		assert.False(t, success)
		assert.Equal(t, "", result)
	})

	t.Run("No placeholders in the string", func(t *testing.T) {
		s := "This is a test string with no placeholders."
		data := map[string]string{"key1": "value1"}

		result, success := replacePlaceholders(s, data)

		assert.True(t, success)
		assert.Equal(t, s, result)
	})

	t.Run("Empty string input", func(t *testing.T) {
		s := ""
		data := map[string]string{"key1": "value1"}

		result, success := replacePlaceholders(s, data)

		assert.True(t, success)
		assert.Equal(t, "", result)
	})

	t.Run("Placeholder with spaces", func(t *testing.T) {
		s := "{{   key1   }}"
		data := map[string]string{"key1": "value1"}

		result, success := replacePlaceholders(s, data)

		assert.True(t, success)
		assert.Equal(t, "value1", result)
	})

	t.Run("Placeholder with no corresponding key", func(t *testing.T) {
		s := "{{ key1 }}"
		data := map[string]string{}

		result, success := replacePlaceholders(s, data)

		assert.False(t, success)
		assert.Equal(t, "", result)
	})
}
