package virtfoundry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) jsonRequest(ctx context.Context, tenantID, method, path string, in any, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	resp, err := c.doWithTenant(ctx, tenantID, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apiError(resp)
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) doWithTenant(ctx context.Context, tenantID, method, path string, body io.Reader) (*http.Response, error) {
	if tenantID != "" {
		prev := c.tenantID
		c.tenantID = tenantID
		defer func() { c.tenantID = prev }()
	}
	return c.do(ctx, method, path, body)
}
