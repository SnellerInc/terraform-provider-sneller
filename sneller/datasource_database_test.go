package sneller

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDatabase(t *testing.T) {
	initVars(t)

	resourceName := "data.sneller_database.test"
	baseConfig := providerConfig + `
		resource "sneller_tenant_region" "test" {
			region   = "` + defaultSnellerRegion + `"
			bucket   = "` + bucket1Name + `"
			role_arn = "` + role1ARN + `"
		}
		
		resource "sneller_table" "db1_tablex" {
			region   = sneller_tenant_region.test.region
			database = "testdb"
			table    = "table-x"
			input {
				pattern = "s3://` + bucket1Name + `/*.ndjson"
				format  = "json"
			}			
		}
		
		resource "sneller_table" "db1_tabley" {
			region   = sneller_tenant_region.test.region
			database = "testdb"
			table    = "table-y"
			input {
				pattern = "s3://` + bucket2Name + `/*.ndjson"
				format  = "json"
			}			
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
					data "sneller_database" "test" {
						region   = sneller_tenant_region.test.region
						database = "testdb"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s", snellerTenantID, defaultSnellerRegion, "testdb")),
					resource.TestCheckResourceAttr(resourceName, "region", defaultSnellerRegion),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/%s%s/", bucket1Name, defaultDbPrefix, "testdb")),
					resource.TestCheckResourceAttr(resourceName, "tables.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tables.0", "table-x"),
					resource.TestCheckResourceAttr(resourceName, "tables.1", "table-y"),
				),
			},
		},
	})
}
