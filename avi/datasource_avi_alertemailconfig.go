/*
 * Copyright (c) 2017. Avi Networks.
 * Author: Gaurav Rastogi (grastogi@avinetworks.com)
 *
 */
package avi

import "github.com/hashicorp/terraform/helper/schema"

func dataSourceAviAlertEmailConfig() *schema.Resource {
	return &schema.Resource{
		Read: ResourceAviAlertEmailConfigRead,
		Schema: map[string]*schema.Schema{
			"cc_emails": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tenant_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"to_emails": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
