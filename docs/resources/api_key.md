---
page_title: "virtfoundry_api_key Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry API key. The secret is only available at creation time.
---

# virtfoundry_api_key (Resource)

Creates an API key for programmatic access. The full secret (`vfd_live_...`) is returned **once** at creation — store it securely.

## Example Usage

```hcl
resource "virtfoundry_api_key" "ci" {
  name            = "github-actions"
  expires_in_days = 90
  scopes          = ["vms:read", "vms:write"]
}

output "api_key_secret" {
  value     = virtfoundry_api_key.ci.secret
  sensitive = true
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Key name. Forces replacement. |
| `user_id` | String | no | Owner user UUID. Defaults to the authenticated user. |
| `expires_in_days` | Number | no | Expiration in days. Forces replacement. |
| `scopes` | List(String) | no | Permission scopes. Forces replacement. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | API key UUID. |
| `prefix` | Key prefix for identification. |
| `secret` | Full API key secret (sensitive; only at create). |

## Import

```shell
terraform import virtfoundry_api_key.ci <tenant_id>/<key_id>
```

> **Note:** Imported keys do not expose the secret. Rotate the key if the secret was lost.
