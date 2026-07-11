package smc

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 3
	defaultRetryWait  = 300 * time.Millisecond
)

type Client struct {
	baseURL      string
	webSocketURL string
	apiKey       string
	apiSecret    string

	mu          sync.RWMutex
	accessToken string
	feedToken   string

	transport *Transport
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.transport.httpClient = httpClient
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.transport.httpClient.Timeout = timeout
		}
	}
}

func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) {
		if maxRetries >= 0 {
			c.transport.maxRetries = maxRetries
		}
	}
}

func WithWebSocketURL(webSocketURL string) Option {
	return func(c *Client) {
		c.webSocketURL = webSocketURL
	}
}

func NewClient(baseURL string, apiKey string, apiSecret string, opts ...Option) *Client {
	client := &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}

	client.transport = NewTransport(client, &http.Client{Timeout: defaultTimeout}, defaultMaxRetries, defaultRetryWait)

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) IsAuthenticated() bool {
	return c.AccessToken() != ""
}

func (c *Client) AccessToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.accessToken
}

func (c *Client) FeedToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.feedToken
}

func (c *Client) SetTokens(accessToken string, feedToken string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.accessToken = accessToken
	c.feedToken = feedToken
}

func (c *Client) baseEndpoint(endpoint string) string {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}

	return c.baseURL + endpoint
}
