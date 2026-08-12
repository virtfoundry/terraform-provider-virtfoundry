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
	_ resource.Resource                = &volumeResource{}
	_ resource.ResourceWithImportState = &volumeResource{}
)

type volumeResource struct {
	client *virtfoundry.Client
}

type volumeModel struct {
	ID        types.String `tfsdk:"id"`
	TenantID  types.String `tfsdk:"tenant_id"`
	Name      types.String `tfsdk:"name"`
	SizeGi    types.Int64  `tfsdk:"size_gi"`
	Namespace types.String `tfsdk:"namespace"`
	PVCName   types.String `tfsdk:"pvc_name"`
	State     types.String `tfsdk:"state"`
	VMID      types.String `tfsdk:"vm_id"`
}

func NewVolumeResource() resource.Resource { return &volumeResource{} }

func (r *volumeResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_volume"
}

func (r *volumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry block storage volume.",
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tenant_id": schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":      schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"size_gi":   schema.Int64Attribute{Required: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
			"namespace": schema.StringAttribute{Computed: true},
			"pvc_name":  schema.StringAttribute{Computed: true},
			"state":     schema.StringAttribute{Computed: true},
			"vm_id":     schema.StringAttribute{Computed: true},
		},
	}
}

func (r *volumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *volumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan volumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	vol, err := r.client.CreateVolume(ctx, tenantID, virtfoundry.CreateVolumeInput{
		Name:   plan.Name.ValueString(),
		SizeGi: int(plan.SizeGi.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create volume failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, volumeToModel(vol, plan))...)
}

func (r *volumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	vol, err := r.client.GetVolume(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read volume failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, volumeToModel(vol, state))...)
}

func (r *volumeResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Volume size cannot be changed after creation.")
}

func (r *volumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVolume(ctx, tenantID, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete volume failed", err.Error())
	}
}

func (r *volumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
}

func volumeToModel(v *virtfoundry.Volume, cfg volumeModel) volumeModel {
	out := volumeModel{
		ID:        types.StringValue(v.ID),
		Name:      types.StringValue(v.Name),
		SizeGi:    types.Int64Value(int64(v.SizeGi)),
		Namespace: types.StringValue(v.Namespace),
		PVCName:   types.StringValue(v.PVCName),
		State:     types.StringValue(v.State),
	}
	if v.VMID != "" {
		out.VMID = types.StringValue(v.VMID)
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	return out
}
