# Terraform Provider for DNScale

The DNScale Terraform provider allows you to manage DNS zones and records on the [DNScale](https://dnscale.eu) platform.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build the provider plugin)

## Installation

### From Terraform Registry

```hcl
terraform {
  required_providers {
    dnscale = {
      source  = "dnscaleou/dnscale"
      version = "~> 1.0"
    }
  }
}
```

### Local Development

```bash
git clone https://github.com/dnscaleou/terraform-provider-dnscale.git
cd terraform-provider-dnscale
go install .
```

## Configuration

Configure the provider with your DNScale API key:

```hcl
provider "dnscale" {
  api_key = var.dnscale_api_key
  # api_url = "https://api.dnscale.eu"  # Optional, defaults to production API
}
```

You can also use environment variables:

```bash
export DNSCALE_API_KEY="your-api-key"
export DNSCALE_API_URL="https://api.dnscale.eu"  # Optional
```

## Usage Examples

### Create a DNS Zone

```hcl
resource "dnscale_zone" "example" {
  name   = "example.com"
  region = "EU"
  type   = "master"
}
```

### Create DNS Records

```hcl
# A Record
resource "dnscale_record" "www" {
  zone_id = dnscale_zone.example.id
  name    = "www.example.com."
  type    = "A"
  content = "192.0.2.1"
  ttl     = 300
}

# MX Record
resource "dnscale_record" "mail" {
  zone_id  = dnscale_zone.example.id
  name     = "example.com."
  type     = "MX"
  content  = "mail.example.com."
  ttl      = 3600
  priority = 10
}

# TXT Record (SPF)
resource "dnscale_record" "spf" {
  zone_id = dnscale_zone.example.id
  name    = "example.com."
  type    = "TXT"
  content = "v=spf1 include:_spf.example.com -all"
  ttl     = 3600
}

# CNAME Record
resource "dnscale_record" "cdn" {
  zone_id = dnscale_zone.example.id
  name    = "cdn.example.com."
  type    = "CNAME"
  content = "example.com."
  ttl     = 3600
}
```

### Enable DNSSEC

```hcl
resource "dnscale_dnssec_key" "ksk" {
  zone_id   = dnscale_zone.example.id
  key_type  = "KSK"
  algorithm = "ECDSAP256SHA256"
  active    = true
  published = true
}

resource "dnscale_dnssec_key" "zsk" {
  zone_id   = dnscale_zone.example.id
  key_type  = "ZSK"
  algorithm = "ECDSAP256SHA256"
  active    = true
  published = true
}
```

### Data Sources

```hcl
# List all zones
data "dnscale_zones" "all" {}

# Get a specific zone
data "dnscale_zone" "example" {
  id = "zone-uuid"
}

# List records in a zone
data "dnscale_records" "all" {
  zone_id = dnscale_zone.example.id
}

# Get DNSSEC status
data "dnscale_dnssec_status" "example" {
  zone_id = dnscale_zone.example.id
}
```

## Resources

| Resource | Description |
|----------|-------------|
| `dnscale_zone` | Manages a DNS zone |
| `dnscale_record` | Manages a DNS record |
| `dnscale_dnssec_key` | Manages DNSSEC cryptographic keys |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `dnscale_zone` | Retrieves a DNS zone by ID |
| `dnscale_zones` | Lists all DNS zones |
| `dnscale_records` | Lists all records in a zone |
| `dnscale_dnssec_status` | Retrieves DNSSEC status for a zone |

## Supported Record Types

- A, AAAA
- CNAME, ALIAS
- MX
- TXT
- NS
- SRV
- CAA
- PTR
- TLSA
- SSHFP
- HTTPS, SVCB

## Import

Resources can be imported using their IDs:

```bash
# Import a zone
terraform import dnscale_zone.example <zone-id>

# Import a record (format: zone_id/record_id)
terraform import dnscale_record.www <zone-id>/<record-id>
```

## Development

### Building

```bash
go build -o terraform-provider-dnscale
```

### Testing

```bash
# Unit tests
go test ./...

# Acceptance tests (requires API key)
DNSCALE_API_KEY=your-key go test -v ./... -run TestAcc
```

### Generating Documentation

```bash
go generate ./...
```

### Releasing

Release and Terraform Registry publishing steps are documented in [docs/RELEASE.md](docs/RELEASE.md).

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.
