# Terraform Provider Terrasquid

The Terrasquid provider allows you to manage resources for Terrasquid (Squid-as-a-Service), including Source/Destination groups, ACL rules, and configuration settings.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25.8 (to build the provider plugin)

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the `GNUmakefile`:

```shell
make default
```

This will produce a `terraform-provider-terrasquid` binary in the current directory.

## Using the Provider

```hcl
provider "terrasquid" {
  endpoint = "https://terrasquid.example.com"
  api_key  = var.terrasquid_api_key
}

# Example: Define a source group
resource "terrasquid_source_group" "example" {
  name        = "internal-networks"
  description = "Internal network ranges"
  cidrs       = ["192.168.1.0/24", "10.0.0.0/8"]
}
```

### Authentication

The provider requires an endpoint and an API key. These can be provided in the provider block or via environment variables:

- `TERRASQUID_ENDPOINT`
- `TERRASQUID_API_KEY`

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](https://golang.org/doc/install) installed on your machine (see [Requirements](#requirements)).

To compile the provider, run `make default`. This will build the binary and put it in the root directory.

```shell
make default
```

### Testing

To run the full suite of unit tests, run:

```shell
make test
```

To run acceptance tests (which interact with a real or simulated API), set the `TF_ACC` environment variable and run:

```shell
make testacc
```

Note: Acceptance tests require `TERRASQUID_ENDPOINT` and `TERRASQUID_API_KEY` to be set.

### Linting & Formatting

```shell
make lint
make fmt
make vet
```
