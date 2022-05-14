package zabbix

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/nzolot/go-zabbix-api"
)

func TestAccZabbixItem_Basic(t *testing.T) {
	strID := acctest.RandString(5)
	groupName := fmt.Sprintf("host_group_%s", strID)
	templateName := fmt.Sprintf("template_%s", strID)
	itemName := fmt.Sprintf("item_%s", strID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixItemConfig(groupName, templateName, itemName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccZabbixItemExists("zabbix_item.my_item1"),
					resource.TestCheckResourceAttr("zabbix_item.my_item1", "name", itemName),
					resource.TestCheckResourceAttr("zabbix_item.my_item1", "key", "bilou.bilou"),
					resource.TestCheckResourceAttr("zabbix_item.my_item1", "delay", "15"),
					resource.TestCheckResourceAttr("zabbix_item.my_item1", "description", fmt.Sprintf("description for item : %s", itemName)),
					testCheckResourceAttrValueFunc("zabbix_item.my_item1", "trends", func(zabbixVersion string) string {
						return fmt.Sprintf("300%s", getZabbixServerUnitDays(zabbixVersion))
					}),
					testCheckResourceAttrValueFunc("zabbix_item.my_item1", "history", func(zabbixVersion string) string {
						return fmt.Sprintf("25%s", getZabbixServerUnitDays(zabbixVersion))
					}),
				),
			},
			{
				Config: testAccZabbixItemUpdate(groupName, templateName, itemName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccZabbixItemExists("zabbix_item.my_item1"),
					resource.TestCheckResourceAttr("zabbix_item.my_item1", "name", fmt.Sprintf("update_%s", itemName)),
					resource.TestCheckResourceAttr("zabbix_item.my_item1", "key", "update.bilou.bilou"),
					resource.TestCheckResourceAttr("zabbix_item.my_item1", "delay", "30"),
					resource.TestCheckResourceAttr("zabbix_item.my_item1", "description", fmt.Sprintf("update description for item : %s", itemName)),
					testCheckResourceAttrValueFunc("zabbix_item.my_item1", "trends", func(zabbixVersion string) string {
						return fmt.Sprintf("3%s", getZabbixServerUnitDays(zabbixVersion))
					}),
					testCheckResourceAttrValueFunc("zabbix_item.my_item1", "history", func(zabbixVersion string) string {
						return fmt.Sprintf("2%s", getZabbixServerUnitDays(zabbixVersion))
					}),
				),
			},
		},
	})
}

func testAccZabbixItemConfig(groupName, templateName, itemName string) string {
	return fmt.Sprintf(`
		data "zabbix_server" "test" {}

		resource "zabbix_host_group" "zabbix" {
			name = "%s"
		}

		resource "zabbix_template" "my_zbx_template" {
			host = "%s"
			groups = ["${zabbix_host_group.zabbix.name}"]
			name = "display name %s"
			description = "description for template %s"
	  	}

		resource "zabbix_item" "my_item1" {
			name = "%s"
			key = "bilou.bilou"
			delay = "15"
			description = "description for item : %s"
			trends = join("", ["300", data.zabbix_server.test.unit_time_days])
			history = join("", ["25", data.zabbix_server.test.unit_time_days])
			host_id = "${zabbix_template.my_zbx_template.id}"
	  	}
	`, groupName, templateName, templateName, templateName, itemName, itemName)
}

func testAccZabbixItemUpdate(groupName, templateName, itemName string) string {
	return fmt.Sprintf(`
		data "zabbix_server" "test" {}

		resource "zabbix_host_group" "zabbix" {
			name = "%s"
		}

		resource "zabbix_template" "my_zbx_template" {
			host = "%s"
			groups = ["${zabbix_host_group.zabbix.name}"]
			name = "display name %s"
			description = "description for template %s"
	  	}

		resource "zabbix_item" "my_item1" {
			name = "update_%s"
			key = "update.bilou.bilou"
			delay = "30"
			description = "update description for item : %s"
			trends = join("", ["3", data.zabbix_server.test.unit_time_days])
			history = join("", ["2", data.zabbix_server.test.unit_time_days])
			host_id = "${zabbix_template.my_zbx_template.id}"
	  	}
	`, groupName, templateName, templateName, templateName, itemName, itemName)
}

func testAccZabbixItemExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found : %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No record ID set")
		}
		return nil
	}
}

func testAccCheckZabbixItemDestroy(s *terraform.State) error {
	api := testAccProvider.Meta().(*zabbix.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zabbix_item" {
			continue
		}

		_, err := api.ItemGetByID(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Item still exists %s", rs.Primary.ID)
		}

		expectedError := "Expected exactly one result, got 0."
		if err.Error() != expectedError {
			return fmt.Errorf("expected error : %s, got : %s", expectedError, err.Error())
		}
	}
	return nil
}
