package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var (
	_ resource.Resource                = &vmResource{}
	_ resource.ResourceWithImportState = &vmResource{}
)

type vmResource struct {
	client *virtfoundry.Client
}

type vmModel struct {
	ID                types.String `tfsdk:"id"`
	TenantID          types.String `tfsdk:"tenant_id"`
	Name              types.String `tfsdk:"name"`
	DisplayName       types.String `tfsdk:"display_name"`
	TemplateID        types.String `tfsdk:"template_id"`
	ServiceOfferingID types.String `tfsdk:"service_offering_id"`
	PublicIP          types.Bool   `tfsdk:"public_ip"`
	NetworkIDs        types.List   `tfsdk:"network_ids"`
	SecurityGroupIDs  types.List   `tfsdk:"security_group_ids"`
	DesiredState      types.String `tfsdk:"desired_state"`
	State             types.String `tfsdk:"state"`
	IP                types.String `tfsdk:"ip"`
	CPU               types.Int64  `tfsdk:"cpu"`
	MemoryMi          types.Int64  `tfsdk:"memory_mi"`
}

func NewVMResource() resource.Resource {
	return &vmResource{}
}

func (r *vmResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_vm"
}

func (r *vmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry virtual machine (KubeVirt).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "VirtFoundry VM UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Tenant UUID. Defaults to the provider `tenant_id`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "VM name (slug) within the tenant namespace.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Human-readable display name.",
			},
			"template_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "VM template UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_offering_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Service offering UUID or name (e.g. `small`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_ip": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Attach the shared public network (requires `security_group_ids`).",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"network_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Private or shared network UUIDs to attach.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"security_group_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Security group UUIDs (required when `public_ip` is true).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"desired_state": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Target power state: `running` or `stopped`. Default: `running`.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current API state (Running, Stopped, Starting, ...).",
			},
			"ip": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Primary IP address when assigned.",
			},
			"cpu": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "vCPU count.",
			},
			"memory_mi": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Memory in MiB.",
			},
		},
	}
}

func (r *vmResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*virtfoundry.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data",
			fmt.Sprintf("Expected *virtfoundry.Client, got %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *vmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "VirtFoundry client is nil")
		return
	}

	var plan vmModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	desired := strings.ToLower(stringValue(plan.DesiredState, "running"))
	in := virtfoundry.DeployVMInput{Name: plan.Name.ValueString()}
	if !plan.DisplayName.IsNull() {
		in.DisplayName = plan.DisplayName.ValueString()
	}
	if !plan.TemplateID.IsNull() {
		in.TemplateID = plan.TemplateID.ValueString()
	}
	if !plan.ServiceOfferingID.IsNull() {
		in.ServiceOfferingID = plan.ServiceOfferingID.ValueString()
	}
	if !plan.PublicIP.IsNull() {
		in.PublicIP = plan.PublicIP.ValueBool()
	}
	if !plan.NetworkIDs.IsNull() {
		var ids []string
		resp.Diagnostics.Append(plan.NetworkIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.NetworkIDs = ids
	}
	if !plan.SecurityGroupIDs.IsNull() {
		var ids []string
		resp.Diagnostics.Append(plan.SecurityGroupIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.SecurityGroupIDs = ids
	}

	vm, err := r.client.DeployVM(ctx, tenantID, in)
	if err != nil {
		resp.Diagnostics.AddError("Create VM failed", err.Error())
		return
	}

	vm, err = r.client.WaitForVMState(ctx, tenantID, vm.Name, desired, 15*time.Minute)
	if err != nil {
		resp.Diagnostics.AddError("Wait for VM state failed", err.Error())
		return
	}

	if desired == "stopped" && !virtfoundry.StateMatches(vm.State, "stopped") {
		vm, err = r.client.StopVM(ctx, tenantID, vm.Name)
		if err != nil {
			resp.Diagnostics.AddError("Stop VM after create failed", err.Error())
			return
		}
		vm, err = r.client.WaitForVMState(ctx, tenantID, vm.Name, "stopped", 5*time.Minute)
		if err != nil {
			resp.Diagnostics.AddError("Wait for stopped state failed", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, vmToModel(vm, plan))...)
}

func (r *vmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "VirtFoundry client is nil")
		return
	}

	var state vmModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.client.GetVM(ctx, tenantID, state.Name.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read VM failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, vmToModel(vm, state))...)
}

