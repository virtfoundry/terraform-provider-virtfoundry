package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var (
	_ resource.Resource                = &vmTemplateResource{}
	_ resource.ResourceWithImportState = &vmTemplateResource{}
)

type vmTemplateResource struct {
	client *virtfoundry.Client
}

type vmTemplateModel struct {
	ID                types.String `tfsdk:"id"`
	TenantID          types.String `tfsdk:"tenant_id"`
	Name              types.String `tfsdk:"name"`
	DisplayName       types.String `tfsdk:"display_name"`
	Description       types.String `tfsdk:"description"`
	Image             types.String `tfsdk:"image"`
	SourceType        types.String `tfsdk:"source_type"`
	OSType            types.String `tfsdk:"os_type"`
	CloudInitUserData types.String `tfsdk:"cloud_init_user_data"`
	ISOVolumeID       types.String `tfsdk:"iso_volume_id"`
	ISOSizeGi         types.Int64  `tfsdk:"iso_size_gi"`
	BootDiskSizeGi    types.Int64  `tfsdk:"boot_disk_size_gi"`
	StorageClass      types.String `tfsdk:"storage_class"`
	WaitForImport     types.Bool   `tfsdk:"wait_for_import"`
	ImportWaitMinutes types.Int64  `tfsdk:"import_wait_timeout_minutes"`
	State             types.String `tfsdk:"state"`
	ImportState       types.String `tfsdk:"import_state"`
	Hypervisor        types.String `tfsdk:"hypervisor"`
}

func NewVMTemplateResource() resource.Resource { return &vmTemplateResource{} }

func (r *vmTemplateResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_vm_template"
}

func (r *vmTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry VM template (container disk or ISO).",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tenant_id":    schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":         schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"display_name": schema.StringAttribute{Optional: true},
			"description":  schema.StringAttribute{Optional: true},
			"image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Container image reference or ISO source.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"source_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "`container` or `iso`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"os_type":              schema.StringAttribute{Optional: true},
			"cloud_init_user_data": schema.StringAttribute{Optional: true},
			"iso_volume_id":        schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"iso_size_gi":          schema.Int64Attribute{Optional: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
			"boot_disk_size_gi":    schema.Int64Attribute{Optional: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
			"storage_class":        schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"wait_for_import": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "When `source_type` is `iso`, block until CDI import reaches `ready` (default: true).",
			},
			"import_wait_timeout_minutes": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Max minutes to wait for ISO import (default: 45).",
			},
			"state":        schema.StringAttribute{Computed: true},
			"import_state": schema.StringAttribute{Computed: true},
			"hypervisor":   schema.StringAttribute{Computed: true},
		},
	}
}

func (r *vmTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *vmTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan vmTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := templateInputFromPlan(plan)
	tmpl, err := r.client.CreateVMTemplate(ctx, tenantID, in)
	if err != nil {
		resp.Diagnostics.AddError("Create VM template failed", err.Error())
		return
	}
	if templateIsISO(plan) && templateWaitForImport(plan) {
		timeout := templateImportTimeout(plan)
		tmpl, err = r.client.WaitForVMTemplateImport(ctx, tenantID, tmpl, timeout)
		if err != nil {
			resp.Diagnostics.AddError("Wait for ISO import failed", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, vmTemplateToModel(tmpl, plan))...)
}

func (r *vmTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state vmTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tmpl, err := r.client.GetVMTemplate(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read VM template failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, vmTemplateToModel(tmpl, state))...)
}

func (r *vmTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan vmTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := virtfoundry.UpdateVMTemplateInput{}
	if !plan.DisplayName.IsNull() {
		in.DisplayName = plan.DisplayName.ValueString()
	}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	if !plan.Image.IsNull() {
		in.Image = plan.Image.ValueString()
	}
	if !plan.SourceType.IsNull() {
		in.SourceType = plan.SourceType.ValueString()
	}
	if !plan.OSType.IsNull() {
		in.OSType = plan.OSType.ValueString()
	}
	if !plan.CloudInitUserData.IsNull() {
		in.CloudInitUserData = plan.CloudInitUserData.ValueString()
	}
	tmpl, err := r.client.UpdateVMTemplate(ctx, tenantID, plan.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Update VM template failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, vmTemplateToModel(tmpl, plan))...)
}

