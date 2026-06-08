resource "terrasquid_destination_group" "example" {
  name         = "internal-services"
  destinations = [terrasquid_destination_config.example.id]
}
