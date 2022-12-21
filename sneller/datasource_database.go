package sneller

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatabase() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDatabaseRead,
		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"database": {
				Type:     schema.TypeString,
				Required: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tables": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceDatabaseRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	region := d.Get("region").(string)
	if region == "" {
		region = c.DefaultRegion
	}
	database := d.Get("database").(string)

	tenantInfo, err := c.Tenant(region)
	if err != nil {
		return diag.FromErr(err)
	}

	tableInfos, err := c.Database(region, database)
	if err != nil {
		return diag.FromErr(err)
	}

	tables := make([]string, 0, len(tableInfos))
	for _, ti := range tableInfos {
		tables = append(tables, ti.Name)
	}

	if err := d.Set("location", fmt.Sprintf("%s/db/%s/", tenantInfo.Regions[region].Bucket, database)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tables", tables); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", tenantInfo.TenantID, region, database))

	return diags
}
