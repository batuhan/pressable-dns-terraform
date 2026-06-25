resource "pressable_site" "example" {
  # Import first:
  # terraform import pressable_site.example 12345
  name        = "example-site"
  php_version = "8.3"
}
