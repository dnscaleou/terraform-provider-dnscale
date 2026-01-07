# Assume zone already exists
data "dnscale_zone" "example" {
  name = "example.com"
}

# Create KSK (Key-Signing-Key) for DNSSEC
resource "dnscale_dnssec_key" "ksk" {
  zone_id   = data.dnscale_zone.example.id
  key_type  = "KSK"
  algorithm = "ECDSAP256SHA256" # Modern ECDSA algorithm
  active    = true
  published = true
}

# Create ZSK (Zone-Signing-Key) for DNSSEC
resource "dnscale_dnssec_key" "zsk" {
  zone_id   = data.dnscale_zone.example.id
  key_type  = "ZSK"
  algorithm = "ECDSAP256SHA256"
  active    = true
  published = true
}

# Output DS records for registrar configuration
output "ds_records" {
  description = "DS records to add at your domain registrar"
  value       = dnscale_dnssec_key.ksk.ds
}

output "ksk_key_tag" {
  description = "KSK key tag"
  value       = dnscale_dnssec_key.ksk.key_tag
}

output "zsk_key_tag" {
  description = "ZSK key tag"
  value       = dnscale_dnssec_key.zsk.key_tag
}
