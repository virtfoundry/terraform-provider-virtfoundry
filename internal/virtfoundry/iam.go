package virtfoundry

import (
	"context"
	"net/http"
)

// --- IAM users ---

type CreateUserInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
	RoleID   string `json:"role_id,omitempty"`
	RoleName string `json:"role_name,omitempty"`
}

type UpdateUserInput struct {
	Email  string `json:"email,omitempty"`
	RoleID string `json:"role_id,omitempty"`
	State  string `json:"state,omitempty"`
}

func (c *Client) ListUsers(ctx context.Context, tenantID string) ([]User, error) {
	var out struct {
		Users []User `json:"users"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/users", nil, &out); err != nil {
		return nil, err
	}
	return out.Users, nil
}

func (c *Client) CreateUser(ctx context.Context, tenantID string, in CreateUserInput) (*User, error) {
	var out struct {
		User User `json:"user"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/users", in, &out); err != nil {
		return nil, err
	}
	return &out.User, nil
}

func (c *Client) UpdateUser(ctx context.Context, tenantID, id string, in UpdateUserInput) (*User, error) {
	var out struct {
		User User `json:"user"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPatch, "/api/v1/users/"+id, in, &out); err != nil {
		return nil, err
	}
	return &out.User, nil
}

func (c *Client) DeleteUser(ctx context.Context, tenantID, id string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodDelete, "/api/v1/users/"+id, nil, nil)
}

func (c *Client) GetUser(ctx context.Context, tenantID, id string) (*User, error) {
	items, err := c.ListUsers(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(u User) string { return u.ID })
}

// --- IAM roles ---

type CreateRoleInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type UpdateRoleInput struct {
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

func (c *Client) ListRoles(ctx context.Context, tenantID string) ([]Role, error) {
	var out struct {
		Roles []Role `json:"roles"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/roles", nil, &out); err != nil {
		return nil, err
	}
	return out.Roles, nil
}

func (c *Client) CreateRole(ctx context.Context, tenantID string, in CreateRoleInput) (*Role, error) {
	var out struct {
		Role Role `json:"role"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/roles", in, &out); err != nil {
		return nil, err
	}
	return &out.Role, nil
}

func (c *Client) UpdateRole(ctx context.Context, tenantID, id string, in UpdateRoleInput) (*Role, error) {
	var out struct {
		Role Role `json:"role"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPatch, "/api/v1/roles/"+id, in, &out); err != nil {
		return nil, err
	}
	return &out.Role, nil
}

func (c *Client) DeleteRole(ctx context.Context, tenantID, id string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodDelete, "/api/v1/roles/"+id, nil, nil)
}

func (c *Client) GetRole(ctx context.Context, tenantID, id string) (*Role, error) {
	items, err := c.ListRoles(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(r Role) string { return r.ID })
}

// --- IAM API keys ---

type CreateAPIKeyInput struct {
	Name          string   `json:"name"`
	UserID        string   `json:"user_id,omitempty"`
	ExpiresInDays int      `json:"expires_in_days,omitempty"`
	Scopes        []string `json:"scopes,omitempty"`
}

type CreateAPIKeyResult struct {
	Key    APIKey
	Secret string
}

func (c *Client) ListAPIKeys(ctx context.Context, tenantID string) ([]APIKey, error) {
	var out struct {
		APIKeys []APIKey `json:"api_keys"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/api-keys", nil, &out); err != nil {
		return nil, err
	}
	return out.APIKeys, nil
}

func (c *Client) CreateAPIKey(ctx context.Context, tenantID string, in CreateAPIKeyInput) (*CreateAPIKeyResult, error) {
	var out struct {
		APIKey APIKey `json:"api_key"`
		Secret string `json:"secret"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/api-keys", in, &out); err != nil {
		return nil, err
	}
	return &CreateAPIKeyResult{Key: out.APIKey, Secret: out.Secret}, nil
}

func (c *Client) DeleteAPIKey(ctx context.Context, tenantID, id string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodDelete, "/api/v1/api-keys/"+id, nil, nil)
}

func (c *Client) GetAPIKey(ctx context.Context, tenantID, id string) (*APIKey, error) {
	items, err := c.ListAPIKeys(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(k APIKey) string { return k.ID })
}
