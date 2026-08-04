package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var (
	_ resource.Resource                = &vmSnapshotResource{}
	_ resource.ResourceWithImportState = &vmSnapshotResource{}
)

type vmSnapshotResource struct {
	client *virtfoundry.Client
}

type vmSnapshotModel struct {
	ID          types.String `tfsdk:"id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	VMName      types.String `tfsdk:"vm_name"`
	Name        types.String `tfsdk:"name"`
	Namespace   types.String `tfsdk:"namespace"`
	SnapshotUID types.String `tfsdk:"snapshot_uid"`
	Phase       types.String `tfsdk:"phase"`
	VMID        types.String `tfsdk:"vm_id"`
}

func NewVMSnapshotResource() resource.Resource { return &vmSnapshotResource{} }

func (r *vmSnapshotResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_vm_snapshot"
}

func (r *vmSnapshotResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a KubeVirt VM snapshot.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tenant_id":    schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"vm_name":      schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":         schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"namespace":    schema.StringAttribute{Computed: true},
			"snapshot_uid": schema.StringAttribute{Computed: true},
			"phase":        schema.StringAttribute{Computed: true},
			"vm_id":        schema.StringAttribute{Computed: true},
		},
	}
}

func (r *vmSnapshotResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *vmSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan vmSnapshotModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	snap, err := r.client.CreateVMSnapshot(ctx, tenantID, virtfoundry.CreateVMSnapshotInput{
		VMName: plan.VMName.ValueString(),
		Name:   plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create VM snapshot failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, vmSnapshotToModel(snap, plan))...)
}

func (r *vmSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state vmSnapshotModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	snap, err := r.client.GetVMSnapshot(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read VM snapshot failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, vmSnapshotToModel(snap, state))...)
}

func (r *vmSnapshotResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *vmSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state vmSnapshotModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVMSnapshot(ctx, tenantID, state.Name.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete VM snapshot failed", err.Error())
	}
}

func (r *vmSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	switch len(parts) {
	case 3:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vm_name"), parts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[2])...)
	case 2:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vm_name"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
	default:
		importTenantName(ctx, resp, req.ID)
	}
}

func vmSnapshotToModel(s *virtfoundry.VMSnapshot, cfg vmSnapshotModel) vmSnapshotModel {
	out := vmSnapshotModel{
		ID:        types.StringValue(s.ID),
		VMName:    types.StringValue(s.VMName),
		Name:      types.StringValue(s.Name),
		Namespace: types.StringValue(s.Namespace),
		Phase:     types.StringValue(s.Phase),
	}
	if s.SnapshotUID != "" {
		out.SnapshotUID = types.StringValue(s.SnapshotUID)
	}
	if s.VMID != "" {
		out.VMID = types.StringValue(s.VMID)
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	return out
}
