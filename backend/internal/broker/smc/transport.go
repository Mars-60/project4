package smc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Transport struct {
	client     *Client
	httpClient *http.Client
	maxRetries int
	retryWait  time.Duration
}

func NewTransport(client *Client, httpClient *http.Client, maxRetries int, retryWait time.Duration) *Transport {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	return &Transport{
		client:     client,
		httpClient: httpClient,
		maxRetries: maxRetries,
		retryWait:  retryWait,
	}
}

func (t *Transport) Get(ctx context.Context, endpoint string, response any) error {
	return t.Do(ctx, http.MethodGet, endpoint, nil, response)
}

func (t *Transport) Post(ctx context.Context, endpoint string, request any, response any) error {
	return t.Do(ctx, http.MethodPost, endpoint, request, response)
}

func (t *Transport) Put(ctx context.Context, endpoint string, request any, response any) error {
	return t.Do(ctx, http.MethodPut, endpoint, request, response)
}

func (t *Transport) Delete(ctx context.Context, endpoint string, request any, response any) error {
	return t.Do(ctx, http.MethodDelete, endpoint, request, response)
}

func (t *Transport) Do(ctx context.Context, method string, endpoint string, request any, response any) error {
	var body []byte
	var err error

	if request != nil {
		body, err = json.Marshal(request)
		if err != nil {
			return fmt.Errorf("marshal smc request: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleep(ctx, t.retryWait*time.Duration(attempt)); err != nil {
				return err
			}
		}

		lastErr = t.doOnce(ctx, method, endpoint, body, response)
		if lastErr == nil {
			return nil
		}

		var apiErr *APIError
		if ok := errors.As(lastErr, &apiErr); ok && !isRetryableStatus(apiErr.StatusCode) {
			return lastErr
		}
	}

	return lastErr
}

func (t *Transport) doOnce(ctx context.Context, method string, endpoint string, body []byte, response any) error {
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		t.client.baseEndpoint(endpoint),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create smc request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if t.client.apiKey != "" {
		req.Header.Set("X-API-Key", t.client.apiKey)
	}
	if token := t.client.AccessToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute smc request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read smc response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(responseBody),
		}
	}

	if response == nil || len(responseBody) == 0 {
		return nil
	}

	if err := json.Unmarshal(responseBody, response); err != nil {
		return fmt.Errorf("decode smc response: %w", err)
	}

	return nil
}

func sleep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
