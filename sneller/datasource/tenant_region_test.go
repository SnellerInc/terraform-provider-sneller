package datasource_test

import (
	"fmt"
	"terraform-provider-sneller/sneller/acctest"
	"terraform-provider-sneller/sneller/api"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const DefaultMaxScanBytes = 1024 * 1024 * 1024 * 1024 * 3

func TestAccDataSourceTenantRegion(t *testing.T) {
	resourceName := "data.sneller_tenant_region.test"
	baseConfig := acctest.ProviderConfig + `
		resource "sneller_tenant_region" "test" {
			region   = "` + api.DefaultSnellerRegion + `"
			bucket   = "` + acctest.Bucket1Name + `"
			role_arn = "` + acctest.Role1ARN + `"
		}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create resource
			{
				Config: baseConfig,
			},
			// Read testing
			{
				Config: baseConfig + `
					data "sneller_tenant_region" "test" {
						region = "` + api.DefaultSnellerRegion + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", acctest.SnellerTenantID, api.DefaultSnellerRegion)),
					resource.TestCheckResourceAttr(resourceName, "region", api.DefaultSnellerRegion),
					resource.TestCheckResourceAttr(resourceName, "bucket", acctest.Bucket1Name),
					resource.TestCheckResourceAttr(resourceName, "prefix", api.DefaultDbPrefix),
					resource.TestCheckResourceAttr(resourceName, "role_arn", acctest.Role1ARN),
					resource.TestCheckResourceAttr(resourceName, "external_id", acctest.SnellerTenantID),
					resource.TestCheckResourceAttr(resourceName, "max_scan_bytes", fmt.Sprintf("%d", DefaultMaxScanBytes)),
					resource.TestCheckResourceAttr(resourceName, "sqs_arn", fmt.Sprintf("arn:aws:sqs:%s:%s:tenant-sdb-%s", api.DefaultSnellerRegion, acctest.SnellerAccountID, acctest.SnellerTenantID)),
				),
			},
		},
	})
}
