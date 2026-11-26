variable "scope" {
  type        = string
  description = "Integration scope: user or project"
  default     = "project"
}

variable "project_name" {
  type        = string
  description = "CodeMie project associated with the integration"
}

variable "alias" {
  type        = string
  description = "Unique integration alias within the scope and project"
}

variable "credential_type" {
  type        = string
  description = "One of Webhook, MCP, Scheduler, Jira, OpenAPI, or Git"
}

variable "credential_values" {
  type        = string
  description = "JSON array of integration credential key/value objects"
  sensitive   = true
}

variable "is_global" {
  type        = bool
  description = "Whether the integration is globally visible"
  default     = false
}