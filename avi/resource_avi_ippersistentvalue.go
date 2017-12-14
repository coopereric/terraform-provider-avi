/*
 * Copyright (c) 2017. Avi Networks.
 * Author: Gaurav Rastogi (grastogi@avinetworks.com)
 *
 */
package avi

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceIpPersistentValueSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"server_ip": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true, Elem: ResourceIpAddrSchema(),
				Set: func(v interface{}) int {
					return 0
				},
			},
		},
	}
}