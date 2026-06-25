---
page_title: "pressable_site Resource"
description: |-
  Adopts and manages an existing Pressable site.
---

# pressable_site

Adopts and manages an existing Pressable site through `/v1/sites`.

Creation is disabled by default. Use import for normal workflows:

```sh
terraform import pressable_site.main 12345
```

Remote deletion is also disabled by default. Destroying this resource removes it from Terraform state only unless `delete_on_destroy = true`.

## Example Usage

```terraform
resource "pressable_site" "main" {
  # Imported with: terraform import pressable_site.main 12345
  name        = "example-site"
  php_version = "8.3"
}
```

To intentionally create a new Pressable site:

```terraform
resource "pressable_site" "created" {
  allow_create = true
  name         = "example-site"
}
```
