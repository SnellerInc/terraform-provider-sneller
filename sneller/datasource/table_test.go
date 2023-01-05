package datasource_test

import (
	"fmt"
	"terraform-provider-sneller/sneller/acctest"
	"terraform-provider-sneller/sneller/api"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTable(t *testing.T) {
	resourceName := "data.sneller_table.test"
	baseConfig := acctest.ProviderConfig + `
		resource "sneller_tenant_region" "test" {
			region   = "` + api.DefaultSnellerRegion + `"
			bucket   = "` + acctest.Bucket1Name + `"
			role_arn = "` + acctest.Role1ARN + `"
		}
		
		resource "sneller_table" "test" {
			region   = sneller_tenant_region.test.region
			database = "testdb"
			table    = "test-table"

			inputs = [
				{
					pattern = "s3://` + acctest.Bucket1Name + `/*.ndjson"
					format  = "json"
				},
				{
					pattern = "s3://` + acctest.Bucket2Name + `/*.ndjson.gz"
					format  = "json.gz"
				},
			]
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
					data "sneller_table" "test" {
						region   = sneller_tenant_region.test.region
						database = "testdb"
						table    = "test-table"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s/%s", acctest.SnellerTenantID, api.DefaultSnellerRegion, "testdb", "test-table")),
					resource.TestCheckResourceAttr(resourceName, "region", api.DefaultSnellerRegion),
					resource.TestCheckResourceAttr(resourceName, "database", "testdb"),
					resource.TestCheckResourceAttr(resourceName, "table", "test-table"),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/%s%s/%s/", acctest.Bucket1Name, api.DefaultDbPrefix, "testdb", "test-table")),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.pattern", "s3://"+acctest.Bucket1Name+"/*.ndjson"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.format", "json"),
					resource.TestCheckResourceAttr(resourceName, "inputs.1.pattern", "s3://"+acctest.Bucket2Name+"/*.ndjson.gz"),
					resource.TestCheckResourceAttr(resourceName, "inputs.1.format", "json.gz"),
				),
			},
		},
	})
}
