---
page_title: "DNScale Provider"
subcategory: ""
description: |-
  The DNScale provider is used to manage DNS zones and records on the DNScale platform.
---

# DNScale Provider

The DNScale provider allows you to manage DNS zones and records on the [DNScale](https://dnscale.eu) platform.

## Example Usage

```terraform
terraform {
  required_providers {
    dnscale = {
      source  = "dnscaleou/dnscale"
      version = "~> 1.0"
    }
  }
}

provider "dnscale" {
  api_key = var.dnscale_api_key
}

# Create a DNS zone
resource "dnscale_zone" "example" {
  name   = "example.com"
  region = "EU"
  type   = "master"
}

# Create an A record
resource "dnscale_record" "www" {
  zone_id = dnscale_zone.example.id
  name    = "www.example.com."
  type    = "A"
  content = "192.0.2.1"
  ttl     = 300
}
```

## Authentication

The DNScale provider requires an API key for authentication. You can obtain an API key from the [DNScale Dashboard](https://dashboard.dnscale.eu).

### Configuration

You can configure the provider in your Terraform configuration:

```terraform
provider "dnscale" {
  api_key = "your-api-key"
  api_url = "https://api.dnscale.eu"  # Optional
}
```

### Environment Variables

Alternatively, you can use environment variables:

- `DNSCALE_API_KEY` - Your DNScale API key (required)
- `DNSCALE_API_URL` - API endpoint URL (optional, defaults to `https://api.dnscale.eu`)

```bash
export DNSCALE_API_KEY="your-api-key"
```

## Schema

### Optional

- `api_key` (String, Sensitive) - DNScale API key for authentication. Can also be set via `DNSCALE_API_KEY` environment variable.
- `api_url` (String) - DNScale API URL. Defaults to `https://api.dnscale.eu`. Can also be set via `DNSCALE_API_URL` environment variable.
