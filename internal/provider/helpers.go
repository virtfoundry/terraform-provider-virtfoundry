package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

func configureClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) *virtfoundry.Client {
	if req.ProviderData == nil {
		return nil
	}
	client, ok := req.ProviderData.(*virtfoundry.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data",
			fmt.Sprintf("Expected *virtfoundry.Client, got %T", req.ProviderData),
		)
		return nil
	}
	return client
}

func requireClient(client *virtfoundry.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	if client == nil {
		diags.AddError("Provider not configured", "VirtFoundry client is nil")
	}
	return diags
}

func requireRootClient(client *virtfoundry.Client) diag.Diagnostics {
	diags := requireClient(client)
	if diags.HasError() {
		return diags
	}
	if client.TenantID() != "" {
		diags.AddError(
			"Root credentials required",
			"virtfoundry_tenant requires a provider without tenant_id. Use a dedicated root provider alias (see modules/tenant).",
		)
	}
	return diags
}

func resolveTenantID(client *virtfoundry.Client, attr types.String) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !attr.IsNull() && attr.ValueString() != "" {
		return attr.ValueString(), diags
	}
	if tid := client.TenantID(); tid != "" {
		return tid, diags
	}
	diags.AddError("Missing tenant_id", "Set tenant_id on the resource or provider block.")
	return "", diags
}

func stringValue(v types.String, fallback string) string {
	if !v.IsNull() && v.ValueString() != "" {
		return v.ValueString()
	}
	return fallback
}

func isNotFound(err error) bool {
	return virtfoundry.IsNotFound(err)
}

func importByID(ctx context.Context, resp *resource.ImportStateResponse, id string) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func importTenantName(ctx context.Context, resp *resource.ImportStateResponse, id string) {
	parts := strings.Split(id, "/")
	switch len(parts) {
	case 2:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
	case 1:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[0])...)
	default:
		resp.Diagnostics.AddError("Invalid import ID", "Use `<tenant_id>/<id>` or `<id>`.")
	}
}
