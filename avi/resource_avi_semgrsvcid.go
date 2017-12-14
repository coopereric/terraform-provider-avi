/*
 * Copyright (c) 2017. Avi Networks.
 * Author: Gaurav Rastogi (grastogi@avinetworks.com)
 *
 */
package avi

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceSeMgrSvcIdSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"svc_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"svc_q_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"svc_uuid": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}