package virtfoundry_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

func TestVKSClusterCreateGetDeleteKubeconfig(t *testing.T) {
	t.Parallel()

	const tenantID = "tenant-1"
	var created bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Tenant-ID") != tenantID {
			http.Error(w, `{"error":"tenant"}`, http.StatusBadRequest)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/vks/clusters":
			var in struct {
				Name    string `json:"name"`
				Workers struct {
					Count       int `json:"count"`
					TemplateRef struct {
						Name string `json:"name"`
					} `json:"template_ref"`
					SSHKeyRefs []struct {
						Name string `json:"name"`
					} `json:"ssh_key_refs"`
				} `json:"workers"`
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name != "dev" ||
				in.Workers.Count != 2 || in.Workers.TemplateRef.Name != "node-v1" || len(in.Workers.SSHKeyRefs) != 1 {
				http.Error(w, `{"error":"bad body"}`, http.StatusBadRequest)
				return
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"cluster":{"name":"dev","tenant_id":"tenant-1","namespace":"vks-dev","kubernetes_version":"v1.31.4","phase":"Provisioning","workers":{"count":2,"template_ref":{"name":"node-v1"},"offering_ref":{"name":"medium"},"network_ref":{"name":"lan"},"ssh_key_refs":[{"name":"admin"}]}}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/vks/clusters/dev":
			if !created {
				http.NotFound(w, r)
				return
			}
			_, _ = w.Write([]byte(`{"cluster":{"name":"dev","namespace":"vks-dev","kubernetes_version":"v1.31.4","phase":"Ready","control_plane_endpoint":"https://10.0.0.9:6443","ready_workers":2,"workers":{"count":2,"template_ref":{"name":"node-v1"},"offering_ref":{"name":"medium"},"network_ref":{"name":"lan"}}}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/vks/clusters/dev/kubeconfig":
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write([]byte("apiVersion: v1\nkind: Config\n"))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/vks/clusters/dev":
			created = false
			_, _ = w.Write([]byte(`{"status":"deleting"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(context.Background(), srv.URL, true)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("vfd_live_test")
	ctx := context.Background()

	got, err := client.CreateVKSCluster(ctx, tenantID, virtfoundry.CreateVKSClusterInput{
		Name:              "dev",
		KubernetesVersion: "v1.31.4",
		Workers: virtfoundry.VKSWorkers{
			Count:       2,
			TemplateRef: virtfoundry.VKSLocalObjectRef{Name: "node-v1"},
			OfferingRef: virtfoundry.VKSLocalObjectRef{Name: "medium"},
			NetworkRef:  virtfoundry.VKSLocalObjectRef{Name: "lan"},
			SSHKeyRefs:  []virtfoundry.VKSLocalObjectRef{{Name: "admin"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateVKSCluster: %v", err)
	}
	if got.Phase != "Provisioning" || got.Namespace != "vks-dev" {
		t.Fatalf("unexpected create result: %+v", got)
	}

	got, err = client.GetVKSCluster(ctx, tenantID, "dev")
	if err != nil {
		t.Fatalf("GetVKSCluster: %v", err)
	}
	if got.Phase != "Ready" || got.ReadyWorkers != 2 || got.ControlPlaneEndpoint != "https://10.0.0.9:6443" {
		t.Fatalf("unexpected get result: %+v", got)
	}

	kc, err := client.GetVKSKubeconfig(ctx, tenantID, "dev")
	if err != nil {
		t.Fatalf("GetVKSKubeconfig: %v", err)
	}
	if !strings.HasPrefix(kc, "apiVersion: v1") {
		t.Fatalf("unexpected kubeconfig: %q", kc)
	}

	if err := client.DeleteVKSCluster(ctx, tenantID, "dev"); err != nil {
		t.Fatalf("DeleteVKSCluster: %v", err)
	}
	if _, err := client.GetVKSCluster(ctx, tenantID, "dev"); !virtfoundry.IsNotFound(err) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
	if _, err := client.GetVKSKubeconfig(ctx, tenantID, "missing"); !virtfoundry.IsNotFound(err) {
		t.Fatalf("expected not found for missing kubeconfig, got %v", err)
	}
}
