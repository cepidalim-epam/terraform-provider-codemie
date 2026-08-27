# ==============================================================================
# Example: Consuming CodeMie Provider from EPAM GitLab Provider Registry
# ==============================================================================

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
  # Configuration can also be provided via environment variables:
  # CODEMIE_HOST, CODEMIE_TOKEN_URL, CODEMIE_CLIENT_ID, CODEMIE_CLIENT_SECRET
}
