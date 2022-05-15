package zabbix

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/nzolot/go-zabbix-api"
)

func resourceZabbixItem() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixItemCreate,
		Read:   resourceZabbixItemRead,
		Exists: resourceZabbixItemExists,
		Update: resourceZabbixItemUpdate,
		Delete: resourceZabbixItemDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"delay": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"host_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the host or template that the item belongs to.",
			},
			"interface_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "0",
			},
			"key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Item key.",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the item.",
			},
			"type": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 0 || v > 16 {
						errs = append(errs, fmt.Errorf("%q, must be between 0 and 16 inclusive, got %d", key, v))
					}
					return
				},
			},
			"value_type": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 0 || v > 4 {
						errs = append(errs, fmt.Errorf("%q, must be between 0 and 4 inclusive, got %d", key, v))
					}
					return
				},
			},
			"data_type": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Data type of the item (Removed in Zabbix 3.4).",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 0 || v > 3 {
						errs = append(errs, fmt.Errorf("%q, must be between 0 and 3 inclusive, got %d", key, v))
					}
					return
				},
			},
			"delta": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Value that will be stored (Removed in Zabbix 3.4).",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 0 || v > 2 {
						errs = append(errs, fmt.Errorf("%q, must be between 0 and 2 inclusive, got %d", key, v))
					}
					return
				},
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the item.",
				Default:     "",
			},
			"history": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Number of days to keep item's history data. From 3.4 version, string is required instead of integer. Default: 90 (90d for 3.4+).",
			},
			"trends": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Number of days to keep item's trends data. From 3.4 version, string is required instead of interger. Default: 365 (365d for 3.4+).",
			},
			"trapper_host": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allowed hosts. Used only by trapper items.",
			},
            "tags": &schema.Schema{
                Type:        schema.TypeMap,
                Elem:        &schema.Schema{Type: schema.TypeString},
                Optional:    true,
                Description: "Tags for item. Support in Zabbix >=6.0",
            },
		},
	}
}

func createItemObject(d *schema.ResourceData) *zabbix.Item {

	item := zabbix.Item{
		Delay:        d.Get("delay").(string),
		HostID:       d.Get("host_id").(string),
		InterfaceID:  d.Get("interface_id").(string),
		Key:          d.Get("key").(string),
		Name:         d.Get("name").(string),
		Type:         zabbix.ItemType(d.Get("type").(int)),
		ValueType:    zabbix.ValueType(d.Get("value_type").(int)),
		DataType:     zabbix.DataType(d.Get("data_type").(int)),
		Delta:        zabbix.DeltaType(d.Get("delta").(int)),
		Description:  d.Get("description").(string),
		History:      d.Get("history").(string),
		Trends:       d.Get("trends").(string),
		TrapperHosts: d.Get("trapper_host").(string),
		Tags:         createZabbixTag(d),
	}

	return &item
}

func resourceZabbixItemCreate(d *schema.ResourceData, meta interface{}) error {
	item := createItemObject(d)

	return createRetry(d, meta, createItem, *item, resourceZabbixItemRead)
}

func resourceZabbixItemRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	item, err := api.ItemGetByID(d.Id())
	if err != nil {
		return err
	}

	d.Set("delay", item.Delay)
	d.Set("host_id", item.HostID)
	d.Set("interface_id", item.InterfaceID)
	d.Set("key", item.Key)
	d.Set("name", item.Name)
	d.Set("type", item.Type)
	d.Set("value_type", item.ValueType)
	d.Set("data_type", item.DataType)
	d.Set("delta", item.Delta)
	d.Set("description", item.Description)
	d.Set("history", item.History)
	d.Set("trends", item.Trends)
	d.Set("trapper_host", item.TrapperHosts)

	terraformTags := make(map[string]interface{}, len(item.Tags))
	for _, tag := range item.Tags {
		terraformTags[tag.TagName] = tag.Value
	}
	d.Set("tags", terraformTags)

	log.Printf("[DEBUG] Item name is %s\n", item.Name)
	return nil
}

func resourceZabbixItemExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	api := meta.(*zabbix.API)

	_, err := api.ItemGetByID(d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "Expected exactly one result") {
			log.Printf("[DEBUG] Item with id %s doesn't exist", d.Id())
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func resourceZabbixItemUpdate(d *schema.ResourceData, meta interface{}) error {
	item := createItemObject(d)

	item.ItemID = d.Id()
	return createRetry(d, meta, updateItem, *item, resourceZabbixItemRead)

}

func resourceZabbixItemDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	return deleteRetry(d.Id(), getItemParentID, api.ItemsDeleteIDs, api)
}

func getItemParentID(api *zabbix.API, id string) (string, error) {
	items, err := api.ItemsGet(zabbix.Params{
		"output":      "extend",
		"selectHosts": "extend",
		"itemids":     id,
	})
	if err != nil {
		return "", fmt.Errorf("%s, with item %s", err.Error(), id)
	}
	if len(items) != 1 {
		return "", fmt.Errorf("Expected one item and got %d items", len(items))
	}
	if len(items[0].ItemParent) != 1 {
		return "", fmt.Errorf("Expected one parent for item %s and got %d", id, len(items[0].ItemParent))
	}
	return items[0].ItemParent[0].HostID, nil
}

func createItem(item interface{}, api *zabbix.API) (id string, err error) {
	items := zabbix.Items{item.(zabbix.Item)}

	err = api.ItemsCreate(items)
	if err != nil {
		return
	}
	id = items[0].ItemID
	return
}

func updateItem(item interface{}, api *zabbix.API) (id string, err error) {
	items := zabbix.Items{item.(zabbix.Item)}

	err = api.ItemsUpdate(items)
	if err != nil {
		return
	}
	id = items[0].ItemID
	return
}
