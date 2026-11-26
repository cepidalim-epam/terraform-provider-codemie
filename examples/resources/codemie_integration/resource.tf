variable "jira_url" {
  type = string
}

variable "jira_token" {
  type      = string
  sensitive = true
}

variable "git_token" {
  type      = string
  sensitive = true
}

terraform {
  required_providers {
    codemie = {
      source = "cepidalim-epam/codemie"
    }
  }
}

resource "codemie_integration" "jira" {
  scope           = "project"
  project_name    = "example-project"
  alias           = "jira-main"
  credential_type = "Jira"

  credential_values = jsonencode([
    { key = "url", value = var.jira_url },
    { key = "token", value = var.jira_token },
    { key = "is_cloud", value = true },
  ])
}

resource "codemie_integration" "git" {
  scope           = "project"
  project_name    = "example-project"
  alias           = "git-main"
  credential_type = "Git"

  credential_values = jsonencode([
    { key = "url", value = "https://git.example.invalid/group/repository" },
    { key = "token", value = var.git_token },
    { key = "auth_type", value = "token" },
    { key = "name", value = "Terraform example" },
  ])
}

resource "codemie_integration" "webhook" {
  scope           = "project"
  project_name    = "example-project"
  alias           = "example-webhook"
  credential_type = "Webhook"

  credential_values = jsonencode([
    { key = "webhook_id", value = "example-webhook-id" },
    { key = "is_enabled", value = true },
    { key = "resource_type", value = "workflow" },
    { key = "resource_id", value = "example-workflow-id" },
  ])
}

resource "codemie_integration" "mcp" {
  scope           = "project"
  project_name    = "example-project"
  alias           = "example-mcp-integration"
  credential_type = "MCP"

  credential_values = jsonencode([
    { key = "url", value = "https://mcp.example.invalid" },
  ])
}

resource "codemie_integration" "scheduler" {
  scope           = "project"
  project_name    = "example-project"
  alias           = "example-scheduler"
  credential_type = "Scheduler"

  credential_values = jsonencode([
    { key = "is_enabled", value = true },
    { key = "schedule", value = "0 9 * * 1-5" },
    { key = "timezone", value = "UTC" },
    { key = "resource_type", value = "workflow" },
    { key = "resource_id", value = "example-workflow-id" },
  ])
}

resource "codemie_integration" "openapi" {
  scope           = "project"
  project_name    = "example-project"
  alias           = "example-openapi"
  credential_type = "OpenAPI"

  credential_values = jsonencode([
    { key = "openapi_spec", value = "{\"openapi\":\"3.0.0\",\"info\":{\"title\":\"Example API\",\"version\":\"1.0.0\"},\"paths\":{}}" },
    { key = "is_basic_auth", value = false },
    { key = "timeout", value = 30 },
  ])
}