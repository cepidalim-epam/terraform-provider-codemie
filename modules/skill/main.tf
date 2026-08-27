resource "codemie_skill" "this" {
  name                      = var.name
  description               = var.description
  content                   = var.content
  project                   = var.project
  visibility                = var.visibility
  categories                = var.categories
  toolkits                  = var.toolkits
  mcp_servers               = var.mcp_servers
  companion_files           = var.companion_files
  enabled_builtin_subagents = var.enabled_builtin_subagents
}
