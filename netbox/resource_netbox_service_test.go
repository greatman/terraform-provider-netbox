package netbox

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/ipam"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxServiceFullDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_cluster_type" "test" {
  name = "%[1]s"
}

resource "netbox_cluster" "test" {
  name = "%[1]s"
  cluster_type_id = netbox_cluster_type.test.id
}

resource "netbox_virtual_machine" "test" {
  name = "%[1]s"
  cluster_id = netbox_cluster.test.id
}

`, testName)
}

func TestAccNetboxService_basic(t *testing.T) {
	testSlug := "svc_basic"
	testName := testAccGetTestName(testSlug)
	testDescription := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxServiceFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_service" "test" {
  name = "%s"
  virtual_machine_id = netbox_virtual_machine.test.id
  ports = [666]
  protocol = "tcp"
  description = "%s"
}`, testName, testDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_service.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_service.test", "description", testDescription),
					resource.TestCheckResourceAttrPair("netbox_service.test", "virtual_machine_id", "netbox_virtual_machine.test", "id"),
					resource.TestCheckResourceAttr("netbox_service.test", "ports.#", "1"),
					resource.TestCheckResourceAttr("netbox_service.test", "ports.0", "666"),
					resource.TestCheckResourceAttr("netbox_service.test", "protocol", "tcp"),
				),
			},
			{
				ResourceName:      "netbox_service.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetboxServiceFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_ip_address" "test" {
  ip_address = "1.1.1.1/24"
  status = "active"
}
resource "netbox_service" "test_ip" {
  name = "%s"
  virtual_machine_id = netbox_virtual_machine.test.id
  ports = [666]
  protocol = "tcp"
  ip_addresses = [netbox_ip_address.test.id]
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_service.test_ip", "ip_addresses.#", "1"),
					resource.TestCheckResourceAttrPair("netbox_service.test_ip", "ip_addresses.0", "netbox_ip_address.test", "id"),
				),
			},
			{
				ResourceName:      "netbox_service.test_ip",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxService_customFields(t *testing.T) {
	testSlug := "svc_custom_fields"
	testName := testAccGetTestName(testSlug)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxServiceFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = "custom_field"
  type          = "text"
  content_types = ["ipam.service"]
}
resource "netbox_service" "test_customfield" {
  name = "%s"
  virtual_machine_id = netbox_virtual_machine.test.id
  ports = [333]
  protocol = "tcp"
  custom_fields = {"${netbox_custom_field.test.name}" = "testtext"}
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_service.test_customfield", "name", testName),
					resource.TestCheckResourceAttrPair("netbox_service.test_customfield", "virtual_machine_id", "netbox_virtual_machine.test", "id"),
					resource.TestCheckResourceAttr("netbox_service.test_customfield", "ports.#", "1"),
					resource.TestCheckResourceAttr("netbox_service.test_customfield", "ports.0", "333"),
					resource.TestCheckResourceAttr("netbox_service.test_customfield", "protocol", "tcp"),
					resource.TestCheckResourceAttr("netbox_service.test_customfield", "custom_fields.custom_field", "testtext"),
				),
			},
			{
				ResourceName:      "netbox_service.test_customfield",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckServiceDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testAccProvider.Meta().(*client.NetBoxAPI)

	// loop through the resources in state, verifying each service
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "netbox_service" {
			continue
		}

		// Retrieve our service by referencing it's state ID for API lookup
		stateID, _ := strconv.ParseInt(rs.Primary.ID, 10, 64)
		params := ipam.NewIpamServicesReadParams().WithID(stateID)
		_, err := conn.Ipam.IpamServicesRead(params, nil)

		if err == nil {
			return fmt.Errorf("service (%s) still exists", rs.Primary.ID)
		}

		if err != nil {
			if errresp, ok := err.(*ipam.IpamServicesReadDefault); ok {
				errorcode := errresp.Code()
				if errorcode == 404 {
					return nil
				}
			}
			return err
		}
	}
	return nil
}

func init() {
	resource.AddTestSweepers("netbox_service", &resource.Sweeper{
		Name:         "netbox_service",
		Dependencies: []string{},
		F: func(region string) error {
			m, err := sharedClientForRegion(region)
			if err != nil {
				return fmt.Errorf("Error getting client: %s", err)
			}
			api := m.(*client.NetBoxAPI)
			params := ipam.NewIpamServicesListParams()
			res, err := api.Ipam.IpamServicesList(params, nil)
			if err != nil {
				return err
			}
			for _, intrface := range res.GetPayload().Results {
				if strings.HasPrefix(*intrface.Name, testPrefix) {
					deleteParams := ipam.NewIpamServicesDeleteParams().WithID(intrface.ID)
					_, err := api.Ipam.IpamServicesDelete(deleteParams, nil)
					if err != nil {
						return err
					}
					log.Print("[DEBUG] Deleted an interface")
				}
			}
			return nil
		},
	})
}
