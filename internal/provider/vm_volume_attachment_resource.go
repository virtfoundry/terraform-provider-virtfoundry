package provider

import (
	"context"
	"fmt"
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
	_ resource.Resource                = &vmVolumeAttachmentResource{}
	_ resource.ResourceWithImportState = &vmVolumeAttachmentResource{}
)

type vmVolumeAttachmentResource struct {
	client *virtfoundry.Client
}

type vmVolumeAttachmentModel struct {
	ID       types.String `tfsdk:"id"`
	TenantID types.String `tfsdk:"tenant_id"`
	VMName   types.String `tfsdk:"vm_name"`
	VolumeID types.String `tfsdk:"volume_id"`
}

func NewVMVolumeAttachmentResource() resource.Resource { return &vmVolumeAttachmentResource{} }

func (r *vmVolumeAttachmentResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_vm_volume_attachment"
}

func (r *vmVolumeAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches an existing volume to a VM (hot-plug). Destroy detaches the volume.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Attachment id `<vm_name>/<volume_id>`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tenant_id": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"vm_name": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"volume_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *vmVolumeAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *vmVolumeAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan vmVolumeAttachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	vmName := plan.VMName.ValueString()
	volumeID := plan.VolumeID.ValueString()
	if _, err := r.client.AttachVolumeToVM(ctx, tenantID, vmName, volumeID); err != nil {
		resp.Diagnostics.AddError("Attach volume failed", err.Error())
		return
	}
	state := vmVolumeAttachmentModel{
		ID:       types.StringValue(attachmentID(vmName, volumeID)),
		VMName:   plan.VMName,
		VolumeID: plan.VolumeID,
	}
	if !plan.TenantID.IsNull() && plan.TenantID.ValueString() != "" {
		state.TenantID = plan.TenantID
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *vmVolumeAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var prev vmVolumeAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prev)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, prev.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	volumes, err := r.client.ListVMVolumes(ctx, tenantID, prev.VMName.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("List VM volumes failed", err.Error())
		return
	}
	volumeID := prev.VolumeID.ValueString()
	for _, v := range volumes {
		if v.ID == volumeID {
			out := vmVolumeAttachmentModel{
				ID:       types.StringValue(attachmentID(prev.VMName.ValueString(), volumeID)),
				VMName:   prev.VMName,
				VolumeID: prev.VolumeID,
			}
			if !prev.TenantID.IsNull() && prev.TenantID.ValueString() != "" {
				out.TenantID = prev.TenantID
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, out)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *vmVolumeAttachmentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Change vm_name or volume_id forces replacement. No in-place update is available.",
	)
}

func (r *vmVolumeAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state vmVolumeAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.DetachVolumeFromVM(ctx, tenantID, state.VMName.ValueString(), state.VolumeID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Detach volume failed", err.Error())
	}
}

func (r *vmVolumeAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Use `<vm_name>/<volume_id>`, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vm_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("volume_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func attachmentID(vmName, volumeID string) string {
	return vmName + "/" + volumeID
}
