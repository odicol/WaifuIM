package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WaifuIMClient is the interface for interacting with the waifu.im API.
type WaifuIMClient interface {
	GET(ctx context.Context, path string, query url.Values) ([]byte, error)
	POST(ctx context.Context, path string, body io.Reader) ([]byte, error)
	PATCH(ctx context.Context, path string, body io.Reader) ([]byte, error)
	DELETE(ctx context.Context, path string) (bool, error)
}

// Client is an HTTP client for the waifu.im API.
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
	maxRetries int
	retryDelay time.Duration
}

// New returns a Client with the provided options applied
func New(opts ...Option) *Client {
	c := &Client{
		apiKey:     "",
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.waifu.im",
		maxRetries: 3,
		retryDelay: time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) do(ctx context.Context, req *http.Request, expectedResponse int) (*http.Response, error) {
	var code int
	var responseBody, contentType string
	for i := 1; i < c.maxRetries+1; i++ {
		if req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("cloning request body for retry %d: %w", i, err)
			}
			req.Body = body
		}

		if c.apiKey != "" {
			req.Header.Set("X-Api-Key", c.apiKey)
		}
		res, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		code = res.StatusCode
		contentType = res.Header.Get("Content-Type")

		body, err := io.ReadAll(res.Body)
		responseBody = strings.TrimRight(string(body), "\n")
		if err != nil {
			return nil, fmt.Errorf("reading error response body: %w", err)
		}

		if code == expectedResponse {
			if !strings.Contains(contentType, "application/json") && expectedResponse != http.StatusNoContent {
				res.Body.Close()
				return nil, &BadResponse{
					StatusCode:   code,
					ContentType:  contentType,
					ResponseBody: responseBody,
				}
			}

			return res, nil
		}

		if code == http.StatusTooManyRequests {
			// 429 errors are retried with provided retry-after header
			res.Body.Close()
			retryAfter := res.Header.Get("Retry-After")
			var toSleep time.Duration
			if retryAfter == "" {
				toSleep = time.Duration(i*2) * c.retryDelay
			} else {
				toSleep, err = time.ParseDuration(retryAfter)
				if err != nil {
					return nil, fmt.Errorf("parsing Retry-After header %q: %w", retryAfter, err)
				}
			}

			select {
			case <-time.After(toSleep):
				// sleep finished, retry
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled waiting to retry: %w", ctx.Err())
			}
			continue
		} else if code >= http.StatusInternalServerError {
			// 500+ errors are retried
			res.Body.Close()
			select {
			case <-time.After(time.Duration(i*2) * c.retryDelay): // Sleeps 2 4 6 seconds
				// sleep finished, retry
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled waiting to retry: %w", ctx.Err())
			}
			continue
		}

		return nil, &BadResponse{
			StatusCode:   code,
			ContentType:  contentType,
			ResponseBody: responseBody,
		}
	}
	return nil, &BadResponse{
		StatusCode:   code,
		ContentType:  contentType,
		ResponseBody: responseBody,
	}
}

// GET sends a GET request to path with the given query parameters and returns the response body.
func (c *Client) GET(ctx context.Context, path string, query url.Values) ([]byte, error) {
	finalUrl := strings.TrimSuffix(c.baseURL, "/") + "/" + strings.Trim(path, "/") + "?" + query.Encode()
	req, err := http.NewRequest(http.MethodGet, finalUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", finalUrl, err)
	}
	res, err := c.do(ctx, req, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", finalUrl, err)
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading GET response body: %w", err)
	}
	return respBody, nil
}

// POST sends a POST request to path with the given payload and returns the response body.
func (c *Client) POST(ctx context.Context, path string, body io.Reader) ([]byte, error) {
	finalUrl := strings.TrimSuffix(c.baseURL, "/") + "/" + strings.Trim(path, "/")
	req, err := http.NewRequest(http.MethodPost, finalUrl, body)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", finalUrl, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	res, err := c.do(ctx, req, http.StatusCreated)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", finalUrl, err)
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading POST response body: %w", err)
	}
	return respBody, nil
}

// PATCH sends a PATCH request to path with the given payload and returns the response body.
func (c *Client) PATCH(ctx context.Context, path string, body io.Reader) ([]byte, error) {
	finalUrl := strings.TrimSuffix(c.baseURL, "/") + "/" + strings.Trim(path, "/")
	req, err := http.NewRequest(http.MethodPatch, finalUrl, body)
	if err != nil {
		return nil, fmt.Errorf("PATCH %s: %w", finalUrl, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	res, err := c.do(ctx, req, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("PATCH %s: %w", finalUrl, err)
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading PATCH response body: %w", err)
	}
	return respBody, nil
}

// DELETE sends a DELETE requests to path and returns the boolean indicator of success.
func (c *Client) DELETE(ctx context.Context, path string) (bool, error) {
	finalUrl := strings.TrimSuffix(c.baseURL, "/") + "/" + strings.Trim(path, "/")
	req, err := http.NewRequest(http.MethodDelete, finalUrl, nil)
	if err != nil {
		return false, fmt.Errorf("DELETE %s: %w", finalUrl, err)
	}
	req.Header.Set("Accept", "application/json")
	res, err := c.do(ctx, req, http.StatusNoContent)
	if err != nil {
		return false, fmt.Errorf("DELETE %s: %w", finalUrl, err)
	}
	defer res.Body.Close()
	return true, nil
}
