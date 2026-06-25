terraform {
  required_providers {
    pressable = {
      source = "batuhan/pressable"
    }
  }
}

provider "pressable" {}

resource "pressable_site" "main" {
  # Import with:
  # terraform import pressable_site.main 12345
  name = "existing-site"
}

resource "pressable_site_domain" "apex" {
  # Pressable domains belong to a site.
  # Import with:
  # terraform import pressable_site_domain.apex 12345/67890
  site_id = pressable_site.main.id
  name    = "example.com"
  primary = true
}

data "pressable_zones" "all" {}

resource "pressable_dns_record" "www" {
  # Import existing records with:
  # terraform import pressable_dns_record.www <zone_id>/<record_id>
  zone_id = 111
  type    = "CNAME"
  name    = "www"
  value   = "example.com."
  ttl     = 3600
}

resource "pressable_site_setting" "wordpress_mcp" {
  # Import with:
  # terraform import pressable_site_setting.wordpress_mcp 12345/wordpress-mcp
  site_id = pressable_site.main.id
  setting = "wordpress-mcp"

  request_body = jsonencode({
    enabled = true
  })
}

