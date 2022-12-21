package sneller

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTenant() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTenantRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"home_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mfa": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"activated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deactivated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTenantRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	tenantInfo, err := c.Tenant("")
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("tenant_id", tenantInfo.TenantID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", tenantInfo.TenantState); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", tenantInfo.TenantName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("home_region", tenantInfo.HomeRegion); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", tenantInfo.Email); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tenant_role_arn", tenantInfo.TenantRoleArn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mfa", tenantInfo.Mfa); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", tenantInfo.CreatedAt.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if tenantInfo.ActivatedAt != nil {
		if err := d.Set("activated_at", tenantInfo.ActivatedAt.Format(time.RFC3339)); err != nil {
			return diag.FromErr(err)
		}
	}
	if tenantInfo.DeactivatedAt != nil {
		if err := d.Set("deactivated_at", tenantInfo.DeactivatedAt.Format(time.RFC3339)); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(tenantInfo.TenantID)

	return diags
}
