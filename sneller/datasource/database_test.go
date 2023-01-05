package datasource_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"terraform-provider-sneller/sneller/acctest"
	"terraform-provider-sneller/sneller/api"
)

func TestAccDataSourceDatabase(t *testing.T) {
	resourceName := "data.sneller_database.test"
	baseConfig := acctest.ProviderConfig + `
		resource "sneller_tenant_region" "test" {
			region   = "` + api.DefaultSnellerRegion + `"
			bucket   = "` + acctest.Bucket1Name + `"
			role_arn = "` + acctest.Role1ARN + `"
		}
		
		resource "sneller_table" "db1_tablex" {
			region   = sneller_tenant_region.test.region
			database = "testdb"
			table    = "table-x"
			inputs   = [{
				pattern = "s3://` + acctest.Bucket1Name + `/*.ndjson"
				format  = "json"
			}]
		}
		
		resource "sneller_table" "db1_tabley" {
			region   = sneller_tenant_region.test.region
			database = "testdb"
			table    = "table-y"
			inputs   = [{
				pattern = "s3://` + acctest.Bucket2Name + `/*.ndjson"
				format  = "json"
			}]
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
					data "sneller_database" "test" {
						region   = sneller_tenant_region.test.region
						database = "testdb"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s", acctest.SnellerTenantID, api.DefaultSnellerRegion, "testdb")),
					resource.TestCheckResourceAttr(resourceName, "region", api.DefaultSnellerRegion),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/%s%s/", acctest.Bucket1Name, api.DefaultDbPrefix, "testdb")),
					resource.TestCheckResourceAttr(resourceName, "tables.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tables.0", "table-x"),
					resource.TestCheckResourceAttr(resourceName, "tables.1", "table-y"),
				),
			},
		},
	})
}
