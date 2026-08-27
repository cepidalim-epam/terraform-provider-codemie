output "id" {
  description = "Unique ID of the created CodeMie workflow"
  value       = codemie_workflow.this.id
}

output "name" {
  description = "Name of the CodeMie workflow"
  value       = codemie_workflow.this.name
}
