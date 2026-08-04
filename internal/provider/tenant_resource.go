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
	_ resource.Resource                = &tenantResource{}
	_ resource.ResourceWithImportState = &tenantResource{}
)

type tenantResource struct {
	client *virtfoundry.Client
}

type tenantModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Slug          types.String `tfsdk:"slug"`
	AdminPassword types.String `tfsdk:"admin_password"`
	Namespace     types.String `tfsdk:"namespace"`
	State         types.String `tfsdk:"state"`
}

func NewTenantResource() resource.Resource { return &tenantResource{} }

func (r *tenantResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_tenant"
}

func (r *tenantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry tenant. Requires root credentials.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Tenant UUID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for the tenant.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "URL-safe slug; derived from name when omitted.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"admin_password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Initial tenant admin password.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"namespace": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Kubernetes namespace for tenant workloads.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Tenant lifecycle state.",
			},
		},
	}
}

func (r *tenantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *tenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireRootClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := virtfoundry.CreateTenantInput{Name: plan.Name.ValueString()}
	if !plan.Slug.IsNull() {
		in.Slug = plan.Slug.ValueString()
	}
	if !plan.AdminPassword.IsNull() {
		in.AdminPassword = plan.AdminPassword.ValueString()
	}
	t, err := r.client.CreateTenant(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Create tenant failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tenantToModel(t, plan))...)
}

func (r *tenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state tenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	t, err := r.client.GetTenant(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read tenant failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tenantToModel(t, state))...)
}

func (r *tenantResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"The VirtFoundry API does not support updating tenants. Recreate the resource instead.",
	)
}

func (r *tenantResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Delete not supported",
		"The VirtFoundry API does not support deleting tenants via Terraform.",
	)
}

func (r *tenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importByID(ctx, resp, req.ID)
}

func tenantToModel(t *virtfoundry.Tenant, cfg tenantModel) tenantModel {
	out := tenantModel{
		ID:        types.StringValue(t.ID),
		Name:      types.StringValue(t.Name),
		Namespace: types.StringValue(t.Namespace),
		State:     types.StringValue(t.State),
	}
	if !cfg.Slug.IsNull() && cfg.Slug.ValueString() != "" {
		out.Slug = cfg.Slug
	} else if t.Slug != "" {
		out.Slug = types.StringValue(t.Slug)
	}
	if !cfg.AdminPassword.IsNull() {
		out.AdminPassword = cfg.AdminPassword
	}
	return out
}
