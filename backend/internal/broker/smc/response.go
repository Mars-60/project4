package smc

import (
	"fmt"
	"net/http"
	"time"
)

type Envelope[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func (e Envelope[T]) OK() bool {
	return e.Status == "success" || e.Status == "ok"
}

type APIError struct {
	StatusCode int
	Status     string
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("smc api error: status_code=%d status=%s message=%s", e.StatusCode, e.Status, e.Message)
	}

	return fmt.Sprintf("smc api error: status_code=%d body=%s", e.StatusCode, e.Body)
}

func isRetryableStatus(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"02-01-2006 15:04:05",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed
		}
	}

	return time.Time{}
}
