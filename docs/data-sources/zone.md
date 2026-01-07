---
page_title: "dnscale_zone Data Source - DNScale"
subcategory: ""
description: |-
  Retrieves information about a DNS zone in DNScale.
---

# dnscale_zone (Data Source)

Retrieves information about a DNS zone in DNScale.

## Example Usage

```terraform
data "dnscale_zone" "example" {
  id = "12345678-1234-1234-1234-123456789abc"
}

output "zone_name" {
  value = data.dnscale_zone.example.name
}
```

## Schema

### Required

- `id` (String) - The unique identifier for the zone (UUID).

### Read-Only

- `name` (String) - The domain name for the zone.
- `region` (String) - The region where the zone is hosted.
- `type` (String) - The zone type.
- `status` (String) - The current status of the zone.
- `created_at` (String) - Timestamp when the zone was created.
- `updated_at` (String) - Timestamp when the zone was last updated.
