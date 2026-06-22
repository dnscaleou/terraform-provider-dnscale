---
page_title: "dnscale_dnssec_status Data Source - DNScale"
subcategory: ""
description: |-
  Retrieves the DNSSEC status for a zone.
---

# dnscale_dnssec_status (Data Source)

Retrieves the DNSSEC status for a zone.

## Example Usage

```terraform
data "dnscale_dnssec_status" "example" {
  zone_id = dnscale_zone.example.id
}

output "dnssec_enabled" {
  value = data.dnscale_dnssec_status.example.enabled
}

output "has_ksk" {
  value = data.dnscale_dnssec_status.example.has_ksk
}
```

## Schema

### Required

- `zone_id` (String) - The zone ID to check DNSSEC status for.

### Read-Only

- `enabled` (Boolean) - Whether DNSSEC is enabled for the zone.
- `keys_count` (Number) - The number of DNSSEC keys configured.
- `has_ksk` (Boolean) - Whether a Key Signing Key (KSK) is configured.
- `has_zsk` (Boolean) - Whether a Zone Signing Key (ZSK) is configured.

Combined Signing Keys (CSK) are reported as both `has_ksk` and `has_zsk`.
