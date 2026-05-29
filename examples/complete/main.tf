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

variable "dnscale_api_key" {
  description = "DNScale API key"
  type        = string
  sensitive   = true
}

variable "domain" {
  description = "Domain name to manage"
  type        = string
  default     = "example.com"
}

# Create the DNS zone
resource "dnscale_zone" "main" {
  name   = var.domain
  region = "EU_GLOBAL"
}

# Root domain A records
resource "dnscale_record" "root_a" {
  zone_id = dnscale_zone.main.id
  name    = "${var.domain}."
  type    = "A"
  content = "192.0.2.1"
  ttl     = 300
}

# WWW CNAME
resource "dnscale_record" "www" {
  zone_id = dnscale_zone.main.id
  name    = "www.${var.domain}."
  type    = "CNAME"
  content = "${var.domain}."
  ttl     = 3600
}

# Mail server records
resource "dnscale_record" "mx_primary" {
  zone_id  = dnscale_zone.main.id
  name     = "${var.domain}."
  type     = "MX"
  content  = "mail.${var.domain}."
  ttl      = 3600
  priority = 10
}

resource "dnscale_record" "mail_a" {
  zone_id = dnscale_zone.main.id
  name    = "mail.${var.domain}."
  type    = "A"
  content = "192.0.2.10"
  ttl     = 3600
}

# SPF record
resource "dnscale_record" "spf" {
  zone_id = dnscale_zone.main.id
  name    = "${var.domain}."
  type    = "TXT"
  content = "v=spf1 mx -all"
  ttl     = 3600
}

# DMARC record
resource "dnscale_record" "dmarc" {
  zone_id = dnscale_zone.main.id
  name    = "_dmarc.${var.domain}."
  type    = "TXT"
  content = "v=DMARC1; p=quarantine; rua=mailto:dmarc@${var.domain}"
  ttl     = 3600
}

# CAA record for Let's Encrypt
resource "dnscale_record" "caa" {
  zone_id = dnscale_zone.main.id
  name    = "${var.domain}."
  type    = "CAA"
  content = "0 issue \"letsencrypt.org\""
  ttl     = 3600
}

# DNSSEC keys
resource "dnscale_dnssec_key" "ksk" {
  zone_id   = dnscale_zone.main.id
  key_type  = "KSK"
  algorithm = "ECDSAP256SHA256"
}

resource "dnscale_dnssec_key" "zsk" {
  zone_id   = dnscale_zone.main.id
  key_type  = "ZSK"
  algorithm = "ECDSAP256SHA256"
}

# Data source examples
data "dnscale_dnssec_status" "main" {
  zone_id = dnscale_zone.main.id
}

data "dnscale_records" "all" {
  zone_id = dnscale_zone.main.id
}

# Outputs
output "zone_id" {
  description = "Zone UUID"
  value       = dnscale_zone.main.id
}

output "nameservers" {
  description = "Configure these nameservers at your registrar"
  value       = "ns1.dnscale.eu, ns2.dnscale.eu"
}

output "ds_records" {
  description = "DS records to add at your registrar for DNSSEC"
  value       = dnscale_dnssec_key.ksk.ds
}

output "dnssec_status" {
  description = "DNSSEC configuration status"
  value = {
    enabled = data.dnscale_dnssec_status.main.enabled
    has_ksk = data.dnscale_dnssec_status.main.has_ksk
    has_zsk = data.dnscale_dnssec_status.main.has_zsk
  }
}

output "record_count" {
  description = "Total number of DNS records"
  value       = length(data.dnscale_records.all.records)
}
