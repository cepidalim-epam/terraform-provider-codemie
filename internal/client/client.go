// Package client implements a hand-rolled HTTP client for the CodeMie REST
// API, including OAuth2 client_credentials authentication with automatic
// token refresh.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Config holds the settings required to build a Client.
type Config struct {
	// Host is the base URL of the CodeMie API, e.g.
	// "https://codemie.lab.epam.com/code-assistant-api".
	Host string
	// TokenURL is the Keycloak/OAuth2 token endpoint used to acquire
	// access tokens via the client_credentials grant.
	TokenURL string
	// ClientID is the OAuth2 client id.
	ClientID string
	// ClientSecret is the OAuth2 client secret.
	ClientSecret string
	// HTTPClient, if set, is used as the base transport (mainly for
	// tests). If nil, http.DefaultClient is used.
	HTTPClient *http.Client
}

// Client is a thin wrapper around net/http that knows how to talk to the
// CodeMie REST API and automatically manages OAuth2 bearer tokens.
type Client struct {
	host       string
	httpClient *http.Client
}

// New builds a new CodeMie API Client from the given Config.
func New(cfg Config) (*Client, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("client: host must not be empty")
	}
	if cfg.TokenURL == "" {
		return nil, fmt.Errorf("client: token_url must not be empty")
	}
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("client: client_id must not be empty")
	}
	if cfg.ClientSecret == "" {
		return nil, fmt.Errorf("client: client_secret must not be empty")
	}

	oauthCfg := &clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenURL,
	}

	base := cfg.HTTPClient
	if base == nil {
		base = &http.Client{Timeout: 60 * time.Second}
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, base)
	tokenClient := oauthCfg.Client(ctx)
	tokenClient.Timeout = base.Timeout

	return &Client{
		host:       strings.TrimRight(cfg.Host, "/"),
		httpClient: tokenClient,
	}, nil
}

// APIError represents a non-2xx response from the CodeMie API.
type APIError struct {
	StatusCode int
	Body       string
	Validation *HTTPValidationError
}

func (e *APIError) Error() string {
	return fmt.Sprintf("codemie api error: status=%d body=%s", e.StatusCode, e.Body)
}

// IsNotFound reports whether err represents an HTTP 404 response.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

type ValidationError struct {
	Location []any  `json:"loc"`
	Message  string `json:"msg"`
	Type     string `json:"type"`
}

type HTTPValidationError struct {
	Detail []ValidationError `json:"detail"`
}

// CategoryList represents a list of category identifiers. When unmarshaling
// from JSON, it accepts either an array of strings (e.g. ["cat1", "cat2"])
// or an array of category objects (e.g. [{"id": "cat1", "name": "Category 1"}]),
// extracting the ID (or name if ID is empty).
type CategoryList []string

// UnmarshalJSON implements json.Unmarshaler.
func (c *CategoryList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*c = []string{}
		return nil
	}

	// Try array of strings first
	var strList []string
	if err := json.Unmarshal(data, &strList); err == nil {
		*c = strList
		return nil
	}

	// Try array of category objects
	var objList []struct {
		ID   *string `json:"id"`
		Name *string `json:"name"`
	}
	if err := json.Unmarshal(data, &objList); err != nil {
		return err
	}

	out := make([]string, 0, len(objList))
	for _, item := range objList {
		if item.ID != nil && *item.ID != "" {
			out = append(out, *item.ID)
		} else if item.Name != nil && *item.Name != "" {
			out = append(out, *item.Name)
		}
	}
	*c = out
	return nil
}

// doJSON issues an HTTP request with an optional JSON body and decodes a
// JSON response into out (if out is non-nil). It returns *APIError for any
// non-2xx response.
func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	url := c.host + path

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("client: marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("client: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("client: request %s %s failed: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("client: read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
		if resp.StatusCode == http.StatusUnprocessableEntity {
			var validation HTTPValidationError
			if json.Unmarshal(respBody, &validation) == nil {
				apiErr.Validation = &validation
			}
		}
		return apiErr
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("client: decode response body: %w (body=%s)", err, string(respBody))
		}
	}

	return nil
}
