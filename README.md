# CodeMie Terraform Provider

Terraform provider for managing CodeMie assistants, workflows, and skills through the CodeMie REST API.

## Repository & GitLab CI/CD

- **Remote URL:** `git@git.epam.com:ismaildalim_cepic/codemie-terraform-provider.git`
- **Registry:** EPAM GitLab Terraform Provider Registry (`git.epam.com`)

### Remote Setup

To push this repository to EPAM GitLab:

```bash
git remote add origin git@git.epam.com:ismaildalim_cepic/codemie-terraform-provider.git
git branch -M main
git push -u origin main
```

---

## Consuming the Provider from GitLab Registry

### 1. Authenticate Terraform with EPAM GitLab

#### For Local Development (`~/.terraformrc`)
Add the following block to `~/.terraformrc` (or `%APPDATA%\terraform.rc` on Windows):

```hcl
credentials "git.epam.com" {
  token = "<YOUR_GITLAB_ACCESS_TOKEN>"
}
```

#### For CI/CD Pipelines
Export the `TF_TOKEN_<hostname>` environment variable in your consumer pipeline:

```bash
export TF_TOKEN_git_epam_com="${CI_JOB_TOKEN}"
# Or using a group/project access token
export TF_TOKEN_git_epam_com="${GITLAB_ACCESS_TOKEN}"
```

### 2. Provider Declaration

In your downstream Terraform project (`versions.tf`):

```hcl
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    codemie = {
      source  = "git.epam.com/ismaildalim_cepic/codemie"
      version = "~> 0.1.0"
    }
  }
}

provider "codemie" {
  # Direct attributes or environment variables:
  # CODEMIE_HOST, CODEMIE_TOKEN_URL, CODEMIE_CLIENT_ID, CODEMIE_CLIENT_SECRET
}
```

---

## Authentication & Credentials

The provider uses OAuth 2.0 client credentials. Configure attributes directly in the provider block or use standard environment variables:

| Parameter | Environment Variable | Description |
|-----------|----------------------|-------------|
| `host` | `CODEMIE_HOST` | Base URL of CodeMie API |
| `token_url` | `CODEMIE_TOKEN_URL` | OAuth2 Token endpoint URL |
| `client_id` | `CODEMIE_CLIENT_ID` | OAuth2 Client ID |
| `client_secret` | `CODEMIE_CLIENT_SECRET` | OAuth2 Client Secret (sensitive) |

```bash
export CODEMIE_HOST="https://codemie-api.example.com"
export CODEMIE_TOKEN_URL="https://auth.example.com/oauth/token"
export CODEMIE_CLIENT_ID="terraform-client"
export CODEMIE_CLIENT_SECRET="secret"
```

> **Security Note:** Never commit client secrets to version control.

---

## CI/CD Pipeline Lifecycle

The included `.gitlab-ci.yml` provides a production-grade automated pipeline:

```mermaid
graph LR
    A[Push / MR] --> B[Lint & Vet]
    B --> C[Unit Tests & Coverage]
    C --> D[Cross-Platform Build]
    D --> E[Publish to GitLab Provider Registry]
    E --> F[Create GitLab Release]
```

1. **Lint**: Validates Go formatting (`gofmt`) and runs static analysis (`go vet`).
2. **Test**: Executes test suite with coverage reports (`go test -coverprofile`).
3. **Build**: Cross-compiles binaries for `linux_amd64`, `linux_arm64`, `darwin_amd64`, `darwin_arm64`, and `windows_amd64`, generating SHA256 checksums and zip archives.
4. **Publish**: Automatically uploads versioned provider packages and checksums to the GitLab Terraform Provider Registry (`/packages/terraform/providers/codemie/...`).
5. **Release**: On Git tags (e.g. `v0.1.0`), creates a tagged GitLab Release with links to package assets.

### Creating a Release

To trigger a production release build and publish to the package registry:

```bash
git tag v0.1.0
git push origin v0.1.0
```

---

## Local Development & Testing

Requires Go 1.24 or newer.

```bash
# Run tests and formatting
make test
make lint

# Package binaries for all architectures into dist/
make package

# Install provider locally for development overrides
make install-local
```

### Development Override (`~/.terraformrc`)

For local testing without querying the remote registry:

```hcl
provider_installation {
  dev_overrides {
    "git.epam.com/ismaildalim_cepic/codemie" = "/path/to/codemie-terraform-provider"
  }
  direct {}
}
```