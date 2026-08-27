terraform {
  required_version = ">=1.5.0"

  cloud {
    
    organization = "ismaildalim-cepic"

    workspaces {
      name = "codemie"
    }
  }
}