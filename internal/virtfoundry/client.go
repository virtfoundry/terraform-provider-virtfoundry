package virtfoundry

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

// Client is a thin HTTP client for the VirtFoundry REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	tenantID   string
}

// NewClient builds a client for the given API endpoint.
func NewClient(endpoint string, insecure bool) (*Client, error) {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // dev-only opt-in
	}

	return &Client{
		baseURL: endpoint,
		httpClient: &http.Client{
			Timeout:   defaultTimeout,
			Transport: transport,
		},
	}, nil
}

// SetAPIKey configures Bearer auth with a VirtFoundry API key (vfd_live_...).
func (c *Client) SetAPIKey(key string) {
	c.token = strings.TrimSpace(key)
}

// SetTenantID sets the default X-Tenant-ID header for root-scoped calls.
func (c *Client) SetTenantID(tenantID string) {
	c.tenantID = strings.TrimSpace(tenantID)
}

// TenantID returns the configured default tenant ID.
func (c *Client) TenantID() string {
	return c.tenantID
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// Login exchanges username/password for a JWT.
func (c *Client) Login(ctx context.Context, username, password string) error {
	body, err := json.Marshal(loginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/auth/login", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}

	var out loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("decode login response: %w", err)
	}
	if out.Token == "" {
		return fmt.Errorf("login response missing token")
	}
	c.token = out.Token
	return nil
}

// Health checks the unauthenticated /health endpoint.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}
	return nil
}

// PingAuth verifies credentials against GET /api/v1/auth/me.
func (c *Client) PingAuth(ctx context.Context) error {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/auth/me", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}
	return nil
}

// Do sends an authenticated API request.
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, method, path, body)
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	if c.token == "" {
		return nil, fmt.Errorf("not authenticated")
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	if c.tenantID != "" {
		req.Header.Set("X-Tenant-ID", c.tenantID)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.httpClient.Do(req)
}

func apiError(resp *http.Response) error {
	msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	text := strings.TrimSpace(string(msg))
	if text == "" {
		return fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}
	return fmt.Errorf("API error: HTTP %d: %s", resp.StatusCode, text)
}
