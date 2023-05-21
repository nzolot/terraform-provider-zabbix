package zabbix

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/nzolot/go-zabbix-api"
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

func createZabbixTag(d *schema.ResourceData) zabbix.Tags {
	var tags zabbix.Tags

	terraformTags := d.Get("tags").(map[string]interface{})
	for i, terraformTag := range terraformTags {
		tag := zabbix.Tag{
			TagName: fmt.Sprintf("%s", i),
			Value:   terraformTag.(string),
		}
		tags = append(tags, tag)
	}
	return tags
}

func createZabbixItemPreProcs(d *schema.ResourceData) zabbix.PreProcs {
	var preprocs zabbix.PreProcs

	terraformItemPreProcs := d.Get("preprocessing").([]interface{})
	for _, terraformItemPreProc := range terraformItemPreProcs {
		preproc := zabbix.PreProc{
			Type:               fmt.Sprintf("%d", terraformItemPreProc.(map[string]interface{})["type"]),
			Params:             fmt.Sprintf("%v", terraformItemPreProc.(map[string]interface{})["params"]),
			ErrorHandler:       fmt.Sprintf("%d", terraformItemPreProc.(map[string]interface{})["error_handler"]),
			ErrorHandlerParams: fmt.Sprintf("%v", terraformItemPreProc.(map[string]interface{})["error_handler_params"]),
		}
		preprocs = append(preprocs, preproc)
	}
	return preprocs
}

func createZabbixLLDMacroPaths(d *schema.ResourceData) zabbix.LLDMacroPaths {
	var macropaths zabbix.LLDMacroPaths

	terraformLLDMacroPaths := d.Get("lld_macros").([]interface{})
	for _, terraformLLDMacroPath := range terraformLLDMacroPaths {
		macropath := zabbix.LLDMacroPath{
			Macro: fmt.Sprintf("%s", terraformLLDMacroPath.(map[string]interface{})["macro"]),
			Path:  fmt.Sprintf("%s", terraformLLDMacroPath.(map[string]interface{})["path"]),
		}
		macropaths = append(macropaths, macropath)
	}
	return macropaths
}

func createZabbixHttpTestHeaders(d *schema.ResourceData) zabbix.Headers {
	var headers zabbix.Headers

	terraformHeaders := d.Get("headers").(map[string]interface{})
	for i, terraformHeader := range terraformHeaders {
		header := zabbix.Header{
			Name:  fmt.Sprintf("%s", i),
			Value: terraformHeader.(string),
		}
		headers = append(headers, header)
	}
	return headers
}
