---
page_title: "pressable_api_action Resource"
description: |-
  Runs any Pressable public API mutation.
---

# pressable_api_action

Runs any public Pressable API mutation. Prefer typed resources for durable objects.

## Example Usage

```terraform
resource "pressable_api_action" "purge_cache" {
  method = "DELETE"
  path   = "/v1/sites/${pressable_site.main.id}/cache"

  triggers_json = jsonencode({
    deploy = var.deploy_id
  })
}
```

