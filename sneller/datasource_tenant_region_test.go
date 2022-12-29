package sneller

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTenantRegion(t *testing.T) {
	initVars(t)

	resourceName := "data.sneller_tenant_region.test"
	baseConfig := providerConfig + `
		resource "sneller_tenant_region" "test" {
			region   = "` + defaultSnellerRegion + `"
			bucket   = "` + bucket1Name + `"
			role_arn = "` + role1ARN + `"
		}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create resource
			{
				Config: baseConfig,
			},
			// Read testing
			{
				Config: baseConfig + `
					data "sneller_tenant_region" "test" {
						region = "` + defaultSnellerRegion + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", snellerTenantID, defaultSnellerRegion)),
					resource.TestCheckResourceAttr(resourceName, "region", defaultSnellerRegion),
					resource.TestCheckResourceAttr(resourceName, "bucket", bucket1Name),
					resource.TestCheckResourceAttr(resourceName, "prefix", defaultDbPrefix),
					resource.TestCheckResourceAttr(resourceName, "role_arn", role1ARN),
					resource.TestCheckResourceAttr(resourceName, "external_id", snellerTenantID),
					resource.TestCheckResourceAttr(resourceName, "sqs_arn", fmt.Sprintf("arn:aws:sqs:%s:%s:tenant-sdb-%s", defaultSnellerRegion, snellerAccountID, snellerTenantID)),
				),
			},
		},
	})
}
