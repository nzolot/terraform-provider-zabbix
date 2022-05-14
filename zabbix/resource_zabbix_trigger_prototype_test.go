package zabbix

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/nzolot/go-zabbix-api"
)

func TestAccZabbixTriggerPrototype_Basic(t *testing.T) {
	strID := acctest.RandString(5)
	groupName := fmt.Sprintf("host_group_%s", strID)
	templateName := fmt.Sprintf("template_%s", strID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixItemPrototypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixTriggerPrototypeConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "description", "trigger_prototype_test"),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "expression", fmt.Sprintf("{%s:test.key.last()}=0", templateName)),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "priority", "5"),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "status", "0"),
				),
			},
			{
				Config: testAccZabbixTriggerPrototypeUpdateConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "description", "trigger_prototype_test_update"),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "expression", fmt.Sprintf("{%s:test.key.last()}=25", templateName)),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "priority", "1"),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "status", "1"),
				),
			},
		},
	})
}

func TestAccZabbixTriggerPrototype_BasicDependencies(t *testing.T) {
	strID := acctest.RandString(5)
	groupName := fmt.Sprintf("host_group_%s", strID)
	templateName := fmt.Sprintf("template_%s", strID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixItemPrototypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixTriggerPrototypeDependenciesConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkServerTriggerPrototypeDependencies(),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test_1", "dependencies.#", "1"),
				),
			},
		},
	})
}

func TestAccZabbixTriggerPrototype_ExpressionUpdate(t *testing.T) {
	strID := acctest.RandString(5)
	groupName := fmt.Sprintf("host_group_%s", strID)
	templateName := fmt.Sprintf("template_%s", strID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixItemPrototypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixTriggerPrototypeConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "description", "trigger_prototype_test"),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "expression", fmt.Sprintf("{%s:test.key.last()}=0", templateName)),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "priority", "5"),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "status", "0"),
				),
			},
			{
				Config: testAccZabbixTriggerPrototypeUpdateKeyConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "description", "trigger_prototype_test"),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "expression", fmt.Sprintf("{%s_update:test.key.last()}=0", templateName)),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "priority", "5"),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "status", "0"),
				),
			},
			{
				Config: testAccZabbixTriggerPrototypeUpdateKeyConfig2(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "description", "trigger_prototype_test"),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "expression", fmt.Sprintf("{%s_update:test.key.update.last()}=0", templateName)),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "priority", "5"),
					resource.TestCheckResourceAttr("zabbix_trigger_prototype.trigger_prototype_test", "status", "0"),
				),
			},
		},
	})
}

func testAccCheckZabbixTriggerPrototypeDestroy(s *terraform.State) error {
	api := testAccProvider.Meta().(*zabbix.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zabbix_trigger_prototype" {
			continue
		}

		_, err := api.TriggerPrototypeGetByID(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Trigger prototype still exist %s", rs.Primary.ID)
		}

		expectedError := "Expected exactly one result, got 0."
		if err.Error() != expectedError {
			return fmt.Errorf("expected error : %s, got : %s", expectedError, err.Error())
		}
	}
	return nil
}

func testAccZabbixTriggerPrototypeConfig(groupName, templateName string) string {
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

		resource "zabbix_item_prototype" "item_prototype_test" {
			delay = 60
			host_id  = zabbix_template.template_test.id
			rule_id = zabbix_lld_rule.lld_rule_test.id
			interface_id = "0"
			key = "test.key"
			name = "item_prototype_test"
			type = 0
			status = 0
		}

		resource "zabbix_trigger_prototype" "trigger_prototype_test" {
			description = "trigger_prototype_test"
			expression = "{${zabbix_template.template_test.host}:${zabbix_item_prototype.item_prototype_test.key}.last()}=0"
			priority = 5
			status = 0
		}
	`, groupName, templateName, templateName)
}

func testAccZabbixTriggerPrototypeUpdateConfig(groupName, templateName string) string {
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

		resource "zabbix_item_prototype" "item_prototype_test" {
			delay = 60
			host_id  = zabbix_template.template_test.id
			rule_id = zabbix_lld_rule.lld_rule_test.id
			interface_id = "0"
			key = "test.key"
			name = "item_prototype_test"
			type = 0
			status = 0
		}

		resource "zabbix_trigger_prototype" "trigger_prototype_test" {
			description = "trigger_prototype_test_update"
			expression = "{${zabbix_template.template_test.host}:${zabbix_item_prototype.item_prototype_test.key}.last()}=25"
			priority = 1
			status = 1
		}
	`, groupName, templateName, templateName)
}

