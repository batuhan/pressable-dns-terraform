---
page_title: "pressable_site_setting Resource"
description: |-
  Applies a durable Pressable site setting endpoint.
---

# pressable_site_setting

Applies durable site settings such as:

- `cdn`
- `edge-cache`
- `basic-authentication`
- `maintenance-mode`
- `multisite_support`
- `light_weight_404`
- `php_fs_permissions`
- `wordpress-mcp`
- `woo-cart-cache`

Import format:

```sh
terraform import pressable_site_setting.maintenance <site_id>/maintenance-mode
```

## Example Usage

```terraform
resource "pressable_site_setting" "maintenance" {
  site_id = pressable_site.main.id
  setting = "maintenance-mode"

  request_body = jsonencode({
    enabled = true
  })
}
```
