package main

import (
	"github.com/nzolot/terraform-provider-zabbix/zabbix"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	p := plugin.ServeOpts{
		ProviderFunc: zabbix.Provider,
	}

	plugin.Serve(&p)
}
