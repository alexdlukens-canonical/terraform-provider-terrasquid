# terrasquid_acl_rule

Manage an ACL rule that links one or more sources to one or more destinations with a priority.

## Example Usage

```hcl
resource "terrasquid_acl_rule" "example" {
  name         = "allow-internal"
  priority     = 100
  sources      = [terrasquid_source_acl.example.id]
  destinations = [terrasquid_destination_config.example.id]
}
```

## Schema

### Required

- `name` (String) Unique name for this ACL rule.
- `sources` (List of String) Source ACL IDs. At least one is required.
- `destinations` (List of String) Destination config IDs. At least one is required.

### Optional

- `priority` (Number) Rule priority. Defaults to `100`.

### Read-Only

- `id` (String) Server-assigned UUID.
- `service` (String) Service namespace.
- `created_at` (String) Creation timestamp.
- `updated_at` (String) Last update timestamp.

## Import

Import using the UUID:

```bash
terraform import terrasquid_acl_rule.example <uuid>
```
