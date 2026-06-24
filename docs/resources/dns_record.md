---
page_title: "pressable_dns_record Resource"
description: |-
  Manages a DNS record in a Pressable zone.
---

# pressable_dns_record

Manages a DNS record through `/v1/zones/{zone_id}/records`.

The public API exposes create and delete semantics. Terraform updates replace the record.

## Example Usage

```terraform
resource "pressable_dns_record" "www" {
  zone_id = 123
  type    = "CNAME"
  name    = "www"
  value   = "example.com."
  ttl     = 3600
}
```

