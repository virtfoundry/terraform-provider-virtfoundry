package virtfoundry_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

func testClient(t *testing.T, h http.HandlerFunc) *virtfoundry.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	client, err := virtfoundry.NewClient(srv.URL, false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("vfd_live_test")
	return client
}

func TestDeleteVolume(t *testing.T) {
	t.Parallel()
	var deleted string
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/volumes/vol-1" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Tenant-ID") != "tenant-1" {
			http.Error(w, `{"error":"tenant"}`, http.StatusBadRequest)
			return
		}
		deleted = "vol-1"
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	})
	if err := client.DeleteVolume(context.Background(), "tenant-1", "vol-1"); err != nil {
		t.Fatalf("DeleteVolume: %v", err)
	}
	if deleted != "vol-1" {
		t.Fatal("delete not called")
	}
}

func TestDeleteVolumeConflict(t *testing.T) {
	t.Parallel()
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"volume attached to a VM"}`, http.StatusConflict)
	})
	err := client.DeleteVolume(context.Background(), "tenant-1", "vol-1")
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "409") {
		t.Fatalf("want HTTP 409, got %v", err)
	}
}

func TestAttachAndDetachVolume(t *testing.T) {
	t.Parallel()
	var attached, detached bool
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/vms/web-01/volumes":
			body, _ := io.ReadAll(r.Body)
			var in map[string]string
			if err := json.Unmarshal(body, &in); err != nil || in["volume_id"] != "vol-1" {
				http.Error(w, `{"error":"volume_id"}`, http.StatusBadRequest)
				return
			}
			attached = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"volume":{"id":"vol-1","name":"data","size_gi":10,"vm_id":"vm-1","state":"attached"}}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/vms/web-01/volumes/vol-1":
			detached = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"volume":{"id":"vol-1","name":"data","size_gi":10,"state":"ready"}}`))
		default:
			http.NotFound(w, r)
		}
	})
	ctx := context.Background()
	vol, err := client.AttachVolume(ctx, "tenant-1", "web-01", "vol-1")
	if err != nil {
		t.Fatalf("AttachVolume: %v", err)
	}
	if vol.VMID != "vm-1" {
		t.Fatalf("vm_id: %s", vol.VMID)
	}
	if err := client.DetachVolume(ctx, "tenant-1", "web-01", "vol-1"); err != nil {
		t.Fatalf("DetachVolume: %v", err)
	}
	if !attached || !detached {
		t.Fatal("attach/detach not called")
	}
}

func TestDeleteTenant(t *testing.T) {
	t.Parallel()
	var gotPath string
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/tenants/ten-1" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"deleted":true,"id":"ten-1"}`))
	})
	if err := client.DeleteTenant(context.Background(), "ten-1"); err != nil {
		t.Fatalf("DeleteTenant: %v", err)
	}
	if gotPath != "/api/v1/tenants/ten-1" {
		t.Fatalf("path: %s", gotPath)
	}
}

func TestServiceOfferingCRUD(t *testing.T) {
	t.Parallel()
	var created, patched, deleted bool
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/service-offerings":
			created = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"service_offering":{"id":"off-1","name":"e2e-xl","display_name":"E2E XL","cpu":2,"memory_mi":2048,"dedicated_cpu":true,"state":"Active"}}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/service-offerings/off-1":
			patched = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"service_offering":{"id":"off-1","name":"e2e-xl","display_name":"XL","cpu":4,"memory_mi":4096,"dedicated_cpu":false,"state":"Active"}}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/service-offerings/off-1":
			deleted = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/service-offerings":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"service_offerings":[{"id":"off-1","name":"e2e-xl","cpu":2,"memory_mi":2048,"dedicated_cpu":true,"state":"Active"}]}`))
		default:
			http.NotFound(w, r)
		}
	})
	ctx := context.Background()
	off, err := client.CreateServiceOffering(ctx, virtfoundry.CreateServiceOfferingInput{
		Name: "e2e-xl", DisplayName: "E2E XL", CPU: 2, MemoryMi: 2048, DedicatedCPU: true,
	})
	if err != nil {
		t.Fatalf("CreateServiceOffering: %v", err)
	}
	if !off.DedicatedCPU || off.ID != "off-1" {
		t.Fatalf("create: %+v", off)
	}
	got, err := client.GetServiceOffering(ctx, "off-1")
	if err != nil {
		t.Fatalf("GetServiceOffering: %v", err)
	}
	if !got.DedicatedCPU {
		t.Fatal("expected dedicated_cpu on get")
	}
	falseVal := false
	updated, err := client.UpdateServiceOffering(ctx, "off-1", virtfoundry.UpdateServiceOfferingInput{
		DisplayName: "XL", CPU: 4, MemoryMi: 4096, DedicatedCPU: &falseVal,
	})
	if err != nil {
		t.Fatalf("UpdateServiceOffering: %v", err)
	}
	if updated.CPU != 4 || updated.DedicatedCPU {
		t.Fatalf("update: %+v", updated)
	}
	if err := client.DeleteServiceOffering(ctx, "off-1"); err != nil {
		t.Fatalf("DeleteServiceOffering: %v", err)
	}
	if !created || !patched || !deleted {
		t.Fatal("CRUD not fully exercised")
	}
}

func TestDeployVMDedicatedCPUAndUpdateOffering(t *testing.T) {
	t.Parallel()
	var deployBody map[string]any
	var patchBody map[string]any
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/vms":
			_ = json.Unmarshal(body, &deployBody)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"vm":{"id":"vm-1","name":"web-01","state":"Running","cpu":1,"memory_mi":1024,"dedicated_cpu":true,"service_offering_id":"small"}}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/vms/web-01":
			_ = json.Unmarshal(body, &patchBody)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"vm":{"id":"vm-1","name":"web-01","state":"Running","cpu":2,"memory_mi":4096,"service_offering_id":"medium"}}`))
		default:
			http.NotFound(w, r)
		}
	})
	ctx := context.Background()
	vm, err := client.DeployVM(ctx, "tenant-1", virtfoundry.DeployVMInput{
		Name: "web-01", DedicatedCPU: true, ServiceOfferingID: "small",
	})
	if err != nil {
		t.Fatalf("DeployVM: %v", err)
	}
	if !vm.DedicatedCPU {
		t.Fatal("expected dedicated_cpu on VM")
	}
	if deployBody["dedicated_cpu"] != true {
		t.Fatalf("deploy body: %+v", deployBody)
	}
	updated, err := client.UpdateVM(ctx, "tenant-1", "web-01", virtfoundry.UpdateVMInput{
		ServiceOfferingID: "medium",
	})
	if err != nil {
		t.Fatalf("UpdateVM: %v", err)
	}
	if updated.ServiceOfferingID != "medium" {
		t.Fatalf("offering: %s", updated.ServiceOfferingID)
	}
	if patchBody["service_offering_id"] != "medium" {
		t.Fatalf("patch body: %+v", patchBody)
	}
	if _, ok := patchBody["cpu"]; ok {
		t.Fatal("resize patch must not send cpu")
	}
	if _, ok := patchBody["memory_mi"]; ok {
		t.Fatal("resize patch must not send memory_mi")
	}
}
