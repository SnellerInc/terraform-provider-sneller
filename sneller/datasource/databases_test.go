package datasource_test

import (
	"fmt"
	"terraform-provider-sneller/sneller/acctest"
	"terraform-provider-sneller/sneller/api"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDatabases(t *testing.T) {
	resourceName := "data.sneller_databases.test"
	baseConfig := acctest.ProviderConfig + `
		resource "sneller_tenant_region" "test" {
			region   = "` + api.DefaultSnellerRegion + `"
			bucket   = "` + acctest.Bucket1Name + `"
			role_arn = "` + acctest.Role1ARN + `"
		}
		
		resource "sneller_table" "db1_tablex" {
			region   = sneller_tenant_region.test.region
			database = "db1"
			table    = "tablex"
			inputs   = [{
				pattern = "s3://` + acctest.Bucket1Name + `/*.ndjson"
				format  = "json"
			}]
		}
		
		resource "sneller_table" "db2_tabley" {
			region   = sneller_tenant_region.test.region
			database = "db2"
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
					data "sneller_databases" "test" {
						region = sneller_tenant_region.test.region
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", acctest.SnellerTenantID, api.DefaultSnellerRegion)),
					resource.TestCheckResourceAttr(resourceName, "region", api.DefaultSnellerRegion),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/%s", acctest.Bucket1Name, api.DefaultDbPrefix)),
					resource.TestCheckResourceAttr(resourceName, "databases.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "databases.0", "db1"),
					resource.TestCheckResourceAttr(resourceName, "databases.1", "db2"),
				),
			},
		},
	})
}
