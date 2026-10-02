package provider

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var (
	_ resource.Resource                = &vksClusterResource{}
	_ resource.ResourceWithImportState = &vksClusterResource{}
)

type vksClusterResource struct {
	client *virtfoundry.Client
}

type vksClusterModel struct {
	ID                   types.String `tfsdk:"id"`
	TenantID             types.String `tfsdk:"tenant_id"`
	Name                 types.String `tfsdk:"name"`
	KubernetesVersion    types.String `tfsdk:"kubernetes_version"`
	Workers              types.Object `tfsdk:"workers"`
	ControlPlane         types.Object `tfsdk:"control_plane"`
	Phase                types.String `tfsdk:"phase"`
	ControlPlaneEndpoint types.String `tfsdk:"control_plane_endpoint"`
	ReadyWorkers         types.Int64  `tfsdk:"ready_workers"`
	Namespace            types.String `tfsdk:"namespace"`
}

type vksWorkersModel struct {
	Count       types.Int64  `tfsdk:"count"`
	TemplateRef types.String `tfsdk:"template_ref"`
	OfferingRef types.String `tfsdk:"offering_ref"`
	NetworkRef  types.String `tfsdk:"network_ref"`
	SSHKeyRefs  types.Set    `tfsdk:"ssh_key_refs"`
}

type vksControlPlaneModel struct {
	ServiceType types.String `tfsdk:"service_type"`
	Address     types.String `tfsdk:"address"`
	Port        types.Int64  `tfsdk:"port"`
}

var vksWorkersAttrTypes = map[string]attr.Type{
	"count":        types.Int64Type,
	"template_ref": types.StringType,
	"offering_ref": types.StringType,
	"network_ref":  types.StringType,
	"ssh_key_refs": types.SetType{ElemType: types.StringType},
}

var vksControlPlaneAttrTypes = map[string]attr.Type{
	"service_type": types.StringType,
	"address":      types.StringType,
	"port":         types.Int64Type,
}

func NewVKSClusterResource() resource.Resource { return &vksClusterResource{} }

func (r *vksClusterResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_vks_cluster"
}

func (r *vksClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replaceString := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry Kubernetes Service (VKS) cluster: a Kamaji-hosted control plane plus worker Instances. " +
			"The API has no in-place update, so changing any argument replaces the cluster. " +
			"Create returns once the API accepts the cluster; use `phase` / `ready_workers` (or the `virtfoundry_vks_kubeconfig` data source) to track readiness.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster identifier (the cluster name; clusters are addressed by name within a tenant).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tenant_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Tenant UUID. Defaults to the provider `tenant_id`.",
				PlanModifiers:       replaceString,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cluster name (DNS label) within the tenant.",
				PlanModifiers:       replaceString,
			},
			"kubernetes_version": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Kubernetes version (for example `v1.31.4`). Must match a published node image version.",
				PlanModifiers:       replaceString,
			},
			"workers": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "Worker pool. Every worker is a VirtFoundry Instance created from the referenced template.",
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.RequiresReplace()},
				Attributes: map[string]schema.Attribute{
					"count": schema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "Number of worker nodes (at least 1).",
						PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
					},
					"template_ref": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Name of the node image VM template.",
					},
					"offering_ref": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Name of the service offering (for example `medium`).",
					},
					"network_ref": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Name of the network the workers attach to.",
					},
					"ssh_key_refs": schema.SetAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						MarkdownDescription: "Names of SSH keys injected into the workers.",
						PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
					},
				},
			},
			"control_plane": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Control plane exposure. Omitted fields use platform defaults.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"service_type": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Kubernetes Service type fronting the API server (for example `LoadBalancer`, `NodePort`, `ClusterIP`).",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"address": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Advertised API server address (IP or hostname).",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"port": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "API server port.",
						PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					},
				},
			},
			"phase": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster phase reported by the platform (for example `Provisioning`, `Ready`).",
			},
			"control_plane_endpoint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Kubernetes API server endpoint once known.",
			},
			"ready_workers": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of workers that have joined and are ready.",
			},
			"namespace": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Management-cluster namespace hosting the cluster resources.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *vksClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *vksClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan vksClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	in, diags := vksCreateInputFromPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.CreateVKSCluster(ctx, tenantID, in)
	if err != nil {
		resp.Diagnostics.AddError("Create VKS cluster failed", err.Error())
		return
	}
	state, diags := vksClusterToModel(ctx, cluster, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *vksClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state vksClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cluster, err := r.client.GetVKSCluster(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read VKS cluster failed", err.Error())
		return
	}
	next, diags := vksClusterToModel(ctx, cluster, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, next)...)
}

func (r *vksClusterResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "VKS clusters cannot be updated in place; every argument change replaces the cluster.")
}

func (r *vksClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state vksClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVKSCluster(ctx, tenantID, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete VKS cluster failed", err.Error())
	}
}

