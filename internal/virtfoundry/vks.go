package virtfoundry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// maxKubeconfigBytes caps the kubeconfig response body we are willing to read.
const maxKubeconfigBytes = 1 << 20

// VKSLocalObjectRef references a sibling object by name.
type VKSLocalObjectRef struct {
	Name string `json:"name"`
}

// VKSControlPlane mirrors the VKSCluster control plane exposure spec.
type VKSControlPlane struct {
	ServiceType string `json:"service_type,omitempty"`
	Address     string `json:"address,omitempty"`
	Port        int32  `json:"port,omitempty"`
}

// VKSWorkers mirrors the VKSCluster worker pool spec.
type VKSWorkers struct {
	Count       int32               `json:"count"`
	TemplateRef VKSLocalObjectRef   `json:"template_ref"`
	OfferingRef VKSLocalObjectRef   `json:"offering_ref"`
	NetworkRef  VKSLocalObjectRef   `json:"network_ref"`
	SSHKeyRefs  []VKSLocalObjectRef `json:"ssh_key_refs,omitempty"`
}

// VKSCondition mirrors a status condition on the cluster.
type VKSCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// VKSCluster is a VirtFoundry Kubernetes Service cluster.
type VKSCluster struct {
	Name                 string          `json:"name"`
	TenantID             string          `json:"tenant_id"`
	Namespace            string          `json:"namespace"`
	KubernetesVersion    string          `json:"kubernetes_version"`
	ControlPlane         VKSControlPlane `json:"control_plane"`
	Workers              VKSWorkers      `json:"workers"`
	Phase                string          `json:"phase,omitempty"`
	ControlPlaneEndpoint string          `json:"control_plane_endpoint,omitempty"`
	ReadyWorkers         int32           `json:"ready_workers,omitempty"`
	CreatedAt            string          `json:"created_at,omitempty"`
	Conditions           []VKSCondition  `json:"conditions,omitempty"`
}

// CreateVKSClusterInput is the create payload for POST /api/v1/vks/clusters.
type CreateVKSClusterInput struct {
	Name              string          `json:"name"`
	KubernetesVersion string          `json:"kubernetes_version"`
	ControlPlane      VKSControlPlane `json:"control_plane"`
	Workers           VKSWorkers      `json:"workers"`
}

func vksClusterPath(name string) string {
	return "/api/v1/vks/clusters/" + url.PathEscape(name)
}

// ListVKSClusters lists clusters in the tenant.
func (c *Client) ListVKSClusters(ctx context.Context, tenantID string) ([]VKSCluster, error) {
	var out struct {
		Clusters []VKSCluster `json:"clusters"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, "/api/v1/vks/clusters", nil, &out); err != nil {
		return nil, err
	}
	return out.Clusters, nil
}

// CreateVKSCluster creates a VKS cluster (Kamaji control plane + worker Instances).
func (c *Client) CreateVKSCluster(ctx context.Context, tenantID string, in CreateVKSClusterInput) (*VKSCluster, error) {
	var out struct {
		Cluster VKSCluster `json:"cluster"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodPost, "/api/v1/vks/clusters", in, &out); err != nil {
		return nil, err
	}
	return &out.Cluster, nil
}

// GetVKSCluster fetches a cluster by name.
func (c *Client) GetVKSCluster(ctx context.Context, tenantID, name string) (*VKSCluster, error) {
	var out struct {
		Cluster VKSCluster `json:"cluster"`
	}
	if err := c.jsonRequest(ctx, tenantID, http.MethodGet, vksClusterPath(name), nil, &out); err != nil {
		return nil, err
	}
	return &out.Cluster, nil
}

// DeleteVKSCluster requests deletion of a cluster by name.
func (c *Client) DeleteVKSCluster(ctx context.Context, tenantID, name string) error {
	return c.jsonRequest(ctx, tenantID, http.MethodDelete, vksClusterPath(name), nil, nil)
}

// GetVKSKubeconfig downloads the admin kubeconfig (YAML) for a cluster.
// Requires the vks:kubeconfig permission.
func (c *Client) GetVKSKubeconfig(ctx context.Context, tenantID, name string) (string, error) {
	resp, err := c.do(ctx, tenantID, http.MethodGet, vksClusterPath(name)+"/kubeconfig", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", apiError(resp)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxKubeconfigBytes+1))
	if err != nil {
		return "", fmt.Errorf("read kubeconfig: %w", err)
	}
	if len(raw) > maxKubeconfigBytes {
		return "", fmt.Errorf("kubeconfig response exceeds %d bytes", maxKubeconfigBytes)
	}
	return string(raw), nil
}
