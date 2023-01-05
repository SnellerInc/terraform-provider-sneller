package resource_test

import (
	"fmt"
	"terraform-provider-sneller/sneller/acctest"
	"terraform-provider-sneller/sneller/api"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceTable(t *testing.T) {
	resourceName := "sneller_table.test"
	baseConfig := acctest.ProviderConfig + `
		resource "sneller_tenant_region" "test" {
			region   = "` + api.DefaultSnellerRegion + `"
			bucket   = "` + acctest.Bucket1Name + `"
			role_arn = "` + acctest.Role1ARN + `"
		}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: baseConfig + `
					resource "sneller_table" "test" {
						region   = sneller_tenant_region.test.region
						database = "` + acctest.DatabaseName + `"
						table    = "` + acctest.TableName + `"
					  
						inputs = [
							{
								pattern    = "s3://` + acctest.Bucket1Name + `/data/{tenant}/{yyyy}/{mm}/{dd}/*.ndjson"
								format     = "json"
								json_hints = [
									{
										field = "path.to.value.a"
										hints = ["ignore"]
									},{
										field = "endTimestamp"
										hints = ["no_index","RFC3339Nano"]
									}
								]
							},
							{
								pattern = "s3://` + acctest.Bucket2Name + `/data/*.ndjson"
								format  = "json"
							},
						]

						partitions = [
							{
								field = "tenant"
							},
							{
								field = "date"
								type  = "datetime"
								value = "{yyyy}-{mm}-{dd}:00:00:00Z"
							}
						]

						beta_features = ["zion"]
						skip_backfill = true
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s/%s", acctest.SnellerTenantID, api.DefaultSnellerRegion, acctest.DatabaseName, acctest.TableName)),
					resource.TestCheckResourceAttr(resourceName, "database", acctest.DatabaseName),
					resource.TestCheckResourceAttr(resourceName, "table", acctest.TableName),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/db/%s/%s/", acctest.Bucket1Name, acctest.DatabaseName, acctest.TableName)),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.pattern", "s3://"+acctest.Bucket1Name+"/data/{tenant}/{yyyy}/{mm}/{dd}/*.ndjson"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.format", "json"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.json_hints.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.json_hints.0.field", "path.to.value.a"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.json_hints.0.hints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.json_hints.0.hints.0", "ignore"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.json_hints.1.field", "endTimestamp"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.json_hints.1.hints.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.json_hints.1.hints.0", "no_index"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.json_hints.1.hints.1", "RFC3339Nano"),
					resource.TestCheckResourceAttr(resourceName, "inputs.1.pattern", "s3://"+acctest.Bucket2Name+"/data/*.ndjson"),
					resource.TestCheckResourceAttr(resourceName, "inputs.1.format", "json"),
					resource.TestCheckResourceAttr(resourceName, "partitions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "partitions.0.field", "tenant"),
					resource.TestCheckResourceAttr(resourceName, "partitions.1.field", "date"),
					resource.TestCheckResourceAttr(resourceName, "partitions.1.type", "datetime"),
					resource.TestCheckResourceAttr(resourceName, "partitions.1.value", "{yyyy}-{mm}-{dd}:00:00:00Z"),
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
						database = "` + acctest.DatabaseName + `"
						table    = "` + acctest.TableName + `"
					  
						inputs = [
							{
								pattern = "s3://` + acctest.Bucket2Name + `/data/*.ndjson"
								format  = "json.gz"
							}
						]
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s/%s/%s", acctest.SnellerTenantID, api.DefaultSnellerRegion, acctest.DatabaseName, acctest.TableName)),
					resource.TestCheckResourceAttr(resourceName, "database", acctest.DatabaseName),
					resource.TestCheckResourceAttr(resourceName, "table", acctest.TableName),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/db/%s/%s/", acctest.Bucket1Name, acctest.DatabaseName, acctest.TableName)),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.pattern", "s3://"+acctest.Bucket2Name+"/data/*.ndjson"),
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
