package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var (
	_ resource.Resource                = &apiKeyResource{}
	_ resource.ResourceWithImportState = &apiKeyResource{}
)

type apiKeyResource struct{ client *virtfoundry.Client }

type apiKeyModel struct {
	ID            types.String `tfsdk:"id"`
	TenantID      types.String `tfsdk:"tenant_id"`
	Name          types.String `tfsdk:"name"`
	UserID        types.String `tfsdk:"user_id"`
	ExpiresInDays types.Int64  `tfsdk:"expires_in_days"`
	Scopes        types.List   `tfsdk:"scopes"`
	Prefix        types.String `tfsdk:"prefix"`
	Secret        types.String `tfsdk:"secret"`
}

func NewAPIKeyResource() resource.Resource { return &apiKeyResource{} }

func (r *apiKeyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_api_key"
}

func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry API key. The secret is only available at creation time.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tenant_id":       schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":            schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"user_id":         schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"expires_in_days": schema.Int64Attribute{Optional: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
			"scopes":          schema.ListAttribute{ElementType: types.StringType, Optional: true, PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()}},
			"prefix":          schema.StringAttribute{Computed: true},
			"secret":          schema.StringAttribute{Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan apiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	scopes, diags := stringListFromModel(ctx, plan.Scopes)
	resp.Diagnostics.Append(diags...)
	in := virtfoundry.CreateAPIKeyInput{Name: plan.Name.ValueString(), Scopes: scopes}
	if !plan.UserID.IsNull() {
		in.UserID = plan.UserID.ValueString()
	}
	if !plan.ExpiresInDays.IsNull() {
		in.ExpiresInDays = int(plan.ExpiresInDays.ValueInt64())
	}
	res, err := r.client.CreateAPIKey(ctx, tenantID, in)
	if err != nil {
		resp.Diagnostics.AddError("Create API key failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, apiKeyToModel(&res.Key, plan, res.Secret))...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state apiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	key, err := r.client.GetAPIKey(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read API key failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, apiKeyToModel(key, state, state.Secret.ValueString()))...)
}

func (r *apiKeyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Revoke and recreate the API key resource.")
}

func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state apiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAPIKey(ctx, tenantID, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete API key failed", err.Error())
	}
}

func (r *apiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
}

func apiKeyToModel(k *virtfoundry.APIKey, cfg apiKeyModel, secret string) apiKeyModel {
	out := apiKeyModel{
		ID:     types.StringValue(k.ID),
		Name:   types.StringValue(k.Name),
		Prefix: types.StringValue(k.Prefix),
	}
	if k.UserID != "" {
		out.UserID = types.StringValue(k.UserID)
	} else if !cfg.UserID.IsNull() {
		out.UserID = cfg.UserID
	}
	if secret != "" {
		out.Secret = types.StringValue(secret)
	} else if !cfg.Secret.IsNull() {
		out.Secret = cfg.Secret
	}
	if len(k.Scopes) > 0 {
		elems := make([]types.String, len(k.Scopes))
		for i, s := range k.Scopes {
			elems[i] = types.StringValue(s)
		}
		out.Scopes, _ = types.ListValueFrom(ctxBackground, types.StringType, elems)
	} else if !cfg.Scopes.IsNull() {
		out.Scopes = cfg.Scopes
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	if !cfg.ExpiresInDays.IsNull() {
		out.ExpiresInDays = cfg.ExpiresInDays
	}
	return out
}
