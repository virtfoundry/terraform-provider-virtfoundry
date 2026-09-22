package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

func TestAPIKeySecretNotPersistedOnRead(t *testing.T) {
	k := &virtfoundry.APIKey{ID: "k1", Name: "test", Prefix: "vfd_test", Scopes: []string{"a"}}
	cfg := apiKeyModel{Secret: types.StringValue("old-secret"), TenantID: types.StringValue("t1")}
	out := apiKeyToModel(k, cfg, "")
	if !out.Secret.IsNull() {
		t.Fatalf("expected Secret to be null after Read, got %q", out.Secret.ValueString())
	}
	out2 := apiKeyToModel(k, cfg, "new-secret")
	if out2.Secret.ValueString() != "new-secret" {
		t.Fatalf("expected Secret to be new-secret on Create, got %q", out2.Secret.ValueString())
	}
}

func TestSSHPrivateKeyNotPersistedOnRead(t *testing.T) {
	k := &virtfoundry.SSHKey{ID: "s1", Name: "test", PublicKey: "ssh-ed25519 AAA", Fingerprint: "fp"}
	cfg := sshKeyModel{PrivateKeyPEM: types.StringValue("old-pem"), TenantID: types.StringValue("t1")}
	out := sshKeyToModel(k, cfg, "")
	if !out.PrivateKeyPEM.IsNull() {
		t.Fatalf("expected PrivateKeyPEM to be null after Read, got %q", out.PrivateKeyPEM.ValueString())
	}
	out2 := sshKeyToModel(k, cfg, "new-pem")
	if out2.PrivateKeyPEM.ValueString() != "new-pem" {
		t.Fatalf("expected PrivateKeyPEM new-pem, got %q", out2.PrivateKeyPEM.ValueString())
	}
}

func TestUserPasswordWriteOnlyNotPersisted(t *testing.T) {
	u := &virtfoundry.User{ID: "u1", Username: "alice", Role: "admin"}
	cfg := userModel{Password: types.StringValue("secret"), TenantID: types.StringValue("t1")}
	out := userToModel(u, cfg)
	if !out.Password.IsNull() {
		t.Fatalf("expected Password to be null (WriteOnly), got %q", out.Password.ValueString())
	}
}

func TestTenantAdminPasswordWriteOnly(t *testing.T) {
	tenant := &virtfoundry.Tenant{ID: "t1", Name: "acme", Namespace: "acme", State: "active"}
	cfg := tenantModel{AdminPassword: types.StringValue("secret"), Slug: types.StringValue("acme")}
	out := tenantToModel(tenant, cfg)
	if !out.AdminPassword.IsNull() {
		t.Fatalf("expected AdminPassword to be null (WriteOnly), got %q", out.AdminPassword.ValueString())
	}
}

func TestProviderHasEphemeralResources(t *testing.T) {
	p := &virtfoundryProvider{}
	ephemerals := p.EphemeralResources(context.Background())
	if len(ephemerals) != 2 {
		t.Fatalf("expected 2 ephemeral resources, got %d", len(ephemerals))
	}
	resources := p.Resources(context.Background())
	if len(resources) != 13 {
		t.Fatalf("expected 13 resources, got %d", len(resources))
	}
}

func TestSchemasHaveCorrectSensitivity(t *testing.T) {
	ctx := context.Background()
	// user password must be WriteOnly + Sensitive
	ur := &userResource{}
	var ureq resource.SchemaRequest
	var uresp resource.SchemaResponse
	ur.Schema(ctx, ureq, &uresp)
	if attr, ok := uresp.Schema.Attributes["password"]; ok {
		if sAttr, ok := attr.(schema.StringAttribute); ok {
			if !sAttr.WriteOnly {
				t.Fatalf("user password should be WriteOnly")
			}
			if !sAttr.Sensitive {
				t.Fatalf("user password should be Sensitive")
			}
		} else {
			t.Fatalf("password not StringAttribute")
		}
	} else {
		t.Fatalf("password attribute missing")
	}
	// tenant admin_password WriteOnly
	tr := &tenantResource{}
	var tresp resource.SchemaResponse
	tr.Schema(ctx, ureq, &tresp)
	if attr, ok := tresp.Schema.Attributes["admin_password"]; ok {
		if sAttr, ok := attr.(schema.StringAttribute); ok {
			if !sAttr.WriteOnly {
				t.Fatalf("tenant admin_password should be WriteOnly")
			}
		}
	}
	// api_key secret should be Sensitive and Computed, not WriteOnly
	ar := &apiKeyResource{}
	var aresp resource.SchemaResponse
	ar.Schema(ctx, ureq, &aresp)
	if attr, ok := aresp.Schema.Attributes["secret"]; ok {
		if sAttr, ok := attr.(schema.StringAttribute); ok {
			if sAttr.WriteOnly {
				t.Fatalf("api_key secret should NOT be WriteOnly (Computed)")
			}
			if !sAttr.Sensitive || !sAttr.Computed {
				t.Fatalf("api_key secret should be Sensitive+Computed")
			}
		}
	}
}

