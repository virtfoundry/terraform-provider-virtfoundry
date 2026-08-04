package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var _ provider.Provider = &virtfoundryProvider{}

type virtfoundryProvider struct {
	version string
}

type virtfoundryProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	APIKey   types.String `tfsdk:"api_key"`
	TenantID types.String `tfsdk:"tenant_id"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &virtfoundryProvider{version: version}
	}
}

func (p *virtfoundryProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "virtfoundry"
	resp.Version = p.version
}

func (p *virtfoundryProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Interact with [VirtFoundry](https://github.com/virtfoundry/core) private cloud IaaS.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "VirtFoundry API base URL (e.g. `https://virtfoundry.example.com`).",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for JWT login. Required when `api_key` is not set.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for JWT login. Required when `api_key` is not set.",
				Optional:            true,
				Sensitive:           true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "VirtFoundry API key (`vfd_live_...`). Preferred for automation.",
				Optional:            true,
				Sensitive:           true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "Default tenant ID for API calls (root users may override per resource later).",
				Optional:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification. For development only.",
				Optional:            true,
			},
		},
	}
}

func (p *virtfoundryProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config virtfoundryProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := firstNonEmpty(config.Endpoint, envString("VIRTFOUNDRY_ENDPOINT"))
	username := firstNonEmpty(config.Username, envString("VIRTFOUNDRY_USERNAME"))
	password := firstNonEmpty(config.Password, envString("VIRTFOUNDRY_PASSWORD"))
	apiKey := firstNonEmpty(config.APIKey, envString("VIRTFOUNDRY_API_KEY"))
	tenantID := firstNonEmpty(config.TenantID, envString("VIRTFOUNDRY_TENANT_ID"))
	insecure := config.Insecure.ValueBool() || os.Getenv("VIRTFOUNDRY_INSECURE") == "true"

	if endpoint.IsNull() || endpoint.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing VirtFoundry API endpoint",
			"Set the endpoint attribute or VIRTFOUNDRY_ENDPOINT environment variable.",
		)
		return
	}

	hasAPIKey := !apiKey.IsNull() && apiKey.ValueString() != ""
	hasUserPass := (!username.IsNull() && username.ValueString() != "") || (!password.IsNull() && password.ValueString() != "")

	if hasAPIKey && hasUserPass {
		resp.Diagnostics.AddError(
			"Conflicting credentials",
			"Set either api_key or username/password, not both.",
		)
		return
	}
	if !hasAPIKey {
		if username.IsNull() || username.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("username"),
				"Missing username",
				"Set username/password or api_key.",
			)
		}
		if password.IsNull() || password.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("password"),
				"Missing password",
				"Set username/password or api_key.",
			)
		}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	client, err := virtfoundry.NewClient(endpoint.ValueString(), insecure)
	if err != nil {
		resp.Diagnostics.AddError("Invalid endpoint", err.Error())
		return
	}

	if !tenantID.IsNull() && tenantID.ValueString() != "" {
		client.SetTenantID(tenantID.ValueString())
	}

	if hasAPIKey {
		client.SetAPIKey(apiKey.ValueString())
	} else {
		if err := client.Login(ctx, username.ValueString(), password.ValueString()); err != nil {
			resp.Diagnostics.AddError("VirtFoundry login failed", err.Error())
			return
		}
	}

	if err := client.Health(ctx); err != nil {
		tflog.Warn(ctx, "Health endpoint unavailable (common when only /api is exposed via Gateway); continuing with auth check", map[string]any{
			"error": err.Error(),
		})
	}
	if err := client.PingAuth(ctx); err != nil {
		resp.Diagnostics.AddError("VirtFoundry authentication failed", err.Error())
		return
	}

	tflog.Info(ctx, "Configured VirtFoundry client", map[string]any{
		"endpoint": endpoint.ValueString(),
	})

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *virtfoundryProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVMResource,
	}
}

func (p *virtfoundryProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func firstNonEmpty(primary types.String, fallback string) types.String {
	if !primary.IsNull() && primary.ValueString() != "" {
		return primary
	}
	if fallback != "" {
		return types.StringValue(fallback)
	}
	return primary
}

func envString(key string) string {
	return os.Getenv(key)
}
