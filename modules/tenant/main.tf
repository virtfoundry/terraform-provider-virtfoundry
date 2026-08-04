# Root-only module. Pass provider "virtfoundry.root" without tenant_id.
resource "virtfoundry_tenant" "this" {
  provider = virtfoundry

  name           = var.name
  slug           = var.slug
  admin_password = var.admin_password
}
