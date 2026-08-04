resource "virtfoundry_role" "this" {
  provider = virtfoundry

  for_each = local.resolved_roles

  tenant_id   = var.tenant_id
  name        = each.key
  description = each.value.description
  permissions = each.value.permissions
}

resource "virtfoundry_user" "this" {
  provider = virtfoundry

  for_each = var.users

  tenant_id = var.tenant_id
  username  = each.value.username
  password  = each.value.password
  email     = each.value.email != "" ? each.value.email : null
  role_id   = local.user_role_ids[each.key]
  state     = each.value.state
}

resource "virtfoundry_api_key" "this" {
  provider = virtfoundry

  for_each = var.api_keys

  tenant_id       = var.tenant_id
  name            = each.value.name
  user_id         = coalesce(each.value.user_id, try(virtfoundry_user.this[each.value.user_key].id, null))
  expires_in_days = each.value.expires_in_days
  scopes          = each.value.scopes
}
