package netbox

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"testing"
)

func TestAccNetboxVirtualMachineV6_basic(t *testing.T) {
	testSlug := "vm_machine"
	testName := testAccGetTestName(testSlug)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_2_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineV6Config(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.only_site", "name", testName),
				),
			},
		},
	})
}

func testAccVirtualMachineV6Config(configurableAttribute string) string {
	return fmt.Sprintf(`
resource "netbox_cluster_type" "test" {
  name = "%[1]s"
}

resource "netbox_cluster" "test" {
  name = "%[1]s"
  cluster_type_id = netbox_cluster_type.test.id
}

resource "netbox_virtual_machine" "only_site" {
  name = "%[1]s"
  status = "active"
cluster_id = netbox_cluster.test.id
}
`, configurableAttribute)
}
