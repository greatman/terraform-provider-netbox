package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"testing"
)

func TestAccNetboxContactRoleV6_basic(t *testing.T) {
	testSlug := "contact_role_basic"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	randomDescription := testAccGetTestName(testSlug)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_contact_role" "test" {
	name = "%s"
	slug = "%s"
	description = "%s"
}
`, testName, randomSlug, randomDescription),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("netbox_contact_role.test", tfjsonpath.New("name"), knownvalue.StringExact(testName)),
					statecheck.ExpectKnownValue("netbox_contact_role.test", tfjsonpath.New("slug"), knownvalue.StringExact(randomSlug)),
					statecheck.ExpectKnownValue("netbox_contact_role.test", tfjsonpath.New("description"), knownvalue.StringExact(randomDescription)),
				},
			},
			{
				ResourceName:      "netbox_contact_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_contact_role" "test" {
	name = "%s_updated"
	slug = "%s_updated"
	description = "%s_updated"
}
`, testName, randomSlug, randomDescription),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("netbox_contact_role.test", tfjsonpath.New("name"), knownvalue.StringExact(testName+"_updated")),
					statecheck.ExpectKnownValue("netbox_contact_role.test", tfjsonpath.New("slug"), knownvalue.StringExact(randomSlug+"_updated")),
					statecheck.ExpectKnownValue("netbox_contact_role.test", tfjsonpath.New("description"), knownvalue.StringExact(randomDescription+"_updated")),
				},
			},
		},
	})
}

func TestAccNetboxContactRole_Tag(t *testing.T) {
	testPrefix := "contact_role_tag"
	testNameContactRole := testAccGetTestName(testPrefix)
	testSlugContactRole := testAccGetTestName(testPrefix)
	testNameTag := testAccGetTestName(testPrefix)
	testSlugTag := testAccGetTestName(testPrefix)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%s"
  slug = "%s"
  color_hex = "112233"
  description = "This is a test"
}
resource "netbox_contact_role" "test" {
	name = "%s"
	slug = "%s"
	tags = [netbox_tag.test.id]
}
`, testNameTag, testSlugTag, testNameContactRole, testSlugContactRole),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValueCollection("netbox_contact_role.test", []tfjsonpath.Path{
						tfjsonpath.New("tags"),
					}, "netbox_tag.test", tfjsonpath.New("id"), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccNetboxContactRole_CustomField(t *testing.T) {
	testPrefix := "region_customfield"
	customFieldName := testAccGetTestCustomFieldName(testPrefix)
	contactRoleName := testAccGetTestName(testPrefix)
	contactRoleSlug := testAccGetTestName(testPrefix)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name = "%s"
  type = "text"
  content_types = ["tenancy.contactrole"]
}
resource "netbox_contact_role" "test" {
	name = "%s"
	slug = "%s"
	custom_fields = {
	"${netbox_custom_field.test.name}" : "testcustomfield"
	}
}
`, customFieldName, contactRoleName, contactRoleSlug),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"netbox_contact_role.test",
						tfjsonpath.New("custom_fields").AtMapKey(customFieldName), knownvalue.StringExact("testcustomfield")),
				},
			},
		},
	})
}
