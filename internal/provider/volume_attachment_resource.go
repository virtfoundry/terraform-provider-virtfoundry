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
	_ resource.Resource                = &volumeAttachmentResource{}
	_ resource.ResourceWithImportState = &volumeAttachmentResource{}
)

type volumeAttachmentResource struct {
	client *virtfoundry.Client
}

type volumeAttachmentModel struct {
	ID       types.String `tfsdk:"id"`
	TenantID types.String `tfsdk:"tenant_id"`
	VMName   types.String `tfsdk:"vm_name"`
	VolumeID types.String `tfsdk:"volume_id"`
}

func NewVolumeAttachmentResource() resource.Resource { return &volumeAttachmentResource{} }

func (r *volumeAttachmentResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_volume_attachment"
}

func (r *volumeAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches a data volume to a VM (hot-plug). Do not use the same volume as `virtfoundry_vm.data_volume_id`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Equals `volume_id` (a volume attaches to at most one VM).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tenant_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Tenant UUID. Defaults to the provider `tenant_id`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"vm_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "VM name (slug) to attach to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"volume_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Volume UUID to attach.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *volumeAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *volumeAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan volumeAttachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	vol, err := r.client.AttachVolume(ctx, tenantID, plan.VMName.ValueString(), plan.VolumeID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Attach volume failed", err.Error())
		return
	}
	plan.ID = types.StringValue(vol.ID)
	if plan.ID.ValueString() == "" {
		plan.ID = plan.VolumeID
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *volumeAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state volumeAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	vol, err := r.client.GetVolume(ctx, tenantID, state.VolumeID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read volume attachment failed", err.Error())
		return
	}
	if vol.VMID == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	if state.VMName.IsNull() || state.VMName.ValueString() == "" {
		vms, listErr := r.client.ListVMs(ctx, tenantID)
		if listErr != nil {
			resp.Diagnostics.AddError("List VMs failed", listErr.Error())
			return
		}
		for i := range vms {
			if vms[i].ID == vol.VMID {
				state.VMName = types.StringValue(vms[i].Name)
				break
			}
		}
	}
	state.ID = types.StringValue(vol.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *volumeAttachmentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Volume attachments are force-new; change vm_name or volume_id to recreate.")
}

func (r *volumeAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state volumeAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DetachVolume(ctx, tenantID, state.VMName.ValueString(), state.VolumeID.ValueString())
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Detach volume failed", err.Error())
	}
}

func (r *volumeAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
	parts := strings.Split(req.ID, "/")
	volID := parts[len(parts)-1]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("volume_id"), volID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), volID)...)
}
