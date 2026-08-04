package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var (
	_ resource.Resource                = &sshKeyResource{}
	_ resource.ResourceWithImportState = &sshKeyResource{}
)

type sshKeyResource struct {
	client *virtfoundry.Client
}

type sshKeyModel struct {
	ID            types.String `tfsdk:"id"`
	TenantID      types.String `tfsdk:"tenant_id"`
	Name          types.String `tfsdk:"name"`
	PublicKey     types.String `tfsdk:"public_key"`
	Generate      types.Bool   `tfsdk:"generate"`
	PrivateKeyPEM types.String `tfsdk:"private_key_pem"`
	Fingerprint   types.String `tfsdk:"fingerprint"`
}

func NewSSHKeyResource() resource.Resource { return &sshKeyResource{} }

func (r *sshKeyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfoundry_ssh_key"
}

func (r *sshKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VirtFoundry tenant SSH key. Register an existing public key or generate a new Ed25519 pair.",
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tenant_id": schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":      schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"public_key": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "OpenSSH authorized_keys line. Omit when `generate` is true.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"generate": schema.BoolAttribute{Optional: true, MarkdownDescription: "Generate a new key pair via the API.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()}},
			"private_key_pem": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Generated private key PEM (only when `generate` is true).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"fingerprint": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *sshKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *sshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan sshKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	generate := !plan.Generate.IsNull() && plan.Generate.ValueBool()
	hasPublic := !plan.PublicKey.IsNull() && plan.PublicKey.ValueString() != ""
	if generate && hasPublic {
		resp.Diagnostics.AddError("Conflicting SSH key input", "Set either `generate = true` or `public_key`, not both.")
		return
	}
	if !generate && !hasPublic {
		resp.Diagnostics.AddError("Missing SSH key material", "Set `public_key` or `generate = true`.")
		return
	}

	tenantID, diags := resolveTenantID(r.client, plan.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var key *virtfoundry.SSHKey
	var privateKey string
	var err error
	if generate {
		key, privateKey, err = r.client.CreateSSHKey(ctx, tenantID, plan.Name.ValueString())
	} else {
		key, err = r.client.RegisterSSHKey(ctx, tenantID, virtfoundry.RegisterSSHKeyInput{
			Name:      plan.Name.ValueString(),
			PublicKey: plan.PublicKey.ValueString(),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("Create SSH key failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, sshKeyToModel(key, plan, privateKey))...)
}

func (r *sshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state sshKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	key, err := r.client.GetSSHKey(ctx, tenantID, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read SSH key failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, sshKeyToModel(key, state, state.PrivateKeyPEM.ValueString()))...)
}

func (r *sshKeyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Recreate the SSH key resource to change name or key material.")
}

func (r *sshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(requireClient(r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state sshKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, diags := resolveTenantID(r.client, state.TenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSSHKey(ctx, tenantID, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete SSH key failed", err.Error())
	}
}

func (r *sshKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTenantName(ctx, resp, req.ID)
}

func sshKeyToModel(k *virtfoundry.SSHKey, cfg sshKeyModel, privateKey string) sshKeyModel {
	out := sshKeyModel{
		ID:          types.StringValue(k.ID),
		Name:        types.StringValue(k.Name),
		PublicKey:   types.StringValue(k.PublicKey),
		Fingerprint: types.StringValue(k.Fingerprint),
		Generate:    cfg.Generate,
	}
	if privateKey != "" {
		out.PrivateKeyPEM = types.StringValue(privateKey)
	} else if !cfg.PrivateKeyPEM.IsNull() {
		out.PrivateKeyPEM = cfg.PrivateKeyPEM
	}
	if !cfg.TenantID.IsNull() && cfg.TenantID.ValueString() != "" {
		out.TenantID = cfg.TenantID
	}
	return out
}
