resource "codemie_workflow" "this" {
  name                  = var.name
  description           = var.description
  start_hint            = var.start_hint
  project               = var.project
  mode                  = var.mode
  icon_url              = var.icon_url
  shared                = var.shared
  yaml_config           = var.yaml_config
  supervisor_prompt     = var.supervisor_prompt
  meta_config           = var.meta_config
  guardrail_assignments = var.guardrail_assignments
  assistants            = var.assistants
  tools                 = var.tools
  states                = var.states
}
