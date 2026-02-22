terraform {
  required_version = ">= 1.0"

  required_providers {
    stripe = {
      source  = "lukasaron/stripe"
      version = "~> 1.10"
    }
  }
}

provider "stripe" {
  api_key = var.stripe_api_key
}
