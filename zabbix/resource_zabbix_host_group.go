package zabbix

import (
	"log"
	"strings"

	"github.com/nzolot/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceZabbixHostGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixHostGroupCreate,
		Read:   resourceZabbixHostGroupRead,
		Exists: resourceZabbixHostGroupExists,
		Update: resourceZabbixHostGroupUpdate,
		Delete: resourceZabbixHostGroupDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the host group.",
			},
			"group_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Computed: true,
			},
		},
	}
}

func resourceZabbixHostGroupCreate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	hostGroup := zabbix.HostGroup{
		Name: d.Get("name").(string),
	}
	groups := zabbix.HostGroups{hostGroup}

	err := api.HostGroupsCreate(groups)
	if err != nil {
		return err
	}

	groupID := groups[0].GroupID

	log.Printf("[DEBUG] Created host group, id is %s", groupID)

	d.Set("group_id", groupID)
	d.SetId(groupID)

	return nil
}

func resourceZabbixHostGroupRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	log.Printf("[DEBUG] Will read host group with id %s", d.Id())

	group, err := api.HostGroupGetByID(d.Id())

	if err != nil {
		return err
	}

	d.Set("name", group.Name)

	return nil
}

func resourceZabbixHostGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	api := meta.(*zabbix.API)

	_, err := api.HostGroupGetByID(d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "Expected exactly one result") {
			log.Printf("[DEBUG] Host group with id %s doesn't exist", d.Id())
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func resourceZabbixHostGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	hostGroup := zabbix.HostGroup{
		Name:    d.Get("name").(string),
		GroupID: d.Id(),
	}

	return api.HostGroupsUpdate(zabbix.HostGroups{hostGroup})
}

func resourceZabbixHostGroupDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	return api.HostGroupsDeleteByIds([]string{d.Id()})
}
