# terrasquid_destination_group

Manage a named group of destination configurations.

## Example Usage

```hcl
resource "terrasquid_destination_config" "github" {
  name  = "github"
  dst   = "github.com"
  type  = "CONNECT"
  ports = [443]
}

resource "terrasquid_destination_group" "common" {
  name         = "common-sites-cloud-access"
  destinations = [terrasquid_destination_config.github.id]
  comment      = "Destinations commonly used by cloud workloads"
}
```

## Schema

### Required

- `name` (String) Unique name for this destination group.
- `destinations` (Set of String) IDs of the destination configurations in this group. At least one is required.

### Optional

- `comment` (String) Comment associated with the destination group. Defaults to an empty string.

### Read-Only

- `id` (String) Server-assigned UUID.
- `service` (String) Service namespace.
- `created_at` (String) Creation timestamp.
- `updated_at` (String) Last update timestamp.

## Import

Import using the UUID:

```bash
terraform import terrasquid_destination_group.example <uuid>
```