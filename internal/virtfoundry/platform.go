package virtfoundry

import (
	"context"
	"fmt"
	"net/http"
)

// --- Tenants (root-scoped) ---

type CreateTenantInput struct {
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	AdminPassword string `json:"admin_password,omitempty"`
}

func (c *Client) ListTenants(ctx context.Context) ([]Tenant, error) {
	var out struct {
		Tenants []Tenant `json:"tenants"`
	}
	if err := c.jsonRequest(ctx, "", http.MethodGet, "/api/v1/tenants", nil, &out); err != nil {
		return nil, err
	}
	return out.Tenants, nil
}

func (c *Client) CreateTenant(ctx context.Context, in CreateTenantInput) (*Tenant, error) {
	var out struct {
		Tenant Tenant `json:"tenant"`
	}
	if err := c.jsonRequest(ctx, "", http.MethodPost, "/api/v1/tenants", in, &out); err != nil {
		return nil, err
	}
	return &out.Tenant, nil
}

func (c *Client) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	items, err := c.ListTenants(ctx)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(t Tenant) string { return t.ID })
}

// --- VPCs ---

type CreateVPCInput struct {
	Name string `json:"name"`
	CIDR string `json:"cidr"`
}

type CreateVPCResult struct {
	VPC            VPC
	DefaultNetwork *Network
}

func (c *Client) ListVPCs(ctx context.Context, tenantID string) ([]VPC, error) {
	var out struct {
		VPCs []VPC `json:"vpcs"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/vpcs", nil, &out); err != nil {
		return nil, err
	}
	return out.VPCs, nil
}

func (c *Client) CreateVPC(ctx context.Context, tenantID string, in CreateVPCInput) (*CreateVPCResult, error) {
	var out struct {
		VPC            VPC     `json:"vpc"`
		DefaultNetwork Network `json:"default_network"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vpcs", in, &out); err != nil {
		return nil, err
	}
	res := &CreateVPCResult{VPC: out.VPC}
	if out.DefaultNetwork.ID != "" {
		res.DefaultNetwork = &out.DefaultNetwork
	}
	return res, nil
}

func (c *Client) UpdateVPC(ctx context.Context, tenantID, id, name string) (*VPC, error) {
	var out struct {
		VPC VPC `json:"vpc"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPatch, "/api/v1/vpcs/"+id, map[string]string{"name": name}, &out); err != nil {
		return nil, err
	}
	return &out.VPC, nil
}

func (c *Client) DeleteVPC(ctx context.Context, tenantID, id string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodDelete, "/api/v1/vpcs/"+id, nil, nil)
}

func (c *Client) GetVPC(ctx context.Context, tenantID, id string) (*VPC, error) {
	items, err := c.ListVPCs(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(v VPC) string { return v.ID })
}

// --- Networks ---

type CreateNetworkInput struct {
	Name   string `json:"name"`
	CIDR   string `json:"cidr,omitempty"`
	VPCID  string `json:"vpc_id"`
	Prefix int    `json:"prefix,omitempty"`
}

func (c *Client) ListNetworks(ctx context.Context, tenantID string) ([]Network, error) {
	var out struct {
		Networks []Network `json:"networks"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/networks", nil, &out); err != nil {
		return nil, err
	}
	return out.Networks, nil
}

func (c *Client) CreateNetwork(ctx context.Context, tenantID string, in CreateNetworkInput) (*Network, error) {
	var out struct {
		Network Network `json:"network"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/networks", in, &out); err != nil {
		return nil, err
	}
	return &out.Network, nil
}

func (c *Client) UpdateNetwork(ctx context.Context, tenantID, id, name string) (*Network, error) {
	var out struct {
		Network Network `json:"network"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPatch, "/api/v1/networks/"+id, map[string]string{"name": name}, &out); err != nil {
		return nil, err
	}
	return &out.Network, nil
}

func (c *Client) DeleteNetwork(ctx context.Context, tenantID, id string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodDelete, "/api/v1/networks/"+id, nil, nil)
}

func (c *Client) GetNetwork(ctx context.Context, tenantID, id string) (*Network, error) {
	items, err := c.ListNetworks(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(n Network) string { return n.ID })
}

// --- Security groups ---

type CreateSecurityGroupInput struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	VPCID       string              `json:"vpc_id,omitempty"`
	Rules       []SecurityGroupRule `json:"rules,omitempty"`
}

func (c *Client) ListSecurityGroups(ctx context.Context, tenantID string) ([]SecurityGroup, error) {
	var out struct {
		SecurityGroups []SecurityGroup `json:"security_groups"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/security-groups", nil, &out); err != nil {
		return nil, err
	}
	return out.SecurityGroups, nil
}

func (c *Client) CreateSecurityGroup(ctx context.Context, tenantID string, in CreateSecurityGroupInput) (*SecurityGroup, error) {
	var out struct {
		SecurityGroup SecurityGroup `json:"security_group"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/security-groups", in, &out); err != nil {
		return nil, err
	}
	return &out.SecurityGroup, nil
}

