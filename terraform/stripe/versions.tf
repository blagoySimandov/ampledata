terraform {
  required_version = "~>1.10.3"

  required_providers {
    stripe = {
      source  = "lukasaron/stripe"
      version = "3.4.1"
    }
  }
}

provider "stripe" {
  api_key = var.stripe_api_key
}
