---
page_title: "dnscale_record Resource - DNScale"
subcategory: ""
description: |-
  Manages a DNS record in DNScale.
---

# dnscale_record (Resource)

Manages a DNS record in DNScale.

## Example Usage

### A Record

```terraform
resource "dnscale_record" "www" {
  zone_id = dnscale_zone.example.id
  name    = "www.example.com."
  type    = "A"
  content = "192.0.2.1"
  ttl     = 300
}
```

### AAAA Record

```terraform
resource "dnscale_record" "ipv6" {
  zone_id = dnscale_zone.example.id
  name    = "ipv6.example.com."
  type    = "AAAA"
  content = "2001:db8::1"
  ttl     = 300
}
```

### CNAME Record

```terraform
resource "dnscale_record" "cdn" {
  zone_id = dnscale_zone.example.id
  name    = "cdn.example.com."
  type    = "CNAME"
  content = "example.com."
  ttl     = 3600
}
```

### MX Record

```terraform
resource "dnscale_record" "mail" {
  zone_id  = dnscale_zone.example.id
  name     = "example.com."
  type     = "MX"
  content  = "mail.example.com."
  ttl      = 3600
  priority = 10
}
```

### TXT Record

```terraform
resource "dnscale_record" "spf" {
  zone_id = dnscale_zone.example.id
  name    = "example.com."
  type    = "TXT"
  content = "v=spf1 include:_spf.example.com -all"
  ttl     = 3600
}
```

### CAA Record

```terraform
resource "dnscale_record" "caa" {
  zone_id = dnscale_zone.example.id
  name    = "example.com."
  type    = "CAA"
  content = "0 issue \"letsencrypt.org\""
  ttl     = 3600
}
```

### SRV Record

```terraform
resource "dnscale_record" "srv" {
  zone_id  = dnscale_zone.example.id
  name     = "_sip._tcp.example.com."
  type     = "SRV"
  content  = "10 5 5060 sip.example.com."
  ttl      = 3600
  priority = 0
}
```

## Schema

### Required

- `zone_id` (String) - Zone UUID that this record belongs to.
- `name` (String) - Full record name with trailing dot (e.g., `www.example.com.`). Changing this value forces replacement.
- `type` (String) - DNS record type. Valid values: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `SRV`, `CAA`, `PTR`, `ALIAS`, `TLSA`, `SSHFP`, `HTTPS`, `SVCB`. Changing this value forces replacement.
- `content` (String) - Record value (IP address, hostname, text, etc.).

### Optional

- `ttl` (Number) - Time-to-live in seconds. Must be between 300 and 86400. Default: `3600`.
- `priority` (Number) - Priority for MX and SRV records. Must be between 0 and 65535.
- `disabled` (Boolean) - Whether the record is disabled. Default: `false`.

### Read-Only

- `id` (String) - Record ID (base64 encoded).

## Import

Records can be imported using the format `zone_id/record_id`:

```bash
terraform import dnscale_record.www <zone-id>/<record-id>
```

Example:

```bash
terraform import dnscale_record.www 12345678-1234-1234-1234-123456789abc/d3d3LmV4YW1wbGUuY29tLg==
```
