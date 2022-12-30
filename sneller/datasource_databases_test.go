package sneller

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDatabases(t *testing.T) {
	initVars(t)

	resourceName := "data.sneller_databases.test"
	baseConfig := providerConfig + `
		resource "sneller_tenant_region" "test" {
			region   = "` + defaultSnellerRegion + `"
			bucket   = "` + bucket1Name + `"
			role_arn = "` + role1ARN + `"
		}
		
		resource "sneller_table" "db1_tablex" {
			region   = sneller_tenant_region.test.region
			database = "db1"
			table    = "tablex"
			input    = [{
				pattern = "s3://` + bucket1Name + `/*.ndjson"
				format  = "json"
			}]
		}
		
		resource "sneller_table" "db2_tabley" {
			region   = sneller_tenant_region.test.region
			database = "db2"
			table    = "table-y"
			input    = [{
				pattern = "s3://` + bucket2Name + `/*.ndjson"
				format  = "json"
			}]
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
					data "sneller_databases" "test" {
						region = sneller_tenant_region.test.region
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", snellerTenantID, defaultSnellerRegion)),
					resource.TestCheckResourceAttr(resourceName, "region", defaultSnellerRegion),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/%s", bucket1Name, defaultDbPrefix)),
					resource.TestCheckResourceAttr(resourceName, "databases.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "databases.0", "db1"),
					resource.TestCheckResourceAttr(resourceName, "databases.1", "db2"),
				),
			},
		},
	})
}
