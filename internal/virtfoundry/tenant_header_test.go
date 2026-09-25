package virtfoundry_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

// TestTenantHeaderIsolationUnderConcurrency asserts that a Client shared by
// parallel Terraform resources never sends one tenant's X-Tenant-ID on another
// tenant's request. Each VM is named after its tenant, so the server can verify
// the header against the requested path.
func TestTenantHeaderIsolationUnderConcurrency(t *testing.T) {
	t.Parallel()

	const (
		tenants             = 8
		requestsPerTenant   = 40
		vmNamePrefix        = "vm-of-"
		defaultTenant       = "tenant-default"
		unexpectedTenantFmt = "path %q carried X-Tenant-ID %q"
	)

	var (
		mu         sync.Mutex
		mismatches []string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/v1/vms/")
		want := strings.TrimPrefix(name, vmNamePrefix)
		if got := r.Header.Get("X-Tenant-ID"); got != want {
			mu.Lock()
			mismatches = append(mismatches, fmt.Sprintf(unexpectedTenantFmt, r.URL.Path, got))
			mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"vm":{"id":%q,"tenant_id":%q,"name":%q,"state":"Running","cpu":1,"memory_mi":1024}}`, name, want, name)
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(context.Background(), srv.URL, true)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("vfd_live_test")
	// A provider-level default tenant must not leak into tenant-scoped calls.
	client.SetTenantID(defaultTenant)

	ctx := context.Background()
	var wg sync.WaitGroup
	for i := range tenants {
		tenantID := fmt.Sprintf("tenant-%02d", i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requestsPerTenant {
				vm, err := client.GetVM(ctx, tenantID, vmNamePrefix+tenantID)
				if err != nil {
					mu.Lock()
					mismatches = append(mismatches, fmt.Sprintf("GetVM(%s): %v", tenantID, err))
					mu.Unlock()
					return
				}
				if vm.TenantID != tenantID {
					mu.Lock()
					mismatches = append(mismatches, fmt.Sprintf("got tenant %q, want %q", vm.TenantID, tenantID))
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(mismatches) > 0 {
		t.Fatalf("tenant header leaked across %d concurrent requests:\n%s", len(mismatches), strings.Join(mismatches, "\n"))
	}
	if got := client.TenantID(); got != defaultTenant {
		t.Fatalf("default tenant mutated by scoped requests: got %q, want %q", got, defaultTenant)
	}
}

// TestTenantScopeFallsBackToDefault documents that an empty per-request tenant
// keeps using the provider-level default.
func TestTenantScopeFallsBackToDefault(t *testing.T) {
	t.Parallel()

	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Header.Get("X-Tenant-ID"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(context.Background(), srv.URL, true)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("vfd_live_test")
	client.SetTenantID("tenant-default")

	ctx := context.Background()
	for _, tenantID := range []string{"", "tenant-scoped", ""} {
		resp, err := client.Do(ctx, tenantID, http.MethodGet, "/api/v1/auth/me", nil)
		if err != nil {
			t.Fatalf("Do(%q): %v", tenantID, err)
		}
		_ = resp.Body.Close()
	}

	want := []string{"tenant-default", "tenant-scoped", "tenant-default"}
	if len(got) != len(want) {
		t.Fatalf("got %d requests, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("request %d: got X-Tenant-ID %q, want %q", i, got[i], want[i])
		}
	}
}
