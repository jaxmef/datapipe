package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jaxmef/datapipe/config"
)

type Handler interface {
	Name() string
	Handle(ctx context.Context, data map[string]string) ([]HandlerResult, error)
}

func newHandler(name string, cfg config.Handler) (Handler, error) {
	switch cfg.Type {
	case "", config.HandlerTypeHTTP:
		return newHTTPHandler(name, cfg.HTTPHandler), nil
	case config.HandlerTypeFilter:
		return newFilterHandler(name, cfg.FilterHandler), nil
	default:
		return nil, fmt.Errorf("unknown handler type: %s", cfg.Type)
	}
}

type httpHandler struct {
	busyMux sync.Mutex

	name string
	cfg  config.HTTPHandler

	httpClient *http.Client
}

func newHTTPHandler(name string, cfg config.HTTPHandler) *httpHandler {
	timeout := 15 * time.Second
	if cfg.Timeout != 0 {
		timeout = cfg.Timeout
	}

	if cfg.ExpectedResponseCode == 0 {
		cfg.ExpectedResponseCode = http.StatusOK
	}

	return &httpHandler{
		name: name,
		cfg:  cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (h *httpHandler) Name() string {
	return h.name
}

func (h *httpHandler) Handle(ctx context.Context, data map[string]string) ([]HandlerResult, error) {
	var lastErr error
	for attempt := 0; attempt <= h.cfg.Retries; attempt++ {
		results, err := h.executeRequest(ctx, data)
		if err == nil {
			return results, nil
		}

		lastErr = err

		if attempt < h.cfg.Retries {
			timer := time.NewTimer(h.cfg.RetryInterval)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("failed to execute HTTP request after %d attempts: %s", h.cfg.Retries, lastErr)
}

func (h *httpHandler) executeRequest(ctx context.Context, data map[string]string) ([]HandlerResult, error) {
	if !h.cfg.ParallelRun {
		h.busyMux.Lock()
		defer h.busyMux.Unlock()
	}

	req, err := h.createRequest(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %s", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != h.cfg.ExpectedResponseCode {
		return nil, fmt.Errorf(
			"unexpected response code: got %d, expected %d",
			resp.StatusCode, h.cfg.ExpectedResponseCode,
		)
	}

	respBody := handlerResponseBody{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %s", err)
	}

	return respBody.Results, nil
}

func (h *httpHandler) createRequest(ctx context.Context, data map[string]string) (*http.Request, error) {
	url, ok := replacePlaceholders(h.cfg.URL, data)
	if !ok {
		return nil, fmt.Errorf("failed to replace placeholders in URL: some data not found")
	}

	body, ok := replacePlaceholders(h.cfg.Body, data)
	if !ok {
		return nil, fmt.Errorf("failed to replace placeholders in body: some data not found")
	}

	req, err := http.NewRequestWithContext(ctx, h.cfg.Method, url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %s", err)
	}

	for key, value := range h.cfg.Headers {
		req.Header.Set(key, value)
	}

	for key, value := range h.cfg.QueryParams {
		req.URL.Query().Add(key, value)
	}

	return req, nil
}

// replacePlaceholders replaces placeholders in the format {{ key }} with the value from the data map.
func replacePlaceholders(s string, data map[string]string) (string, bool) {
	re := regexp.MustCompile(`\{\{\s*([^\s}]+)\s*\}\}`)

	result := re.ReplaceAllStringFunc(s, func(m string) string {
		key := strings.TrimSpace(m[2 : len(m)-2])
		value, exists := data[key]
		if !exists {
			return m
		}
		return value
	})

	if re.MatchString(result) {
		return "", false
	}

	return result, true
}
