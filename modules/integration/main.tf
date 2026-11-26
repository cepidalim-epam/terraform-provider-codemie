resource "codemie_integration" "this" {
  scope             = var.scope
  project_name      = var.project_name
  alias             = var.alias
  credential_type   = var.credential_type
  credential_values = var.credential_values
  is_global         = var.is_global
}