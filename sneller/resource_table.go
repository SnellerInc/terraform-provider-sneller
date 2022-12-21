package sneller

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTableCreate,
		ReadContext:   resourceTableRead,
		UpdateContext: resourceTableUpdate,
		DeleteContext: resourceTableDelete,
		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"database": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"table": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"input": {
				Type:     schema.TypeList,
				Optional: true,
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
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceTableCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return resourceTableCreateOrUpdate(ctx, d, m, false)
}

func resourceTableUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return resourceTableCreateOrUpdate(ctx, d, m, true)
}

func resourceTableCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m any, update bool) diag.Diagnostics {
	c := m.(*Client)

	tenantInfo, err := c.Tenant("")
	if err != nil {
		return diag.FromErr(err)
	}

	// determine region
	region := d.Get("region").(string)
	if region == "" {
		region = tenantInfo.HomeRegion
	}

	database := d.Get("database").(string)
	table := d.Get("table").(string)

	if !update || d.HasChange("input") {
		var inputs []TableInput

		for _, input := range d.Get("input").([]any) {
			i := input.(map[string]any)
			inputs = append(inputs, TableInput{
				Pattern: i["pattern"].(string),
				Format:  i["format"].(string),
			})
		}

		c.SetTable(region, database, table, inputs)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s/%s", tenantInfo.TenantID, region, database, table))

	return resourceTableRead(ctx, d, m)
}

func resourceTableRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 4 {
		return diag.Errorf("invalid id %q", d.Id())
	}
	tenantID := parts[0]
	region := parts[1]
	database := parts[2]
	table := parts[3]

	tenantInfo, err := c.withTenantID(tenantID).Tenant(region)
	if err != nil {
		return diag.FromErr(err)
	}

	tableDescription, err := c.withTenantID(tenantID).Table(region, database, table)
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

	return diags
}

func resourceTableDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 4 {
		return diag.Errorf("invalid id %q", d.Id())
	}
	tenantID := parts[0]
	region := parts[1]
	database := parts[2]
	table := parts[3]

	err := c.withTenantID(tenantID).DeleteTable(region, database, table, true)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
