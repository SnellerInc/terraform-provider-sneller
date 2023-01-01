package sneller

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTable(t *testing.T) {
	initVars(t)

	resourceName := "data.sneller_table.test"
	baseConfig := providerConfig + `
		resource "sneller_tenant_region" "test" {
			region   = "` + defaultSnellerRegion + `"
			bucket   = "` + bucket1Name + `"
			role_arn = "` + role1ARN + `"
		}
		
		resource "sneller_table" "test" {
			region   = sneller_tenant_region.test.region
			database = "testdb"
			table    = "test-table"

			inputs = [
				{
					pattern = "s3://` + bucket1Name + `/*.ndjson"
					format  = "json"
				},
				{
					pattern = "s3://` + bucket2Name + `/*.ndjson.gz"
					format  = "json.gz"
				},
			]
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
					data "sneller_table" "test" {
						region   = sneller_tenant_region.test.region
						database = "testdb"
						table    = "test-table"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s/%s", snellerTenantID, defaultSnellerRegion, "testdb", "test-table")),
					resource.TestCheckResourceAttr(resourceName, "region", defaultSnellerRegion),
					resource.TestCheckResourceAttr(resourceName, "database", "testdb"),
					resource.TestCheckResourceAttr(resourceName, "table", "test-table"),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/%s%s/%s/", bucket1Name, defaultDbPrefix, "testdb", "test-table")),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.pattern", "s3://"+bucket1Name+"/*.ndjson"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.format", "json"),
					resource.TestCheckResourceAttr(resourceName, "inputs.1.pattern", "s3://"+bucket2Name+"/*.ndjson.gz"),
					resource.TestCheckResourceAttr(resourceName, "inputs.1.format", "json.gz"),
				),
			},
		},
	})
}
