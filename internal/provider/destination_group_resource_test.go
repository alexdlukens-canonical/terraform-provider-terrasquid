package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDestinationGroupResourceAndDataSource(t *testing.T) {
	srv, _ := newMockServer(t)
	t.Setenv("TERRASQUID_ENDPOINT", srv.URL)
	t.Setenv("TERRASQUID_API_KEY", "valid-key")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "terrasquid_destination_config" "github" {
	name  = "github"
	dst   = "github.com"
	type  = "CONNECT"
	ports = [443]
}

resource "terrasquid_destination_group" "common" {
  name         = "common-sites-cloud-access"
	destinations = [terrasquid_destination_config.github.id]
}

data "terrasquid_destination_group" "common" {
  name = terrasquid_destination_group.common.name
}

resource "terrasquid_acl_rule" "consumer" {
  name               = "allow-common-sites"
  sources            = ["local-source"]
  destination_groups = [data.terrasquid_destination_group.common.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("terrasquid_destination_group.common", "destinations.#", "1"),
					resource.TestCheckResourceAttrSet("data.terrasquid_destination_group.common", "id"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.consumer", "destination_groups.#", "1"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "terrasquid_destination_config" "github" {
	name  = "github"
	dst   = "github.com"
	type  = "CONNECT"
	ports = [443]
}

resource "terrasquid_destination_config" "google" {
	name  = "google"
	dst   = "google.com"
	type  = "CONNECT"
	ports = [443]
}

resource "terrasquid_destination_group" "common" {
	name = "common-sites-cloud-access"
	destinations = [
		terrasquid_destination_config.google.id,
		terrasquid_destination_config.github.id,
	]
}

data "terrasquid_destination_group" "common" {
	name = terrasquid_destination_group.common.name
}

resource "terrasquid_acl_rule" "consumer" {
	name               = "allow-common-sites"
	sources            = ["local-source"]
	destinations       = [terrasquid_destination_config.github.id]
	destination_groups = [data.terrasquid_destination_group.common.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("terrasquid_destination_group.common", "destinations.#", "2"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.consumer", "destinations.#", "1"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.consumer", "destination_groups.#", "1"),
				),
			},
		},
	})
}
