package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

func TestVolumeDeleteCallsAPI(t *testing.T) {
	t.Parallel()

	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	t.Cleanup(srv.Close)

	ctx := context.Background()
	client, err := virtfoundry.NewClient(ctx, srv.URL, true)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("test-token")

	var schemaResp resource.SchemaResponse
	(&volumeResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	ot := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)

	stateRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"id":        tftypes.NewValue(tftypes.String, "vol-1"),
		"tenant_id": tftypes.NewValue(tftypes.String, "t1"),
		"name":      tftypes.NewValue(tftypes.String, "data"),
		"size_gi":   tftypes.NewValue(tftypes.Number, 10),
		"namespace": tftypes.NewValue(tftypes.String, "ns"),
		"pvc_name":  tftypes.NewValue(tftypes.String, "pvc"),
		"state":     tftypes.NewValue(tftypes.String, "ready"),
		"vm_id":     tftypes.NewValue(tftypes.String, nil),
	})

	r := &volumeResource{client: client}
	req := resource.DeleteRequest{
		State: tfsdk.State{Schema: schemaResp.Schema, Raw: stateRaw},
	}
	var resp resource.DeleteResponse
	resp.State = tfsdk.State{Schema: schemaResp.Schema, Raw: stateRaw}

	r.Delete(ctx, req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete diagnostics: %v", resp.Diagnostics)
	}
	if gotMethod != http.MethodDelete {
		t.Fatalf("method: got %q want DELETE", gotMethod)
	}
	if gotPath != "/api/v1/volumes/vol-1" {
		t.Fatalf("path: got %q want /api/v1/volumes/vol-1", gotPath)
	}
}
