variable "name" {
  type        = string
  description = "Name of the workflow"
}

variable "description" {
  type        = string
  description = "Description of the workflow"
  default     = null
}

variable "start_hint" {
  type        = string
  description = "Hint for starting the workflow"
  default     = null
}

variable "project" {
  type        = string
  description = "Project key associated with the workflow"
  default     = null
}

variable "mode" {
  type        = string
  description = "Execution mode (Sequential, Hierarchical, etc.)"
  default     = "Sequential"
}

variable "icon_url" {
  type        = string
  description = "Icon URL or identifier"
  default     = null
}

variable "shared" {
  type        = bool
  description = "Whether the workflow is shared"
  default     = null
}

variable "yaml_config" {
  type        = string
  description = "YAML configuration defining the workflow graph"
  default     = null
}

variable "supervisor_prompt" {
  type        = string
  description = "Supervisor prompt instructions"
  default     = null
}

variable "meta_config" {
  type        = string
  description = "Metadata configuration"
  default     = null
}

variable "guardrail_assignments" {
  type        = string
  description = "JSON string for guardrails"
  default     = null
}

variable "assistants" {
  type        = string
  description = "JSON string for assistants mapping"
  default     = null
}

variable "tools" {
  type        = string
  description = "JSON string for tools mapping"
  default     = null
}

variable "states" {
  type        = string
  description = "JSON string for states mapping"
  default     = null
}