func (c *Client) UpdateSecurityGroup(ctx context.Context, tenantID, id string, in CreateSecurityGroupInput) (*SecurityGroup, error) {
	body := map[string]any{
		"name":        in.Name,
		"description": in.Description,
		"rules":       in.Rules,
	}
	var out struct {
		SecurityGroup SecurityGroup `json:"security_group"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPatch, "/api/v1/security-groups/"+id, body, &out); err != nil {
		return nil, err
	}
	return &out.SecurityGroup, nil
}

func (c *Client) DeleteSecurityGroup(ctx context.Context, tenantID, id string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodDelete, "/api/v1/security-groups/"+id, nil, nil)
}

func (c *Client) GetSecurityGroup(ctx context.Context, tenantID, id string) (*SecurityGroup, error) {
	items, err := c.ListSecurityGroups(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(sg SecurityGroup) string { return sg.ID })
}

// --- Volumes ---

type CreateVolumeInput struct {
	Name   string `json:"name"`
	SizeGi int    `json:"size_gi"`
}

func (c *Client) ListVolumes(ctx context.Context, tenantID string) ([]Volume, error) {
	var out struct {
		Volumes []Volume `json:"volumes"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/volumes", nil, &out); err != nil {
		return nil, err
	}
	return out.Volumes, nil
}

func (c *Client) CreateVolume(ctx context.Context, tenantID string, in CreateVolumeInput) (*Volume, error) {
	var out struct {
		Volume Volume `json:"volume"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/volumes", in, &out); err != nil {
		return nil, err
	}
	return &out.Volume, nil
}

func (c *Client) GetVolume(ctx context.Context, tenantID, id string) (*Volume, error) {
	items, err := c.ListVolumes(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(v Volume) string { return v.ID })
}

// ErrDeleteNotSupported indicates the API has no delete endpoint for this resource type.
var ErrDeleteNotSupported = fmt.Errorf("VirtFoundry API does not support deleting this resource type yet")

// --- Volume snapshots ---

type CreateVolumeSnapshotInput struct {
	VolumeID string `json:"volume_id"`
	Name     string `json:"name"`
}

func (c *Client) ListVolumeSnapshots(ctx context.Context, tenantID string) ([]VolumeSnapshot, error) {
	var out struct {
		Snapshots []VolumeSnapshot `json:"snapshots"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/snapshots", nil, &out); err != nil {
		return nil, err
	}
	return out.Snapshots, nil
}

func (c *Client) CreateVolumeSnapshot(ctx context.Context, tenantID string, in CreateVolumeSnapshotInput) (*VolumeSnapshot, error) {
	var out struct {
		Snapshot VolumeSnapshot `json:"snapshot"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/snapshots", in, &out); err != nil {
		return nil, err
	}
	return &out.Snapshot, nil
}

func (c *Client) GetVolumeSnapshot(ctx context.Context, tenantID, id string) (*VolumeSnapshot, error) {
	items, err := c.ListVolumeSnapshots(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(s VolumeSnapshot) string { return s.ID })
}

// --- VM templates ---

type CreateVMTemplateInput struct {
	Name              string `json:"name"`
	DisplayName       string `json:"display_name,omitempty"`
	Description       string `json:"description,omitempty"`
	Image             string `json:"image"`
	SourceType        string `json:"source_type,omitempty"`
	OSType            string `json:"os_type,omitempty"`
	CloudInitUserData string `json:"cloud_init_user_data,omitempty"`
	ISOVolumeID       string `json:"iso_volume_id,omitempty"`
	ISOSizeGi         int    `json:"iso_size_gi,omitempty"`
	BootDiskSizeGi    int    `json:"boot_disk_size_gi,omitempty"`
	StorageClass      string `json:"storage_class,omitempty"`
}

type UpdateVMTemplateInput struct {
	DisplayName       string `json:"display_name,omitempty"`
	Description       string `json:"description,omitempty"`
	Image             string `json:"image,omitempty"`
	SourceType        string `json:"source_type,omitempty"`
	OSType            string `json:"os_type,omitempty"`
	CloudInitUserData string `json:"cloud_init_user_data,omitempty"`
	State             string `json:"state,omitempty"`
}

func (c *Client) ListVMTemplates(ctx context.Context, tenantID string) ([]VMTemplate, error) {
	var out struct {
		VMTemplates []VMTemplate `json:"vm_templates"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/vm-templates", nil, &out); err != nil {
		return nil, err
	}
	return out.VMTemplates, nil
}

