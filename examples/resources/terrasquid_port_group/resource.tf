resource "terrasquid_port_group" "example" {
  name  = "web-ports"
  ports = [80, 443, 8080]
}
