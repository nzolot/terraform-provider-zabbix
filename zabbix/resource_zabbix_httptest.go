package zabbix

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/nzolot/go-zabbix-api"
)

var stepsSchema *schema.Resource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		"order": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		"url": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		"status_codes": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"search_string": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"headers": &schema.Schema{
			Type:     schema.TypeMap,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
		},
	},
}

func resourceZabbixHttpTest() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixHttpTestCreate,
		Read:   resourceZabbixHttpTestRead,
		Update: resourceZabbixHttpTestUpdate,
		Delete: resourceZabbixHttpTestDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"host_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the host or template that the httptest belongs to.",
				ForceNew:    true,
			},
			"httptest_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the web check",
				ForceNew:    true,
			},
			"delay": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Execution interval of the web scenario. Accepts seconds, time unit with suffix and user macro.",
				ForceNew:    false,
				Default:     "1m",
			},
			"retries": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Number of times a web scenario will try to execute each step before failing.",
				ForceNew:    false,
				Default:     "1",
			},
			"headers": &schema.Schema{
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Headers for the httptest",
			},
			"steps": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     stepsSchema,
				Required: true,
				ForceNew: false,
			},
		},
	}
}

func resourceZabbixHttpTestCreate(d *schema.ResourceData, meta interface{}) error {
	httptest, err := createHttpTestObj(d)

	if err != nil {
		return err
	}

	return createRetry(d, meta, createHttpTest, httptest, resourceZabbixHttpTestRead)
}

func createHttpTest(httptest interface{}, api *zabbix.API) (id string, err error) {
	httptests := zabbix.HttpTests{httptest.(zabbix.HttpTest)}

	err = api.HttpTestsCreate(httptests)
	if err != nil {
		return
	}
	id = httptests[0].HttpTestID
	return
}

func createHttpTestObj(d *schema.ResourceData) (zabbix.HttpTest, error) {
	httptest := zabbix.HttpTest{
		HttpTestID: d.Get("httptest_id").(string),
		Name:       d.Get("name").(string),
		HostID:     d.Get("host_id").(string),
		Delay:      d.Get("delay").(string),
		Retries:    d.Get("retries").(string),
		Headers:    createZabbixHttpTestHeaders(d),
	}

	steps, err := createStepsObj(d)
	if err != nil {
		return zabbix.HttpTest{}, err
	}
	httptest.Steps = steps

	return httptest, nil
}

func createStepsObj(d *schema.ResourceData) (zabbix.Steps, error) {
	stepsCount := d.Get("steps.#").(int)

	steps := make(zabbix.Steps, stepsCount)

	for i := 0; i < stepsCount; i++ {
		prefix := fmt.Sprintf("steps.%d.", i)
		steps[i] = zabbix.Step{
			Name: d.Get(prefix + "name").(string),
			No:   d.Get(prefix + "order").(string),
			Url:  d.Get(prefix + "url").(string),

			StatusCodes: d.Get(prefix + "status_codes").(string),
			RequiredStr: d.Get(prefix + "search_string").(string),
		}

		headers := zabbix.Headers{}
		terraformHeaders := d.Get(prefix + "headers").(map[string]interface{})
		for i, terraformHeader := range terraformHeaders {
			header := zabbix.Header{
				Name:  fmt.Sprintf("%s", i),
				Value: terraformHeader.(string),
			}
			headers = append(headers, header)
		}
		steps[i].Headers = headers

	}

	return steps, nil
}

func resourceZabbixHttpTestRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	httptest, err := api.HttpTestGetByID(d.Id())
	if err != nil {
		return err
	}

	d.Set("name", httptest.Name)
	d.Set("host_id", httptest.HostID)
	d.Set("httptest_id", httptest.HttpTestID)
	d.Set("retries", httptest.Retries)
	d.Set("delay", httptest.Delay)
	d.Set("steps", httptest.Steps)

	log.Printf("[DEBUG] httptest name is %s\n", httptest.Name)
	return nil
}

// Update
func resourceZabbixHttpTestUpdate(d *schema.ResourceData, meta interface{}) error {
	httptest, err := createHttpTestObj(d)

	if err != nil {
		return err
	}
	httptest.HostID = ""

	return createRetry(d, meta, updateHttpTest, httptest, resourceZabbixHttpTestRead)
}

func updateHttpTest(httptest interface{}, api *zabbix.API) (id string, err error) {
	httptests := zabbix.HttpTests{httptest.(zabbix.HttpTest)}

	err = api.HttpTestsUpdate(httptests)
	if err != nil {
		return
	}
	id = httptests[0].HttpTestID
	return
}

// Delete
func resourceZabbixHttpTestDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	return api.HttpTestsDeleteByIds([]string{d.Id()})
}
