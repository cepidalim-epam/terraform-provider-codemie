resource "codemie_skill" "example" {
  name        = "sample-skill"
  description = "A sample CodeMie skill deployed via Terraform"
  content     = "# Sample Skill\nProvides utility capabilities for CodeMie assistants."
  visibility  = "private"
  project     = "talent-turbocharge"
}

resource "codemie_assistant" "example" {
  name          = "Production Assistant"
  description   = "CodeMie Assistant provisioned from GitLab Registry"
  system_prompt = "You are a helpful coding assistant for the team."
  project       = "talent-turbocharge"
  skill_ids     = [codemie_skill.example.id]
}
