# terrasquid_acl_rule

Manage an ACL rule that links one or more sources to one or more destinations with a priority.

## Example Usage

```hcl
resource "terrasquid_acl_rule" "example" {
  name               = "allow-internal"
  priority           = 100
  comment            = "Allow internal access"
  sources            = [terrasquid_source_acl.example.id]
  destination_groups = [terrasquid_destination_group.example.id]
}
```

## Schema

### Required

- `name` (String) Unique name for this ACL rule.
- `sources` (Set of String) Source ACL IDs. At least one is required.

### Optional

- `destinations` (Set of String) Destination config IDs. Defaults to an empty set.
- `destination_groups` (Set of String) Destination group IDs. Defaults to an empty set.
- `priority` (Number) Rule priority. Lower values are evaluated first. Rules at the same priority are ordered by destination type (`DENY`, `CONNECT`, `ALLOW`), then creation time. Defaults to `100`.
- `comment` (String) Single-line comment emitted before the Squid access rules. Defaults to an empty string, which emits no comment.

At least one of `destinations` or `destination_groups` must contain an ID.

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
