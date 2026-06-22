---
page_title: "dnscale_dnssec_key Resource - DNScale"
subcategory: ""
description: |-
  Manages a DNSSEC cryptographic key in DNScale.
---

# dnscale_dnssec_key (Resource)

Manages a DNSSEC cryptographic key in DNScale.

DNSSEC (Domain Name System Security Extensions) adds cryptographic signatures to DNS records to protect against spoofing attacks. This resource allows you to create and manage the cryptographic keys used for DNSSEC signing.

## Example Usage

### Key Signing Key (KSK)

```terraform
resource "dnscale_dnssec_key" "ksk" {
  zone_id   = dnscale_zone.example.id
  key_type  = "KSK"
  algorithm = "ECDSAP256SHA256"
  active    = true
  published = true
}
```

### Zone Signing Key (ZSK)

```terraform
resource "dnscale_dnssec_key" "zsk" {
  zone_id   = dnscale_zone.example.id
  key_type  = "ZSK"
  algorithm = "ECDSAP256SHA256"
  active    = true
  published = true
}
```

### Combined Signing Key (CSK)

```terraform
resource "dnscale_dnssec_key" "csk" {
  zone_id   = dnscale_zone.example.id
  key_type  = "CSK"
  algorithm = "ECDSAP256SHA256"
  active    = true
  published = true
}
```

### Complete DNSSEC Setup

```terraform
resource "dnscale_zone" "example" {
  name   = "example.com"
  region = "EU"
  type   = "master"
}

resource "dnscale_dnssec_key" "ksk" {
  zone_id   = dnscale_zone.example.id
  key_type  = "KSK"
  algorithm = "ECDSAP256SHA256"
  active    = true
  published = true
}

resource "dnscale_dnssec_key" "zsk" {
  zone_id   = dnscale_zone.example.id
  key_type  = "ZSK"
  algorithm = "ECDSAP256SHA256"
  active    = true
  published = true
}

# Output the DS record for registrar configuration
output "ds_records" {
  value = dnscale_dnssec_key.ksk.ds
}
```

## Schema

### Required

- `zone_id` (String) - Zone UUID that this key belongs to.
- `key_type` (String) - Type of DNSSEC key. Valid values: `KSK` (Key Signing Key), `ZSK` (Zone Signing Key), `CSK` (Combined Signing Key).

### Optional

- `algorithm` (String) - DNSSEC algorithm. Valid values: `ECDSAP256SHA256`, `ECDSAP384SHA384`, `RSASHA256`, `RSASHA512`. Default: `ECDSAP256SHA256`.
- `bits` (Number) - Key size in bits. Only applicable for RSA algorithms.
- `active` (Boolean) - Whether the key is active for signing. Default: `true`.
- `published` (Boolean) - Whether the key is published in DNS. Default: `true`.

### Read-Only

- `id` (Number) - The unique identifier for the key.
- `key_tag` (Number) - The DNSSEC key tag.
- `dnskey` (String) - The DNSKEY record content.
- `ds` (List of String) - The DS record(s) for this key (only for KSK).

## Import

DNSSEC keys can be imported using the format `zone_id/key_id`:

```bash
terraform import dnscale_dnssec_key.ksk <zone-id>/<key-id>
```
