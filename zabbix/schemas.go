package zabbix

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var interfaceSchema *schema.Resource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"dns": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			//ForceNew: true,
		},
		"ip": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			//ForceNew: true,
		},
		"main": &schema.Schema{
			Type:     schema.TypeBool,
			Required: true,
			ForceNew: true,
		},
		"port": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  "10050",
			//ForceNew: true,
		},
		"type": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  "agent",
			ForceNew: true,
		},
		"interface_id": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			ForceNew: true,
		},
	},
}

var proxyInterfaceSchema *schema.Resource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"dns": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"ip": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"port": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  "10050",
		},
		"useip": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  "1",
		},
	},
}

var itemPreprocessingSchema *schema.Resource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"type": &schema.Schema{
			Type:     schema.TypeInt,
			Required: true,
		},
		"params": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"error_handler": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
		},
		"error_handler_params": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
	},
}

var LLDMacroPathsSchema *schema.Resource = &schema.Resource{
	// LLDMacro string `json:"macro"` // Required
	// Path     string `json:"path"` // Required
	Schema: map[string]*schema.Schema{
		"macro": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		"path": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
	},
}
