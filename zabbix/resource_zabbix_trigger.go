package zabbix

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/mcuadros/go-version"
	"github.com/nzolot/go-zabbix-api"
)

func resourceZabbixTrigger() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixTriggerCreate,
		Read:   resourceZabbixTriggerRead,
		Exists: resourceZabbixTriggerExists,
		Update: resourceZabbixTriggerUpdate,
		Delete: resourceZabbixTriggerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"expression": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"recovery_mode": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 0 || v > 2 {
						errs = append(errs, fmt.Errorf("%q, must be between 0 and 2 inclusive, got %d", key, v))
					}
					return
				},
			},
			"recovery_expression": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"comment": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"priority": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 0 || v > 5 {
						errs = append(errs, fmt.Errorf("%q, must be between 0 and 5 inclusive, got %d", key, v))
					}
					return
				},
			},
			"status": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 0 || v > 1 {
						errs = append(errs, fmt.Errorf("%q, must be between 0 and 1 inclusive, got %d", key, v))
					}
					return
				},
			},
			"dependencies": &schema.Schema{
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "ID of the trigger it depands",
			},
			"tags": &schema.Schema{
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Tags for trigger. Support in Zabbix >=6.0",
			},
		},
	}
}

func resourceZabbixTriggerCreate(d *schema.ResourceData, meta interface{}) error {
	trigger := createTriggerObj(d)

	return createRetry(d, meta, createTrigger, trigger, resourceZabbixTriggerRead)
}

func resourceZabbixTriggerRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	params := zabbix.Params{
		"output":             "extend",
		"selectDependencies": "extend",
		"selectFunctions":    "extend",
		"selectItems":        "extend",
		"selectTags":         "extend",
		"triggerids":         d.Id(),
	}
	res, err := api.TriggersGet(params)
	if err != nil {
		return err
	}
	if len(res) != 1 {
		return fmt.Errorf("Expected one result got : %d", len(res))
	}
	trigger := res[0]
	err = getTriggerExpression(&trigger, api)
	log.Printf("[DEBUG] trigger expression: %s", trigger.Expression)
	d.Set("description", trigger.Description)
	d.Set("expression", trigger.Expression)

	d.Set("recovery_mode", trigger.RecoveryMode)
	d.Set("recovery_expression", trigger.RecoveryExpression)

	if trigger.Comments != "" {
		d.Set("comment", trigger.Comments)
	}
	d.Set("priority", trigger.Priority)
	if trigger.Status != 0 {
		d.Set("status", trigger.Status)
	} else {
		d.Set("value", 0)
	}

	var dependencies []string
	log.Printf("[DEBUG] var dependencies: %s", dependencies)
	for _, dependencie := range trigger.Dependencies {
		log.Printf("[DEBUG] loop dependencies: %s", dependencies)
		dependencies = append(dependencies, dependencie.TriggerID)
	}
	log.Printf("[DEBUG] loop end dependencies: %s", dependencies)
	d.Set("dependencies", dependencies)

	terraformTags := make(map[string]interface{}, len(trigger.Tags))
	for _, tag := range trigger.Tags {
		terraformTags[tag.TagName] = tag.Value
	}
	d.Set("tags", terraformTags)

	return nil
}

func resourceZabbixTriggerExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	api := meta.(*zabbix.API)

	_, err := api.TriggerGetByID(d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "Expected exactly one result") {
			log.Printf("[DEBUG] Trigger with id %s doesn't exist", d.Id())
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func resourceZabbixTriggerUpdate(d *schema.ResourceData, meta interface{}) error {
	trigger := createTriggerObj(d)

	trigger.TriggerID = d.Id()
	if !d.HasChange("dependencies") {
		trigger.Dependencies = nil
	}
	return createRetry(d, meta, updateTrigger, trigger, resourceZabbixTriggerRead)
}

func resourceZabbixTriggerDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	return deleteRetry(d.Id(), getTriggerParentID, api.TriggersDeleteIDs, api)
}

