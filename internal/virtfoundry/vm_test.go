package virtfoundry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

func TestDeployGetDeleteVM(t *testing.T) {
	t.Parallel()

	tenantID := "tenant-1"
	var created bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/vms":
			if r.Header.Get("X-Tenant-ID") != tenantID {
				http.Error(w, `{"error":"tenant"}`, http.StatusBadRequest)
				return
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"vm":{"id":"vm-1","tenant_id":"tenant-1","name":"web-01","state":"Starting","cpu":1,"memory_mi":1024}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/vms/web-01":
			if !created {
				http.NotFound(w, r)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"vm":{"id":"vm-1","tenant_id":"tenant-1","name":"web-01","state":"Running","cpu":1,"memory_mi":1024,"ip":"10.0.0.5"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/vms/delete":
			created = false
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(srv.URL, false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("vfd_live_test")

	ctx := context.Background()
	vm, err := client.DeployVM(ctx, tenantID, virtfoundry.DeployVMInput{
		Name:              "web-01",
		TemplateID:        "tmpl-1",
		ServiceOfferingID: "small",
	})
	if err != nil {
		t.Fatalf("DeployVM: %v", err)
	}
	if vm.State != "Starting" {
		t.Fatalf("state: %s", vm.State)
	}

	got, err := client.GetVM(ctx, tenantID, "web-01")
	if err != nil {
		t.Fatalf("GetVM: %v", err)
	}
	if got.IP != "10.0.0.5" {
		t.Fatalf("ip: %s", got.IP)
	}

	if err := client.DeleteVM(ctx, tenantID, "web-01"); err != nil {
		t.Fatalf("DeleteVM: %v", err)
	}
}

func TestStateMatches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		actual, want string
		ok           bool
	}{
		{"Running", "running", true},
		{"Stopped", "stopped", true},
		{"Shutoff", "stopped", true},
		{"Starting", "running", true},
	}
	for _, tc := range cases {
		if virtfoundry.StateMatches(tc.actual, tc.want) != tc.ok {
			t.Fatalf("%s/%s", tc.actual, tc.want)
		}
	}
}
