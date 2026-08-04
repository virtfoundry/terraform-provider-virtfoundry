package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var (
	_ resource.Resource                = &securityGroupResource{}
	_ resource.ResourceWithImportState = &securityGroupResource{}
)

type securityGroupResource struct {
	client *virtfoundry.Client
}

type securityGroupModel struct {
	ID          types.String `tfsdk:"id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	VPCID       types.String `tfsdk:"vpc_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Rules       types.List   `tfsdk:"rule"`
}

type securityGroupRuleModel struct {
	Direction types.String `tfsdk:"direction"`
	Protocol  types.String `tfsdk:"protocol"`
	PortFrom  types.Int64  `tfsdk:"port_from"`
	PortTo    types.Int64  `tfsdk:"port_to"`
	CIDR      types.String `tfsdk:"cidr"`
}

var sgRuleAttrTypes = map[string]attr.Type{
	"direction": types.StringType,
	"protocol":  types.StringType,
	"port_from": types.Int64Type,
	"port_to":   types.Int64Type,
	"cidr":      types.StringType,
}

func NewSecurityGroupResource() resource.Resource { return &securityGroupResource{} }

func (r *securityGroupResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_security_group"
}

func (r *securityGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry security group with ingress/egress rules.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tenant_id":   schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"vpc_id":      schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":        schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{Optional: true},
		},
		Blocks: map[string]schema.Block{
			"rule": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"direction": schema.StringAttribute{Required: true, MarkdownDescription: "`ingress` or `egress`."},
						"protocol":  schema.StringAttribute{Required: true, MarkdownDescription: "`tcp`, `udp`, or `icmp`."},
						"port_from": schema.Int64Attribute{Optional: true},
						"port_to":   schema.Int64Attribute{Optional: true},
						"cidr":      schema.StringAttribute{Required: true},
					},
				},
			},
		},
	}
}

func (r *securityGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *securityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan securityGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	rules, diags := sgRulesFromModel(ctx, plan.Rules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := virtfoundry.CreateSecurityGroupInput{
		Name:  plan.Name.ValueString(),
		Rules: rules,
	}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	if !plan.VPCID.IsNull() {
		in.VPCID = plan.VPCID.ValueString()
	}
	sg, err := r.client.CreateSecurityGroup(ctx, tenantID, in)
	if err != nil {
		resp.Diagnostics.AddError("Create security group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, sgToModel(sg, plan))...)
}

func (r *securityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state securityGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	sg, err := r.client.GetSecurityGroup(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read security group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, sgToModel(sg, state))...)
}

func (r *securityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan securityGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	rules, diags := sgRulesFromModel(ctx, plan.Rules)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := virtfoundry.CreateSecurityGroupInput{
		Name:  plan.Name.ValueString(),
		Rules: rules,
	}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	sg, err := r.client.UpdateSecurityGroup(ctx, tenantID, plan.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Update security group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, sgToModel(sg, plan))...)
}

func (r *securityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state securityGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSecurityGroup(ctx, tenantID, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete security group failed", err.Error())
	}
}

func (r *securityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
}

func sgRulesFromModel(ctx context.Context, list types.List) ([]virtfoundry.SecurityGroupRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}
	var models []securityGroupRuleModel
	diags.Append(list.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil, diags
	}
	rules := make([]virtfoundry.SecurityGroupRule, len(models))
	for i, m := range models {
		r := virtfoundry.SecurityGroupRule{
			Direction: m.Direction.ValueString(),
			Protocol:  m.Protocol.ValueString(),
			CIDR:      m.CIDR.ValueString(),
		}
		if !m.PortFrom.IsNull() {
			r.PortFrom = int(m.PortFrom.ValueInt64())
		}
		if !m.PortTo.IsNull() {
			r.PortTo = int(m.PortTo.ValueInt64())
		}
		rules[i] = r
	}
	return rules, diags
}

func sgToModel(sg *virtfoundry.SecurityGroup, cfg securityGroupModel) securityGroupModel {
	out := securityGroupModel{
		ID:   types.StringValue(sg.ID),
		Name: types.StringValue(sg.Name),
	}
	if sg.VPCID != "" {
		out.VPCID = types.StringValue(sg.VPCID)
	} else if !cfg.VPCID.IsNull() {
		out.VPCID = cfg.VPCID
	}
	if sg.Description != "" {
		out.Description = types.StringValue(sg.Description)
	} else if !cfg.Description.IsNull() {
		out.Description = cfg.Description
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	elems := make([]attr.Value, len(sg.Rules))
	for i, rule := range sg.Rules {
		obj, diags := types.ObjectValue(sgRuleAttrTypes, map[string]attr.Value{
			"direction": types.StringValue(rule.Direction),
			"protocol":  types.StringValue(rule.Protocol),
			"port_from": types.Int64Value(int64(rule.PortFrom)),
			"port_to":   types.Int64Value(int64(rule.PortTo)),
			"cidr":      types.StringValue(rule.CIDR),
		})
		if diags.HasError() {
			continue
		}
		elems[i] = obj
	}
	out.Rules, _ = types.ListValue(types.ObjectType{AttrTypes: sgRuleAttrTypes}, elems)
	return out
}
