package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/nzolot/terraform-provider-zabbix/zabbix"
)

func main() {
	p := plugin.ServeOpts{
		ProviderFunc: zabbix.Provider,
	}

	plugin.Serve(&p)
}
