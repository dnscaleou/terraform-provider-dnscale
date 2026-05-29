terraform {
  required_providers {
    dnscale = {
      source  = "dnscaleou/dnscale"
      version = "~> 1.0"
    }
  }
}

# Configure the DNScale Provider
# API key can be set via DNSCALE_API_KEY environment variable
provider "dnscale" {
  api_key = var.dnscale_api_key
  # api_url = "https://api.dnscale.eu" # Optional, defaults to production
}

variable "dnscale_api_key" {
  description = "DNScale API key for authentication"
  type        = string
  sensitive   = true
}
