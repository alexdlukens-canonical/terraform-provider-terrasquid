resource "terrasquid_acl_rule" "example" {
  name     = "allow-internal"
  priority = 100
  src      = terrasquid_source_acl.example.id
  dst      = terrasquid_destination_config.example.id
}
