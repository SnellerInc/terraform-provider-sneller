package sneller

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const defaultDbPrefix = "db/"

func dataSourceTenantRegion() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTenantRegionRead,
		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sqs_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTenantRegionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// determine region
	region := d.Get("region").(string)
	if region == "" {
		tenantInfo, err := c.Tenant("")
		if err != nil {
			return diag.FromErr(err)
		}
		region = tenantInfo.HomeRegion
	}

	tenantInfo, err := c.Tenant(region)
	if err != nil {
		return diag.FromErr(err)
	}

	tenantRegionInfo := tenantInfo.Regions[region]

	sqsARN := tenantRegionInfo.SqsArn
	if tenantRegionInfo.SqsArn == "" {
		// workaround for older API versions that
		// don't expose the SQS queue ARN
		tenantRoleARN, err := arn.Parse(tenantInfo.TenantRoleArn)
		if err == nil {
			sqsARN = fmt.Sprintf("arn:aws:sqs:%s:%s:tenant-sdb-%s", region, tenantRoleARN.AccountID, tenantInfo.TenantID)
		}
	}

	if err := d.Set("region", region); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("bucket", strings.TrimPrefix(tenantRegionInfo.Bucket, "s3://")); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("prefix", defaultDbPrefix); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("role_arn", tenantRegionInfo.RegionRoleArn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("external_id", tenantRegionInfo.RegionExternalID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("sqs_arn", sqsARN); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", tenantInfo.TenantID, region))

	return diags
}
