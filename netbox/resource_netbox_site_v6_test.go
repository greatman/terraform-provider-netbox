package netbox

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"testing"
)

func TestAccNetboxSiteV6_basic(t *testing.T) {
	testSlug := testAccGetTestName("")
	testName := testAccGetTestName(testSlug)
	testSlug = testAccGetTestName(testSlug)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_2_0),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_site_group" "test" {
  name = "%[1]s"
}
resource "netbox_rir" "test" {
  name = "%[1]s"
}

resource "netbox_asn" "test" {
  asn = 1338
  rir_id = netbox_rir.test.id
}

resource "netbox_site" "test" {
  name = "%[1]s"
  slug = "%[2]s"
  status = "planned"
  description = "%[1]s"
  facility = "%[1]s"
  physical_address = "%[1]s"
  shipping_address = "%[1]s"
  asn_ids = [netbox_asn.test.id]
  group_id = netbox_site_group.test.id
}`, testName, testSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_site.test", "slug", testSlug),
					resource.TestCheckResourceAttr("netbox_site.test", "status", "planned"),
					resource.TestCheckResourceAttr("netbox_site.test", "description", testName),
					resource.TestCheckResourceAttr("netbox_site.test", "facility", testName),
					resource.TestCheckResourceAttr("netbox_site.test", "physical_address", testName),
					resource.TestCheckResourceAttr("netbox_site.test", "shipping_address", testName),
					resource.TestCheckResourceAttr("netbox_site.test", "asn_ids.#", "1"),
					resource.TestCheckResourceAttrPair("netbox_site.test", "asn_ids.0", "netbox_asn.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_site.test", "group_id", "netbox_site_group.test", "id"),
				),
			},
		},
	})
}

func TestAccNetboxSite_fieldUpdateV6(t *testing.T) {
	testSlug := "site_field_update"
	testName := testAccGetTestName(testSlug)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_site" "test" {
	name        = "%[2]s"
	slug		= "%[1]s"
	description = "Test site description"
	physical_address = "Physical address"
	shipping_address = "Shipping address"
	latitude      = "12.123456"
  	longitude     = "-13.123456"

}`, testSlug, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site.test", "description", "Test site description"),
					resource.TestCheckResourceAttr("netbox_site.test", "physical_address", "Physical address"),
					resource.TestCheckResourceAttr("netbox_site.test", "shipping_address", "Shipping address"),
					resource.TestCheckResourceAttr("netbox_site.test", "latitude", "12.123456"),
					resource.TestCheckResourceAttr("netbox_site.test", "longitude", "-13.123456"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_site" "test" {
	name = "%[1]s"
	slug = "%[2]s"
}`, testName, testSlug),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("netbox_site.test", tfjsonpath.New("description"), knownvalue.Null()),
					statecheck.ExpectKnownValue("netbox_site.test", tfjsonpath.New("physical_address"), knownvalue.Null()),
					statecheck.ExpectKnownValue("netbox_site.test", tfjsonpath.New("shipping_address"), knownvalue.Null()),
					statecheck.ExpectKnownValue("netbox_site.test", tfjsonpath.New("latitude"), knownvalue.Null()),
					statecheck.ExpectKnownValue("netbox_site.test", tfjsonpath.New("longitude"), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccSiteV6Config(configurableAttribute ...any) string {
	return fmt.Sprintf(`
resource "netbox_site" "test" {
	name = "%[1]s"
    slug = "%[2]s"
	status = "planned"
}
`, configurableAttribute...)
}
