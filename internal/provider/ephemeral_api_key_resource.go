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
	_ ephemeral.EphemeralResource              = &ephemeralAPIKeyResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &ephemeralAPIKeyResource{}
	_ ephemeral.EphemeralResourceWithClose     = &ephemeralAPIKeyResource{}
)

type ephemeralAPIKeyResource struct {
	client *virtfoundry.Client
}

type ephemeralAPIKeyModel struct {
	ID            types.String `tfsdk:"id"`
	TenantID      types.String `tfsdk:"tenant_id"`
	Name          types.String `tfsdk:"name"`
	UserID        types.String `tfsdk:"user_id"`
	ExpiresInDays types.Int64  `tfsdk:"expires_in_days"`
	Scopes        types.List   `tfsdk:"scopes"`
	Prefix        types.String `tfsdk:"prefix"`
	Secret        types.String `tfsdk:"secret"`
}

func NewEphemeralAPIKeyResource() ephemeral.EphemeralResource {
	return &ephemeralAPIKeyResource{}
}

func (r *ephemeralAPIKeyResource) Metadata(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = "virtfoundry_api_key"
}

func (r *ephemeralAPIKeyResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = ephemeralSchema.Schema{
		MarkdownDescription: "Ephemeral VirtFoundry API key. The secret is never persisted in state — it exists only in memory for the current Terraform operation. The key is deleted on `Close` (end of apply). For long-lived keys use the managed `virtfoundry_api_key` resource and enable state encryption.",
		Attributes: map[string]ephemeralSchema.Attribute{
			"tenant_id": ephemeralSchema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Tenant UUID. Defaults to provider `tenant_id`.",
			},
			"name": ephemeralSchema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Key name.",
			},
			"user_id": ephemeralSchema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Owner user UUID. Defaults to the authenticated user.",
			},
			"expires_in_days": ephemeralSchema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Expiration in days.",
			},
			"scopes": ephemeralSchema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Permission scopes.",
			},
			"id": ephemeralSchema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "API key UUID.",
			},
			"prefix": ephemeralSchema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Key prefix for identification.",
			},
			"secret": ephemeralSchema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Full API key secret (`vfd_live_...`). Ephemeral only — not written to state.",
			},
		},
	}
}

func (r *ephemeralAPIKeyResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	r.client = configureClientEphemeral(req, resp)
}

func (r *ephemeralAPIKeyResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var cfg ephemeralAPIKeyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID, diags := resolveTenantID(r.client, cfg.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	scopes, diags := stringListFromModel(ctx, cfg.Scopes)
	resp.Diagnostics.Append(diags...)

	in := virtfoundry.CreateAPIKeyInput{Name: cfg.Name.ValueString(), Scopes: scopes}
	if !cfg.UserID.IsNull() && cfg.UserID.ValueString() != "" {
		in.UserID = cfg.UserID.ValueString()
	}
	if !cfg.ExpiresInDays.IsNull() {
		in.ExpiresInDays = int(cfg.ExpiresInDays.ValueInt64())
	}
	res, err := r.client.CreateAPIKey(ctx, tenantID, in)
	if err != nil {
		resp.Diagnostics.AddError("Create ephemeral API key failed", err.Error())
		return
	}

	// Build result — never stored in state, only in ephemeral result
	result := ephemeralAPIKeyModel{
		ID:     types.StringValue(res.Key.ID),
		Prefix: types.StringValue(res.Key.Prefix),
		Secret: types.StringValue(res.Secret),
		Name:   cfg.Name,
	}
	if res.Key.UserID != "" {
		result.UserID = types.StringValue(res.Key.UserID)
	} else if !cfg.UserID.IsNull() {
		result.UserID = cfg.UserID
	}
	if len(res.Key.Scopes) > 0 {
		elems := make([]types.String, len(res.Key.Scopes))
		for i, s := range res.Key.Scopes {
			elems[i] = types.StringValue(s)
		}
		result.Scopes, _ = types.ListValueFrom(ctx, types.StringType, elems)
	} else if !cfg.Scopes.IsNull() {
		result.Scopes = cfg.Scopes
	}
	if !cfg.TenantID.IsNull() {
		result.TenantID = cfg.TenantID
	}
	if !cfg.ExpiresInDays.IsNull() {
		result.ExpiresInDays = cfg.ExpiresInDays
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &result)...)

	// Store private data for Close (auto-delete)
	priv := map[string]string{"tenant_id": tenantID, "id": res.Key.ID}
	b, _ := json.Marshal(priv)
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, "api_key", b)...)
}

func (r *ephemeralAPIKeyResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	if r.client == nil {
		return
	}
	b, diags := req.Private.GetKey(ctx, "api_key")
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
	// Best-effort delete — ephemeral keys are short-lived; ignore not-found
	if err := r.client.DeleteAPIKey(ctx, tenantID, id); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete ephemeral API key", err.Error())
	}
}
