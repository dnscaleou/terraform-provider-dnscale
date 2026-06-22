---
page_title: "dnscale_records Data Source - DNScale"
subcategory: ""
description: |-
  Retrieves a list of all DNS records in a zone.
---

# dnscale_records (Data Source)

Retrieves a list of all DNS records in a zone. The provider follows DNScale API pagination automatically, so zones with more than one API page of records are returned as a single combined `records` list.

## Example Usage

```terraform
data "dnscale_records" "all" {
  zone_id = dnscale_zone.example.id
}

output "record_count" {
  value = length(data.dnscale_records.all.records)
}

# Filter A records
output "a_records" {
  value = [for r in data.dnscale_records.all.records : r if r.type == "A"]
}
```

## Schema

### Required

- `zone_id` (String) - The zone ID to list records from.

### Read-Only

- `records` (List of Object) - List of records. Each record has the following attributes:
  - `id` (String) - The unique identifier for the record.
  - `name` (String) - The full record name with trailing dot.
  - `type` (String) - The DNS record type.
  - `content` (String) - The record value.
  - `ttl` (Number) - Time-to-live in seconds.
  - `priority` (Number) - Priority (for MX and SRV records).
  - `disabled` (Boolean) - Whether the record is disabled.

## Pagination

DNScale API responses include pagination metadata (`total`, `offset`, `limit`, `count`, and `has_more`). This data source requests records in 1000-record pages and continues until the API reports there are no more records. This covers large zones, including zones with more than 10,000 records.
