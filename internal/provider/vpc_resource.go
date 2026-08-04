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
	_ resource.Resource                = &vpcResource{}
	_ resource.ResourceWithImportState = &vpcResource{}
)

type vpcResource struct {
	client *virtfoundry.Client
}

type vpcModel struct {
	ID               types.String `tfsdk:"id"`
	TenantID         types.String `tfsdk:"tenant_id"`
	Name             types.String `tfsdk:"name"`
	CIDR             types.String `tfsdk:"cidr"`
	Namespace        types.String `tfsdk:"namespace"`
	State            types.String `tfsdk:"state"`
	DefaultNetworkID types.String `tfsdk:"default_network_id"`
}

func NewVPCResource() resource.Resource { return &vpcResource{} }

func (r *vpcResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_vpc"
}

func (r *vpcResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry VPC. Creating a VPC also provisions a default network.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tenant_id": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"cidr": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "VPC CIDR block (e.g. `10.0.0.0/16`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"namespace": schema.StringAttribute{Computed: true},
			"state":     schema.StringAttribute{Computed: true},
			"default_network_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the auto-created default network.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *vpcResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *vpcResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan vpcModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.CreateVPC(ctx, tenantID, virtfoundry.CreateVPCInput{
		Name: plan.Name.ValueString(),
		CIDR: plan.CIDR.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create VPC failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, vpcToModel(&res.VPC, res.DefaultNetwork, plan))...)
}

func (r *vpcResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state vpcModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	vpc, err := r.client.GetVPC(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read VPC failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, vpcToModel(vpc, nil, state))...)
}

func (r *vpcResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan, state vpcModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Name.Equal(state.Name) {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	vpc, err := r.client.UpdateVPC(ctx, tenantID, plan.ID.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update VPC failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, vpcToModel(vpc, nil, plan))...)
}

func (r *vpcResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state vpcModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVPC(ctx, tenantID, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete VPC failed", err.Error())
	}
}

func (r *vpcResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
}

func vpcToModel(v *virtfoundry.VPC, defNet *virtfoundry.Network, cfg vpcModel) vpcModel {
	out := vpcModel{
		ID:        types.StringValue(v.ID),
		Name:      types.StringValue(v.Name),
		CIDR:      types.StringValue(v.CIDR),
		Namespace: types.StringValue(v.Namespace),
		State:     types.StringValue(v.State),
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	if defNet != nil && defNet.ID != "" {
		out.DefaultNetworkID = types.StringValue(defNet.ID)
	} else if !cfg.DefaultNetworkID.IsNull() {
		out.DefaultNetworkID = cfg.DefaultNetworkID
	}
	return out
}
