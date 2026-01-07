# Assume zone already exists
data "dnscale_zone" "example" {
  name = "example.com"
}

# A record for root domain
resource "dnscale_record" "root" {
  zone_id = data.dnscale_zone.example.id
  name    = "example.com."
  type    = "A"
  content = "192.0.2.1"
  ttl     = 300
}

# AAAA record for IPv6
resource "dnscale_record" "root_ipv6" {
  zone_id = data.dnscale_zone.example.id
  name    = "example.com."
  type    = "AAAA"
  content = "2001:db8::1"
  ttl     = 300
}

# CNAME record for www subdomain
resource "dnscale_record" "www" {
  zone_id = data.dnscale_zone.example.id
  name    = "www.example.com."
  type    = "CNAME"
  content = "example.com."
  ttl     = 3600
}

# MX records for mail
resource "dnscale_record" "mx_primary" {
  zone_id  = data.dnscale_zone.example.id
  name     = "example.com."
  type     = "MX"
  content  = "mail.example.com."
  ttl      = 3600
  priority = 10
}

resource "dnscale_record" "mx_secondary" {
  zone_id  = data.dnscale_zone.example.id
  name     = "example.com."
  type     = "MX"
  content  = "mail2.example.com."
  ttl      = 3600
  priority = 20
}

# TXT record for SPF
resource "dnscale_record" "spf" {
  zone_id = data.dnscale_zone.example.id
  name    = "example.com."
  type    = "TXT"
  content = "v=spf1 include:_spf.google.com ~all"
  ttl     = 3600
}

# TXT record for DKIM
resource "dnscale_record" "dkim" {
  zone_id = data.dnscale_zone.example.id
  name    = "google._domainkey.example.com."
  type    = "TXT"
  content = "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA..."
  ttl     = 3600
}

# CAA record
resource "dnscale_record" "caa" {
  zone_id = data.dnscale_zone.example.id
  name    = "example.com."
  type    = "CAA"
  content = "0 issue \"letsencrypt.org\""
  ttl     = 3600
}
