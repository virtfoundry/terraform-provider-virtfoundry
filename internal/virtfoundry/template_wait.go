package virtfoundry

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// WaitForVMTemplateImport polls until an ISO template finishes CDI import.
// last is the template returned by Create; it may be omitted from list APIs while State is Inactive.
func (c *Client) WaitForVMTemplateImport(ctx context.Context, tenantID string, last *VMTemplate, timeout time.Duration) (*VMTemplate, error) {
	if last == nil {
		return nil, fmt.Errorf("template is nil")
	}
	if !strings.EqualFold(last.SourceType, "iso") {
		return last, nil
	}

	deadline := time.Now().Add(timeout)
	current := last
	for {
		if tmpl, err := c.GetVMTemplate(ctx, tenantID, current.ID); err == nil {
			current = tmpl
		} else if !strings.Contains(err.Error(), "not found") {
			return nil, err
		}

		switch strings.ToLower(strings.TrimSpace(current.ImportState)) {
		case "ready":
			return current, nil
		case "failed":
			msg := strings.TrimSpace(current.Description)
			if msg == "" {
				msg = "ISO import failed"
			}
			return current, fmt.Errorf("%s", msg)
		}

		if time.Now().After(deadline) {
			return current, fmt.Errorf("timeout waiting for ISO import on template %q (last import_state=%q)", current.ID, current.ImportState)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
}
