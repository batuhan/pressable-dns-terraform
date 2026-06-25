# Terraform Provider for Pressable

Terraform provider for the public Pressable API.

The provider has two layers:

- Typed resources/data sources for durable objects: imported sites, DNS records, site-scoped domains, site settings, zones, and site lists.
- Generic API primitives for complete day-one coverage of the public API: `pressable_api_request` and `pressable_api_action`.

Use typed resources wherever possible. Use the generic primitives for public endpoints that are operational, read-only, or not yet worth a dedicated Terraform abstraction.

## Provider Address

```hcl
terraform {
  required_providers {
    pressable = {
      source  = "batuhan/pressable"
      version = "~> 0.1"
    }
  }
}
```

The GitHub repository must be named `terraform-provider-pressable` for Terraform Registry publishing.

## Authentication

Use either a bearer token:

```hcl
provider "pressable" {
  access_token = var.pressable_access_token
}
```

or a Pressable API application:

```hcl
provider "pressable" {
  client_id     = var.pressable_client_id
  client_secret = var.pressable_client_secret
}
```

Environment variables are also supported:

- `PRESSABLE_BASE_URL`
- `PRESSABLE_ACCESS_TOKEN`
- `PRESSABLE_CLIENT_ID`
- `PRESSABLE_CLIENT_SECRET`

## Resources

- `pressable_site` (import-first; creation and remote deletion require explicit opt-ins)
- `pressable_dns_record`
- `pressable_site_domain`
- `pressable_site_setting`
- `pressable_api_action`

## Data Sources

- `pressable_api_request`
- `pressable_sites`
- `pressable_zones`
- `pressable_zone_records`

## Public API Coverage

Typed resources intentionally cover durable state. The generic primitives cover every public API path exposed by Pressable:

```hcl
data "pressable_api_request" "php_versions" {
  path = "/v1/sites/php-versions"
}

resource "pressable_api_action" "purge_cache" {
  method = "DELETE"
  path   = "/v1/sites/${pressable_site.main.id}/cache"

  triggers_json = jsonencode({
    version = var.deploy_version
  })
}
```

## Existing Site Workflow

The provider is optimized for existing Pressable sites. Import the site, then manage DNS, site domains, and settings around it:

```sh
terraform import pressable_site.main 12345
terraform import pressable_site_domain.main 12345/67890
terraform import pressable_dns_record.www 111/222
terraform import pressable_site_setting.maintenance 12345/maintenance-mode
```

Import IDs:

| Resource | Import ID |
| --- | --- |
| `pressable_site` | `site_id` |
| `pressable_site_domain` | `site_id/domain_id` |
| `pressable_dns_record` | `zone_id/record_id` |
| `pressable_site_setting` | `site_id/setting` |

`pressable_site` does not create sites unless `allow_create = true`. Destroying an imported `pressable_site` removes it from Terraform state only; set `delete_on_destroy = true` to delete the remote Pressable site.

## Local Development

```sh
make test
make install
```

Then use:

```hcl
terraform {
  required_providers {
    pressable = {
      source = "batuhan/pressable"
    }
  }
}
```

For a development override, add a Terraform CLI config:

```hcl
provider_installation {
  dev_overrides {
    "batuhan/pressable" = "/Users/batuhan/Projects/terraform-provider-pressable/bin"
  }
  direct {}
}
```

## Release

Registry releases are built with GoReleaser from `v*` tags. Configure these
GitHub Actions secrets first:

- `GPG_PRIVATE_KEY`
- `GPG_PASSPHRASE`
- `GPG_FINGERPRINT`

Then tag and push:

```sh
git tag v0.1.0
git push origin v0.1.0
```

After the signed GitHub release exists, publish `batuhan/pressable` through the
Terraform Registry UI.