func testAccZabbixTriggerPrototypeDependenciesConfig(groupName, templateName string) string {
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

		resource "zabbix_item_prototype" "item_prototype_test" {
			delay = 60
			host_id  = zabbix_template.template_test.id
			rule_id = zabbix_lld_rule.lld_rule_test.id
			interface_id = "0"
			key = "test.key"
			name = "item_prototype_test"
			type = 0
			status = 0
		}

		resource "zabbix_trigger_prototype" "trigger_prototype_test_0" {
			description = "trigger_prototype_test_update_0"
			expression = "{${zabbix_template.template_test.host}:${zabbix_item_prototype.item_prototype_test.key}.last()}=25"
			priority = 1
		}

		resource "zabbix_trigger_prototype" "trigger_prototype_test_1" {
			description = "trigger_prototype_test_update_1"
			expression = "{${zabbix_template.template_test.host}:${zabbix_item_prototype.item_prototype_test.key}.last()}=25"
			priority = 1
			dependencies = [
				zabbix_trigger_prototype.trigger_prototype_test_0.id
			]
		}
	`, groupName, templateName, templateName)
}

func testAccZabbixTriggerPrototypeUpdateKeyConfig(groupName, templateName string) string {
	return fmt.Sprintf(`
		resource "zabbix_host_group" "zabbix" {
			name = "host group test %s"
		}

		resource "zabbix_template" "template_test" {
			host = "%s_update"
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

		resource "zabbix_item_prototype" "item_prototype_test" {
			delay = 60
			host_id  = zabbix_template.template_test.id
			rule_id = zabbix_lld_rule.lld_rule_test.id
			interface_id = "0"
			key = "test.key"
			name = "item_prototype_test"
			type = 0
			status = 0
		}

		resource "zabbix_trigger_prototype" "trigger_prototype_test" {
			description = "trigger_prototype_test"
			expression = "{${zabbix_template.template_test.host}:${zabbix_item_prototype.item_prototype_test.key}.last()}=0"
			priority = 5
		}
	`, groupName, templateName, templateName)
}

func testAccZabbixTriggerPrototypeUpdateKeyConfig2(groupName, templateName string) string {
	return fmt.Sprintf(`
		resource "zabbix_host_group" "zabbix" {
			name = "host group test %s"
		}

		resource "zabbix_template" "template_test" {
			host = "%s_update"
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

		resource "zabbix_item_prototype" "item_prototype_test" {
			delay = 60
			host_id  = zabbix_template.template_test.id
			rule_id = zabbix_lld_rule.lld_rule_test.id
			interface_id = "0"
			key = "test.key.update"
			name = "item_prototype_test"
			type = 0
			status = 0
		}

		resource "zabbix_trigger_prototype" "trigger_prototype_test" {
			description = "trigger_prototype_test"
			expression = "{${zabbix_template.template_test.host}:${zabbix_item_prototype.item_prototype_test.key}.last()}=0"
			priority = 5
		}
	`, groupName, templateName, templateName)
}

func checkServerTriggerPrototypeDependencies() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		api := testAccProvider.Meta().(*zabbix.API)

		trigger0, ok := state.RootModule().Resources["zabbix_trigger_prototype.trigger_prototype_test_0"]
		if !ok {
			return fmt.Errorf("Cannot found trigger_prototype_0 in state")
		}
		trigger1, ok := state.RootModule().Resources["zabbix_trigger_prototype.trigger_prototype_test_1"]
		if !ok {
			return fmt.Errorf("Cannot found trigger_prototype_1 in state")
		}

		result, err := api.TriggerPrototypesGet(zabbix.Params{
			"triggerids":         trigger1.Primary.ID,
			"selectDependencies": "extend",
		})
		if err != nil {
			return err
		}
		if len(result) != 1 {
			return fmt.Errorf("Expected one trigger prototype and got %d", len(result))
		}
		trigger := result[0]
		if len(trigger.Dependencies) != 1 {
			return fmt.Errorf("Expected one dependencies for trigger_prototype_1 and got %d dependencies", len(trigger.Dependencies))
		}
		if trigger.Dependencies[0].TriggerID != trigger0.Primary.ID {
			return fmt.Errorf("Expected dependencies on trigger %s but got %s id", trigger0.Primary.ID, trigger.Dependencies[0].TriggerID)
		}
		return nil
	}
}
