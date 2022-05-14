package zabbix

import (
	"encoding/json"
	//"errors"
	"fmt"
	"log"
	//"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/nzolot/go-zabbix-api"
	//"github.com/mcuadros/go-version"
)

func dataSourceZabbixHost() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceZabbixHostRead,
		Schema: map[string]*schema.Schema{
			"host_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the host",
			},
			"main_interface_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Get host main interface ID",
			},
			"interfaces": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     interfaceSchema,
				Computed: true,
			},
			"tags": &schema.Schema{
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataSourceZabbixHostRead(d *schema.ResourceData, meta interface{}) (err error) {
	host_id := d.Get("host_id").(string)

	d.SetId(host_id)

	api := meta.(*zabbix.API)
	log.Printf("[DEBUG] Will read host with id %s", d.Get("host_id").(string))

	host, err := api.HostGetByID(d.Get("host_id").(string))

	if err != nil {
		return err
	}

	interfaces_str, err := json.Marshal(host.Interfaces)
	log.Printf("[DEBUG] host interfaces =  %s", fmt.Sprintf(string(interfaces_str)))

	//d.Set("interfaces", host.Interfaces)

	// single id
	log.Printf("[DEBUG] testing host.Interfaces[0].InterfaceId =  %d", host.Interfaces[0].InterfaceId)
	d.Set("main_interface_id", fmt.Sprintf("%d", host.Interfaces[0].InterfaceId))

	// get all interfaces
	interfaceCount := len(host.Interfaces)
	log.Printf("[DEBUG] interfaceCount =  %d", interfaceCount)
	interfaces := make(zabbix.HostInterfaces, interfaceCount)

	for i := 0; i < interfaceCount; i++ {
		prefix := fmt.Sprintf("interfaces.%d.", i)
		log.Printf("[DEBUG] Interface id =  %d", i)
		log.Printf("[DEBUG] Prefix id =  %s", prefix)

		interfaces[i] = zabbix.HostInterface{
			//IP:    ip,
			//DNS:   dns,
			//Main:  main,
			//Port:  d.Get(prefix + "port").(string),
			//Type:  typeID,
			//UseIP: useip,
		}
	}

	d.Set("interfaces", interfaces)

	d.Set("tags", host.Tags)
	return nil
}
