resource "terrasquid_destination_config" "example" {
  name  = "internal-domains"
  dst   = ".internal.example.com"
  type  = "ALLOW"
  ports = [80, 443]
}
