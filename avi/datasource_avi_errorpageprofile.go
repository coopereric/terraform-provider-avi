/*
 * Copyright (c) 2017. Avi Networks.
 * Author: Gaurav Rastogi (grastogi@avinetworks.com)
 *
 */
package avi

import "github.com/hashicorp/terraform/helper/schema"

func dataSourceAviErrorPageProfile() *schema.Resource {
	return &schema.Resource{
		Read: ResourceAviErrorPageProfileRead,
		Schema: map[string]*schema.Schema{
			"error_pages": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     ResourceErrorPageSchema(),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tenant_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
