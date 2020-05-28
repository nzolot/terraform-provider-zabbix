package zabbix

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nzolot/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func sqlError(err error) bool {
	if strings.Contains(err.Error(), "SQL statement execution") || strings.Contains(err.Error(), "DBEXECUTE_ERROR") {
		return true
	}
	return false
}

type deleteFunc func([]string) ([]interface{}, error)
type createFunc func(interface{}, *zabbix.API) (string, error)
type getParentFunc func(*zabbix.API, string) (string, error)

func deleteRetry(id string, get getParentFunc, delete deleteFunc, api *zabbix.API) error {
	return resource.Retry(time.Minute, func() *resource.RetryError {
		parentID, err := get(api, id)
		if err != nil {
			if sqlError(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		templates, err := api.TemplatesGet(zabbix.Params{
			"output":            "extend",
			"selectHosts":       "extend",
			"parentTemplateids": parentID,
		})

		nbExpected := 1
		for _, template := range templates {
			nbExpected += len(template.LinkedHosts) + 1
		}

		deleteIDs, err := delete([]string{id})
		if err == nil {
			if len(deleteIDs) != nbExpected {
				return resource.NonRetryableError(fmt.Errorf("Expected to delete %d object and %d were deleted", nbExpected, len(deleteIDs)))
			}
			return nil
		} else if sqlError(err) {
			log.Printf("[DEBUG] Deletion failed. Got error %s, with id %s", err.Error(), id)
			return resource.RetryableError(fmt.Errorf("Failed to delete object with id: %s, got error %s", id, err.Error()))
		} else {
			return resource.NonRetryableError(err)
		}
	})
}

func createRetry(d *schema.ResourceData, meta interface{}, create createFunc, createArg interface{}, read schema.ReadFunc) error {
	return resource.Retry(time.Minute, func() *resource.RetryError {
		api := meta.(*zabbix.API)
		id, err := create(createArg, api)
		if err != nil {
			if sqlError(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		if d.Id() == "" {
			d.SetId(id)
		}

		return resource.NonRetryableError(read(d, meta))
	})
}

func createZabbixMacro(d *schema.ResourceData) zabbix.Macros {
	var macros zabbix.Macros

	terraformMacros := d.Get("macro").(map[string]interface{})
	for i, terraformMacro := range terraformMacros {
		macro := zabbix.Macro{
			MacroName: fmt.Sprintf("{$%s}", i),
			Value:     terraformMacro.(string),
		}
		macros = append(macros, macro)
	}
	return macros
}
