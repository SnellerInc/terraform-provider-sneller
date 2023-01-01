package sneller

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceTable(t *testing.T) {
	initVars(t)

	resourceName := "sneller_table.test"
	baseConfig := providerConfig + `
		resource "sneller_tenant_region" "test" {
			region   = "` + defaultSnellerRegion + `"
			bucket   = "` + bucket1Name + `"
			role_arn = "` + role1ARN + `"
		}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: baseConfig + `
					resource "sneller_table" "test" {
						region   = sneller_tenant_region.test.region
						database = "` + databaseName + `"
						table    = "` + tableName + `"
					  
						inputs = [
							{
								pattern = "s3://` + bucket1Name + `/data/*.ndjson"
								format  = "json"
							},
							{
								pattern = "s3://` + bucket2Name + `/data/*.ndjson"
								format  = "json"
							},
						]

						beta_features = ["zion"]
						skip_backfill = true
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s/%s", snellerTenantID, defaultSnellerRegion, databaseName, tableName)),
					resource.TestCheckResourceAttr(resourceName, "database", databaseName),
					resource.TestCheckResourceAttr(resourceName, "table", tableName),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/db/%s/%s/", bucket1Name, databaseName, tableName)),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.pattern", "s3://"+bucket1Name+"/data/*.ndjson"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.format", "json"),
					resource.TestCheckResourceAttr(resourceName, "inputs.1.pattern", "s3://"+bucket2Name+"/data/*.ndjson"),
					resource.TestCheckResourceAttr(resourceName, "inputs.1.format", "json"),
					resource.TestCheckResourceAttr(resourceName, "beta_features.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "beta_features.0", "zion"),
					resource.TestCheckResourceAttr(resourceName, "skip_backfill", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
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
				Config: baseConfig + `
					resource "sneller_table" "test" {
						region   = sneller_tenant_region.test.region
						database = "` + databaseName + `"
						table    = "` + tableName + `"
					  
						inputs = [
							{
								pattern = "s3://` + bucket2Name + `/data/*.ndjson"
								format  = "json.gz"
							}
						]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s/%s", snellerTenantID, defaultSnellerRegion, databaseName, tableName)),
					resource.TestCheckResourceAttr(resourceName, "database", databaseName),
					resource.TestCheckResourceAttr(resourceName, "table", tableName),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/db/%s/%s/", bucket1Name, databaseName, tableName)),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.pattern", "s3://"+bucket2Name+"/data/*.ndjson"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.format", "json.gz"),
					resource.TestCheckNoResourceAttr(resourceName, "beta_features"),
					resource.TestCheckResourceAttr(resourceName, "skip_backfill", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Delete is automatically tested
		},
	})
}
