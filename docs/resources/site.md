---
page_title: "pressable_site Resource"
description: |-
  Manages a Pressable site.
---

# pressable_site

Manages a Pressable site through `/v1/sites`.

## Example Usage

```terraform
resource "pressable_site" "main" {
  name        = "example-site"
  php_version = "8.3"
}
```

