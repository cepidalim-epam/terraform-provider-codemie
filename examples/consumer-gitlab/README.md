# Consuming CodeMie Terraform Provider from GitLab Registry

This example demonstrates how to consume the `codemie` Terraform provider published to the EPAM GitLab Terraform Provider Registry (`git.epam.com`).

## 1. Authentication with GitLab Registry

Terraform CLI requires credentials to download providers from private GitLab registries.

### Option A: Using `~/.terraformrc` (Local Development)

Add your GitLab Personal Access Token (with `read_api` and `read_package_registry` scopes) to `~/.terraformrc` (or `%APPDATA%\terraform.rc` on Windows):

```hcl
credentials "git.epam.com" {
  token = "your-gitlab-personal-access-token"
}
```

### Option B: Using Environment Variables (CI/CD pipelines)

Set the Terraform token environment variable before running `terraform init`:

```bash
# In GitLab CI of consumer repositories:
export TF_TOKEN_git_epam_com="${CI_JOB_TOKEN}"
# Or using Personal Access Token:
export TF_TOKEN_git_epam_com="your-gitlab-personal-access-token"
```

## 2. Terraform Provider Configuration

Declare the provider in `versions.tf`:

```hcl
terraform {
  required_providers {
    codemie = {
      source  = "git.epam.com/ismaildalim_cepic/codemie"
      version = "~> 0.1.0"
    }
  }
}

provider "codemie" {
  # Provider credentials can be supplied here or via environment variables
}
```

## 3. Initialize & Apply

```bash
terraform init
terraform plan
terraform apply
```
