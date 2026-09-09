package virtfoundry

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

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
	SSHKeyID          string   `json:"ssh_key_id,omitempty"`
	DataVolumeID      string   `json:"data_volume_id,omitempty"`
	ExposeSSH         bool     `json:"expose_ssh,omitempty"`
	DedicatedCPU      bool     `json:"dedicated_cpu,omitempty"`
}

// UpdateVMInput is the PATCH /vms/{name} payload.
type UpdateVMInput struct {
	DisplayName       string `json:"display_name,omitempty"`
	CPU               int    `json:"cpu,omitempty"`
	MemoryMi          int64  `json:"memory_mi,omitempty"`
	ServiceOfferingID string `json:"service_offering_id,omitempty"`
}

type vmNameRequest struct {
	Name string `json:"name"`
}

func (c *Client) ListVMs(ctx context.Context, tenantID string) ([]VM, error) {
	var out struct {
		VMs []VM `json:"vms"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/vms", nil, &out); err != nil {
		return nil, err
	}
	return out.VMs, nil
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
func (c *Client) UpdateVM(ctx context.Context, tenantID, name string, in UpdateVMInput) (*VM, error) {
	var out struct {
		VM VM `json:"vm"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPatch, "/api/v1/vms/"+name, in, &out); err != nil {
		return nil, err
	}
	return &out.VM, nil
}

type attachVolumeInput struct {
	VolumeID string `json:"volume_id"`
}

// AttachVolume hot-plugs a volume onto a running VM.
func (c *Client) AttachVolume(ctx context.Context, tenantID, vmName, volumeID string) (*Volume, error) {
	var out struct {
		Volume Volume `json:"volume"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vms/"+vmName+"/volumes", attachVolumeInput{VolumeID: volumeID}, &out); err != nil {
		return nil, err
	}
	return &out.Volume, nil
}

// DetachVolume removes a hot-plugged volume from a VM.
func (c *Client) DetachVolume(ctx context.Context, tenantID, vmName, volumeID string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodDelete, "/api/v1/vms/"+vmName+"/volumes/"+volumeID, nil, nil)
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

type exposeSSHRequest struct {
	NodePort int32 `json:"node_port,omitempty"`
}

// GetVMSSH returns NodePort SSH exposure details for a VM.
func (c *Client) GetVMSSH(ctx context.Context, tenantID, name string) (*VMSSHInfo, error) {
	var out VMSSHInfo
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/vms/"+name+"/ssh", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ExposeVMSSH creates or updates a NodePort Service for VM SSH access.
func (c *Client) ExposeVMSSH(ctx context.Context, tenantID, name string, nodePort int32) (int32, error) {
	var out struct {
		NodePort int32 `json:"node_port"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vms/"+name+"/ssh", exposeSSHRequest{NodePort: nodePort}, &out); err != nil {
		return 0, err
	}
	return out.NodePort, nil
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

// WaitForVMExactRunning polls until the guest is Running (not Starting).
func (c *Client) WaitForVMExactRunning(ctx context.Context, tenantID, name string, timeout time.Duration) (*VM, error) {
	deadline := time.Now().Add(timeout)
	for {
		vm, err := c.GetVM(ctx, tenantID, name)
		if err != nil {
			return nil, err
		}
		if IsFullyRunning(vm.State) {
			return vm, nil
		}
		if StateMatches(vm.State, "stopped") {
			return vm, nil
		}
		if time.Now().After(deadline) {
			return vm, fmt.Errorf("timeout waiting for VM %q to reach Running (last=%q)", name, vm.State)
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

// IsFullyRunning is true only when the guest reports Running (not Starting).
func IsFullyRunning(actual string) bool {
	return strings.ToLower(strings.TrimSpace(actual)) == "running"
}
