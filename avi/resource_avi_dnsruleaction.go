/*
 * Copyright (c) 2017. Avi Networks.
 * Author: Gaurav Rastogi (grastogi@avinetworks.com)
 *
 */
package avi

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceDnsRuleActionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"allow": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     ResourceDnsRuleActionAllowDropSchema(),
				Set: func(v interface{}) int {
					return 0
				},
			},
			"gslb_site_selection": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     ResourceDnsRuleActionGslbSiteSelectionSchema(),
				Set: func(v interface{}) int {
					return 0
				},
			},
			"response": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     ResourceDnsRuleActionResponseSchema(),
				Set: func(v interface{}) int {
					return 0
				},
			},
		},
	}
}