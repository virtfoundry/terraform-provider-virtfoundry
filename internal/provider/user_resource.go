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
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

type userResource struct{ client *virtfoundry.Client }

type userModel struct {
	ID       types.String `tfsdk:"id"`
	TenantID types.String `tfsdk:"tenant_id"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Email    types.String `tfsdk:"email"`
	RoleID   types.String `tfsdk:"role_id"`
	RoleName types.String `tfsdk:"role_name"`
	Role     types.String `tfsdk:"role"`
	State    types.String `tfsdk:"state"`
}

func NewUserResource() resource.Resource { return &userResource{} }

func (r *userResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry tenant IAM user. Use modules/tenant-iam or a tenant-scoped provider.",
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tenant_id": schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"username":  schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"password":  schema.StringAttribute{Required: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"email":     schema.StringAttribute{Optional: true},
			"role_id":   schema.StringAttribute{Optional: true},
			"role_name": schema.StringAttribute{Optional: true, MarkdownDescription: "Role name when `role_id` is omitted."},
			"role":      schema.StringAttribute{Computed: true},
			"state":     schema.StringAttribute{Optional: true},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := virtfoundry.CreateUserInput{Username: plan.Username.ValueString(), Password: plan.Password.ValueString()}
	if !plan.Email.IsNull() {
		in.Email = plan.Email.ValueString()
	}
	if !plan.RoleID.IsNull() {
		in.RoleID = plan.RoleID.ValueString()
	}
	if !plan.RoleName.IsNull() {
		in.RoleName = plan.RoleName.ValueString()
	}
	u, err := r.client.CreateUser(ctx, tenantID, in)
	if err != nil {
		resp.Diagnostics.AddError("Create user failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(u, plan))...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	u, err := r.client.GetUser(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read user failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(u, state))...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := virtfoundry.UpdateUserInput{}
	if !plan.Email.IsNull() {
		in.Email = plan.Email.ValueString()
	}
	if !plan.RoleID.IsNull() {
		in.RoleID = plan.RoleID.ValueString()
	}
	if !plan.State.IsNull() {
		in.State = plan.State.ValueString()
	}
	u, err := r.client.UpdateUser(ctx, tenantID, plan.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Update user failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(u, plan))...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteUser(ctx, tenantID, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete user failed", err.Error())
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
}

func userToModel(u *virtfoundry.User, cfg userModel) userModel {
	out := userModel{
		ID:       types.StringValue(u.ID),
		Username: types.StringValue(u.Username),
		Role:     types.StringValue(u.Role),
	}
	if u.Email != "" {
		out.Email = types.StringValue(u.Email)
	} else if !cfg.Email.IsNull() {
		out.Email = cfg.Email
	}
	if u.RoleID != "" {
		out.RoleID = types.StringValue(u.RoleID)
	} else if !cfg.RoleID.IsNull() {
		out.RoleID = cfg.RoleID
	}
	if u.State != "" {
		out.State = types.StringValue(u.State)
	} else if !cfg.State.IsNull() {
		out.State = cfg.State
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	if !cfg.Password.IsNull() {
		out.Password = cfg.Password
	}
	if !cfg.RoleName.IsNull() {
		out.RoleName = cfg.RoleName
	}
	return out
}
