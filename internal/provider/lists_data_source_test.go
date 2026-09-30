package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

func TestListDataSourcesReadAPIShapes(t *testing.T) {
	t.Parallel()

	fixtures := map[string]string{
		"/api/v1/vpcs":            `{"vpcs":[{"id":"v1","name":"main","cidr":"10.0.0.0/16","state":"ready"}]}`,
		"/api/v1/networks":        `{"networks":[{"id":"n1","name":"default","vpc_id":"v1","cidr":"10.0.1.0/24","state":"ready"}]}`,
		"/api/v1/security-groups": `{"security_groups":[{"id":"sg1","name":"web","vpc_id":"v1","description":"http"}]}`,
		"/api/v1/ssh-keys":        `{"ssh_keys":[{"id":"k1","name":"laptop","public_key":"ssh-ed25519 AAA","fingerprint":"SHA256:abc"}]}`,
		"/api/v1/roles":           `{"roles":[{"id":"r1","name":"admin","description":"full","is_system":true,"permissions":["*"]}]}`,
		"/api/v1/users":           `{"users":[{"id":"u1","username":"alice","email":"a@ex.com","role":"admin","role_id":"r1","state":"active"}]}`,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := fixtures[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	ctx := context.Background()
	client, err := virtfoundry.NewClient(ctx, srv.URL, true)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetAPIKey("test-token")
	client.SetTenantID("t1")

	cases := []struct {
		name       string
		ds         datasource.DataSource
		listAttr   string
		configType tftypes.Object
	}{
		{
			name:     "vpcs",
			ds:       &vpcsDataSource{client: client},
			listAttr: "vpcs",
			configType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tenant_id": tftypes.String,
					"vpcs": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"id": tftypes.String, "name": tftypes.String, "cidr": tftypes.String, "state": tftypes.String,
					}}},
				},
			},
		},
		{
			name:     "networks",
			ds:       &networksDataSource{client: client},
			listAttr: "networks",
			configType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tenant_id": tftypes.String,
					"networks": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"id": tftypes.String, "name": tftypes.String, "vpc_id": tftypes.String,
						"cidr": tftypes.String, "state": tftypes.String,
					}}},
				},
			},
		},
		{
			name:     "security_groups",
			ds:       &securityGroupsDataSource{client: client},
			listAttr: "security_groups",
			configType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tenant_id": tftypes.String,
					"security_groups": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"id": tftypes.String, "name": tftypes.String, "vpc_id": tftypes.String, "description": tftypes.String,
					}}},
				},
			},
		},
		{
			name:     "ssh_keys",
			ds:       &sshKeysDataSource{client: client},
			listAttr: "ssh_keys",
			configType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tenant_id": tftypes.String,
					"ssh_keys": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"id": tftypes.String, "name": tftypes.String, "public_key": tftypes.String, "fingerprint": tftypes.String,
					}}},
				},
			},
		},
		{
			name:     "roles",
			ds:       &rolesDataSource{client: client},
			listAttr: "roles",
			configType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tenant_id": tftypes.String,
					"roles": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"id": tftypes.String, "name": tftypes.String, "description": tftypes.String,
						"is_system": tftypes.Bool, "permissions": tftypes.List{ElementType: tftypes.String},
					}}},
				},
			},
		},
		{
			name:     "users",
			ds:       &usersDataSource{client: client},
			listAttr: "users",
			configType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tenant_id": tftypes.String,
					"users": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"id": tftypes.String, "username": tftypes.String, "email": tftypes.String,
						"role": tftypes.String, "role_id": tftypes.String, "state": tftypes.String,
					}}},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var schemaResp datasource.SchemaResponse
			tc.ds.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("Schema: %v", schemaResp.Diagnostics)
			}

			configVals := map[string]tftypes.Value{
				"tenant_id": tftypes.NewValue(tftypes.String, nil),
				tc.listAttr: tftypes.NewValue(tc.configType.AttributeTypes[tc.listAttr], nil),
			}
			configRaw := tftypes.NewValue(tc.configType, configVals)

			req := datasource.ReadRequest{
				Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: configRaw},
			}
			var resp datasource.ReadResponse
			resp.State = tfsdk.State{
				Schema: schemaResp.Schema,
				Raw:    tftypes.NewValue(tc.configType, nil),
			}

			tc.ds.Read(ctx, req, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Read diagnostics: %v", resp.Diagnostics)
			}
			if resp.State.Raw.IsNull() {
				t.Fatal("expected non-null state after Read")
			}
		})
	}
}
