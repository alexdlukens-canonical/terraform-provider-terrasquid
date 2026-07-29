package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccACLRuleResource_basic(t *testing.T) {
	srv, _ := newMockServer(t)
	t.Setenv("TERRASQUID_ENDPOINT", srv.URL)
	t.Setenv("TERRASQUID_API_KEY", "valid-key")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "terrasquid_acl_rule" "test" {
  name         = "acl-rule"
  priority     = 100
	comment      = "Allow default traffic"
  sources      = ["src-1"]
  destinations = ["dst-1"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "name", "acl-rule"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "priority", "100"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "comment", "Allow default traffic"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "sources.#", "1"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "sources.0", "src-1"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "destinations.#", "1"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "destinations.0", "dst-1"),
					resource.TestCheckResourceAttrSet("terrasquid_acl_rule.test", "id"),
				),
			},
			{
				ResourceName:      "terrasquid_acl_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACLRuleResource_update(t *testing.T) {
	srv, _ := newMockServer(t)
	t.Setenv("TERRASQUID_ENDPOINT", srv.URL)
	t.Setenv("TERRASQUID_API_KEY", "valid-key")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "terrasquid_acl_rule" "test" {
  name         = "acl-rule"
  priority     = 100
  sources      = ["src-1"]
  destinations = ["dst-1"]
}
`,
			},
			{
				Config: testAccProviderConfig() + `
resource "terrasquid_acl_rule" "test" {
  name         = "acl-rule"
  priority     = 200
	comment      = "Allow updated traffic"
  sources      = ["src-2"]
  destinations = ["dst-2"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "priority", "200"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "comment", "Allow updated traffic"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "sources.0", "src-2"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "destinations.0", "dst-2"),
				),
			},
			{
				ResourceName:      "terrasquid_acl_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACLRuleResource_multipleSourcesAndDestinations(t *testing.T) {
	srv, _ := newMockServer(t)
	t.Setenv("TERRASQUID_ENDPOINT", srv.URL)
	t.Setenv("TERRASQUID_API_KEY", "valid-key")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "terrasquid_acl_rule" "test" {
  name         = "acl-rule"
  priority     = 150
  sources      = ["src-1", "src-2"]
  destinations = ["dst-1", "dst-2"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "priority", "150"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "sources.#", "2"),
					resource.TestCheckResourceAttr("terrasquid_acl_rule.test", "destinations.#", "2"),
				),
			},
			{
				ResourceName:      "terrasquid_acl_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
