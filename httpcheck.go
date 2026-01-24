package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const maxBodyBytes = 64 << 10

func checkURL(client *http.Client, raw string, retries int) (int, int, error) {
	normalized, err := normalizeURL(raw)
	if err != nil {
		return 0, 0, err
	}

	var lastErr error
	attempts := 0
	for i := 0; i <= retries; i++ {
		attempts++
		status, err := doRequest(client, normalized)
		if err == nil {
			return status, attempts, nil
		}
		lastErr = err
	}
	return 0, attempts, lastErr
}

func doRequest(client *http.Client, raw string) (int, error) {
	// Per-request timeout for a single attempt.
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Drain a small portion of the body to reuse keep-alive connections.
	_, _ = io.CopyN(io.Discard, resp.Body, maxBodyBytes)
	return resp.StatusCode, nil
}

func normalizeURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("empty URL")
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid URL: %s", raw)
	}
	return u.String(), nil
}
