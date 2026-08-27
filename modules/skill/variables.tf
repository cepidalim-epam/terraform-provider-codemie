variable "name" {
  type        = string
  description = "Unique kebab-case name of the skill"
}

variable "description" {
  type        = string
  description = "Description of what the skill does"
  default     = null
}

variable "content" {
  type        = string
  description = "Markdown / instruction body of the skill"
  default     = null
}

variable "project" {
  type        = string
  description = "Project key associated with the skill"
  default     = null
}

variable "visibility" {
  type        = string
  description = "Visibility of skill (private or public)"
  default     = "private"
}

variable "categories" {
  type        = list(string)
  description = "Categories (max 3)"
  default     = null
}

variable "toolkits" {
  type        = string
  description = "JSON string for toolkits"
  default     = null
}

variable "mcp_servers" {
  type        = string
  description = "JSON string for MCP servers"
  default     = null
}

variable "companion_files" {
  type        = string
  description = "JSON string for companion files"
  default     = null
}

variable "enabled_builtin_subagents" {
  type        = list(string)
  description = "Enabled built-in subagents"
  default     = null
}
