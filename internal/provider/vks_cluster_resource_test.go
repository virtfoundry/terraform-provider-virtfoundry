package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

func TestVKSClusterSchemaValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var resp resource.SchemaResponse
	NewVKSClusterResource().Schema(ctx, resource.SchemaRequest{}, &resp)
	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("invalid schema: %v", diags)
	}
}

func testVKSCluster() *virtfoundry.VKSCluster {
	return &virtfoundry.VKSCluster{
		Name:                 "dev",
		Namespace:            "vks-dev",
		KubernetesVersion:    "v1.31.4",
		Phase:                "Ready",
		ControlPlaneEndpoint: "https://10.0.0.9:6443",
		ReadyWorkers:         2,
		Workers: virtfoundry.VKSWorkers{
			Count:       2,
			TemplateRef: virtfoundry.VKSLocalObjectRef{Name: "node-v1"},
			OfferingRef: virtfoundry.VKSLocalObjectRef{Name: "medium"},
			NetworkRef:  virtfoundry.VKSLocalObjectRef{Name: "lan"},
			SSHKeyRefs:  []virtfoundry.VKSLocalObjectRef{{Name: "admin"}},
		},
	}
}

func TestVKSClusterToModel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cfg := vksClusterModel{
		TenantID:     types.StringValue("t1"),
		Workers:      types.ObjectNull(vksWorkersAttrTypes),
		ControlPlane: types.ObjectUnknown(vksControlPlaneAttrTypes),
	}
	out, diags := vksClusterToModel(ctx, testVKSCluster(), cfg)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if out.ID.ValueString() != "dev" || out.Name.ValueString() != "dev" || out.Namespace.ValueString() != "vks-dev" {
		t.Fatalf("unexpected identity: %+v", out)
	}
	if out.Phase.ValueString() != "Ready" || out.ReadyWorkers.ValueInt64() != 2 {
		t.Fatalf("unexpected status: %+v", out)
	}
	if out.TenantID.ValueString() != "t1" {
		t.Fatalf("tenant_id not preserved: %v", out.TenantID)
	}
	if out.ControlPlane.IsUnknown() || out.ControlPlane.IsNull() {
		t.Fatalf("control_plane must be known after apply: %v", out.ControlPlane)
	}
	var w vksWorkersModel
	if d := out.Workers.As(ctx, &w, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("workers: %v", d)
	}
	if w.Count.ValueInt64() != 2 || len(w.SSHKeyRefs.Elements()) != 1 {
		t.Fatalf("unexpected workers: %+v", w)
	}
}

func TestVKSClusterToModelKeepsConfiguredControlPlaneWhenAPIOmits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cp, diags := types.ObjectValue(vksControlPlaneAttrTypes, map[string]attr.Value{
		"service_type": types.StringValue("LoadBalancer"),
		"address":      types.StringNull(),
		"port":         types.Int64Value(6443),
	})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	out, diags := vksClusterToModel(ctx, testVKSCluster(), vksClusterModel{
		Workers:      types.ObjectNull(vksWorkersAttrTypes),
		ControlPlane: cp,
	})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	var got vksControlPlaneModel
	if d := out.ControlPlane.As(ctx, &got, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("control_plane: %v", d)
	}
	if got.ServiceType.ValueString() != "LoadBalancer" || got.Port.ValueInt64() != 6443 || !got.Address.IsNull() {
		t.Fatalf("configured control_plane not preserved: %+v", got)
	}
}

func TestVKSClusterToModelNullSSHKeysWhenAbsent(t *testing.T) {
	t.Parallel()
	c := testVKSCluster()
	c.Workers.SSHKeyRefs = nil
	out, diags := vksClusterToModel(context.Background(), c, vksClusterModel{
		Workers:      types.ObjectNull(vksWorkersAttrTypes),
		ControlPlane: types.ObjectNull(vksControlPlaneAttrTypes),
	})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	var w vksWorkersModel
	if d := out.Workers.As(context.Background(), &w, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("workers: %v", d)
	}
	if !w.SSHKeyRefs.IsNull() {
		t.Fatalf("expected null ssh_key_refs, got %v", w.SSHKeyRefs)
	}
}

func TestVKSCreateInputRejectsZeroWorkers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workers, diags := types.ObjectValue(vksWorkersAttrTypes, map[string]attr.Value{
		"count":        types.Int64Value(0),
		"template_ref": types.StringValue("t"),
		"offering_ref": types.StringValue("o"),
		"network_ref":  types.StringValue("n"),
		"ssh_key_refs": types.SetNull(types.StringType),
	})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	_, diags = vksCreateInputFromPlan(ctx, vksClusterModel{
		Name:              types.StringValue("dev"),
		KubernetesVersion: types.StringValue("v1.31.4"),
		Workers:           workers,
		ControlPlane:      types.ObjectUnknown(vksControlPlaneAttrTypes),
	})
	if !diags.HasError() {
		t.Fatal("expected error for count=0")
	}
}
