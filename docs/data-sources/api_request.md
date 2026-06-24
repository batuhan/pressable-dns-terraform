---
page_title: "pressable_api_request Data Source"
description: |-
  Reads any Pressable public API endpoint.
---

# pressable_api_request

Reads any public Pressable API endpoint and exposes the response envelope.

## Example Usage

```terraform
data "pressable_api_request" "php_versions" {
  path = "/v1/sites/php-versions"
}
```

