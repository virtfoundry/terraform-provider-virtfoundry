---
page_title: "virtfoundry_api_key Ephemeral Resource - virtfoundry"
subcategory: ""
description: |-
  Ephemeral VirtFoundry API key. The secret is never persisted in state.
---

# virtfoundry_api_key (Ephemeral Resource)

Ephemeral API key for short-lived automation. The secret is **never written to state** — it is available only via `ephemeral` result during the current `terraform apply` and the key is **deleted on Close**. For long-lived keys use `virtfoundry_api_key` managed resource with state encryption.

Requires Terraform >= 1.10 and provider `>= 0.3`.

## Example Usage

```hcl
ephemeral "virtfoundry_api_key" "ci" {
  name            = "ci-ephemeral"
  expires_in_days = 1
  scopes          = ["vms:read"]
}

# Use elsewhere via ephemeral reference (never in state):
# provider "virtfoundry" { api_key = ephemeral.virtfoundry_api_key.ci.secret }
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Key name. |
| `user_id` | String | no | Owner user UUID. Defaults to the authenticated user. |
| `expires_in_days` | Number | no | Expiration in days. Short-lived recommended. |
| `scopes` | List(String) | no | Permission scopes. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | API key UUID. |
| `prefix` | Key prefix for identification. |
| `secret` | Full API key secret (sensitive, ephemeral only). |
