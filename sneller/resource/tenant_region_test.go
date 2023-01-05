package resource_test

import (
	"fmt"
	"terraform-provider-sneller/sneller/acctest"
	"terraform-provider-sneller/sneller/api"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const DefaultMaxScanBytes = 1024 * 1024 * 1024 * 1024 * 3

func TestAccResourceTenantRegion(t *testing.T) {
	resourceName := "sneller_tenant_region.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: acctest.ProviderConfig + `
					resource "sneller_tenant_region" "test" {
						bucket         = "` + acctest.Bucket1Name + `"
						role_arn       = "` + acctest.Role1ARN + `"
						max_scan_bytes = 123456789
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", acctest.SnellerTenantID, api.DefaultSnellerRegion)),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
					resource.TestCheckResourceAttr(resourceName, "region", api.DefaultSnellerRegion),
					resource.TestCheckResourceAttr(resourceName, "bucket", acctest.Bucket1Name),
					resource.TestCheckResourceAttr(resourceName, "prefix", api.DefaultDbPrefix),
					resource.TestCheckResourceAttr(resourceName, "role_arn", acctest.Role1ARN),
					resource.TestCheckResourceAttr(resourceName, "external_id", acctest.SnellerTenantID),
					resource.TestCheckResourceAttr(resourceName, "max_scan_bytes", "123456789"),
					resource.TestCheckResourceAttr(resourceName, "effective_max_scan_bytes", "123456789"),
					resource.TestCheckResourceAttr(resourceName, "sqs_arn", fmt.Sprintf("arn:aws:sqs:%s:%s:tenant-sdb-%s", api.DefaultSnellerRegion, acctest.SnellerAccountID, acctest.SnellerTenantID)),
				),
			},
			// Import testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"}, // TF only
			},
			// Update and Read testing
			{
				Config: acctest.ProviderConfig + `
					resource "sneller_tenant_region" "test" {
						region   = "` + api.DefaultSnellerRegion + `"
						bucket   = "` + acctest.Bucket2Name + `"
						role_arn = "` + acctest.Role2ARN + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", acctest.SnellerTenantID, api.DefaultSnellerRegion)),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
					resource.TestCheckResourceAttr(resourceName, "region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "bucket", acctest.Bucket2Name),
					resource.TestCheckResourceAttr(resourceName, "prefix", api.DefaultDbPrefix),
					resource.TestCheckResourceAttr(resourceName, "role_arn", acctest.Role2ARN),
					resource.TestCheckResourceAttr(resourceName, "external_id", acctest.SnellerTenantID),
					resource.TestCheckNoResourceAttr(resourceName, "max_scan_bytes"),
					resource.TestCheckResourceAttr(resourceName, "effective_max_scan_bytes", fmt.Sprintf("%d", DefaultMaxScanBytes)),
					resource.TestCheckResourceAttr(resourceName, "sqs_arn", fmt.Sprintf("arn:aws:sqs:%s:%s:tenant-sdb-%s", api.DefaultSnellerRegion, acctest.SnellerAccountID, acctest.SnellerTenantID)),
				),
			},
			// Delete is automatically tested
		},
	})
}
