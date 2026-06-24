---
page_title: "Pressable Provider"
subcategory: ""
description: |-
  Manage Pressable sites, DNS, domains, site settings, and public API operations.
---

# Pressable Provider

The Pressable provider manages resources exposed by the public Pressable API.

Use typed resources for durable objects. Use `pressable_api_request` and `pressable_api_action` for public API endpoints that are read-only, operational, or not yet modeled as first-class Terraform resources.

## Example Usage

```terraform
provider "pressable" {
  client_id     = var.pressable_client_id
  client_secret = var.pressable_client_secret
}
```

## Schema

### Optional

- `base_url` (String) Pressable API base URL. Defaults to `https://my.pressable.com`.
- `access_token` (String, Sensitive) Pressable bearer token.
- `client_id` (String, Sensitive) Pressable API application client id.
- `client_secret` (String, Sensitive) Pressable API application client secret.

