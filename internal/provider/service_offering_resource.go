package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
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
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	DisplayName  types.String `tfsdk:"display_name"`
	CPU          types.Int64  `tfsdk:"cpu"`
	MemoryMi     types.Int64  `tfsdk:"memory_mi"`
	DedicatedCPU types.Bool   `tfsdk:"dedicated_cpu"`
	State        types.String `tfsdk:"state"`
}

func NewServiceOfferingResource() resource.Resource { return &serviceOfferingResource{} }

func (r *serviceOfferingResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_service_offering"
}

func (r *serviceOfferingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a platform service offering (CPU/memory catalog). Requires root credentials. Changing catalog size does not resize existing VMs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Offering slug. Forces replacement.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"display_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Human-readable name. Defaults to `name` on create.",
			},
			"cpu": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "vCPU count.",
			},
			"memory_mi": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Memory in MiB.",
			},
			"dedicated_cpu": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Guaranteed CPU (request equals limit).",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"state": schema.StringAttribute{Computed: true},
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
	in := virtfoundry.CreateServiceOfferingInput{
		Name:         plan.Name.ValueString(),
		CPU:          int(plan.CPU.ValueInt64()),
		MemoryMi:     plan.MemoryMi.ValueInt64(),
		DedicatedCPU: !plan.DedicatedCPU.IsNull() && plan.DedicatedCPU.ValueBool(),
	}
	if !plan.DisplayName.IsNull() && plan.DisplayName.ValueString() != "" {
		in.DisplayName = plan.DisplayName.ValueString()
	}
	off, err := r.client.CreateServiceOffering(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Create service offering failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, offeringToModel(off, plan))...)
}

func (r *serviceOfferingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
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
	resp.Diagnostics.Append(resp.State.Set(ctx, offeringToModel(off, state))...)
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
	ded := !plan.DedicatedCPU.IsNull() && plan.DedicatedCPU.ValueBool()
	in := virtfoundry.UpdateServiceOfferingInput{
		CPU:          int(plan.CPU.ValueInt64()),
		MemoryMi:     plan.MemoryMi.ValueInt64(),
		DedicatedCPU: &ded,
	}
	if !plan.DisplayName.IsNull() {
		in.DisplayName = plan.DisplayName.ValueString()
	}
	off, err := r.client.UpdateServiceOffering(ctx, plan.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Update service offering failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, offeringToModel(off, plan))...)
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
	err := r.client.DeleteServiceOffering(ctx, state.ID.ValueString())
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete service offering failed", err.Error())
	}
}

func (r *serviceOfferingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importByID(ctx, resp, req.ID)
}

func offeringToModel(o *virtfoundry.ServiceOffering, cfg serviceOfferingResourceModel) serviceOfferingResourceModel {
	out := serviceOfferingResourceModel{
		ID:           types.StringValue(o.ID),
		Name:         types.StringValue(o.Name),
		DisplayName:  types.StringValue(o.DisplayName),
		CPU:          types.Int64Value(int64(o.CPU)),
		MemoryMi:     types.Int64Value(o.MemoryMi),
		DedicatedCPU: types.BoolValue(o.DedicatedCPU),
		State:        types.StringValue(o.State),
	}
	if !cfg.Name.IsNull() && cfg.Name.ValueString() != "" {
		out.Name = cfg.Name
	}
	return out
}
