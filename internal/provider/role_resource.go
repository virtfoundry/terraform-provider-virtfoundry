package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var (
	_ resource.Resource                = &roleResource{}
	_ resource.ResourceWithImportState = &roleResource{}
)

type roleResource struct{ client *virtfoundry.Client }

type roleModel struct {
	ID          types.String `tfsdk:"id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.List   `tfsdk:"permissions"`
	IsSystem    types.Bool   `tfsdk:"is_system"`
}

func NewRoleResource() resource.Resource { return &roleResource{} }

func (r *roleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_role"
}

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry tenant IAM role and permissions. Use modules/tenant-iam or a tenant-scoped provider.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tenant_id":   schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":        schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"description": schema.StringAttribute{Optional: true},
			"permissions": schema.ListAttribute{ElementType: types.StringType, Optional: true},
			"is_system":   schema.BoolAttribute{Computed: true},
		},
	}
}

func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan roleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	perms, diags := stringListFromModel(ctx, plan.Permissions)
	resp.Diagnostics.Append(diags...)
	in := virtfoundry.CreateRoleInput{Name: plan.Name.ValueString(), Permissions: perms}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	role, err := r.client.CreateRole(ctx, tenantID, in)
	if err != nil {
		resp.Diagnostics.AddError("Create role failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, roleToModel(role, plan))...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state roleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := r.client.GetRole(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read role failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, roleToModel(role, state))...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan roleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	perms, diags := stringListFromModel(ctx, plan.Permissions)
	resp.Diagnostics.Append(diags...)
	in := virtfoundry.UpdateRoleInput{Permissions: perms}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	role, err := r.client.UpdateRole(ctx, tenantID, plan.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Update role failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, roleToModel(role, plan))...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state roleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.IsSystem.ValueBool() {
		resp.Diagnostics.AddError("Delete not allowed", "System roles cannot be deleted.")
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRole(ctx, tenantID, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete role failed", err.Error())
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
}

func roleToModel(role *virtfoundry.Role, cfg roleModel) roleModel {
	out := roleModel{
		ID:       types.StringValue(role.ID),
		Name:     types.StringValue(role.Name),
		IsSystem: types.BoolValue(role.IsSystem),
	}
	if role.Description != "" {
		out.Description = types.StringValue(role.Description)
	} else if !cfg.Description.IsNull() {
		out.Description = cfg.Description
	}
	if len(role.Permissions) > 0 {
		elems := make([]types.String, len(role.Permissions))
		for i, p := range role.Permissions {
			elems[i] = types.StringValue(p)
		}
		out.Permissions, _ = types.ListValueFrom(ctxBackground, types.StringType, elems)
	} else if !cfg.Permissions.IsNull() {
		out.Permissions = cfg.Permissions
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	return out
}
