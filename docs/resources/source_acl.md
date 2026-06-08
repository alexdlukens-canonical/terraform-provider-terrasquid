# terrasquid_source_acl

Manage a source ACL (IP/CIDR-based access list).

## Example Usage

```hcl
resource "terrasquid_source_acl" "example" {
  name = "office-network"
  cidr = ["10.0.0.0/8", "192.168.1.0/24"]
}
```

## Schema

### Required

- `name` (String) Unique name for this source ACL.
- `cidr` (List of String) List of CIDR blocks.

## Import

Import using the UUID:

```bash
terraform import terrasquid_source_acl.example <uuid>
```
