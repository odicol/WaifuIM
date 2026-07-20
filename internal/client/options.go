package client

import "time"

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithAPIKey sets the API keu used for authenticated requests.
func WithAPIKey(key string) Option {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithRetryDelay encodes the amount of time to wait between retries.
func WithRetryDelay(retryDelay time.Duration) Option {
	return func(c *Client) {
		c.retryDelay = retryDelay
	}
}
