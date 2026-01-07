# First, look up the zone
data "dnscale_zone" "example" {
  name = "example.com"
}

# Get DNSSEC status
data "dnscale_dnssec_status" "example" {
  zone_id = data.dnscale_zone.example.id
}

# Output DNSSEC status
output "dnssec_status" {
  value = {
    enabled    = data.dnscale_dnssec_status.example.enabled
    keys_count = data.dnscale_dnssec_status.example.keys_count
    has_ksk    = data.dnscale_dnssec_status.example.has_ksk
    has_zsk    = data.dnscale_dnssec_status.example.has_zsk
  }
}

# Conditional check - warn if DNSSEC is not fully configured
output "dnssec_ready" {
  value = data.dnscale_dnssec_status.example.has_ksk && data.dnscale_dnssec_status.example.has_zsk
}
