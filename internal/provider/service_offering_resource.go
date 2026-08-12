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
	_ resource.Resource                = &serviceOfferingResource{}
	_ resource.ResourceWithImportState = &serviceOfferingResource{}
)

type serviceOfferingResource struct {
	client *virtfoundry.Client
}

type serviceOfferingResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	CPU         types.Int64  `tfsdk:"cpu"`
	MemoryMi    types.Int64  `tfsdk:"memory_mi"`
	State       types.String `tfsdk:"state"`
}

func NewServiceOfferingResource() resource.Resource { return &serviceOfferingResource{} }

func (r *serviceOfferingResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_service_offering"
}

func (r *serviceOfferingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry service offering (root-only). Soft-delete sets state to Inactive.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":         schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"display_name": schema.StringAttribute{Required: true},
			"cpu":          schema.Int64Attribute{Required: true},
			"memory_mi":    schema.Int64Attribute{Required: true},
			"state":        schema.StringAttribute{Computed: true},
		},
	}
}

func (r *serviceOfferingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *serviceOfferingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireRootClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan serviceOfferingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	off, err := r.client.CreateServiceOffering(ctx, virtfoundry.CreateServiceOfferingInput{
		Name:        plan.Name.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
		CPU:         int(plan.CPU.ValueInt64()),
		MemoryMi:    plan.MemoryMi.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create service offering failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, offeringToModel(off))...)
}

func (r *serviceOfferingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireRootClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceOfferingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	off, err := r.client.GetServiceOffering(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read service offering failed", err.Error())
		return
	}
	if off.State != "" && off.State != "Active" {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, offeringToModel(off))...)
}

func (r *serviceOfferingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(requireRootClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan serviceOfferingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	off, err := r.client.UpdateServiceOffering(ctx, plan.ID.ValueString(), virtfoundry.UpdateServiceOfferingInput{
		DisplayName: plan.DisplayName.ValueString(),
		CPU:         int(plan.CPU.ValueInt64()),
		MemoryMi:    plan.MemoryMi.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update service offering failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, offeringToModel(off))...)
}

func (r *serviceOfferingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireRootClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceOfferingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteServiceOffering(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete service offering failed", err.Error())
	}
}

func (r *serviceOfferingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importByID(ctx, resp, req.ID)
}

func offeringToModel(o *virtfoundry.ServiceOffering) serviceOfferingResourceModel {
	return serviceOfferingResourceModel{
		ID:          types.StringValue(o.ID),
		Name:        types.StringValue(o.Name),
		DisplayName: types.StringValue(o.DisplayName),
		CPU:         types.Int64Value(int64(o.CPU)),
		MemoryMi:    types.Int64Value(o.MemoryMi),
		State:       types.StringValue(o.State),
	}
}
