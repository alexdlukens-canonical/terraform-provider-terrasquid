resource "terrasquid_source_group" "example" {
  name    = "trusted-sources"
  sources = [terrasquid_source_acl.example.id]
}
