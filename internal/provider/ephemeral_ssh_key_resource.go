package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	ephemeralSchema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var (
	_ ephemeral.EphemeralResource              = &ephemeralSSHKeyResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &ephemeralSSHKeyResource{}
	_ ephemeral.EphemeralResourceWithClose     = &ephemeralSSHKeyResource{}
)

type ephemeralSSHKeyResource struct {
	client *virtfoundry.Client
}

type ephemeralSSHKeyModel struct {
	ID            types.String `tfsdk:"id"`
	TenantID      types.String `tfsdk:"tenant_id"`
	Name          types.String `tfsdk:"name"`
	PublicKey     types.String `tfsdk:"public_key"`
	Generate      types.Bool   `tfsdk:"generate"`
	PrivateKeyPEM types.String `tfsdk:"private_key_pem"`
	Fingerprint   types.String `tfsdk:"fingerprint"`
}

func NewEphemeralSSHKeyResource() ephemeral.EphemeralResource {
	return &ephemeralSSHKeyResource{}
}

func (r *ephemeralSSHKeyResource) Metadata(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = "virtfoundry_ssh_key"
}

func (r *ephemeralSSHKeyResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = ephemeralSchema.Schema{
		MarkdownDescription: "Ephemeral VirtFoundry SSH key. When `generate` is true the private key PEM is exposed only in the ephemeral result and the key is deleted on `Close`. Prefer BYO `public_key` with `tls_private_key` outside Terraform for production; use ephemeral generation only for short-lived provisioners.",
		Attributes: map[string]ephemeralSchema.Attribute{
			"tenant_id": ephemeralSchema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Tenant UUID. Defaults to provider `tenant_id`.",
			},
			"name": ephemeralSchema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Key name within the tenant.",
			},
			"public_key": ephemeralSchema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "OpenSSH authorized_keys line. Omit when `generate` is true.",
			},
			"generate": ephemeralSchema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Generate a new Ed25519 key pair via the API (ephemeral).",
			},
			"id": ephemeralSchema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SSH key UUID.",
			},
			"private_key_pem": ephemeralSchema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Generated private key PEM (only when `generate` is true). Ephemeral only.",
			},
			"fingerprint": ephemeralSchema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Key fingerprint.",
			},
		},
	}
}

func (r *ephemeralSSHKeyResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	r.client = configureClientEphemeral(req, resp)
}

func (r *ephemeralSSHKeyResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var cfg ephemeralSSHKeyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	generate := !cfg.Generate.IsNull() && cfg.Generate.ValueBool()
	hasPublic := !cfg.PublicKey.IsNull() && cfg.PublicKey.ValueString() != ""
	if generate && hasPublic {
		resp.Diagnostics.AddError("Conflicting SSH key input", "Set either `generate = true` or `public_key`, not both.")
		return
	}
	if !generate && !hasPublic {
		resp.Diagnostics.AddError("Missing SSH key material", "Set `public_key` or `generate = true`.")
		return
	}

	tenantID, diags := resolveTenantID(r.client, cfg.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var key *virtfoundry.SSHKey
	var privateKey string
	var err error
	if generate {
		key, privateKey, err = r.client.CreateSSHKey(ctx, tenantID, cfg.Name.ValueString())
	} else {
		key, err = r.client.RegisterSSHKey(ctx, tenantID, virtfoundry.RegisterSSHKeyInput{
			Name:      cfg.Name.ValueString(),
			PublicKey: cfg.PublicKey.ValueString(),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("Create ephemeral SSH key failed", err.Error())
		return
	}

	result := ephemeralSSHKeyModel{
		ID:          types.StringValue(key.ID),
		Name:        types.StringValue(key.Name),
		PublicKey:   types.StringValue(key.PublicKey),
		Fingerprint: types.StringValue(key.Fingerprint),
		Generate:    cfg.Generate,
	}
	if privateKey != "" {
		result.PrivateKeyPEM = types.StringValue(privateKey)
	} else {
		result.PrivateKeyPEM = types.StringNull()
	}
	if !cfg.TenantID.IsNull() {
		result.TenantID = cfg.TenantID
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &result)...)

	priv := map[string]string{"tenant_id": tenantID, "id": key.ID}
	b, _ := json.Marshal(priv)
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, "ssh_key", b)...)
}

func (r *ephemeralSSHKeyResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	if r.client == nil {
		return
	}
	b, diags := req.Private.GetKey(ctx, "ssh_key")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || len(b) == 0 {
		return
	}
	var priv map[string]string
	if err := json.Unmarshal(b, &priv); err != nil {
		return
	}
	tenantID := priv["tenant_id"]
	id := priv["id"]
	if id == "" {
		return
	}
	if err := r.client.DeleteSSHKey(ctx, tenantID, id); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete ephemeral SSH key", err.Error())
	}
}
