resource "codemie_skill" "review" {
  name        = "terraform-review"
  description = "Reviews Terraform configuration for correctness and maintainability."
  content     = <<-EOT
    # Terraform review

    Inspect Terraform configuration for correctness, security, maintainability,
    provider version constraints, state safety, and accidental resource replacement.
    Return findings with severity, location, explanation, and a concrete remediation.
  EOT
  project    = "example-project"
  visibility = "private"
  categories = ["development"]
}