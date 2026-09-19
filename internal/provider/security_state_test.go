package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
