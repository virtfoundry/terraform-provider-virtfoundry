package virtfoundry

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// VM is a VirtFoundry virtual machine.
type VM struct {
	ID                string `json:"id"`
	TenantID          string `json:"tenant_id"`
	Name              string `json:"name"`
	DisplayName       string `json:"display_name"`
	State             string `json:"state"`
	ErrorMessage      string `json:"error_message"`
	CPU               int    `json:"cpu"`
	MemoryMi          int64  `json:"memory_mi"`
	Image             string `json:"image"`
	Template          string `json:"template"`
	IP                string `json:"ip"`
	ServiceOfferingID string `json:"service_offering_id"`
}

// DeployVMInput is the POST /vms payload.
type DeployVMInput struct {
	Name              string   `json:"name"`
	DisplayName       string   `json:"display_name,omitempty"`
	TemplateID        string   `json:"template_id,omitempty"`
	ServiceOfferingID string   `json:"service_offering_id,omitempty"`
	CPU               int      `json:"cpu,omitempty"`
	MemoryMi          int64    `json:"memory_mi,omitempty"`
	Image             string   `json:"image,omitempty"`
	PublicIP          bool     `json:"public_ip,omitempty"`
	NetworkIDs        []string `json:"network_ids,omitempty"`
	SecurityGroupIDs  []string `json:"security_group_ids,omitempty"`
}

type vmNameRequest struct {
	Name string `json:"name"`
}

// GetVM returns a VM by name within a tenant.
func (c *Client) GetVM(ctx context.Context, tenantID, name string) (*VM, error) {
	var out struct {
		VM VM `json:"vm"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/vms/"+name, nil, &out); err != nil {
		return nil, err
	}
	return &out.VM, nil
}

// DeployVM creates a VM synchronously.
func (c *Client) DeployVM(ctx context.Context, tenantID string, in DeployVMInput) (*VM, error) {
	var out struct {
		VM VM `json:"vm"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vms", in, &out); err != nil {
		return nil, err
	}
	return &out.VM, nil
}

// UpdateVM patches VM metadata/resources.
func (c *Client) UpdateVM(ctx context.Context, tenantID, name string, displayName string, cpu int, memoryMi int64) (*VM, error) {
	body := map[string]any{}
	if displayName != "" {
		body["display_name"] = displayName
	}
	if cpu > 0 {
		body["cpu"] = cpu
	}
	if memoryMi > 0 {
		body["memory_mi"] = memoryMi
	}
	var out struct {
		VM VM `json:"vm"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPatch, "/api/v1/vms/"+name, body, &out); err != nil {
		return nil, err
	}
	return &out.VM, nil
}

// StartVM powers on a VM.
func (c *Client) StartVM(ctx context.Context, tenantID, name string) (*VM, error) {
	var out struct {
		VM VM `json:"vm"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vms/start", vmNameRequest{Name: name}, &out); err != nil {
		return nil, err
	}
	return &out.VM, nil
}

// StopVM powers off a VM.
func (c *Client) StopVM(ctx context.Context, tenantID, name string) (*VM, error) {
	var out struct {
		VM VM `json:"vm"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vms/stop", vmNameRequest{Name: name}, &out); err != nil {
		return nil, err
	}
	return &out.VM, nil
}

// DeleteVM removes a VM.
func (c *Client) DeleteVM(ctx context.Context, tenantID, name string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vms/delete", vmNameRequest{Name: name}, nil)
}

// WaitForVMState polls until the VM reaches the target state or timeout.
func (c *Client) WaitForVMState(ctx context.Context, tenantID, name, want string, timeout time.Duration) (*VM, error) {
	deadline := time.Now().Add(timeout)
	want = strings.ToLower(strings.TrimSpace(want))
	for {
		vm, err := c.GetVM(ctx, tenantID, name)
		if err != nil {
			return nil, err
		}
		if stateMatches(vm.State, want) {
			return vm, nil
		}
		if time.Now().After(deadline) {
			msg := fmt.Sprintf("timeout waiting for VM %q state %q (last=%q)", name, want, vm.State)
			if vm.ErrorMessage != "" {
				msg += ": " + vm.ErrorMessage
			}
			return vm, fmt.Errorf("%s", msg)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}

func stateMatches(actual, want string) bool {
	return StateMatches(actual, want)
}

// StateMatches reports whether an API VM state satisfies the desired Terraform state.
func StateMatches(actual, want string) bool {
	actual = strings.ToLower(strings.TrimSpace(actual))
	switch want {
	case "running":
		return actual == "running" || actual == "starting" || actual == "scheduled"
	case "stopped":
		return actual == "stopped" || actual == "shutoff" || actual == "shutdown"
	default:
		return actual == want
	}
}
