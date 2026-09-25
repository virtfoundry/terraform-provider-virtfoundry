package virtfoundry

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const defaultTimeout = 30 * time.Second

// Client is a thin HTTP client for the VirtFoundry REST API.
//
// A single Client is shared by every resource in a provider run, and Terraform
// walks the graph in parallel. Per-request state (such as the tenant scope) is
// therefore passed as an argument and never stored on the Client.
type Client struct {
	baseURL    string
	httpClient *http.Client

	// mu guards the credentials below, which are written at provider configure
	// time and read by concurrent requests.
	mu              sync.RWMutex
	token           string
	defaultTenantID string
}

// NewClient builds a client for the given API endpoint.
func NewClient(ctx context.Context, endpoint string, insecure bool) (*Client, error) {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	parsedURL, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	if parsedURL.Scheme == "http" && !insecure {
		return nil, fmt.Errorf("HTTP endpoints are not allowed unless insecure=true is explicitly set; use HTTPS or set insecure=true")
	}

	if insecure && !isLoopbackHost(parsedURL.Hostname()) {
		tflog.Warn(ctx, "TLS certificate verification disabled (insecure=true) for non-loopback host; this is insecure", map[string]any{
			"host": parsedURL.Hostname(),
		})
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

func isLoopbackHost(host string) bool {
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return true
	}
	return false
}

// SetAPIKey configures Bearer auth with a VirtFoundry API key (vfd_live_...).
func (c *Client) SetAPIKey(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = strings.TrimSpace(key)
}

// SetTenantID sets the default X-Tenant-ID header used by requests that do not
// carry their own tenant scope.
func (c *Client) SetTenantID(tenantID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultTenantID = strings.TrimSpace(tenantID)
}

// TenantID returns the configured default tenant ID.
func (c *Client) TenantID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.defaultTenantID
}

// credentials snapshots the shared auth state for a single request.
func (c *Client) credentials() (token, defaultTenantID string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token, c.defaultTenantID
}

func (c *Client) setToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
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
	c.setToken(out.Token)
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
	resp, err := c.do(ctx, "", http.MethodGet, "/api/v1/auth/me", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}
	return nil
}

// Do sends an authenticated API request scoped to the given tenant. An empty
// tenantID falls back to the provider-level default tenant.
func (c *Client) Do(ctx context.Context, tenantID, method, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, tenantID, method, path, body)
}

// do sends an authenticated API request. tenantID scopes this request only and
// is never written back to the Client, so parallel requests cannot observe each
// other's tenant scope.
func (c *Client) do(ctx context.Context, tenantID, method, path string, body io.Reader) (*http.Response, error) {
	token, defaultTenantID := c.credentials()
	if token == "" {
		return nil, fmt.Errorf("not authenticated")
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		tenantID = defaultTenantID
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if tenantID != "" {
		req.Header.Set("X-Tenant-ID", tenantID)
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
