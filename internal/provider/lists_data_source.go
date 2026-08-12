package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

type tenantListDataSource struct {
	client *virtfoundry.Client
}

type tenantListModel struct {
	TenantID types.String `tfsdk:"tenant_id"`
	Items    types.List   `tfsdk:"items"`
}

type tenantIDModel struct {
	TenantID types.String `tfsdk:"tenant_id"`
}

func configureDataSource(client **virtfoundry.Client, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*virtfoundry.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *virtfoundry.Client, got %T", req.ProviderData))
		return
	}
	*client = c
}

func NewVPCsDataSource() datasource.DataSource { return &vpcsDataSource{} }

type vpcsDataSource struct{ client *virtfoundry.Client }

func (d *vpcsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "virtfoundry_vpcs"
}

func (d *vpcsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists VPCs in a tenant.",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{Optional: true},
			"vpcs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true},
					"cidr": schema.StringAttribute{Computed: true}, "state": schema.StringAttribute{Computed: true},
				}},
			},
		},
	}
}

func (d *vpcsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(&d.client, req, resp)
}

func (d *vpcsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "client nil")
		return
	}
	var cfg tenantIDModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	tenantID, diags := resolveTenantID(d.client, cfg.TenantID)
	resp.Diagnostics.Append(diags...)
	items, err := d.client.ListVPCs(ctx, tenantID)
	if err != nil {
		resp.Diagnostics.AddError("List VPCs failed", err.Error())
		return
	}
	attrTypes := map[string]attr.Type{"id": types.StringType, "name": types.StringType, "cidr": types.StringType, "state": types.StringType}
	elems := make([]attr.Value, len(items))
	for i, v := range items {
		obj, objDiags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id": types.StringValue(v.ID), "name": types.StringValue(v.Name),
			"cidr": types.StringValue(v.CIDR), "state": types.StringValue(v.State),
		})
		resp.Diagnostics.Append(objDiags...)
		elems[i] = obj
	}
	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elems)
	resp.Diagnostics.Append(listDiags...)
	out := tenantListModel{Items: list}
	if !cfg.TenantID.IsNull() {
		out.TenantID = cfg.TenantID
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &struct {
		TenantID types.String `tfsdk:"tenant_id"`
		VPCs     types.List   `tfsdk:"vpcs"`
	}{TenantID: out.TenantID, VPCs: list})...)
}

func NewNetworksDataSource() datasource.DataSource { return &networksDataSource{} }

type networksDataSource struct{ client *virtfoundry.Client }

func (d *networksDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "virtfoundry_networks"
}

func (d *networksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists networks (subnets) in a tenant.",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{Optional: true},
			"networks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true},
					"vpc_id": schema.StringAttribute{Computed: true}, "cidr": schema.StringAttribute{Computed: true},
					"state": schema.StringAttribute{Computed: true},
				}},
			},
		},
	}
}

func (d *networksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(&d.client, req, resp)
}

func (d *networksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		return
	}
	var cfg tenantIDModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	tenantID, diags := resolveTenantID(d.client, cfg.TenantID)
	resp.Diagnostics.Append(diags...)
	items, err := d.client.ListNetworks(ctx, tenantID)
	if err != nil {
		resp.Diagnostics.AddError("List networks failed", err.Error())
		return
	}
	attrTypes := map[string]attr.Type{"id": types.StringType, "name": types.StringType, "vpc_id": types.StringType, "cidr": types.StringType, "state": types.StringType}
	elems := make([]attr.Value, len(items))
	for i, n := range items {
		obj, objDiags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id": types.StringValue(n.ID), "name": types.StringValue(n.Name),
			"vpc_id": types.StringValue(n.VPCID), "cidr": types.StringValue(n.CIDR), "state": types.StringValue(n.State),
		})
		resp.Diagnostics.Append(objDiags...)
		elems[i] = obj
	}
	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elems)
	resp.Diagnostics.Append(listDiags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &struct {
		TenantID types.String `tfsdk:"tenant_id"`
		Networks types.List   `tfsdk:"networks"`
	}{TenantID: cfg.TenantID, Networks: list})...)
}

func NewSecurityGroupsDataSource() datasource.DataSource { return &securityGroupsDataSource{} }

type securityGroupsDataSource struct{ client *virtfoundry.Client }

type securityGroupsDataSourceModel struct {
	TenantID       types.String `tfsdk:"tenant_id"`
	SecurityGroups types.List   `tfsdk:"security_groups"`
}

func (d *securityGroupsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "virtfoundry_security_groups"
}

func (d *securityGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists security groups in a tenant.",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{Optional: true},
			"security_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true},
					"vpc_id": schema.StringAttribute{Computed: true}, "description": schema.StringAttribute{Computed: true},
				}},
			},
		},
	}
}

func (d *securityGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(&d.client, req, resp)
}

func (d *securityGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		return
	}
	var state securityGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	tenantID, diags := resolveTenantID(d.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	items, err := d.client.ListSecurityGroups(ctx, tenantID)
	if err != nil {
		resp.Diagnostics.AddError("List security groups failed", err.Error())
		return
	}
	attrTypes := map[string]attr.Type{"id": types.StringType, "name": types.StringType, "vpc_id": types.StringType, "description": types.StringType}
	elems := make([]attr.Value, len(items))
	for i, sg := range items {
		obj, objDiags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id": types.StringValue(sg.ID), "name": types.StringValue(sg.Name),
			"vpc_id": types.StringValue(sg.VPCID), "description": types.StringValue(sg.Description),
		})
		resp.Diagnostics.Append(objDiags...)
		elems[i] = obj
	}
	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elems)
	resp.Diagnostics.Append(listDiags...)
	state.SecurityGroups = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func NewSSHKeysDataSource() datasource.DataSource { return &sshKeysDataSource{} }

