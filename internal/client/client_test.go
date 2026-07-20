package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func newTestClient(t *testing.T, handler http.HandlerFunc, opts ...Option) (WaifuIMClient, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	opts = append(opts, WithBaseURL(srv.URL))
	opts = append(opts, WithRetryDelay(time.Millisecond))
	return New(opts...), srv
}

func TestCorrectPathCreated(t *testing.T) {
	tests := []struct {
		name   string
		status int
		call   func(c WaifuIMClient, ctx context.Context, path string, body io.Reader, query url.Values) (any, error)
		query  url.Values
		path   string
	}{
		{
			name:   "GET request path build",
			status: http.StatusOK,
			call: func(c WaifuIMClient, ctx context.Context, path string, body io.Reader, query url.Values) (any, error) {
				return c.GET(ctx, path, query)
			},
			query: url.Values{
				"limit": []string{"1"},
				"page":  []string{"1"},
			},
			path: "/some/path/get",
		},
		{
			name:   "POST request path build",
			status: http.StatusCreated,
			call: func(c WaifuIMClient, ctx context.Context, path string, body io.Reader, query url.Values) (any, error) {
				return c.POST(ctx, path, body)
			},
			query: url.Values{},
			path:  "/some/path/post",
		},
		{
			name:   "PATCH request path build",
			status: http.StatusOK,
			call: func(c WaifuIMClient, ctx context.Context, path string, body io.Reader, query url.Values) (any, error) {
				return c.PATCH(ctx, path, body)
			},
			query: url.Values{},
			path:  "/some/path/patch",
		},
		{
			name:   "DELETE request path build",
			status: http.StatusNoContent,
			call: func(c WaifuIMClient, ctx context.Context, path string, body io.Reader, query url.Values) (any, error) {
				return c.DELETE(ctx, path)
			},
			query: url.Values{},
			path:  "/some/path/delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotUrl string
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUrl = r.URL.String()
				writeJSON(w, tt.status, map[string]string{})
			})
			client, _ := newTestClient(t, handler, WithMaxRetries(1))
			_, err := tt.call(client, context.Background(), tt.path, bytes.NewBuffer(nil), tt.query)
			if err != nil {
				t.Fatal(err)
			}

			expectedPath := tt.path
			if tt.query.Get("limit") != "" {
				expectedPath += "?" + tt.query.Encode()
			}
			if !strings.Contains(gotUrl, expectedPath) {
				t.Fatalf("invalid url received: %s", gotUrl)
			}
		})
	}
}

func TestBodyEncoded(t *testing.T) {
	tests := []struct {
		name   string
		status int
		call   func(c WaifuIMClient, ctx context.Context, path string, body io.Reader) ([]byte, error)
	}{
		{
			name:   "POST request encodes body",
			status: http.StatusCreated,
			call: func(c WaifuIMClient, ctx context.Context, path string, body io.Reader) ([]byte, error) {
				return c.POST(ctx, path, body)
			},
		},
		{
			name:   "PATCH request encodes body",
			status: http.StatusOK,
			call: func(c WaifuIMClient, ctx context.Context, path string, body io.Reader) ([]byte, error) {
				return c.PATCH(ctx, path, body)
			},
		},
	}

	payload := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	expectedBody := string(jsonBytes)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody string

			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Body != nil {
					defer r.Body.Close()
					bodyBytes, err := io.ReadAll(r.Body)
					if err != nil {
						t.Error(err)
						return
					}
					gotBody = string(bodyBytes)
				}
				writeJSON(w, tt.status, map[string]string{})
			}

			client, _ := newTestClient(t, http.HandlerFunc(handler), WithMaxRetries(1))

			bodyReader := bytes.NewReader(jsonBytes)

			_, err := tt.call(client, context.Background(), "/", bodyReader)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if gotBody != expectedBody {
				t.Errorf("got body %q, want %q", gotBody, expectedBody)
			}
		})
	}
}

func TestBadStatusCodeRetry(t *testing.T) {
	tests := []struct {
		name             string
		code             int
		expectRetryCount int
		expectError      bool
	}{
		{name: "200", code: http.StatusOK, expectRetryCount: 1, expectError: false},
		{name: "400", code: http.StatusBadRequest, expectRetryCount: 1, expectError: true},
		{name: "401", code: http.StatusUnauthorized, expectRetryCount: 1, expectError: true},
		{name: "500", code: http.StatusInternalServerError, expectRetryCount: 2, expectError: false},
		{name: "502", code: http.StatusBadGateway, expectRetryCount: 2, expectError: false},
		{name: "429", code: http.StatusTooManyRequests, expectRetryCount: 2, expectError: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attempts := 0
			handler := func(w http.ResponseWriter, r *http.Request) {
				attempts++
				if attempts < test.expectRetryCount {
					w.WriteHeader(test.code)
					return
				}
				if test.expectError {
					writeJSON(w, test.code, map[string]string{})
				} else {
					writeJSON(w, http.StatusOK, map[string]string{})
				}
			}
			c, _ := newTestClient(t, handler, WithMaxRetries(2))
			_, err := c.GET(context.Background(), "/", nil)

			if test.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !test.expectError && err != nil {
				t.Fatal(err)
			}
			if test.expectRetryCount != attempts {
				t.Fatalf("wrong attempt count: got %d, want %d", attempts, test.expectRetryCount)
			}
		})
	}
}

