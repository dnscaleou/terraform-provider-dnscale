terraform {
  required_providers {
    dnscale = {
      source  = "dnscale/dnscale"
      version = "1.0.0"
    }
  }
}

provider "dnscale" {
  api_key = var.api_key
  api_url = "https://api.dnscale.eu"
}

variable "api_key" {
  description = "DNScale API key"
  type        = string
  sensitive   = true
}

# Create a test zone
resource "dnscale_zone" "test" {
  name   = "terraform-test-zone.com"
  region = "EU"
  type   = "master"
}

# Add some test records
resource "dnscale_record" "root_a" {
  zone_id = dnscale_zone.test.id
  name    = "terraform-test-zone.com."
  type    = "A"
  content = "192.0.2.1"
  ttl     = 300
}

resource "dnscale_record" "www" {
  zone_id = dnscale_zone.test.id
  name    = "www.terraform-test-zone.com."
  type    = "CNAME"
  content = "terraform-test-zone.com."
  ttl     = 3600
}

resource "dnscale_record" "txt" {
  zone_id = dnscale_zone.test.id
  name    = "terraform-test-zone.com."
  type    = "TXT"
  content = "v=spf1 -all"
  ttl     = 3600
}

# Data sources
data "dnscale_zones" "all" {
  depends_on = [dnscale_zone.test]
}

data "dnscale_records" "test_records" {
  zone_id    = dnscale_zone.test.id
  depends_on = [dnscale_record.root_a, dnscale_record.www, dnscale_record.txt]
}

# Outputs
output "zone_id" {
  value = dnscale_zone.test.id
}

output "zone_name" {
  value = dnscale_zone.test.name
}

output "total_zones" {
  value = length(data.dnscale_zones.all.zones)
}

output "test_records_count" {
  value = length(data.dnscale_records.test_records.records)
}
