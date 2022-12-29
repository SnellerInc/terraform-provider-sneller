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
					  
						input {
							pattern = "s3://` + bucket1Name + `/data/*.ndjson"
							format  = "json"
						  }
					  
						input {
							pattern = "s3://` + bucket2Name + `/data/*.ndjson"
							format  = "json"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s/%s", snellerTenantID, defaultSnellerRegion, databaseName, tableName)),
					resource.TestCheckResourceAttr(resourceName, "database", databaseName),
					resource.TestCheckResourceAttr(resourceName, "table", tableName),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/db/%s/%s/", bucket1Name, databaseName, tableName)),
					resource.TestCheckResourceAttr(resourceName, "input.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "input.0.pattern", "s3://"+bucket1Name+"/data/*.ndjson"),
					resource.TestCheckResourceAttr(resourceName, "input.0.format", "json"),
					resource.TestCheckResourceAttr(resourceName, "input.1.pattern", "s3://"+bucket2Name+"/data/*.ndjson"),
					resource.TestCheckResourceAttr(resourceName, "input.1.format", "json"),
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
					  
						input {
							pattern = "s3://` + bucket2Name + `/data/*.ndjson"
							format  = "json.gz"
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s/%s", snellerTenantID, defaultSnellerRegion, databaseName, tableName)),
					resource.TestCheckResourceAttr(resourceName, "database", databaseName),
					resource.TestCheckResourceAttr(resourceName, "table", tableName),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/db/%s/%s/", bucket1Name, databaseName, tableName)),
					resource.TestCheckResourceAttr(resourceName, "input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input.0.pattern", "s3://"+bucket2Name+"/data/*.ndjson"),
					resource.TestCheckResourceAttr(resourceName, "input.0.format", "json.gz"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Delete is automatically tested
		},
	})
}
