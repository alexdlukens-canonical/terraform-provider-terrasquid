resource "terrasquid_acl_rule" "example" {
  name         = "allow-internal"
  priority     = 100
  sources      = [terrasquid_source_acl.example.id]
  destinations = [terrasquid_destination_config.example.id]
}
