resource "pressable_dns_record" "www" {
  zone_id = 123
  type    = "CNAME"
  name    = "www"
  value   = "example.com."
  ttl     = 3600
}