func (r *vmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "VirtFoundry client is nil")
		return
	}

	var plan, state vmModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	var vm *virtfoundry.VM
	var err error

	if !plan.DisplayName.Equal(state.DisplayName) {
		display := ""
		if !plan.DisplayName.IsNull() {
			display = plan.DisplayName.ValueString()
		}
		vm, err = r.client.UpdateVM(ctx, tenantID, name, display, 0, 0)
		if err != nil {
			resp.Diagnostics.AddError("Update VM failed", err.Error())
			return
		}
	}

	desired := strings.ToLower(stringValue(plan.DesiredState, "running"))
	current := vm
	if current == nil {
		current, err = r.client.GetVM(ctx, tenantID, name)
		if err != nil {
			resp.Diagnostics.AddError("Read VM failed", err.Error())
			return
		}
	}

	if !virtfoundry.StateMatches(current.State, desired) {
		switch desired {
		case "running":
			vm, err = r.client.StartVM(ctx, tenantID, name)
		case "stopped":
			vm, err = r.client.StopVM(ctx, tenantID, name)
		default:
			err = fmt.Errorf("unsupported desired_state %q", desired)
		}
		if err != nil {
			resp.Diagnostics.AddError("Change VM power state failed", err.Error())
			return
		}
		vm, err = r.client.WaitForVMState(ctx, tenantID, name, desired, 5*time.Minute)
		if err != nil {
			resp.Diagnostics.AddError("Wait for VM state failed", err.Error())
			return
		}
	} else if vm == nil {
		vm = current
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, vmToModel(vm, plan))...)
}

func (r *vmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "VirtFoundry client is nil")
		return
	}

	var state vmModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteVM(ctx, tenantID, state.Name.ValueString()); err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			return
		}
		resp.Diagnostics.AddError("Delete VM failed", err.Error())
	}
}

func (r *vmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	switch len(parts) {
	case 2:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
	case 1:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[0])...)
	default:
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Use `<tenant_id>/<name>` or `<name>` when provider tenant_id is set.",
		)
	}
}

func resolveTenantID(client *virtfoundry.Client, attr types.String) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !attr.IsNull() && attr.ValueString() != "" {
		return attr.ValueString(), diags
	}
	if tid := client.TenantID(); tid != "" {
		return tid, diags
	}
	diags.AddError("Missing tenant_id", "Set tenant_id on the resource or provider block.")
	return "", diags
}

func vmToModel(vm *virtfoundry.VM, cfg vmModel) vmModel {
	out := vmModel{
		ID:                types.StringValue(vm.ID),
		Name:              types.StringValue(vm.Name),
		State:             types.StringValue(vm.State),
		CPU:               types.Int64Value(int64(vm.CPU)),
		MemoryMi:          types.Int64Value(vm.MemoryMi),
		TemplateID:        cfg.TemplateID,
		ServiceOfferingID: cfg.ServiceOfferingID,
		PublicIP:          cfg.PublicIP,
		NetworkIDs:        cfg.NetworkIDs,
		SecurityGroupIDs:  cfg.SecurityGroupIDs,
		DesiredState:      types.StringValue(stringValue(cfg.DesiredState, "running")),
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	if !cfg.DisplayName.IsNull() && cfg.DisplayName.ValueString() != "" {
		out.DisplayName = cfg.DisplayName
	} else if vm.DisplayName != "" {
		out.DisplayName = types.StringValue(vm.DisplayName)
	}
	if vm.IP != "" {
		out.IP = types.StringValue(vm.IP)
	}
	return out
}

func stringValue(v types.String, fallback string) string {
	if !v.IsNull() && v.ValueString() != "" {
		return v.ValueString()
	}
	return fallback
}