func (r *vksClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
	if resp.Diagnostics.HasError() {
		return
	}
	// Read resolves the cluster by id; seed name so a bare `<name>` import has a complete identity.
	name := req.ID
	if idx := strings.LastIndex(req.ID, "/"); idx >= 0 {
		name = req.ID[idx+1:]
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}

func vksCreateInputFromPlan(ctx context.Context, plan vksClusterModel) (virtfoundry.CreateVKSClusterInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	in := virtfoundry.CreateVKSClusterInput{
		Name:              plan.Name.ValueString(),
		KubernetesVersion: plan.KubernetesVersion.ValueString(),
	}

	var workers vksWorkersModel
	diags.Append(plan.Workers.As(ctx, &workers, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return in, diags
	}
	if workers.Count.ValueInt64() < 1 {
		diags.AddAttributeError(path.Root("workers").AtName("count"), "Invalid worker count", "`workers.count` must be at least 1.")
		return in, diags
	}
	in.Workers = virtfoundry.VKSWorkers{
		Count:       int32(workers.Count.ValueInt64()),
		TemplateRef: virtfoundry.VKSLocalObjectRef{Name: workers.TemplateRef.ValueString()},
		OfferingRef: virtfoundry.VKSLocalObjectRef{Name: workers.OfferingRef.ValueString()},
		NetworkRef:  virtfoundry.VKSLocalObjectRef{Name: workers.NetworkRef.ValueString()},
	}
	if !workers.SSHKeyRefs.IsNull() && !workers.SSHKeyRefs.IsUnknown() {
		var names []string
		diags.Append(workers.SSHKeyRefs.ElementsAs(ctx, &names, false)...)
		sort.Strings(names)
		for _, n := range names {
			in.Workers.SSHKeyRefs = append(in.Workers.SSHKeyRefs, virtfoundry.VKSLocalObjectRef{Name: n})
		}
	}

	if !plan.ControlPlane.IsNull() && !plan.ControlPlane.IsUnknown() {
		var cp vksControlPlaneModel
		diags.Append(plan.ControlPlane.As(ctx, &cp, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return in, diags
		}
		in.ControlPlane = virtfoundry.VKSControlPlane{
			ServiceType: knownString(cp.ServiceType),
			Address:     knownString(cp.Address),
		}
		if !cp.Port.IsNull() && !cp.Port.IsUnknown() {
			in.ControlPlane.Port = int32(cp.Port.ValueInt64())
		}
	}
	return in, diags
}

func knownString(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

// vksClusterToModel maps an API cluster to Terraform state. cfg (plan or prior
// state) supplies tenant_id and fills fields the API omits so state never
// contradicts what the practitioner configured.
func vksClusterToModel(ctx context.Context, c *virtfoundry.VKSCluster, cfg vksClusterModel) (vksClusterModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	var cfgWorkers vksWorkersModel
	if !cfg.Workers.IsNull() && !cfg.Workers.IsUnknown() {
		diags.Append(cfg.Workers.As(ctx, &cfgWorkers, basetypes.ObjectAsOptions{})...)
	}
	var cfgCP vksControlPlaneModel
	if !cfg.ControlPlane.IsNull() && !cfg.ControlPlane.IsUnknown() {
		diags.Append(cfg.ControlPlane.As(ctx, &cfgCP, basetypes.ObjectAsOptions{})...)
	}
	if diags.HasError() {
		return vksClusterModel{}, diags
	}

	// ssh_key_refs: null when the API has none, so an omitted attribute stays null.
	sshRefs := types.SetNull(types.StringType)
	if len(c.Workers.SSHKeyRefs) > 0 {
		elems := make([]attr.Value, len(c.Workers.SSHKeyRefs))
		for i, ref := range c.Workers.SSHKeyRefs {
			elems[i] = types.StringValue(ref.Name)
		}
		set, d := types.SetValue(types.StringType, elems)
		diags.Append(d...)
		sshRefs = set
	} else if !cfgWorkers.SSHKeyRefs.IsNull() && !cfgWorkers.SSHKeyRefs.IsUnknown() && len(cfgWorkers.SSHKeyRefs.Elements()) == 0 {
		sshRefs = cfgWorkers.SSHKeyRefs
	}
	workers, d := types.ObjectValue(vksWorkersAttrTypes, map[string]attr.Value{
		"count":        types.Int64Value(int64(c.Workers.Count)),
		"template_ref": types.StringValue(c.Workers.TemplateRef.Name),
		"offering_ref": types.StringValue(c.Workers.OfferingRef.Name),
		"network_ref":  types.StringValue(c.Workers.NetworkRef.Name),
		"ssh_key_refs": sshRefs,
	})
	diags.Append(d...)

	controlPlane, d := types.ObjectValue(vksControlPlaneAttrTypes, map[string]attr.Value{
		"service_type": stringOrFallback(c.ControlPlane.ServiceType, cfgCP.ServiceType),
		"address":      stringOrFallback(c.ControlPlane.Address, cfgCP.Address),
		"port":         portOrFallback(c.ControlPlane.Port, cfgCP.Port),
	})
	diags.Append(d...)

	out := vksClusterModel{
		ID:                   types.StringValue(c.Name),
		Name:                 types.StringValue(c.Name),
		KubernetesVersion:    types.StringValue(c.KubernetesVersion),
		Workers:              workers,
		ControlPlane:         controlPlane,
		Phase:                types.StringValue(c.Phase),
		ControlPlaneEndpoint: types.StringValue(c.ControlPlaneEndpoint),
		ReadyWorkers:         types.Int64Value(int64(c.ReadyWorkers)),
		Namespace:            types.StringValue(c.Namespace),
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	return out, diags
}

// stringOrFallback prefers the API value; when the API omits it, falls back to
// the configured value, and finally to null (known, empty).
func stringOrFallback(api string, cfg types.String) types.String {
	if api != "" {
		return types.StringValue(api)
	}
	if !cfg.IsNull() && !cfg.IsUnknown() {
		return cfg
	}
	return types.StringNull()
}

func portOrFallback(api int32, cfg types.Int64) types.Int64 {
	if api != 0 {
		return types.Int64Value(int64(api))
	}
	if !cfg.IsNull() && !cfg.IsUnknown() {
		return cfg
	}
	return types.Int64Null()
}
