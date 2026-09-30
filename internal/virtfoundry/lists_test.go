package virtfoundry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

func TestListEndpointsDecodeAPIKeys(t *testing.T) {
	t.Parallel()

	fixtures := map[string]string{
		"/api/v1/vpcs":            `{"vpcs":[{"id":"v1","name":"main","cidr":"10.0.0.0/16","state":"ready"}]}`,
		"/api/v1/networks":        `{"networks":[{"id":"n1","name":"default","vpc_id":"v1","cidr":"10.0.1.0/24","state":"ready"}]}`,
		"/api/v1/security-groups": `{"security_groups":[{"id":"sg1","name":"web","vpc_id":"v1","description":"http"}]}`,
		"/api/v1/ssh-keys":        `{"ssh_keys":[{"id":"k1","name":"laptop","public_key":"ssh-ed25519 AAA","fingerprint":"SHA256:abc"}]}`,
		"/api/v1/roles":           `{"roles":[{"id":"r1","name":"admin","description":"full","is_system":true,"permissions":["*"]}]}`,
		"/api/v1/users":           `{"users":[{"id":"u1","username":"alice","email":"a@ex.com","role":"admin","role_id":"r1","state":"active"}]}`,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := fixtures[r.URL.Path]
		if !ok || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Tenant-ID") != "t1" {
			t.Errorf("%s: expected X-Tenant-ID=t1, got %q", r.URL.Path, r.Header.Get("X-Tenant-ID"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(context.Background(), srv.URL, true)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("test-token")
	ctx := context.Background()
	tenantID := "t1"

	vpcs, err := client.ListVPCs(ctx, tenantID)
	if err != nil {
		t.Fatalf("ListVPCs: %v", err)
	}
	if len(vpcs) != 1 || vpcs[0].ID != "v1" || vpcs[0].Name != "main" {
		t.Fatalf("ListVPCs unexpected: %+v", vpcs)
	}

	nets, err := client.ListNetworks(ctx, tenantID)
	if err != nil {
		t.Fatalf("ListNetworks: %v", err)
	}
	if len(nets) != 1 || nets[0].ID != "n1" || nets[0].VPCID != "v1" {
		t.Fatalf("ListNetworks unexpected: %+v", nets)
	}

	sgs, err := client.ListSecurityGroups(ctx, tenantID)
	if err != nil {
		t.Fatalf("ListSecurityGroups: %v", err)
	}
	if len(sgs) != 1 || sgs[0].ID != "sg1" {
		t.Fatalf("ListSecurityGroups unexpected: %+v", sgs)
	}

	keys, err := client.ListSSHKeys(ctx, tenantID)
	if err != nil {
		t.Fatalf("ListSSHKeys: %v", err)
	}
	if len(keys) != 1 || keys[0].ID != "k1" {
		t.Fatalf("ListSSHKeys unexpected: %+v", keys)
	}

	roles, err := client.ListRoles(ctx, tenantID)
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if len(roles) != 1 || roles[0].ID != "r1" || !roles[0].IsSystem {
		t.Fatalf("ListRoles unexpected: %+v", roles)
	}

	users, err := client.ListUsers(ctx, tenantID)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 1 || users[0].ID != "u1" || users[0].Username != "alice" {
		t.Fatalf("ListUsers unexpected: %+v", users)
	}
}

func TestDeleteVolume(t *testing.T) {
	t.Parallel()

	var gotMethod, gotPath, gotTenant string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotTenant = r.Header.Get("X-Tenant-ID")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(context.Background(), srv.URL, true)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("test-token")

	if err := client.DeleteVolume(context.Background(), "t1", "vol-1"); err != nil {
		t.Fatalf("DeleteVolume: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Fatalf("method: got %q want DELETE", gotMethod)
	}
	if gotPath != "/api/v1/volumes/vol-1" {
		t.Fatalf("path: got %q want /api/v1/volumes/vol-1", gotPath)
	}
	if gotTenant != "t1" {
		t.Fatalf("tenant: got %q want t1", gotTenant)
	}
}
