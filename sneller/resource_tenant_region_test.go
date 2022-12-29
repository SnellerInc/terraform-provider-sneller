package sneller

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceTenantRegion(t *testing.T) {
	initVars(t)

	resourceName := "sneller_tenant_region.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
					resource "sneller_tenant_region" "test" {
						region   = "` + defaultSnellerRegion + `"
						bucket   = "` + bucket1Name + `"
						role_arn = "` + role1ARN + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", snellerTenantID, defaultSnellerRegion)),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
					resource.TestCheckResourceAttr(resourceName, "region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "bucket", bucket1Name),
					resource.TestCheckResourceAttr(resourceName, "prefix", defaultDbPrefix),
					resource.TestCheckResourceAttr(resourceName, "role_arn", role1ARN),
					resource.TestCheckResourceAttr(resourceName, "external_id", snellerTenantID),
					resource.TestCheckResourceAttr(resourceName, "sqs_arn", fmt.Sprintf("arn:aws:sqs:%s:%s:tenant-sdb-%s", defaultSnellerRegion, snellerAccountID, snellerTenantID)),
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
				Config: providerConfig + `
					resource "sneller_tenant_region" "test" {
						region   = "` + defaultSnellerRegion + `"
						bucket   = "` + bucket2Name + `"
						role_arn = "` + role2ARN + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", snellerTenantID, defaultSnellerRegion)),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
					resource.TestCheckResourceAttr(resourceName, "region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "bucket", bucket2Name),
					resource.TestCheckResourceAttr(resourceName, "prefix", defaultDbPrefix),
					resource.TestCheckResourceAttr(resourceName, "role_arn", role2ARN),
					resource.TestCheckResourceAttr(resourceName, "external_id", snellerTenantID),
					resource.TestCheckResourceAttr(resourceName, "sqs_arn", fmt.Sprintf("arn:aws:sqs:%s:%s:tenant-sdb-%s", defaultSnellerRegion, snellerAccountID, snellerTenantID)),
				),
			},
			// Delete is automatically tested
		},
	})
}
