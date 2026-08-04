package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var (
	_ resource.Resource                = &networkResource{}
	_ resource.ResourceWithImportState = &networkResource{}
)

type networkResource struct {
	client *virtfoundry.Client
}

type networkModel struct {
	ID          types.String `tfsdk:"id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	VPCID       types.String `tfsdk:"vpc_id"`
	Name        types.String `tfsdk:"name"`
	CIDR        types.String `tfsdk:"cidr"`
	Prefix      types.Int64  `tfsdk:"prefix"`
	Gateway     types.String `tfsdk:"gateway"`
	NetworkType types.String `tfsdk:"network_type"`
	State       types.String `tfsdk:"state"`
}

func NewNetworkResource() resource.Resource { return &networkResource{} }

func (r *networkResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_network"
}

func (r *networkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry tenant network within a VPC.",
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tenant_id": schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"vpc_id":    schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":      schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"cidr": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Explicit CIDR; omit to auto-allocate from the VPC.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prefix": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Subnet prefix when CIDR is auto-allocated (default 24).",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"gateway":      schema.StringAttribute{Computed: true},
			"network_type": schema.StringAttribute{Computed: true},
			"state":        schema.StringAttribute{Computed: true},
		},
	}
}

func (r *networkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan networkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := virtfoundry.CreateNetworkInput{
		Name:  plan.Name.ValueString(),
		VPCID: plan.VPCID.ValueString(),
	}
	if !plan.CIDR.IsNull() {
		in.CIDR = plan.CIDR.ValueString()
	}
	if !plan.Prefix.IsNull() {
		in.Prefix = int(plan.Prefix.ValueInt64())
	}
	net, err := r.client.CreateNetwork(ctx, tenantID, in)
	if err != nil {
		resp.Diagnostics.AddError("Create network failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, networkToModel(net, plan))...)
}

func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state networkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	net, err := r.client.GetNetwork(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read network failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, networkToModel(net, state))...)
}

func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan, state networkModel
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
	net, err := r.client.UpdateNetwork(ctx, tenantID, plan.ID.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update network failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, networkToModel(net, plan))...)
}

func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state networkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteNetwork(ctx, tenantID, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete network failed", err.Error())
	}
}

func (r *networkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
}

func networkToModel(n *virtfoundry.Network, cfg networkModel) networkModel {
	out := networkModel{
		ID:          types.StringValue(n.ID),
		VPCID:       types.StringValue(n.VPCID),
		Name:        types.StringValue(n.Name),
		CIDR:        types.StringValue(n.CIDR),
		State:       types.StringValue(n.State),
		NetworkType: types.StringValue(n.NetworkType),
	}
	if n.Gateway != "" {
		out.Gateway = types.StringValue(n.Gateway)
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	if !cfg.CIDR.IsNull() && cfg.CIDR.ValueString() != "" {
		out.CIDR = cfg.CIDR
	}
	if !cfg.Prefix.IsNull() {
		out.Prefix = cfg.Prefix
	}
	return out
}
