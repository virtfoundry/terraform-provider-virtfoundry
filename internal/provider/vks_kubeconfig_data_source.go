package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var _ datasource.DataSource = &vksKubeconfigDataSource{}

type vksKubeconfigDataSource struct {
	client *virtfoundry.Client
}

type vksKubeconfigDataSourceModel struct {
	TenantID   types.String `tfsdk:"tenant_id"`
	Name       types.String `tfsdk:"name"`
	Kubeconfig types.String `tfsdk:"kubeconfig"`
}

func NewVKSKubeconfigDataSource() datasource.DataSource { return &vksKubeconfigDataSource{} }

func (d *vksKubeconfigDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "virtfoundry_vks_kubeconfig"
}

func (d *vksKubeconfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the admin kubeconfig of a VKS cluster. Requires the `vks:kubeconfig` permission. The kubeconfig is a credential and is stored in state (marked sensitive) — enable state encryption.",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Tenant UUID. Defaults to provider tenant_id.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "VKS cluster name.",
			},
			"kubeconfig": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Kubeconfig YAML for the cluster.",
			},
		},
	}
}

func (d *vksKubeconfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vksKubeconfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(d.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var config vksKubeconfigDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(d.client, config.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	kubeconfig, err := d.client.GetVKSKubeconfig(ctx, tenantID, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read VKS kubeconfig failed", err.Error())
		return
	}
	config.Kubeconfig = types.StringValue(kubeconfig)
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
