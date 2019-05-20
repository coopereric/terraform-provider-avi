/*
 * Copyright (c) 2017. Avi Networks.
 * Author: Gaurav Rastogi (grastogi@avinetworks.com)
 *
 */
package avi

import "github.com/hashicorp/terraform/helper/schema"

func dataSourceAviVrfContext() *schema.Resource {
	return &schema.Resource{
		Read: ResourceAviVrfContextRead,
		Schema: map[string]*schema.Schema{
			"bgp_profile": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     ResourceBgpProfileSchema(),
			},
			"cloud_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"debugvrfcontext": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     ResourceDebugVrfContextSchema(),
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway_mon": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     ResourceGatewayMonitorSchema(),
			},
			"internal_gateway_monitor": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     ResourceInternalGatewayMonitorSchema(),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"static_routes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     ResourceStaticRouteSchema(),
			},
			"system_default": {
				Type:     schema.TypeBool,
				Computed: true,
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
