# Create a DNS zone
resource "dnscale_zone" "example" {
  name   = "example.com"
  region = "EU_GLOBAL" # Optional: EU, GLOBAL, or EU_GLOBAL
  type   = "master"    # Optional: master (default) or slave
  status = "active"    # Optional: active (default) or paused
}

# Output the zone ID for use with records
output "zone_id" {
  value = dnscale_zone.example.id
}

output "zone_name" {
  value = dnscale_zone.example.name
}
