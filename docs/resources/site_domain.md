---
page_title: "pressable_site_domain Resource"
description: |-
  Attaches a domain to a Pressable site.
---

# pressable_site_domain

Attaches a domain through `/v1/sites/{site_id}/domains`.

## Example Usage

```terraform
resource "pressable_site_domain" "main" {
  site_id = pressable_site.main.id
  name    = "example.com"
  primary = true
}
```