func userObjectType(t *testing.T, ctx context.Context) tftypes.Object {
	t.Helper()
	var schemaResp resource.SchemaResponse
	(&userResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	typ := schemaResp.Schema.Type().TerraformType(ctx)
	obj, ok := typ.(tftypes.Object)
	if !ok {
		t.Fatalf("user schema TerraformType is not Object: %T", typ)
	}
	return obj
}

func tenantObjectType(t *testing.T, ctx context.Context) tftypes.Object {
	t.Helper()
	var schemaResp resource.SchemaResponse
	(&tenantResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	typ := schemaResp.Schema.Type().TerraformType(ctx)
	obj, ok := typ.(tftypes.Object)
	if !ok {
		t.Fatalf("tenant schema TerraformType is not Object: %T", typ)
	}
	return obj
}

func userSchemaRef(ctx context.Context) resource.SchemaResponse {
	var resp resource.SchemaResponse
	(&userResource{}).Schema(ctx, resource.SchemaRequest{}, &resp)
	return resp
}

func tenantSchemaRef(ctx context.Context) resource.SchemaResponse {
	var resp resource.SchemaResponse
	(&tenantResource{}).Schema(ctx, resource.SchemaRequest{}, &resp)
	return resp
}

func TestUserWriteOnlyPasswordReachesAPI(t *testing.T) {
	ctx := context.Background()
	sr := userSchemaRef(ctx)
	ot := userObjectType(t, ctx)

	configRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"id":        tftypes.NewValue(tftypes.String, nil),
		"tenant_id": tftypes.NewValue(tftypes.String, "t1"),
		"username":  tftypes.NewValue(tftypes.String, "alice"),
		"password":  tftypes.NewValue(tftypes.String, "hunter2-correct"),
		"email":     tftypes.NewValue(tftypes.String, nil),
		"role_id":   tftypes.NewValue(tftypes.String, nil),
		"role_name": tftypes.NewValue(tftypes.String, nil),
		"role":      tftypes.NewValue(tftypes.String, nil),
		"state":     tftypes.NewValue(tftypes.String, nil),
	})
	planRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"id":        tftypes.NewValue(tftypes.String, nil),
		"tenant_id": tftypes.NewValue(tftypes.String, "t1"),
		"username":  tftypes.NewValue(tftypes.String, "alice"),
		"password":  tftypes.NewValue(tftypes.String, nil),
		"email":     tftypes.NewValue(tftypes.String, nil),
		"role_id":   tftypes.NewValue(tftypes.String, nil),
		"role_name": tftypes.NewValue(tftypes.String, nil),
		"role":      tftypes.NewValue(tftypes.String, nil),
		"state":     tftypes.NewValue(tftypes.String, nil),
	})

	var receivedBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/users" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &receivedBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if r.Header.Get("X-Tenant-ID") != "t1" {
			t.Errorf("expected X-Tenant-ID=t1, got %q", r.Header.Get("X-Tenant-ID"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"user":{"id":"u1","username":"alice","role":"admin"}}`))
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(ctx, srv.URL, true)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("test-token")

	r := &userResource{client: client}
	req := resource.CreateRequest{
		Config: tfsdk.Config{Schema: sr.Schema, Raw: configRaw},
		Plan:   tfsdk.Plan{Schema: sr.Schema, Raw: planRaw},
	}
	var resp resource.CreateResponse
	resp.State = tfsdk.State{
		Schema: sr.Schema,
		Raw:    tftypes.NewValue(ot, nil),
	}

	r.Create(ctx, req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create diagnostics: %v", resp.Diagnostics)
	}
	if receivedBody == nil {
		t.Fatal("API was never called")
	}
	if got, _ := receivedBody["password"].(string); got != "hunter2-correct" {
		t.Fatalf("password sent to API: got %q want %q (received body: %s)", got, "hunter2-correct", asJSON(receivedBody))
	}
	if got, _ := receivedBody["username"].(string); got != "alice" {
		t.Fatalf("username sent to API: got %q want %q", got, "alice")
	}
}

func TestTenantWriteOnlyAdminPasswordReachesAPI(t *testing.T) {
	ctx := context.Background()
	sr := tenantSchemaRef(ctx)
	ot := tenantObjectType(t, ctx)

	configRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, nil),
		"name":           tftypes.NewValue(tftypes.String, "acme"),
		"slug":           tftypes.NewValue(tftypes.String, "acme"),
		"admin_password": tftypes.NewValue(tftypes.String, "rootpw-correct"),
		"namespace":      tftypes.NewValue(tftypes.String, nil),
		"state":          tftypes.NewValue(tftypes.String, nil),
	})
	planRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, nil),
		"name":           tftypes.NewValue(tftypes.String, "acme"),
		"slug":           tftypes.NewValue(tftypes.String, "acme"),
		"admin_password": tftypes.NewValue(tftypes.String, nil),
		"namespace":      tftypes.NewValue(tftypes.String, nil),
		"state":          tftypes.NewValue(tftypes.String, nil),
	})

	var receivedBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/api/v1/tenants") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &receivedBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tenant":{"id":"t1","name":"acme","slug":"acme","namespace":"acme","state":"active"}}`))
	}))
	t.Cleanup(srv.Close)

	client, err := virtfoundry.NewClient(ctx, srv.URL, true)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("test-token")

	r := &tenantResource{client: client}
	req := resource.CreateRequest{
		Config: tfsdk.Config{Schema: sr.Schema, Raw: configRaw},
		Plan:   tfsdk.Plan{Schema: sr.Schema, Raw: planRaw},
	}
	var resp resource.CreateResponse
	resp.State = tfsdk.State{
		Schema: sr.Schema,
		Raw:    tftypes.NewValue(ot, nil),
	}

	r.Create(ctx, req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create diagnostics: %v", resp.Diagnostics)
	}
	if receivedBody == nil {
		t.Fatal("API was never called")
	}
	if got, _ := receivedBody["admin_password"].(string); got != "rootpw-correct" {
		t.Fatalf("admin_password sent to API: got %q want %q (received body: %s)", got, "rootpw-correct", asJSON(receivedBody))
	}
}

func asJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "<marshal-error>"
	}
	return string(b)
}
