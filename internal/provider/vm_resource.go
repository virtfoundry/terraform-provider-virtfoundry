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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	SSHKeyID          types.String `tfsdk:"ssh_key_id"`
	DataVolumeID      types.String `tfsdk:"data_volume_id"`
	ExposeSSH         types.Bool   `tfsdk:"expose_ssh"`
	SSHNodePort       types.Int64  `tfsdk:"ssh_node_port"`
	SSHExposed        types.Bool   `tfsdk:"ssh_exposed"`
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
			"ssh_key_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "SSH key UUID injected into cloud-init for this VM.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_volume_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional data volume UUID to attach at deploy time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expose_ssh": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Expose VM SSH via a Kubernetes NodePort Service.",
			},
			"ssh_node_port": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Desired NodePort for SSH exposure (auto-assigned when omitted).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"ssh_exposed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether SSH is currently exposed via NodePort.",
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
	r.client = configureClient(req, resp)
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
	in, diags := deployInputFromPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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

	sshInfo, diags := r.readSSHInfo(ctx, tenantID, vm.Name, plan.ExposeSSH)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, vmToModel(vm, plan, sshInfo))...)
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
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read VM failed", err.Error())
		return
	}

	sshInfo, _ := r.client.GetVMSSH(ctx, tenantID, state.Name.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, vmToModel(vm, state, sshInfo))...)
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

	sshInfo, diags := r.ensureSSH(ctx, tenantID, name, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, vmToModel(vm, plan, sshInfo))...)
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
		if isNotFound(err) {
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

func deployInputFromPlan(ctx context.Context, plan vmModel) (virtfoundry.DeployVMInput, diag.Diagnostics) {
	var diags diag.Diagnostics
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
		diags.Append(plan.NetworkIDs.ElementsAs(ctx, &ids, false)...)
		in.NetworkIDs = ids
	}
	if !plan.SecurityGroupIDs.IsNull() {
		var ids []string
		diags.Append(plan.SecurityGroupIDs.ElementsAs(ctx, &ids, false)...)
		in.SecurityGroupIDs = ids
	}
	if !plan.SSHKeyID.IsNull() {
		in.SSHKeyID = plan.SSHKeyID.ValueString()
	}
	if !plan.DataVolumeID.IsNull() {
		in.DataVolumeID = plan.DataVolumeID.ValueString()
	}
	if !plan.ExposeSSH.IsNull() {
		in.ExposeSSH = plan.ExposeSSH.ValueBool()
	}
	return in, diags
}

func (r *vmResource) readSSHInfo(ctx context.Context, tenantID, vmName string, expose types.Bool) (*virtfoundry.VMSSHInfo, diag.Diagnostics) {
	var diags diag.Diagnostics
	wantExpose := !expose.IsNull() && expose.ValueBool()
	deadline := time.Now().Add(2 * time.Minute)
	for {
		info, err := r.client.GetVMSSH(ctx, tenantID, vmName)
		if err != nil {
			if isNotFound(err) {
				return nil, diags
			}
			diags.AddError("Read VM SSH failed", err.Error())
			return nil, diags
		}
		if !wantExpose || info.Exposed || info.NodePort > 0 {
			return info, diags
		}
		if time.Now().After(deadline) {
			return info, diags
		}
		select {
		case <-ctx.Done():
			diags.AddError("Read VM SSH failed", ctx.Err().Error())
			return nil, diags
		case <-time.After(5 * time.Second):
		}
	}
}

func (r *vmResource) ensureSSH(ctx context.Context, tenantID, vmName string, plan, state vmModel) (*virtfoundry.VMSSHInfo, diag.Diagnostics) {
	var diags diag.Diagnostics
	wantExpose := !plan.ExposeSSH.IsNull() && plan.ExposeSSH.ValueBool()
	hadExpose := !state.ExposeSSH.IsNull() && state.ExposeSSH.ValueBool()
	if wantExpose && !hadExpose {
		var port int32
		if !plan.SSHNodePort.IsNull() {
			port = int32(plan.SSHNodePort.ValueInt64())
		}
		deadline := time.Now().Add(3 * time.Minute)
		for {
			_, err := r.client.ExposeVMSSH(ctx, tenantID, vmName, port)
			if err == nil {
				break
			}
			if strings.Contains(err.Error(), "no IP") && time.Now().Before(deadline) {
				select {
				case <-ctx.Done():
					diags.AddError("Expose VM SSH failed", ctx.Err().Error())
					return nil, diags
				case <-time.After(5 * time.Second):
					continue
				}
			}
			diags.AddError("Expose VM SSH failed", err.Error())
			return nil, diags
		}
	}
	return r.readSSHInfo(ctx, tenantID, vmName, plan.ExposeSSH)
}

func vmToModel(vm *virtfoundry.VM, cfg vmModel, ssh *virtfoundry.VMSSHInfo) vmModel {
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
		SSHKeyID:          cfg.SSHKeyID,
		DataVolumeID:      cfg.DataVolumeID,
		ExposeSSH:         cfg.ExposeSSH,
		SSHNodePort:       cfg.SSHNodePort,
		DesiredState:      types.StringValue(stringValue(cfg.DesiredState, "running")),
	}
	if ssh != nil {
		out.SSHExposed = types.BoolValue(ssh.Exposed)
		if ssh.NodePort > 0 {
			out.SSHNodePort = types.Int64Value(int64(ssh.NodePort))
		}
	} else if !cfg.SSHExposed.IsNull() {
		out.SSHExposed = cfg.SSHExposed
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
