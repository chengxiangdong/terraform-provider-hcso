package workspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/workspace/v2/services"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"

	"github.com/huaweicloud/terraform-provider-hcso/internal/services/acceptance"
)

func getServiceFunc(conf *config.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.WorkspaceV2Client(acceptance.HCSO_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating Workspace v2 client: %s", err)
	}
	resp, err := services.Get(client)
	if resp.Status == "CLOSED" {
		return nil, golangsdk.ErrDefault404{}
	}
	return resp, err
}

func TestAccService_basic(t *testing.T) {
	var (
		service      services.Service
		resourceName = "hcso_workspace_service.test"
		rName        = acceptance.RandomAccResourceNameWithDash()
	)

	rc := acceptance.InitResourceCheck(
		resourceName,
		&service,
		getServiceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acceptance.TestAccPreCheck(t)
		},
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccService_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "hcso_vpc.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "network_ids.0",
						"hcso_vpc_subnet.master", "id"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "LITE_AS"),
					resource.TestCheckResourceAttr(resourceName, "access_mode", "INTERNET"),
					resource.TestCheckResourceAttrSet(resourceName, "management_subnet_cidr"),
					resource.TestCheckResourceAttrSet(resourceName, "infrastructure_security_group.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "infrastructure_security_group.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "desktop_security_group.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "desktop_security_group.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "internet_access_port"),
					resource.TestCheckResourceAttrSet(resourceName, "internet_access_address"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
				),
			},
			{
				Config: testAccService_update(rName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPair(resourceName, "network_ids.0",
						"hcso_vpc_subnet.master", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "network_ids.1",
						"hcso_vpc_subnet.standby", "id"),
					resource.TestCheckResourceAttr(resourceName, "internet_access_port", "9001"),
					resource.TestCheckResourceAttrSet(resourceName, "internet_access_address"),
					resource.TestCheckResourceAttr(resourceName, "enterprise_id", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccService_localAD(t *testing.T) {
	var (
		service      services.Service
		resourceName = "hcso_workspace_service.test"
		rName        = acceptance.RandomAccResourceNameWithDash()
	)

	rc := acceptance.InitResourceCheck(
		resourceName,
		&service,
		getServiceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acceptance.TestAccPreCheck(t)
			acceptance.TestAccPreCheckWorkspaceAD(t)
		},
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccService_localAD_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "LOCAL_AD"),
					resource.TestCheckResourceAttr(resourceName, "ad_domain.0.name", acceptance.HCSO_WORKSPACE_AD_DOMAIN_NAME),
					resource.TestCheckResourceAttr(resourceName, "ad_domain.0.admin_account", "Administrator"),
					resource.TestCheckResourceAttr(resourceName, "ad_domain.0.password", acceptance.HCSO_WORKSPACE_AD_SERVER_PWD),
					resource.TestCheckResourceAttr(resourceName, "ad_domain.0.active_domain_ip", acceptance.HCSO_WORKSPACE_AD_DOMAIN_IP),
					resource.TestCheckResourceAttr(resourceName, "ad_domain.0.active_domain_name",
						fmt.Sprintf("server.%s", acceptance.HCSO_WORKSPACE_AD_DOMAIN_NAME)),
					resource.TestCheckResourceAttr(resourceName, "ad_domain.0.active_dns_ip", acceptance.HCSO_WORKSPACE_AD_DOMAIN_IP),
					resource.TestCheckResourceAttr(resourceName, "access_mode", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, "vpc_id", acceptance.HCSO_WORKSPACE_AD_VPC_ID),
					resource.TestCheckResourceAttr(resourceName, "network_ids.0", acceptance.HCSO_WORKSPACE_AD_NETWORK_ID),
					resource.TestCheckResourceAttrSet(resourceName, "infrastructure_security_group.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "infrastructure_security_group.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "desktop_security_group.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "desktop_security_group.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "internet_access_port"),
					resource.TestCheckResourceAttrSet(resourceName, "internet_access_address"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
				),
			},
			{
				Config: testAccService_localAD_update(rName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "network_ids.0", acceptance.HCSO_WORKSPACE_AD_NETWORK_ID),
					resource.TestCheckResourceAttrPair(resourceName, "network_ids.1",
						"hcso_vpc_subnet.master", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "network_ids.2",
						"hcso_vpc_subnet.standby", "id"),
					resource.TestCheckResourceAttr(resourceName, "internet_access_port", "9001"),
					resource.TestCheckResourceAttrSet(resourceName, "internet_access_address"),
					resource.TestCheckResourceAttr(resourceName, "enterprise_id", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ad_domain.0.password",
				},
			},
		},
	})
}

