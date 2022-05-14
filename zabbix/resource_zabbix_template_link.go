package zabbix

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/nzolot/go-zabbix-api"
)

func resourceZabbixTemplateLink() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixTemplateLinkCreate,
		Read:   resourceZabbixTemplateLinkRead,
		Exists: resourceZabbixTemplateLinkExists,
		Update: resourceZabbixTemplateLinkUpdate,
		Delete: resourceZabbixTemplateLinkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"template_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"item": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     schemaTemplateItem(),
				Optional: true,
			},
			"trigger": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     schemaTemplateTrigger(),
				Optional: true,
			},
			"lld_rule": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     schemaTemplatelldRule(),
				Optional: true,
			},
		},
	}
}

func schemaTemplateItem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"local": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"item_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func schemaTemplateTrigger() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"local": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"trigger_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func schemaTemplatelldRule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"local": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"lld_rule_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceZabbixTemplateLinkCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceZabbixTemplateLinkRead(d, meta)
}

func resourceZabbixTemplateLinkRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	itemsTerraform, err := getTerraformTemplateItems(d, api)
	if err != nil {
		return err
	}
	d.Set("item", itemsTerraform)

	triggersTerraform, err := getTerraformTemplateTriggers(d, api)
	if err != nil {
		return err
	}
	d.Set("trigger", triggersTerraform)

	lldRulesTerraform, err := getTerraformTemplateLLDRules(d, api)
	if err != nil {
		return err
	}
	d.Set("lld_rule", lldRulesTerraform)

	d.SetId(d.Get("template_id").(string))
	return nil
}

func resourceZabbixTemplateLinkExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return true, nil
}

func resourceZabbixTemplateLinkUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	err := updateZabbixTemplateItems(d, api)
	if err != nil {
		return err
	}
	err = updateZabbixTemplateTriggers(d, api)
	if err != nil {
		return err
	}
	err = updateZabbixTemplateDiscoveryRules(d, api)
	if err != nil {
		return err
	}
	return resourceZabbixTemplateLinkRead(d, meta)
}

func resourceZabbixTemplateLinkDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func getTerraformTemplateItems(d *schema.ResourceData, api *zabbix.API) ([]interface{}, error) {
	params := zabbix.Params{
		"output": "extend",
		"templateids": []string{
			d.Get("template_id").(string),
		},
		"inherited": false,
	}
	items, err := api.ItemsGet(params)
	if err != nil {
		return nil, err
	}

	itemsTerraform := make([]interface{}, len(items))
	for i, item := range items {
		var itemTerraform = make(map[string]interface{})

		itemTerraform["local"] = true
		itemTerraform["item_id"] = item.ItemID
		itemsTerraform[i] = itemTerraform
	}
	return itemsTerraform, nil
}

func getTerraformTemplateTriggers(d *schema.ResourceData, api *zabbix.API) ([]interface{}, error) {
	params := zabbix.Params{
		"output": "extend",
		"templateids": []string{
			d.Get("template_id").(string),
		},
		"inherited": false,
	}
	triggers, err := api.TriggersGet(params)
	if err != nil {
		return nil, err
	}

	triggersTerraform := make([]interface{}, len(triggers))
	for i, trigger := range triggers {
		var triggerTerraform = make(map[string]interface{})

		triggerTerraform["local"] = true
		triggerTerraform["trigger_id"] = trigger.TriggerID
		triggersTerraform[i] = triggerTerraform
	}
	return triggersTerraform, nil
}

func getTerraformTemplateLLDRules(d *schema.ResourceData, api *zabbix.API) ([]interface{}, error) {
	params := zabbix.Params{
		"output": "extend",
		"templateids": []string{
			d.Get("template_id").(string),
		},
		"inherited": false,
	}
	lldRules, err := api.DiscoveryRulesGet(params)
	if err != nil {
		return nil, err
	}

	lldRulesTerraform := make([]interface{}, len(lldRules))
	for i, lldRule := range lldRules {
		var lldRuleTerraform = make(map[string]interface{})

		lldRuleTerraform["local"] = true
		lldRuleTerraform["lld_rule_id"] = lldRule.ItemID
		lldRulesTerraform[i] = lldRuleTerraform
	}
	return lldRulesTerraform, nil
}

