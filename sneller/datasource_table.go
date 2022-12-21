package sneller

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTable() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTableRead,
		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"database": {
				Type:     schema.TypeString,
				Required: true,
			},
			"table": {
				Type:     schema.TypeString,
				Required: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"input": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pattern": {
							Type:     schema.TypeString,
							Required: true,
						},
						"format": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			// TODO: Add features and partitioning information
		},
	}
}

func dataSourceTableRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	region := d.Get("region").(string)
	if region == "" {
		region = c.DefaultRegion
	}
	database := d.Get("database").(string)
	table := d.Get("table").(string)

	tenantInfo, err := c.Tenant(region)
	if err != nil {
		return diag.FromErr(err)
	}

	tableDescription, err := c.Table(region, database, table)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("location", fmt.Sprintf("%s/db/%s/%s/", tenantInfo.Regions[region].Bucket, database, table)); err != nil {
		return diag.FromErr(err)
	}

	var inputs []any
	for _, input := range tableDescription.Input {
		i := make(map[string]any)
		i["pattern"] = input.Pattern
		i["format"] = input.Format
		inputs = append(inputs, i)
	}

	if err := d.Set("input", inputs); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s/%s", tenantInfo.TenantID, region, database, table))

	return diags
}
