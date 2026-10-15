# terrasquid_destination_config

Manage a destination configuration (domain, CIDR, wildcard with ALLOW/DENY/CONNECT action).

## Example Usage

```hcl
resource "terrasquid_destination_config" "example" {
  name       = "internal-domains"
  dst        = "*.internal.example.com"
  type       = "ALLOW"
  ports      = [80, 443]
  comment    = "Internal web services"
}
```

## Schema

### Required

- `name` (String) Unique name for this destination configuration.
- `dst` (String) Destination address (domain, wildcard, or CIDR).
- `type` (String) Action type: `ALLOW`, `DENY`, or `CONNECT`.

### Optional

- `ports` (List of Number) Individual integer ports between `1` and `65535`, combined with `port_ranges`. When both fields are unset or empty, defaults to `443` for `CONNECT`, otherwise `80`.
- `port_ranges` (List of String) Optional inclusive ranges in `start-end` format, such as `["30000-32767"]`. Endpoints must be between `1` and `65535`, with start less than or equal to end. Defaults to unset (`null`). Ranges are emitted directly in Squid port ACLs, not expanded into individual ports.
- `comment` (String) Single-line comment emitted before the Squid destination ACLs. Defaults to an empty string, which emits no comment.

Existing configurations and state do not need changes. Set `port_ranges` alongside
`ports`, or use ranges alone without adding an implicit default port. Set
`port_ranges = []` or remove the attribute to clear previously configured ranges.

### Read-Only

- `id` (String) Server-assigned UUID.
- `service` (String) Service namespace.
- `created_at` (String) Creation timestamp.
- `updated_at` (String) Last update timestamp.

## Import

Import using the UUID:

```bash
terraform import terrasquid_destination_config.example <uuid>
```
