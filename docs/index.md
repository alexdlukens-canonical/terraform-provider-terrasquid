# Terrasquid Provider

The Terrasquid provider is used to manage Terrasquid (Squid-as-a-Service) resources.

## Example Usage

```hcl
provider "terrasquid" {
  endpoint = "https://terrasquid.example.com"
  api_key  = var.terrasquid_api_key
}
```

## Schema

### Optional

- `endpoint` (String) Terrasquid API endpoint URL. Falls back to `TERRASQUID_ENDPOINT` environment variable.
- `api_key` (String, Sensitive) API key for authentication. Falls back to `TERRASQUID_API_KEY` environment variable.

## Resources

- `terrasquid_source_acl`
- `terrasquid_destination_config`
- `terrasquid_acl_rule`

## Data Sources

- `terrasquid_status`
