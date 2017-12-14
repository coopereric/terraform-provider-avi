/*
 * Copyright (c) 2017. Avi Networks.
 * Author: Gaurav Rastogi (grastogi@avinetworks.com)
 *
 */
package avi

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceServerOperationalAnalysisSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"pool_ref": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"server_ip": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true, Elem: ResourceIpAddrSchema(),
				Set: func(v interface{}) int {
					return 0
				},
			},
			"server_oper_status": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     ResourceOperationalStatusSchema(),
				Set: func(v interface{}) int {
					return 0
				},
			},
			"server_port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}