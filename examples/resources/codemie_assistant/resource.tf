resource "codemie_assistant" "terraform" {
  name          = "Terraform Test Assistant"
  description   = "Validates Terraform syntax and recommends improvements."
  system_prompt = "## Instructions"
  project       = "example-project"
  llm_model_type = "gpt-5-2-2025-12-11"
  shared        = true
  is_react      = true
  categories    = ["business-analysis", "project-management"]
  conversation_starters = ["Review this Terraform configuration"]
  context = jsonencode([{ context_type = "code", name = "example-repository" }])
}