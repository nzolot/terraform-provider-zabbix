package zabbix

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zxzharmlesszxz/go-zabbix-api"
)

func TestAccZabbixTemplateLink_Basic(t *testing.T) {
	strID := acctest.RandString(5)
	groupName := fmt.Sprintf("host_group_%s", strID)
	templateName := fmt.Sprintf("template_%s", strID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixTemplateLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixTemplateLinkConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "item.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_item.#", "0"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "trigger.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_trigger.#", "0"),
				),
			},
			{
				Config: testAccZabbixTemplateLinkDeleteTrigger(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "item.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_item.#", "0"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "trigger.#", "0"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_trigger.#", "0"),
				),
			},
			{
				Config: testAccZabbixTemplateLinkDeleteItem(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "item.#", "0"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_item.#", "0"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "trigger.#", "0"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_trigger.#", "0"),
				),
			},
			{
				Config: testAccZabbixTemplateLinkConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "item.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_item.#", "0"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "trigger.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_trigger.#", "0"),
				),
			},
		},
	})
}

func TestAccZabbixTemplateLink_DeleteServerItem(t *testing.T) {
	strID := acctest.RandString(5)
	groupName := fmt.Sprintf("host_group_%s", strID)
	templateName := fmt.Sprintf("template_%s", strID)

	var template zabbix.Template
	item := zabbix.Item{
		Name:  "server_item",
		Key:   "server.key",
		Type:  zabbix.ZabbixAgent,
		Delay: "30",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixTemplateLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixTemplateLinkConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists("zabbix_template.template_test", &template),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "item.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_item.#", "0"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "trigger.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_trigger.#", "0"),
				),
			},
			{
				PreConfig: testAccZabbixTemplateLinkCreateServerItem(template, &item),
				Config:    testAccZabbixTemplateLinkConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateServerItemDelete(&item),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "item.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_item.#", "0"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "trigger.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_trigger.#", "0"),
				),
			},
		},
	})
}

func TestAccZabbixTemplateLink_DeleteServerTrigger(t *testing.T) {
	strID := acctest.RandString(5)
	groupName := fmt.Sprintf("host_group_%s", strID)
	templateName := fmt.Sprintf("template_%s", strID)

	var template zabbix.Template
	item := zabbix.Item{
		Name:  "server_item",
		Key:   "server.key",
		Type:  zabbix.ZabbixAgent,
		Delay: "30",
	}
	trigger := zabbix.Trigger{
		Description: "server_trigger",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixTemplateLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixTemplateLinkConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists("zabbix_template.template_test", &template),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "item.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_item.#", "0"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "trigger.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_trigger.#", "0"),
				),
			},
			{
				PreConfig: testAccZabbixTemplateLinkCreateServerTrigger(template, item, &trigger),
				Config:    testAccZabbixTemplateLinkConfig(groupName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateServerTriggerDelete(&trigger),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "item.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_item.#", "0"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "trigger.#", "1"),
					resource.TestCheckResourceAttr("zabbix_template_link.template_link_test", "server_trigger.#", "0"),
				),
			},
		},
	})
}

func testAccZabbixTemplateLinkConfig(groupName, templateName string) string {
	return fmt.Sprintf(`
		data "zabbix_server" "test" {}

		resource "zabbix_host_group" "zabbix" {
			name = "host group test %s"
		}

		resource "zabbix_template" "template_test" {
			host = "%s"
			groups = [ zabbix_host_group.zabbix.name ]
			name = "display name for template test %s"
	  	}

		resource "zabbix_item" "item_test_0" {
			name = "item_test_0"
			key = "bilou.bilou"
			delay = "34"
			trends = join("", ["300", data.zabbix_server.test.unit_time_days])
			history = join("", ["25", data.zabbix_server.test.unit_time_days])
			host_id = "${zabbix_template.template_test.id}"
		}

		resource "zabbix_trigger" "trigger_test_0" {
			description = "trigger_test_0"
			expression  = "{${zabbix_template.template_test.host}:${zabbix_item.item_test_0.key}.last()} = 0"
			priority    = 5
		}

		resource "zabbix_template_link" "template_link_test" {
			template_id = zabbix_template.template_test.id
			item {
				item_id = zabbix_item.item_test_0.id
			}
			trigger {
				trigger_id = zabbix_trigger.trigger_test_0.id
			}
		}
	`, groupName, templateName, templateName)
}