func (c *Client) CreateVMTemplate(ctx context.Context, tenantID string, in CreateVMTemplateInput) (*VMTemplate, error) {
	var out struct {
		VMTemplate VMTemplate `json:"vm_template"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vm-templates", in, &out); err != nil {
		return nil, err
	}
	return &out.VMTemplate, nil
}

func (c *Client) UpdateVMTemplate(ctx context.Context, tenantID, id string, in UpdateVMTemplateInput) (*VMTemplate, error) {
	var out struct {
		VMTemplate VMTemplate `json:"vm_template"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPatch, "/api/v1/vm-templates/"+id, in, &out); err != nil {
		return nil, err
	}
	return &out.VMTemplate, nil
}

func (c *Client) DeleteVMTemplate(ctx context.Context, tenantID, id string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodDelete, "/api/v1/vm-templates/"+id, nil, nil)
}

func (c *Client) GetVMTemplate(ctx context.Context, tenantID, id string) (*VMTemplate, error) {
	items, err := c.ListVMTemplates(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(t VMTemplate) string { return t.ID })
}

// --- VM snapshots ---

type CreateVMSnapshotInput struct {
	VMName string `json:"vm_name"`
	Name   string `json:"name"`
}

type vmSnapshotNameRequest struct {
	Name string `json:"name"`
}

func (c *Client) ListVMSnapshots(ctx context.Context, tenantID string) ([]VMSnapshot, error) {
	var out struct {
		VMSnapshots []VMSnapshot `json:"vm_snapshots"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/vm-snapshots", nil, &out); err != nil {
		return nil, err
	}
	return out.VMSnapshots, nil
}

func (c *Client) CreateVMSnapshot(ctx context.Context, tenantID string, in CreateVMSnapshotInput) (*VMSnapshot, error) {
	var out struct {
		VMSnapshot VMSnapshot `json:"vm_snapshot"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vm-snapshots", in, &out); err != nil {
		return nil, err
	}
	return &out.VMSnapshot, nil
}

func (c *Client) DeleteVMSnapshot(ctx context.Context, tenantID, name string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vm-snapshots/delete", vmSnapshotNameRequest{Name: name}, nil)
}

type restoreVMSnapshotInput struct {
	Name   string `json:"name"`
	VMName string `json:"vm_name"`
}

// RestoreVMSnapshot restores a VM from a snapshot.
func (c *Client) RestoreVMSnapshot(ctx context.Context, tenantID, snapshotName, vmName string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vm-snapshots/restore", restoreVMSnapshotInput{
		Name: snapshotName, VMName: vmName,
	}, nil)
}

func (c *Client) GetVMSnapshot(ctx context.Context, tenantID, id string) (*VMSnapshot, error) {
	items, err := c.ListVMSnapshots(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(s VMSnapshot) string { return s.ID })
}

func (c *Client) GetVMSnapshotByName(ctx context.Context, tenantID, name string) (*VMSnapshot, error) {
	items, err := c.ListVMSnapshots(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, name, func(s VMSnapshot) string { return s.Name })
}

// --- SSH keys ---

type RegisterSSHKeyInput struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

func (c *Client) ListSSHKeys(ctx context.Context, tenantID string) ([]SSHKey, error) {
	var out struct {
		SSHKeys []SSHKey `json:"ssh_keys"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/ssh-keys", nil, &out); err != nil {
		return nil, err
	}
	return out.SSHKeys, nil
}

func (c *Client) RegisterSSHKey(ctx context.Context, tenantID string, in RegisterSSHKeyInput) (*SSHKey, error) {
	var out struct {
		Key SSHKey `json:"key"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/ssh-keys/register", in, &out); err != nil {
		return nil, err
	}
	return &out.Key, nil
}

// CreateSSHKey generates a new Ed25519 key pair in the tenant.
func (c *Client) CreateSSHKey(ctx context.Context, tenantID, name string) (*SSHKey, string, error) {
	var out struct {
		Key        SSHKey `json:"key"`
		PrivateKey string `json:"private_key_pem"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/ssh-keys", map[string]string{"name": name}, &out); err != nil {
		return nil, "", err
	}
	return &out.Key, out.PrivateKey, nil
}

func (c *Client) DeleteSSHKey(ctx context.Context, tenantID, id string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodDelete, "/api/v1/ssh-keys/"+id, nil, nil)
}

func (c *Client) GetSSHKey(ctx context.Context, tenantID, id string) (*SSHKey, error) {
	items, err := c.ListSSHKeys(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return findByID(items, id, func(k SSHKey) string { return k.ID })
}

// --- Service offerings (read-only) ---

func (c *Client) ListServiceOfferings(ctx context.Context) ([]ServiceOffering, error) {
	var out struct {
		ServiceOfferings []ServiceOffering `json:"service_offerings"`
	}
	if err := c.jsonRequest(ctx, "", http.MethodGet, "/api/v1/service-offerings", nil, &out); err != nil {
		return nil, err
	}
	return out.ServiceOfferings, nil
}
