package zabbix

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/nzolot/go-zabbix-api"
)

func resourceZabbixLLDRuleLink() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixLLDRuleLinkCreate,
		Read:   resourceZabbixLLDRuleLinkRead,
		Exists: resourceZabbixLLDRuleLinkExists,
		Update: resourceZabbixLLDRuleLinkUpdate,
		Delete: resourceZabbixLLDRuleLinkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"lld_rule_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"item_prototype": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     schemaTemplateItemPrototype(),
				Optional: true,
			},
			"trigger_prototype": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     schemaTemplateTriggerPrototype(),
				Optional: true,
			},
		},
	}
}

func schemaTemplateItemPrototype() *schema.Resource {
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

func schemaTemplateTriggerPrototype() *schema.Resource {
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

func resourceZabbixLLDRuleLinkCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceZabbixLLDRuleLinkRead(d, meta)
}

func resourceZabbixLLDRuleLinkRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	itemsTerraform, err := getTerraformTemplateItemPrototypes(d, api)
	if err != nil {
		return err
	}
	d.Set("item_prototype", itemsTerraform)

	triggersTerraform, err := getTerraformTemplateTriggerPrototypes(d, api)
	if err != nil {
		return err
	}
	d.Set("trigger_prototype", triggersTerraform)

	d.SetId(d.Get("lld_rule_id").(string))
	return nil
}

func resourceZabbixLLDRuleLinkExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return true, nil
}

func resourceZabbixLLDRuleLinkUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	err := updateZabbixTemplateItemPrototypes(d, api)
	if err != nil {
		return err
	}

	err = updateZabbixTemplateTriggerPrototypes(d, api)
	if err != nil {
		return err
	}
	return resourceZabbixLLDRuleLinkRead(d, meta)
}

func resourceZabbixLLDRuleLinkDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func getTerraformTemplateItemPrototypes(d *schema.ResourceData, api *zabbix.API) ([]interface{}, error) {
	params := zabbix.Params{
		"output": "extend",
		"discoveryids": []string{
			d.Get("lld_rule_id").(string),
		},
		"inherited": false,
	}
	items, err := api.ItemPrototypesGet(params)
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

func getTerraformTemplateTriggerPrototypes(d *schema.ResourceData, api *zabbix.API) ([]interface{}, error) {
	params := zabbix.Params{
		"output": "extend",
		"discoveryids": []string{
			d.Get("lld_rule_id").(string),
		},
		"inherited": false,
	}
	triggers, err := api.TriggerPrototypesGet(params)
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

func updateZabbixTemplateItemPrototypes(d *schema.ResourceData, api *zabbix.API) error {
	if d.HasChange("item_prototype") {
		oldV, newV := d.GetChange("item_prototype")
		oldItems := oldV.(*schema.Set).List()
		newItems := newV.(*schema.Set).List()
		var deletedItems []string
		templatedItems, err := api.ItemPrototypesGet(zabbix.Params{
			"discoveryids": []string{
				d.Get("lld_rule_id").(string),
			},
			"inherited": true,
		})

		if err != nil {
			return err
		}
		log.Printf("[DEBUG] Found templated item prototype %#v", templatedItems)
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
			log.Printf("[DEBUG] template link will delete item prototype with ids : %#v", deletedItems)
			_, err := api.ItemPrototypesDeleteIDs(deletedItems)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func updateZabbixTemplateTriggerPrototypes(d *schema.ResourceData, api *zabbix.API) error {
	if d.HasChange("trigger_prototype") {
		oldV, newV := d.GetChange("trigger_prototype")
		oldTriggers := oldV.(*schema.Set).List()
		newTriggers := newV.(*schema.Set).List()
		var deletedTriggers []string
		templatedTriggers, err := api.TriggerPrototypesGet(zabbix.Params{
			"output": "extend",
			"discoveryids": []string{
				d.Get("lld_rule_id").(string),
			},
			"inherited": true,
		})

		if err != nil {
			return err
		}
		log.Printf("[DEBUG] found templated trigger prototype %#v", templatedTriggers)
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
			log.Printf("[DEBUG] template link will delete trigger prototype with ids : %#v", deletedTriggers)
			_, err := api.TriggerPrototypesDeleteIDs(deletedTriggers)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