func testAccService_base(rName string) string {
	return fmt.Sprintf(`
resource "hcso_vpc" "test" {
  name = "%[1]s"
  cidr = "192.168.0.0/20"
}

resource "hcso_vpc_subnet" "master" {
  vpc_id = hcso_vpc.test.id

  name       = "%[1]s-master"
  cidr       = cidrsubnet(hcso_vpc.test.cidr, 4, 1)
  gateway_ip = cidrhost(cidrsubnet(hcso_vpc.test.cidr, 4, 1), 1)
}

resource "hcso_vpc_subnet" "standby" {
  vpc_id = hcso_vpc.test.id

  name       = "%[1]s-standby"
  cidr       = cidrsubnet(hcso_vpc.test.cidr, 4, 2)
  gateway_ip = cidrhost(cidrsubnet(hcso_vpc.test.cidr, 4, 2), 1)
}
`, rName)
}

func testAccService_basic(rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "hcso_workspace_service" "test" {
  access_mode = "INTERNET"
  vpc_id      = hcso_vpc.test.id
  network_ids = [
    hcso_vpc_subnet.master.id,
  ]
}
`, testAccService_base(rName))
}

func testAccService_update(rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "hcso_workspace_service" "test" {
  access_mode = "INTERNET"
  vpc_id      = hcso_vpc.test.id
  network_ids = [
    hcso_vpc_subnet.master.id,
    hcso_vpc_subnet.standby.id,
  ]

  internet_access_port = 9001
  enterprise_id        = "%[2]s"
}
`, testAccService_base(rName), rName)
}

func testAccService_localAD_base(rName string) string {
	return fmt.Sprintf(`
data "hcso_vpc" "test" {
  id = "%[1]s"
}

resource "hcso_vpc_subnet" "master" {
  vpc_id = "%[1]s"

  name       = "%[2]s-master"
  cidr       = cidrsubnet(data.hcso_vpc.test.cidr, 4, 1)
  gateway_ip = cidrhost(cidrsubnet(data.hcso_vpc.test.cidr, 4, 1), 1)
}

resource "hcso_vpc_subnet" "standby" {
  vpc_id = "%[1]s"

  name       = "%[2]s-standby"
  cidr       = cidrsubnet(data.hcso_vpc.test.cidr, 4, 2)
  gateway_ip = cidrhost(cidrsubnet(data.hcso_vpc.test.cidr, 4, 2), 1)
}
`, acceptance.HCSO_WORKSPACE_AD_VPC_ID, rName)
}

func testAccService_localAD_basic(rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "hcso_workspace_service" "test" {
  ad_domain {
    name               = "%[2]s"
    admin_account      = "Administrator"
    password           = "%[3]s"
    active_domain_ip   = "%[4]s"
    active_domain_name = "server.%[2]s"
    active_dns_ip      = "%[4]s"
  }

  auth_type   = "LOCAL_AD"
  access_mode = "INTERNET"
  vpc_id      = "%[5]s"
  network_ids = ["%[6]s"]
}
`, testAccService_localAD_base(rName), acceptance.HCSO_WORKSPACE_AD_DOMAIN_NAME, acceptance.HCSO_WORKSPACE_AD_SERVER_PWD,
		acceptance.HCSO_WORKSPACE_AD_DOMAIN_IP, acceptance.HCSO_WORKSPACE_AD_VPC_ID, acceptance.HCSO_WORKSPACE_AD_NETWORK_ID)
}

func testAccService_localAD_update(rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "hcso_workspace_service" "test" {
  depends_on = [
    hcso_vpc_subnet.master,
	hcso_vpc_subnet.standby,
  ]

  ad_domain {
    name               = "%[2]s"
    admin_account      = "Administrator"
    password           = "%[3]s"
    active_domain_ip   = "%[4]s"
    active_domain_name = "server.%[2]s"
    active_dns_ip      = "%[4]s"
  }

  auth_type   = "LOCAL_AD"
  access_mode = "INTERNET"
  vpc_id      = "%[5]s"
  network_ids = [
    "%[6]s",
    hcso_vpc_subnet.master.id,
    hcso_vpc_subnet.standby.id,
  ]

  internet_access_port = 9001
  enterprise_id        = "%[7]s"
}
`, testAccService_localAD_base(rName), acceptance.HCSO_WORKSPACE_AD_DOMAIN_NAME, acceptance.HCSO_WORKSPACE_AD_SERVER_PWD,
		acceptance.HCSO_WORKSPACE_AD_DOMAIN_IP, acceptance.HCSO_WORKSPACE_AD_VPC_ID, acceptance.HCSO_WORKSPACE_AD_NETWORK_ID,
		rName)
}