func createTriggerDependencies(d *schema.ResourceData) zabbix.DependencyTriggers {
	size := d.Get("dependencies.#").(int)
	dependencies := make(zabbix.DependencyTriggers, size)

	terraformDependencies := d.Get("dependencies").(*schema.Set)
	for i, terraformDependencie := range terraformDependencies.List() {
		dependencies[i].TriggerID = terraformDependencie.(string)
	}
	return dependencies
}

func createTriggerObj(d *schema.ResourceData) zabbix.Trigger {
	return zabbix.Trigger{
		Description:        d.Get("description").(string),
		Expression:         d.Get("expression").(string),
		RecoveryMode:       d.Get("recovery_mode").(int),
		RecoveryExpression: d.Get("recovery_expression").(string),
		Comments:           d.Get("comment").(string),
		Priority:           zabbix.SeverityType(d.Get("priority").(int)),
		Status:             zabbix.StatusType(d.Get("status").(int)),
		Dependencies:       createTriggerDependencies(d),
		Tags:               createZabbixTag(d),
	}
}

func getTriggerExpression(trigger *zabbix.Trigger, api *zabbix.API) error {
	for _, function := range trigger.Functions {
		var item zabbix.Item

		items, err := api.ItemsGet(zabbix.Params{
			"output":      "extend",
			"selectHosts": "extend",
			"webitems":    "extend",
			"itemids":     function.ItemID,
		})
		if err != nil {
			return err
		}
		if len(items) != 1 {
			return fmt.Errorf("Expected one item with id : %s and got : %d", function.ItemID, len(items))
		}
		item = items[0]
		if len(item.ItemParent) != 1 {
			return fmt.Errorf("Expected one parent host for item with id %s, and got : %d", function.ItemID, len(item.ItemParent))
		}
		idstr := fmt.Sprintf("{%s}", function.FunctionID)

		apiVersion, err := api.Version()
		log.Printf("[DEBUG] apiVersion: %s", apiVersion)

		if version.Compare(apiVersion, "5.4", ">=") {
			expendValue := fmt.Sprintf("%s(%s)", function.Function, strings.Replace(function.Parameter, "$", "/"+item.ItemParent[0].Host+"/"+item.Key, 1))
			trigger.Expression = strings.ReplaceAll(trigger.Expression, idstr, expendValue)
			trigger.RecoveryExpression = strings.ReplaceAll(trigger.RecoveryExpression, idstr, expendValue)
		} else {
			expendValue := fmt.Sprintf("{%s:%s.%s(%s)}", item.ItemParent[0].Host, item.Key, function.Function, function.Parameter)
			trigger.Expression = strings.ReplaceAll(trigger.Expression, idstr, expendValue)
			trigger.RecoveryExpression = strings.ReplaceAll(trigger.RecoveryExpression, idstr, expendValue)
		}
	}
	return nil
}

func getTriggerParentID(api *zabbix.API, id string) (string, error) {
	triggers, err := api.TriggersGet(zabbix.Params{
		"ouput":       "extend",
		"selectHosts": "extend",
		"triggerids":  id,
	})
	if err != nil {
		return "", err
	}
	if len(triggers) != 1 {
		return "", fmt.Errorf("Expected one item and got %d items", len(triggers))
	}
	if len(triggers[0].ParentHosts) != 1 {
		return "", fmt.Errorf("Expected one parent for item %s and got %d", id, len(triggers[0].ParentHosts))
	}
	return triggers[0].ParentHosts[0].HostID, nil
}

func createTrigger(trigger interface{}, api *zabbix.API) (id string, err error) {
	triggers := zabbix.Triggers{trigger.(zabbix.Trigger)}

	err = api.TriggersCreate(triggers)
	if err != nil {
		return
	}
	id = triggers[0].TriggerID
	return
}

func updateTrigger(trigger interface{}, api *zabbix.API) (id string, err error) {
	triggers := zabbix.Triggers{trigger.(zabbix.Trigger)}

	err = api.TriggersUpdate(triggers)
	if err != nil {
		return
	}
	id = triggers[0].TriggerID
	return
}