func TestBadResponse_Error(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		expectErr     bool
		expectedError error
		contentType   string
	}{
		{
			name:          "200",
			status:        http.StatusOK,
			expectErr:     false,
			expectedError: nil,
			contentType:   "application/json",
		},
		{
			name:      "400",
			status:    http.StatusBadRequest,
			expectErr: true,
			expectedError: &BadResponse{
				StatusCode:   http.StatusBadRequest,
				ContentType:  "application/problem+json",
				ResponseBody: `{"detail":"some detail"}`,
			},
			contentType: "application/problem+json",
		},
		{
			name:      "401",
			status:    http.StatusUnauthorized,
			expectErr: true,
			expectedError: &BadResponse{
				StatusCode:   http.StatusUnauthorized,
				ContentType:  "application/problem+json",
				ResponseBody: `{"detail":"some detail"}`,
			},
			contentType: "application/problem+json",
		},
		{
			name:      "429",
			status:    http.StatusTooManyRequests,
			expectErr: true,
			expectedError: &BadResponse{
				StatusCode:   http.StatusTooManyRequests,
				ContentType:  "application/problem+json",
				ResponseBody: `{"detail":"some detail"}`,
			},
			contentType: "application/problem+json",
		},
		{
			name:      "500",
			status:    http.StatusInternalServerError,
			expectErr: true,
			expectedError: &BadResponse{
				StatusCode:   http.StatusInternalServerError,
				ContentType:  "application/problem+json",
				ResponseBody: `{"detail":"some detail"}`,
			},
			contentType: "application/problem+json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.status)
				json.NewEncoder(w).Encode(map[string]string{"detail": "some detail"})
			}
			c, _ := newTestClient(t, handler, WithMaxRetries(1))
			_, err := c.GET(context.Background(), "/", nil)
			if tt.expectErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatal(err)
			}
			if tt.expectErr && !errors.Is(err, tt.expectedError) {
				t.Errorf("got error %v, want %v", err, tt.expectedError)
			}
		})
	}
}

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		cancel()
	}
	c, _ := newTestClient(t, handler, WithMaxRetries(2))
	_, err := c.GET(ctx, "/", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatal("expected context.Canceled error")
	}
}

func TestAuthHeader(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
	}{
		{
			name:       "X-Api-Key present",
			authHeader: "1234",
		},
		{
			name:       "X-Api-Key missing",
			authHeader: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotHeader string
			handler := func(w http.ResponseWriter, r *http.Request) {
				gotHeader = r.Header.Get("X-Api-Key")
				writeJSON(w, http.StatusOK, map[string]string{})
			}
			opts := make([]Option, 0)
			opts = append(opts, WithMaxRetries(1))
			if tt.authHeader != "" {
				opts = append(opts, WithAPIKey(tt.authHeader))
			}

			c, _ := newTestClient(t, handler, opts...)
			_, err := c.GET(context.Background(), "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.authHeader != gotHeader {
				t.Errorf("got header %v, want %v", gotHeader, tt.authHeader)
			}
		})
	}
}

func TestRequestHeader(t *testing.T) {
	tests := []struct {
		name            string
		status          int
		call            func(c WaifuIMClient, ctx context.Context, path string, body io.Reader) (any, error)
		expectedHeaders map[string]string
		body            io.Reader
	}{
		{
			name:   "POST",
			status: http.StatusCreated,
			call: func(c WaifuIMClient, ctx context.Context, path string, body io.Reader) (any, error) {
				return c.POST(ctx, path, body)
			},
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
			body: bytes.NewBuffer(nil),
		},
		{
			name:   "PATCH",
			status: http.StatusOK,
			call: func(c WaifuIMClient, ctx context.Context, path string, body io.Reader) (any, error) {
				return c.PATCH(ctx, path, body)
			},
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
			body: bytes.NewBuffer(nil),
		},
		{
			name:   "DELETE",
			status: http.StatusNoContent,
			call: func(c WaifuIMClient, ctx context.Context, path string, body io.Reader) (any, error) {
				return c.DELETE(ctx, path)
			},
			expectedHeaders: map[string]string{
				"Accept": "application/json",
			},
			body: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				for key, wantVal := range tt.expectedHeaders {
					if got := r.Header.Get(key); got != wantVal {
						t.Errorf("header %s: got %q, want %q", key, got, wantVal)
					}
				}
				writeJSON(w, tt.status, map[string]string{})
			}

			c, _ := newTestClient(t, handler, WithMaxRetries(1))
			_, err := tt.call(c, context.Background(), "/", tt.body)
			if err != nil {
				t.Fatal(err)
			}

		})
	}
}

func TestNonJSONResponse(t *testing.T) {
	contentType := "text/html"
	plain := "NON JSON response"
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("NON JSON response"))
	}
	c, _ := newTestClient(t, handler, WithMaxRetries(1))
	_, err := c.GET(context.Background(), "/", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedError := &BadResponse{
		StatusCode:   http.StatusOK,
		ContentType:  contentType,
		ResponseBody: plain,
	}

	if !errors.Is(err, expectedError) {
		t.Errorf("got error %v, want %v", err, expectedError)
	}
}
