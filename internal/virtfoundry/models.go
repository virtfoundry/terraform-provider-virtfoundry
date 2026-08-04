package virtfoundry

import "time"

// Tenant is a VirtFoundry tenant.
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Namespace string    `json:"namespace"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
}

// VPC is a tenant virtual private cloud.
type VPC struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	CIDR      string    `json:"cidr"`
	Namespace string    `json:"namespace"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
}

// Network is a tenant network (isolated or shared).
type Network struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id,omitempty"`
	VPCID        string    `json:"vpc_id,omitempty"`
	Name         string    `json:"name"`
	NetworkType  string    `json:"network_type,omitempty"`
	CIDR         string    `json:"cidr"`
	Gateway      string    `json:"gateway,omitempty"`
	NADNamespace string    `json:"nad_namespace,omitempty"`
	NADName      string    `json:"nad_name,omitempty"`
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
}

// SecurityGroupRule is a firewall rule.
type SecurityGroupRule struct {
	Direction string `json:"direction"`
	Protocol  string `json:"protocol"`
	PortFrom  int    `json:"port_from,omitempty"`
	PortTo    int    `json:"port_to,omitempty"`
	CIDR      string `json:"cidr"`
}

// SecurityGroup is a tenant security group.
type SecurityGroup struct {
	ID          string              `json:"id"`
	TenantID    string              `json:"tenant_id"`
	VPCID       string              `json:"vpc_id,omitempty"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Rules       []SecurityGroupRule `json:"rules"`
	CreatedAt   time.Time           `json:"created_at"`
}

// Volume is a block storage volume.
type Volume struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	SizeGi    int       `json:"size_gi"`
	Namespace string    `json:"namespace"`
	PVCName   string    `json:"pvc_name"`
	State     string    `json:"state"`
	VMID      string    `json:"vm_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// VolumeSnapshot is a point-in-time copy of a volume.
type VolumeSnapshot struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	VolumeID    string    `json:"volume_id"`
	Name        string    `json:"name"`
	Namespace   string    `json:"namespace"`
	SnapshotUID string    `json:"snapshot_uid,omitempty"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
}

// VMTemplate is a deployable VM image profile.
type VMTemplate struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id,omitempty"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description,omitempty"`
	Image       string    `json:"image"`
	SourceType  string    `json:"source_type,omitempty"`
	OSType      string    `json:"os_type,omitempty"`
	ImportState string    `json:"import_state,omitempty"`
	Hypervisor  string    `json:"hypervisor"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
}

// VMSnapshot is a KubeVirt VM snapshot.
type VMSnapshot struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	VMID        string    `json:"vm_id"`
	VMName      string    `json:"vm_name"`
	Name        string    `json:"name"`
	Namespace   string    `json:"namespace"`
	SnapshotUID string    `json:"snapshot_uid,omitempty"`
	Phase       string    `json:"phase"`
	CreatedAt   time.Time `json:"created_at"`
}

// SSHKey is a tenant SSH public key.
type SSHKey struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	PublicKey   string    `json:"public_key"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at"`
}

// ServiceOffering is a CPU/memory catalog entry.
type ServiceOffering struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	CPU         int       `json:"cpu"`
	MemoryMi    int64     `json:"memory_mi"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
}

// VMSSHInfo describes NodePort SSH exposure for a VM.
type VMSSHInfo struct {
	Exposed  bool   `json:"exposed"`
	NodePort int32  `json:"node_port,omitempty"`
	VMIP     string `json:"vm_ip,omitempty"`
}

// User is a tenant IAM user.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	RoleID   string `json:"role_id,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
	Email    string `json:"email,omitempty"`
	State    string `json:"state,omitempty"`
}

// Role is a tenant IAM role.
type Role struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenant_id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	IsSystem    bool     `json:"is_system"`
	Permissions []string `json:"permissions,omitempty"`
}

// APIKey is a VirtFoundry API key metadata record.
type APIKey struct {
	ID       string   `json:"id"`
	UserID   string   `json:"user_id"`
	TenantID string   `json:"tenant_id,omitempty"`
	Name     string   `json:"name"`
	Prefix   string   `json:"prefix"`
	Scopes   []string `json:"scopes,omitempty"`
}

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
