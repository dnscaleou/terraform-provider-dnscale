---
page_title: "dnscale_zone Resource - DNScale"
subcategory: ""
description: |-
  Manages a DNS zone in DNScale.
---

# dnscale_zone (Resource)

Manages a DNS zone in DNScale.

## Example Usage

```terraform
resource "dnscale_zone" "example" {
  name   = "example.com"
  region = "EU"
  type   = "master"
}
```

## Schema

### Required

- `name` (String) - The domain name for the zone (e.g., `example.com`).

### Optional

- `region` (String) - The region where the zone will be hosted. Valid values: `EU`, `GLOBAL`, `EU_GLOBAL`.
- `type` (String) - The zone type. Valid values: `master`, `slave`. Default: `master`.
- `status` (String) - The zone status. Valid values: `active`, `paused`, `pending`, `error`. Default: `active`.

### Read-Only

- `id` (String) - The unique identifier for the zone (UUID).
- `customer_id` (String) - Customer UUID that owns this zone.
- `created_at` (String) - Timestamp when the zone was created.
- `updated_at` (String) - Timestamp when the zone was last updated.

## Import

Zones can be imported using their ID:

```bash
terraform import dnscale_zone.example <zone-id>
```

Example:

```bash
terraform import dnscale_zone.example 12345678-1234-1234-1234-123456789abc
```
