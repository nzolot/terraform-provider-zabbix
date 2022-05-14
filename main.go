package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/zxzharmlesszxz/terraform-provider-zabbix/zabbix"
)

func main() {
	p := plugin.ServeOpts{
		ProviderFunc: zabbix.Provider,
	}

	plugin.Serve(&p)
}
