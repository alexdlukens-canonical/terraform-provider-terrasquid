package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

type APIClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(baseURL, apiKey string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *APIClient) SetInsecureTLS(insecure bool) {
	if !insecure {
		return
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c.HTTPClient.Transport = transport
}

func (c *APIClient) ValidateCredentials(ctx context.Context) error {
	_, err := c.ListSourceACLs(ctx)
	return err
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Timeout()
	}
	return false
}

func isRetryableStatus(code int) bool {
	return (code >= 500 && code < 600) || code == http.StatusTooManyRequests
}

func (c *APIClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
	}

	var lastErr error
	backoff := time.Second
	const maxRetries = 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Accept", "application/json")

		if c.APIKey != "" {
			req.Header.Set("Authorization", "Api-Key "+c.APIKey)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			if !isRetryableError(err) || attempt == maxRetries {
				return nil, fmt.Errorf("request failed: %w", err)
			}
			if err := waitForRetry(ctx, backoff); err != nil {
				return nil, err
			}
			backoff *= 2
			continue
		}

		if !isRetryableStatus(resp.StatusCode) || attempt == maxRetries {
			return resp, nil
		}

		_ = resp.Body.Close()
		lastErr = &APIError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}
		if err := waitForRetry(ctx, backoff); err != nil {
			return nil, err
		}
		backoff *= 2
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *APIClient) doUnauthenticatedRequest(ctx context.Context, method, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	return c.HTTPClient.Do(req)
}

func parseResponse(resp *http.Response, target interface{}) error {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if target != nil {
			return json.NewDecoder(resp.Body).Decode(target)
		}
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: failed to read error body", resp.StatusCode),
		}
	}

	var apiErr APIError
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
		apiErr.StatusCode = resp.StatusCode
		return &apiErr
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    string(body),
	}
}