type sshKeysDataSource struct{ client *virtfoundry.Client }

func (d *sshKeysDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "virtfoundry_ssh_keys"
}

func (d *sshKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists SSH keys registered in a tenant.",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{Optional: true},
			"ssh_keys": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true},
					"public_key": schema.StringAttribute{Computed: true}, "fingerprint": schema.StringAttribute{Computed: true},
				}},
			},
		},
	}
}

func (d *sshKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(&d.client, req, resp)
}

func (d *sshKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		return
	}
	var cfg tenantIDModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	tenantID, diags := resolveTenantID(d.client, cfg.TenantID)
	resp.Diagnostics.Append(diags...)
	items, err := d.client.ListSSHKeys(ctx, tenantID)
	if err != nil {
		resp.Diagnostics.AddError("List SSH keys failed", err.Error())
		return
	}
	attrTypes := map[string]attr.Type{"id": types.StringType, "name": types.StringType, "public_key": types.StringType, "fingerprint": types.StringType}
	elems := make([]attr.Value, len(items))
	for i, k := range items {
		obj, objDiags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id": types.StringValue(k.ID), "name": types.StringValue(k.Name),
			"public_key": types.StringValue(k.PublicKey), "fingerprint": types.StringValue(k.Fingerprint),
		})
		resp.Diagnostics.Append(objDiags...)
		elems[i] = obj
	}
	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elems)
	resp.Diagnostics.Append(listDiags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &struct {
		TenantID types.String `tfsdk:"tenant_id"`
		SSHKeys  types.List   `tfsdk:"ssh_keys"`
	}{TenantID: cfg.TenantID, SSHKeys: list})...)
}

func NewRolesDataSource() datasource.DataSource { return &rolesDataSource{} }

type rolesDataSource struct{ client *virtfoundry.Client }

func (d *rolesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "virtfoundry_roles"
}

func (d *rolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists IAM roles in a tenant.",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{Optional: true},
			"roles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true},
					"description": schema.StringAttribute{Computed: true}, "is_system": schema.BoolAttribute{Computed: true},
					"permissions": schema.ListAttribute{ElementType: types.StringType, Computed: true},
				}},
			},
		},
	}
}

func (d *rolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(&d.client, req, resp)
}

func (d *rolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "client nil")
		return
	}
	var cfg tenantIDModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	tenantID, diags := resolveTenantID(d.client, cfg.TenantID)
	resp.Diagnostics.Append(diags...)
	items, err := d.client.ListRoles(ctx, tenantID)
	if err != nil {
		resp.Diagnostics.AddError("List roles failed", err.Error())
		return
	}
	attrTypes := map[string]attr.Type{
		"id": types.StringType, "name": types.StringType, "description": types.StringType,
		"is_system": types.BoolType, "permissions": types.ListType{ElemType: types.StringType},
	}
	elems := make([]attr.Value, len(items))
	for i, role := range items {
		perms := make([]types.String, len(role.Permissions))
		for j, p := range role.Permissions {
			perms[j] = types.StringValue(p)
		}
		permList, permDiags := types.ListValueFrom(ctx, types.StringType, perms)
		resp.Diagnostics.Append(permDiags...)
		obj, objDiags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id": types.StringValue(role.ID), "name": types.StringValue(role.Name),
			"description": types.StringValue(role.Description), "is_system": types.BoolValue(role.IsSystem),
			"permissions": permList,
		})
		resp.Diagnostics.Append(objDiags...)
		elems[i] = obj
	}
	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elems)
	resp.Diagnostics.Append(listDiags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &struct {
		TenantID types.String `tfsdk:"tenant_id"`
		Roles    types.List   `tfsdk:"roles"`
	}{TenantID: cfg.TenantID, Roles: list})...)
}

func NewUsersDataSource() datasource.DataSource { return &usersDataSource{} }

type usersDataSource struct{ client *virtfoundry.Client }

func (d *usersDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "virtfoundry_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists IAM users in a tenant.",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{Optional: true},
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{Computed: true}, "username": schema.StringAttribute{Computed: true},
					"email": schema.StringAttribute{Computed: true}, "role": schema.StringAttribute{Computed: true},
					"role_id": schema.StringAttribute{Computed: true}, "state": schema.StringAttribute{Computed: true},
				}},
			},
		},
	}
}

func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(&d.client, req, resp)
}

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "client nil")
		return
	}
	var cfg tenantIDModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	tenantID, diags := resolveTenantID(d.client, cfg.TenantID)
	resp.Diagnostics.Append(diags...)
	items, err := d.client.ListUsers(ctx, tenantID)
	if err != nil {
		resp.Diagnostics.AddError("List users failed", err.Error())
		return
	}
	attrTypes := map[string]attr.Type{
		"id": types.StringType, "username": types.StringType, "email": types.StringType,
		"role": types.StringType, "role_id": types.StringType, "state": types.StringType,
	}
	elems := make([]attr.Value, len(items))
	for i, u := range items {
		obj, objDiags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id": types.StringValue(u.ID), "username": types.StringValue(u.Username),
			"email": types.StringValue(u.Email), "role": types.StringValue(u.Role),
			"role_id": types.StringValue(u.RoleID), "state": types.StringValue(u.State),
		})
		resp.Diagnostics.Append(objDiags...)
		elems[i] = obj
	}
	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, elems)
	resp.Diagnostics.Append(listDiags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &struct {
		TenantID types.String `tfsdk:"tenant_id"`
		Users    types.List   `tfsdk:"users"`
	}{TenantID: cfg.TenantID, Users: list})...)
}
