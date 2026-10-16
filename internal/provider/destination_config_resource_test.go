package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDestinationConfigPortRangesState(t *testing.T) {
	ctx := context.Background()
	empty, diags := types.ListValueFrom(ctx, types.StringType, []string{})
	if diags.HasError() {
		t.Fatal(diags)
	}
	for _, test := range []struct {
		name     string
		current  types.List
		values   []string
		wantNull bool
	}{
		{"legacy absent", types.ListNull(types.StringType), nil, true},
		{"server cleared", types.ListNull(types.StringType), []string{}, true},
		{"explicit empty", empty, nil, false},
		{"import ranges", types.ListNull(types.StringType), []string{"30000-32767"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, diags := destinationConfigStringListValue(ctx, test.current, test.values)
			if diags.HasError() {
				t.Fatal(diags)
			}
			if value.IsNull() != test.wantNull {
				t.Errorf("IsNull = %v, want %v", value.IsNull(), test.wantNull)
			}
			if len(value.Elements()) != len(test.values) {
				t.Errorf("elements = %v, want %v", value.Elements(), test.values)
			}
		})
	}
}

func TestAccDestinationConfigResource_portRanges(t *testing.T) {
	srv, _ := newMockServer(t)
	t.Setenv("TERRASQUID_ENDPOINT", srv.URL)
	t.Setenv("TERRASQUID_API_KEY", "valid-key")
	config := func(ranges string) string {
		return testAccProviderConfig() + fmt.Sprintf(`
resource "terrasquid_destination_config" "test" {
	name = "ranges"
	dst = "example.com"
	type = "ALLOW"
	ports = [443]
	%s
}
`, ranges)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config(""),
				Check:  resource.TestCheckNoResourceAttr("terrasquid_destination_config.test", "port_ranges.#"),
			},
			{
				Config: config(`port_ranges = ["30000-32767", "65535-65535"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "ports.0", "443"),
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "port_ranges.#", "2"),
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "port_ranges.0", "30000-32767"),
				),
			},
			{
				ResourceName: "terrasquid_destination_config.test", ImportState: true, ImportStateVerify: true,
			},
			{
				Config: config(`port_ranges = ["8080-8090"]`),
				Check:  resource.TestCheckResourceAttr("terrasquid_destination_config.test", "port_ranges.0", "8080-8090"),
			},
			{
				Config: config(`port_ranges = []`),
				Check:  resource.TestCheckResourceAttr("terrasquid_destination_config.test", "port_ranges.#", "0"),
			},
			{Config: config(`port_ranges = ["30000-32767"]`)},
			{
				Config: config(""),
				Check:  resource.TestCheckNoResourceAttr("terrasquid_destination_config.test", "port_ranges.#"),
			},
			{
				ResourceName: "terrasquid_destination_config.test", ImportState: true, ImportStateVerify: true,
			},
		},
	})
}

func TestAccDestinationConfigResource_invalidPortRestrictions(t *testing.T) {
	for _, restriction := range []string{
		"ports = [0]", "ports = [65536]", `port_ranges = ["0-80"]`,
		`port_ranges = ["80-65536"]`, `port_ranges = ["443-80"]`, `port_ranges = ["80"]`,
	} {
		t.Run(restriction, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: testAccProviderConfig() + fmt.Sprintf(`
resource "terrasquid_destination_config" "test" {
	name = "invalid"
	dst = "example.com"
	type = "ALLOW"
	%s
}
`, restriction),
					ExpectError: regexp.MustCompile(`Invalid (Port Range|Attribute Value)`),
				}},
			})
		})
	}
}

func TestAccDestinationConfigResource_basic(t *testing.T) {
	srv, _ := newMockServer(t)
	t.Setenv("TERRASQUID_ENDPOINT", srv.URL)
	t.Setenv("TERRASQUID_API_KEY", "valid-key")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "terrasquid_destination_config" "test" {
	name    = "test-dest"
	dst     = "192.168.1.1"
	type    = "ALLOW"
	ports   = [80, 443]
	comment = "Internal web service"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "name", "test-dest"),
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "dst", "192.168.1.1"),
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "type", "ALLOW"),
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "ports.#", "2"),
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "ports.0", "80"),
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "comment", "Internal web service"),
					resource.TestCheckResourceAttrSet("terrasquid_destination_config.test", "id"),
				),
			},
			{
				ResourceName:      "terrasquid_destination_config.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDestinationConfigResource_update(t *testing.T) {
	srv, _ := newMockServer(t)
	t.Setenv("TERRASQUID_ENDPOINT", srv.URL)
	t.Setenv("TERRASQUID_API_KEY", "valid-key")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "terrasquid_destination_config" "test" {
  name = "test-dest"
  dst  = "192.168.1.1"
  type = "ALLOW"
}
`,
			},
			{
				Config: testAccProviderConfig() + `
resource "terrasquid_destination_config" "test" {
	name    = "test-dest"
	dst     = "10.0.0.1"
	type    = "DENY"
	ports   = [22]
	comment = "Block SSH"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "dst", "10.0.0.1"),
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "type", "DENY"),
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "ports.#", "1"),
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "ports.0", "22"),
					resource.TestCheckResourceAttr("terrasquid_destination_config.test", "comment", "Block SSH"),
				),
			},
			{
				ResourceName:      "terrasquid_destination_config.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDestinationConfigResource_rejectsAsteriskWildcard(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "terrasquid_destination_config" "test" {
	name = "test-dest"
	dst  = "*.example.com"
	type = "ALLOW"
}
`,
				ExpectError: regexp.MustCompile(`Squid wildcard domains use a leading dot`),
			},
		},
	})
}