func (r *vmTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state vmTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVMTemplate(ctx, tenantID, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete VM template failed", err.Error())
	}
}

func (r *vmTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
}

func templateInputFromPlan(plan vmTemplateModel) virtfoundry.CreateVMTemplateInput {
	in := virtfoundry.CreateVMTemplateInput{
		Name:  plan.Name.ValueString(),
		Image: plan.Image.ValueString(),
	}
	if !plan.DisplayName.IsNull() {
		in.DisplayName = plan.DisplayName.ValueString()
	}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	if !plan.SourceType.IsNull() {
		in.SourceType = plan.SourceType.ValueString()
	}
	if !plan.OSType.IsNull() {
		in.OSType = plan.OSType.ValueString()
	}
	if !plan.CloudInitUserData.IsNull() {
		in.CloudInitUserData = plan.CloudInitUserData.ValueString()
	}
	if !plan.ISOVolumeID.IsNull() {
		in.ISOVolumeID = plan.ISOVolumeID.ValueString()
	}
	if !plan.ISOSizeGi.IsNull() {
		in.ISOSizeGi = int(plan.ISOSizeGi.ValueInt64())
	}
	if !plan.BootDiskSizeGi.IsNull() {
		in.BootDiskSizeGi = int(plan.BootDiskSizeGi.ValueInt64())
	}
	if !plan.StorageClass.IsNull() {
		in.StorageClass = plan.StorageClass.ValueString()
	}
	return in
}

func vmTemplateToModel(t *virtfoundry.VMTemplate, cfg vmTemplateModel) vmTemplateModel {
	out := vmTemplateModel{
		ID:         types.StringValue(t.ID),
		Name:       types.StringValue(t.Name),
		Image:      types.StringValue(t.Image),
		State:      types.StringValue(t.State),
		Hypervisor: types.StringValue(t.Hypervisor),
	}
	if t.DisplayName != "" {
		out.DisplayName = types.StringValue(t.DisplayName)
	} else if !cfg.DisplayName.IsNull() {
		out.DisplayName = cfg.DisplayName
	}
	if t.Description != "" {
		out.Description = types.StringValue(t.Description)
	} else if !cfg.Description.IsNull() {
		out.Description = cfg.Description
	}
	if t.SourceType != "" {
		out.SourceType = types.StringValue(t.SourceType)
	} else if !cfg.SourceType.IsNull() {
		out.SourceType = cfg.SourceType
	}
	if t.OSType != "" {
		out.OSType = types.StringValue(t.OSType)
	} else if !cfg.OSType.IsNull() {
		out.OSType = cfg.OSType
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	if !cfg.CloudInitUserData.IsNull() {
		out.CloudInitUserData = cfg.CloudInitUserData
	}
	if !cfg.ISOVolumeID.IsNull() {
		out.ISOVolumeID = cfg.ISOVolumeID
	}
	if !cfg.ISOSizeGi.IsNull() {
		out.ISOSizeGi = cfg.ISOSizeGi
	}
	if !cfg.BootDiskSizeGi.IsNull() {
		out.BootDiskSizeGi = cfg.BootDiskSizeGi
	}
	if !cfg.StorageClass.IsNull() {
		out.StorageClass = cfg.StorageClass
	}
	if !cfg.WaitForImport.IsNull() {
		out.WaitForImport = cfg.WaitForImport
	}
	if !cfg.ImportWaitMinutes.IsNull() {
		out.ImportWaitMinutes = cfg.ImportWaitMinutes
	}
	if t.ImportState != "" {
		out.ImportState = types.StringValue(t.ImportState)
	}
	return out
}

func templateIsISO(plan vmTemplateModel) bool {
	return strings.EqualFold(stringValue(plan.SourceType, ""), "iso")
}

func templateWaitForImport(plan vmTemplateModel) bool {
	if plan.WaitForImport.IsNull() {
		return true
	}
	return plan.WaitForImport.ValueBool()
}

func templateImportTimeout(plan vmTemplateModel) time.Duration {
	if !plan.ImportWaitMinutes.IsNull() && plan.ImportWaitMinutes.ValueInt64() > 0 {
		return time.Duration(plan.ImportWaitMinutes.ValueInt64()) * time.Minute
	}
	return 45 * time.Minute
}
