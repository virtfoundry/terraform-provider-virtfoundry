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

var _ datasource.DataSource = &vmTemplatesDataSource{}

type vmTemplatesDataSource struct {
	client *virtfoundry.Client
}

type vmTemplatesDataSourceModel struct {
	TenantID  types.String `tfsdk:"tenant_id"`
	Templates types.List   `tfsdk:"templates"`
}

var vmTemplateAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"name":         types.StringType,
	"display_name": types.StringType,
	"image":        types.StringType,
	"source_type":  types.StringType,
	"os_type":      types.StringType,
	"state":        types.StringType,
	"import_state": types.StringType,
	"hypervisor":   types.StringType,
}

func NewVMTemplatesDataSource() datasource.DataSource { return &vmTemplatesDataSource{} }

func (d *vmTemplatesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "virtfoundry_vm_templates"
}

func (d *vmTemplatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists VM templates available in a tenant.",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Tenant UUID. Defaults to provider tenant_id.",
			},
			"templates": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true},
						"name":         schema.StringAttribute{Computed: true},
						"display_name": schema.StringAttribute{Computed: true},
						"image":        schema.StringAttribute{Computed: true},
						"source_type":  schema.StringAttribute{Computed: true},
						"os_type":      schema.StringAttribute{Computed: true},
						"state":        schema.StringAttribute{Computed: true},
						"import_state": schema.StringAttribute{Computed: true},
						"hypervisor":   schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *vmTemplatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*virtfoundry.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *virtfoundry.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *vmTemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "VirtFoundry client is nil")
		return
	}
	var config vmTemplatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(d.client, config.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.ListVMTemplates(ctx, tenantID)
	if err != nil {
		resp.Diagnostics.AddError("List VM templates failed", err.Error())
		return
	}
	elems := make([]attr.Value, len(items))
	for i, t := range items {
		obj, diags := types.ObjectValue(vmTemplateAttrTypes, map[string]attr.Value{
			"id":           types.StringValue(t.ID),
			"name":         types.StringValue(t.Name),
			"display_name": types.StringValue(t.DisplayName),
			"image":        types.StringValue(t.Image),
			"source_type":  types.StringValue(t.SourceType),
			"os_type":      types.StringValue(t.OSType),
			"state":        types.StringValue(t.State),
			"import_state": types.StringValue(t.ImportState),
			"hypervisor":   types.StringValue(t.Hypervisor),
		})
		resp.Diagnostics.Append(diags...)
		elems[i] = obj
	}
	list, diags := types.ListValue(types.ObjectType{AttrTypes: vmTemplateAttrTypes}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	out := vmTemplatesDataSourceModel{Templates: list}
	if !config.TenantID.IsNull() && config.TenantID.ValueString() != "" {
		out.TenantID = config.TenantID
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, out)...)
}
