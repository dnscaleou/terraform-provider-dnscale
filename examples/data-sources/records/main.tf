# First, look up the zone
data "dnscale_zone" "example" {
  name = "example.com"
}

# List all records in the zone
data "dnscale_records" "all" {
  zone_id = data.dnscale_zone.example.id
}

# List only A records
data "dnscale_records" "a_records" {
  zone_id = data.dnscale_zone.example.id
  type    = "A"
}

# List only MX records
data "dnscale_records" "mx_records" {
  zone_id = data.dnscale_zone.example.id
  type    = "MX"
}

# Output all records
output "all_records" {
  value = data.dnscale_records.all.records
}

# Output A record IPs
output "a_record_ips" {
  value = [for r in data.dnscale_records.a_records.records : r.content]
}
