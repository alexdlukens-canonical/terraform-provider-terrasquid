# terrasquid_destination_group

Look up a destination group by name.

## Example Usage

```hcl
data "terrasquid_destination_group" "common" {
  name = "common-sites-cloud-access"
}
```

## Schema

### Required

- `name` (String) Name of the destination group to look up.

### Read-Only

- `id` (String) Server-assigned UUID.
- `destinations` (Set of String) IDs of the destination configurations in the group.
- `comment` (String) Comment associated with the destination group.
- `service` (String) Service namespace.
- `created_at` (String) Creation timestamp.
- `updated_at` (String) Last update timestamp.