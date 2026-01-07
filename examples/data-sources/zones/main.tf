# List all zones
data "dnscale_zones" "all" {}

# Output all zone names
output "zone_names" {
  value = [for z in data.dnscale_zones.all.zones : z.name]
}

# Output zones with their regions
output "zone_regions" {
  value = {
    for z in data.dnscale_zones.all.zones : z.name => z.region
  }
}

# Count zones
output "total_zones" {
  value = length(data.dnscale_zones.all.zones)
}
