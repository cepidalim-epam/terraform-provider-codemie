variable "name" {
  type        = string
  description = "Name of the CodeMie assistant"
}

variable "description" {
  type        = string
  description = "Description of the assistant"
  default     = null
}

variable "system_prompt" {
  type        = string
  description = "System prompt instructions for the assistant"
  default     = null
}

variable "project" {
  type        = string
  description = "Project key associated with the assistant"
  default     = null
}

variable "icon_url" {
  type        = string
  description = "URL or identifier for the assistant icon"
  default     = null
}

variable "llm_model_type" {
  type        = string
  description = "LLM model identifier"
  default     = null
}

variable "enable_image_generation" {
  type        = bool
  description = "Enable DALL-E image generation"
  default     = null
}

variable "image_generation_model" {
  type        = string
  description = "Image generation model name"
  default     = null
}

variable "conversation_starters" {
  type        = list(string)
  description = "List of conversation starter suggestions"
  default     = null
}

variable "shared" {
  type        = bool
  description = "Whether the assistant is shared"
  default     = null
}

variable "is_react" {
  type        = bool
  description = "Whether the assistant uses ReAct loop"
  default     = null
}

variable "is_global" {
  type        = bool
  description = "Whether the assistant is globally visible"
  default     = null
}

variable "agent_mode" {
  type        = string
  description = "Agent execution mode"
  default     = null
}

variable "plan_prompt" {
  type        = string
  description = "Planning prompt for the agent"
  default     = null
}

variable "slug" {
  type        = string
  description = "Unique slug for URL routing"
  default     = null
}

variable "temperature" {
  type        = number
  description = "Sampling temperature (0.0 to 2.0)"
  default     = null
}

variable "top_p" {
  type        = number
  description = "Nucleus sampling top_p (0.0 to 1.0)"
  default     = null
}

variable "tools_tokens_size_limit" {
  type        = number
  description = "Maximum token limit for tools payload"
  default     = null
}

variable "smart_tool_selection_enabled" {
  type        = bool
  description = "Enable smart tool selection"
  default     = null
}

variable "assistant_ids" {
  type        = list(string)
  description = "List of sub-assistant IDs"
  default     = null
}

variable "enabled_builtin_subagents" {
  type        = list(string)
  description = "List of enabled built-in subagents"
  default     = null
}

variable "skill_ids" {
  type        = list(string)
  description = "List of attached skill IDs"
  default     = null
}

variable "type" {
  type        = string
  description = "Type of assistant"
  default     = "codemie"
}

variable "categories" {
  type        = list(string)
  description = "List of categories (max 3)"
  default     = null
}

variable "source_assistant_id" {
  type        = string
  description = "Source assistant ID for cloning"
  default     = null
}

variable "skip_integration_validation" {
  type        = bool
  description = "Skip tool integration validation during save"
  default     = null
}

variable "context" {
  type        = string
  description = "JSON string for context definitions"
  default     = null
}

variable "toolkits" {
  type        = string
  description = "JSON string for toolkits configuration"
  default     = null
}

variable "mcp_servers" {
  type        = string
  description = "JSON string for MCP server definitions"
  default     = null
}

variable "hedging_config" {
  type        = string
  description = "JSON string for hedging config"
  default     = null
}

variable "interactive_features" {
  type        = string
  description = "JSON string for interactive features"
  default     = null
}

variable "bedrock" {
  type        = string
  description = "JSON string for Bedrock configuration"
  default     = null
}

variable "bedrock_agentcore_runtime" {
  type        = string
  description = "JSON string for Bedrock AgentCore runtime"
  default     = null
}

variable "agent_card" {
  type        = string
  description = "JSON string for AgentCard metadata"
  default     = null
}

variable "prompt_variables" {
  type        = string
  description = "JSON string for prompt variables"
  default     = null
}

variable "custom_metadata" {
  type        = string
  description = "JSON string for custom metadata"
  default     = null
}

variable "guardrail_assignments" {
  type        = string
  description = "JSON string for guardrails"
  default     = null
}
