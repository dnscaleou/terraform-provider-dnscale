# Look up a zone by name
data "dnscale_zone" "by_name" {
  name = "example.com"
}

# Look up a zone by ID
data "dnscale_zone" "by_id" {
  id = "12345678-1234-1234-1234-123456789abc"
}

# Output zone information
output "zone_details" {
  value = {
    id         = data.dnscale_zone.by_name.id
    name       = data.dnscale_zone.by_name.name
    region     = data.dnscale_zone.by_name.region
    status     = data.dnscale_zone.by_name.status
    created_at = data.dnscale_zone.by_name.created_at
  }
}
