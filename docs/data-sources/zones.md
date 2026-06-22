---
page_title: "dnscale_zones Data Source - DNScale"
subcategory: ""
description: |-
  Retrieves a list of all DNS zones in DNScale.
---

# dnscale_zones (Data Source)

Retrieves a list of all DNS zones in DNScale. The provider automatically follows the DNScale API's paginated `/v1/zones` response until every visible zone has been read.

## Example Usage

```terraform
data "dnscale_zones" "all" {}

output "zone_count" {
  value = length(data.dnscale_zones.all.zones)
}

output "zone_names" {
  value = [for z in data.dnscale_zones.all.zones : z.name]
}
```

## Schema

### Read-Only

- `zones` (List of Object) - List of zones. Each zone has the following attributes:
  - `id` (String) - The unique identifier for the zone (UUID).
  - `name` (String) - The domain name for the zone.
  - `region` (String) - The region where the zone is hosted.
  - `type` (String) - The zone type.
  - `status` (String) - The current status of the zone.

## Pagination

DNScale API responses include pagination metadata (`total`, `offset`, `limit`, `count`, and `has_more`). This data source handles pagination internally, using the API maximum page size and requesting additional pages while `has_more` is true. Terraform configurations receive a single combined `zones` list.
