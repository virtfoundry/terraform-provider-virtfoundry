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
	_ resource.Resource                = &volumeSnapshotResource{}
	_ resource.ResourceWithImportState = &volumeSnapshotResource{}
)

type volumeSnapshotResource struct {
	client *virtfoundry.Client
}

type volumeSnapshotModel struct {
	ID          types.String `tfsdk:"id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	VolumeID    types.String `tfsdk:"volume_id"`
	Name        types.String `tfsdk:"name"`
	Namespace   types.String `tfsdk:"namespace"`
	SnapshotUID types.String `tfsdk:"snapshot_uid"`
	State       types.String `tfsdk:"state"`
}

func NewVolumeSnapshotResource() resource.Resource { return &volumeSnapshotResource{} }

func (r *volumeSnapshotResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_volume_snapshot"
}

func (r *volumeSnapshotResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry volume snapshot. Destroy removes Terraform state only; the API has no delete endpoint yet.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tenant_id":    schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"volume_id":    schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":         schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"namespace":    schema.StringAttribute{Computed: true},
			"snapshot_uid": schema.StringAttribute{Computed: true},
			"state":        schema.StringAttribute{Computed: true},
		},
	}
}

func (r *volumeSnapshotResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *volumeSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan volumeSnapshotModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	snap, err := r.client.CreateVolumeSnapshot(ctx, tenantID, virtfoundry.CreateVolumeSnapshotInput{
		VolumeID: plan.VolumeID.ValueString(),
		Name:     plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create volume snapshot failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, volumeSnapshotToModel(snap, plan))...)
}

func (r *volumeSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state volumeSnapshotModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	snap, err := r.client.GetVolumeSnapshot(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read volume snapshot failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, volumeSnapshotToModel(snap, state))...)
}

func (r *volumeSnapshotResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *volumeSnapshotResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Volume snapshot not deleted in VirtFoundry",
		"The API has no snapshot delete endpoint. Only Terraform state is cleared.",
	)
}

func (r *volumeSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
}

func volumeSnapshotToModel(s *virtfoundry.VolumeSnapshot, cfg volumeSnapshotModel) volumeSnapshotModel {
	out := volumeSnapshotModel{
		ID:        types.StringValue(s.ID),
		VolumeID:  types.StringValue(s.VolumeID),
		Name:      types.StringValue(s.Name),
		Namespace: types.StringValue(s.Namespace),
		State:     types.StringValue(s.State),
	}
	if s.SnapshotUID != "" {
		out.SnapshotUID = types.StringValue(s.SnapshotUID)
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	return out
}
