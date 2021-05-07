![Build Status](https://github.com/zxzharmlesszxz/terraform-provider-gitlab/workflows/Test%20and%20release/badge.svg)

Terraform Provider for Zabbix
=============================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.12 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/zxzharmlesszxz/terraform-provider-zabbix`

```sh
$ mkdir -p $GOPATH/src/github.com/zxzharmlesszxz; cd $GOPATH/src/github.com/zxzharmlesszxz
$ git clone git@github.com:zxzharmlesszxz/terraform-provider-zabbix
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/zxzharmlesszxz/terraform-provider-zabbix
$ make build
```

**Note**: For contributions created from forks, the repository should still be cloned under the `$GOPATH/src/github.com/zxzharmlesszxz/terraform-provider-zabbix` directory to allow the provided `make` commands to properly run, build, and test this project.

Using the provider
------------------
If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory,  run `terraform init` to initialize it.

Further [usage documentation is available on the Terraform website](https://www.terraform.io/docs/providers/zabbix/index.html).

Developing the Provider
-----------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.11+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-zabbix
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources.

```sh
$ make testacc
```

Notes
-----

### Template linking

Template linking is used as a way to track template items and triggers in an authoritative way. A template link resource must contain all the ids of your local items and triggers otherwise they will be deleted during the next apply.

* If you use the template link resource to track template dependencies you should pay attention to always have your local item and trigger referenced otherwise they will be deleted and will need to be recreated, causing a permanent diff.
* If you have template dependencies you should use the `template_id` value of the `zabbix_template_link` resource to link the children templates to the parent else the child template could be updated before parent item or trigger which lead to errors like `Error: Expected to delete 2 trigger and 1 were delete`.

Examples
--------

### Host

```hcl
provider "zabbix" {
  user = "Admin"
  password = "zabbix"
  server_url = "http://localhost/api_jsonrpc.php"
}

resource "zabbix_host" "zabbix1" {
  host = "127.0.0.1"
  name = "the best name"
  interfaces {
    ip = "127.0.0.1"
    main = true
  }
  groups = ["Linux servers", "${zabbix_host_group.zabbix.name}"]
  templates = ["Template ICMP Ping"]
}

resource "zabbix_host_group" "zabbix" {
  name = "something"
}
```

### Template

The template link resource is required if you want to track your template item and trigger in an authoritative way.

```hcl
provider "zabbix" {
  user = "Admin"
  password = "zabbix"
  server_url = "http://localhost/api_jsonrpc.php"
}

resource "zabbix_host_group" "demo_group" {
  name = "Template demo group"
}

resource "zabbix_template" "demo_template" {
  host        = "template"
  name        = "template demo"
  description = "An exemple of template with item and trigger"
  groups      = [zabbix_host_group.demo_group.name]
  macro = {
    MACRO_TEMPLATE = "12"
  }
}

# This virtual resource is responsible of ensuring no item and trigger is directly associated to the template
resource "zabbix_template_link" "demo_template_link" {
  template_id = zabbix_template.demo_template.id
}
```

### Template with item and trigger

```hcl
provider "zabbix" {
  user = "Admin"
  password = "zabbix"
  server_url = "http://localhost/api_jsonrpc.php"
}

resource "zabbix_host_group" "demo_group" {
  name = "Template demo group"
}

resource "zabbix_template" "demo_template" {
  host        = "template"
  name        = "template demo"
  description = "An exemple of template with item and trigger"
  groups      = [zabbix_host_group.demo_group.name]
  macro = {
    MACRO_TEMPLATE = "12"
  }
}

resource "zabbix_item" "demo_item" {
  name        = "demo item"
  key         = "demo.key"
  delay       = "34"
  description = "Item for the demo template"
  trends      = "300"
  history     = "25"
  host_id     = zabbix_template.demo_template.template_id
}

resource "zabbix_trigger" "demo_trigger" {
  description = "demo trigger"
  expression  = "{${zabbix_template.demo_template.host}:${zabbix_item.demo_item.key}.last()}={$MACRO_TEMPLATE}"
  priority    = 5
  status      = 0
}

# This virtual resource is responsible of ensuring no item and trigger is directly associated to the template
resource "zabbix_template_link" "demo_template_link" {
  template_id = zabbix_template.demo_template.id
  item {
    item_id = zabbix_item.demo_item.id
  }
  trigger {
    trigger_id = zabbix_trigger.demo_trigger.id
  }
}
```

### Template dependencies

```hcl
provider "zabbix" {
  user = "Admin"
  password = "zabbix"
  server_url = "http://localhost/api_jsonrpc.php"
}

resource "zabbix_host_group" "demo_group" {
  name = "Template demo group"
}

resource "zabbix_template" "template_1" {
  host        = "template_1"
  groups      = [zabbix_host_group.demo_group.name]
}

resource "zabbix_template_link" "demo_template_1_link" {
  template_id = zabbix_template.template_1.id
}

resource "zabbix_template" "template_2" {
  host = "template_2"
  groups = [zabbix_host_group.demo_group.name]
  linked_template = [ # use the template link template_id value to be sure that all template_1 dependencies has been updated
    zabbix_template.demo_template_1_link.template_id
  ]
}

resource "zabbix_template_link" "demo_template_2_link" {
  template_id = zabbix_template.template_2.id
}
```
