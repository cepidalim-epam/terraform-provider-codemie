resource "codemie_workflow" "example" {
  name        = "Review workflow"
  description = "Runs an automated code review workflow."
  project     = "example-project"
  mode        = "Sequential"
  shared      = true
  yaml_config = file("${path.module}/workflow.yaml")
}