func updateZabbixTemplateItems(d *schema.ResourceData, api *zabbix.API) error {
	if d.HasChange("item") {
		oldV, newV := d.GetChange("item")
		oldItems := oldV.(*schema.Set).List()
		newItems := newV.(*schema.Set).List()
		var deletedItems []string
		templatedItems, err := api.ItemsGet(zabbix.Params{
			"templateids": []string{
				d.Get("template_id").(string),
			},
			"inherited": true,
		})

		if err != nil {
			return err
		}
		log.Printf("[DEBUG] Found templated item %#v", templatedItems)
		for _, oldItem := range oldItems {
			oldItemValue := oldItem.(map[string]interface{})
			exist := false

			if oldItemValue["local"] == true {
				continue
			}

			for _, newItem := range newItems {
				newItemValue := newItem.(map[string]interface{})
				if newItemValue["item_id"].(string) == oldItemValue["item_id"].(string) {
					exist = true
				}
			}

			if !exist {
				templated := false

				for _, templatedItem := range templatedItems {
					if templatedItem.ItemID == oldItemValue["item_id"].(string) {
						templated = true
						break
					}
				}
				if !templated {
					deletedItems = append(deletedItems, oldItemValue["item_id"].(string))
				}
			}
		}
		if len(deletedItems) > 0 {
			log.Printf("[DEBUG] template link will delete item with ids : %#v", deletedItems)
			_, err := api.ItemsDeleteIDs(deletedItems)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func updateZabbixTemplateTriggers(d *schema.ResourceData, api *zabbix.API) error {
	if d.HasChange("trigger") {
		oldV, newV := d.GetChange("trigger")
		oldTriggers := oldV.(*schema.Set).List()
		newTriggers := newV.(*schema.Set).List()
		var deletedTriggers []string
		templatedTriggers, err := api.TriggersGet(zabbix.Params{
			"output": "extend",
			"templateids": []string{
				d.Get("template_id").(string),
			},
			"inherited": true,
		})

		if err != nil {
			return err
		}
		log.Printf("[DEBUG] found templated trigger %#v", templatedTriggers)
		for _, oldTrigger := range oldTriggers {
			oldTriggerValue := oldTrigger.(map[string]interface{})
			exist := false

			if oldTriggerValue["local"] == true {
				continue
			}

			for _, newTrigger := range newTriggers {
				newTriggerValue := newTrigger.(map[string]interface{})
				if oldTriggerValue["trigger_id"].(string) == newTriggerValue["trigger_id"].(string) {
					exist = true
				}
			}

			if !exist {
				templated := false

				for _, templatedTrigger := range templatedTriggers {
					if templatedTrigger.TriggerID == oldTriggerValue["trigger_id"].(string) {
						templated = true
						break
					}
				}
				if !templated {
					deletedTriggers = append(deletedTriggers, oldTriggerValue["trigger_id"].(string))
				}
			}
		}
		if len(deletedTriggers) > 0 {
			log.Printf("[DEBUG] template link will delete trigger with ids : %#v", deletedTriggers)
			_, err := api.TriggersDeleteIDs(deletedTriggers)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func updateZabbixTemplateDiscoveryRules(d *schema.ResourceData, api *zabbix.API) error {
	if d.HasChange("lld_rule") {
		oldV, newV := d.GetChange("lld_rule")
		oldlldRules := oldV.(*schema.Set).List()
		newlldRules := newV.(*schema.Set).List()
		var deletedlldRules []string
		templatedlldRules, err := api.DiscoveryRulesGet(zabbix.Params{
			"output": "extend",
			"templateids": []string{
				d.Get("template_id").(string),
			},
			"inherited": true,
		})

		if err != nil {
			return err
		}
		log.Printf("[DEBUG] found templated lldRule %#v", templatedlldRules)
		for _, oldlldRule := range oldlldRules {
			oldlldRuleValue := oldlldRule.(map[string]interface{})
			exist := false

			if oldlldRuleValue["local"] == true {
				continue
			}

			for _, newlldRule := range newlldRules {
				newlldRuleValue := newlldRule.(map[string]interface{})
				if oldlldRuleValue["lld_rule_id"].(string) == newlldRuleValue["lld_rule_id"].(string) {
					exist = true
				}
			}

			if !exist {
				templated := false

				for _, templatedlldRule := range templatedlldRules {
					if templatedlldRule.ItemID == oldlldRuleValue["lld_rule_id"].(string) {
						templated = true
						break
					}
				}
				if !templated {
					deletedlldRules = append(deletedlldRules, oldlldRuleValue["lld_rule_id"].(string))
				}
			}
		}
		if len(deletedlldRules) > 0 {
			log.Printf("[DEBUG] template link will delete lldRule with ids : %#v", deletedlldRules)
			_, err := api.DiscoveryRulesDeletesIDs(deletedlldRules)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
