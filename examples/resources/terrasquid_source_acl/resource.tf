resource "terrasquid_source_acl" "example" {
  name = "office-network"
  cidr = ["10.0.0.0/8", "192.168.1.0/24"]
}