func testAccZabbixTemplateLinkDeleteTrigger(groupName, templateName string) string {
	return fmt.Sprintf(`
		resource "zabbix_host_group" "zabbix" {
			name = "host group test %s"
		}

		resource "zabbix_template" "template_test" {
			host = "%s"
			groups = ["${zabbix_host_group.zabbix.name}"]
			name = "display name for template test %s"
	  	}

		resource "zabbix_item" "item_test_0" {
			name = "item_test_0"
			key = "bilou.bilou"
			delay = "34"
			host_id = "${zabbix_template.template_test.id}"
		}

		resource "zabbix_template_link" "template_link_test" {
			template_id = zabbix_template.template_test.id
			item {
				item_id = zabbix_item.item_test_0.id
			}
		}
	`, groupName, templateName, templateName)
}

func testAccZabbixTemplateLinkDeleteItem(groupName, templateName string) string {
	return fmt.Sprintf(`
		resource "zabbix_host_group" "zabbix" {
			name = "host group test %s"
		}

		resource "zabbix_template" "template_test" {
			host = "%s"
			groups = ["${zabbix_host_group.zabbix.name}"]
			name = "display name for template test %s"
		  }

		resource "zabbix_template_link" "template_link_test" {
			template_id = zabbix_template.template_test.id
		}
	`, groupName, templateName, templateName)
}

func testAccZabbixTemplateLinkCreateServerItem(template zabbix.Template, item *zabbix.Item) func() {
	return func() {
		api := testAccProvider.Meta().(*zabbix.API)

		item.HostID = template.TemplateID
		items := zabbix.Items{*item}
		err := api.ItemsCreate(items)
		if err != nil {
			return
		}
		item.ItemID = items[0].ItemID
	}
}

func testAccZabbixTemplateLinkCreateServerTrigger(template zabbix.Template, item zabbix.Item, trigger *zabbix.Trigger) func() {
	return func() {
		api := testAccProvider.Meta().(*zabbix.API)

		trigger.Expression = fmt.Sprintf("{%s:%s.last()} = 0", template.Host, item.Key)
		triggers := zabbix.Triggers{*trigger}
		err := api.TriggersCreate(triggers)
		if err != nil {
			log.Print(err)
			return
		}
		trigger.TriggerID = triggers[0].TriggerID
	}
}

func testAccCheckZabbixTemplateLinkDestroy(s *terraform.State) error {
	return nil
}

func testAccCheckTemplateExists(n string, template *zabbix.Template) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		api := testAccProvider.Meta().(*zabbix.API)
		templates, err := api.TemplateGetByID(rs.Primary.ID)
		if err != nil {
			return err
		}
		template = templates
		return nil
	}
}

func testAccCheckTemplateServerItemDelete(item *zabbix.Item) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := testAccProvider.Meta().(*zabbix.API)

		_, err := api.ItemGetByID(item.ItemID)
		if err == nil {
			return fmt.Errorf("Expected an error")
		}

		expectedErr := "Expected exactly one result, got 0."
		if err.Error() != expectedErr {
			return fmt.Errorf("expected error : %s, got : %s", expectedErr, err.Error())
		}
		return nil
	}
}

func testAccCheckTemplateServerTriggerDelete(trigger *zabbix.Trigger) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := testAccProvider.Meta().(*zabbix.API)

		_, err := api.TriggerGetByID(trigger.TriggerID)
		if err == nil {
			return fmt.Errorf("Expected an error")
		}

		expectedErr := "Expected exactly one result, got 0."
		if err.Error() != expectedErr {
			return fmt.Errorf("expected error : %s, got : %s", expectedErr, err.Error())
		}
		return nil
	}
}
