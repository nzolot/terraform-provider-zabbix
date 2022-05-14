package zabbix

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/nzolot/go-zabbix-api"
)

func TestAccZabbixLLDRule_Basic(t *testing.T) {
	strID := acctest.RandString(5)
	groupName := fmt.Sprintf("host_group_%s", strID)
	templateName := fmt.Sprintf("template_%s", strID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixLLDRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixLLDRuleConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "delay", "60"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "interface_id", "0"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "key", "key.lolo"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "name", "test_low_level_discovery_rule"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "type", "0"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "filter.#", "1"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "filter.3189296381.#", "0"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "filter.3189296381.condition.23998414.macro", "{#TESTMACRO}"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "filter.3189296381.condition.23998414.value", "^lo$"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "filter.3189296381.condition.23998414.operator", "8"),
				),
			},
			{
				Config: testAccZabbixLLDRuleUpdateConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "delay", "90"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "interface_id", "0"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "key", "key.update"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "name", "test_low_level_discovery_rule_update"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "type", "0"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "filter.#", "1"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "filter.1755271774.#", "0"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "filter.1755271774.condition.1739239139.macro", "{#UPDATE}"),
					resource.TestCheckResourceAttr("zabbix_lld_rule.lld_rule_test", "filter.1755271774.condition.1739239139.value", "^lo$"),
				),
			},
		},
	})
}

func testAccCheckZabbixLLDRuleDestroy(s *terraform.State) error {
	api := testAccProvider.Meta().(*zabbix.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zabbix_lld_rule" {
			continue
		}

		_, err := api.ItemGetByID(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("LLD rule still exist %s", rs.Primary.ID)
		}

		expectedError := "Expected exactly one result, got 0."
		if err.Error() != expectedError {
			return fmt.Errorf("expected error : %s, got : %s", expectedError, err.Error())
		}
	}
	return nil
}

func testAccZabbixLLDRuleConfig(groupName, templateName string) string {
	return fmt.Sprintf(`
		resource "zabbix_host_group" "zabbix" {
			name = "host group test %s"
		}

		resource "zabbix_template" "template_test" {
			host = "%s"
			groups = ["${zabbix_host_group.zabbix.name}"]
			name = "display name for template test %s"
	  	}


		resource "zabbix_lld_rule" "lld_rule_test" {
			delay = 60
			host_id = zabbix_template.template_test.id
			interface_id = "0"
			key = "key.lolo"
			name = "test_low_level_discovery_rule"
			type = 0
			filter {
				condition {
					macro = "{#TESTMACRO}"
					value = "^lo$"
				}
				eval_type = 0
			}
		}
	`, groupName, templateName, templateName)
}

func testAccZabbixLLDRuleUpdateConfig(groupName, templateName string) string {
	return fmt.Sprintf(`
		resource "zabbix_host_group" "zabbix" {
			name = "host group test %s"
		}

		resource "zabbix_template" "template_test" {
			host = "%s"
			groups = ["${zabbix_host_group.zabbix.name}"]
			name = "display name for template test %s"
	  	}

		resource "zabbix_lld_rule" "lld_rule_test" {
			delay = 90
			host_id = zabbix_template.template_test.id
			interface_id = "0"
			key = "key.update"
			name = "test_low_level_discovery_rule_update"
			type = 0
			filter {
				condition {
					macro = "{#UPDATE}"
					value = "^lo$"
				}
				eval_type = 0
			}
		}
	`, groupName, templateName, templateName)
}
