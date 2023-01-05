package datasource_test

import (
	"fmt"
	"terraform-provider-sneller/sneller/acctest"
	"terraform-provider-sneller/sneller/api"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceElasticProxy(t *testing.T) {
	resourceName := "data.sneller_elastic_proxy.test"
	baseConfig := acctest.ProviderConfig + `
		resource "sneller_tenant_region" "test" {
			region   = "` + api.DefaultSnellerRegion + `"
			bucket   = "` + acctest.Bucket1Name + `"
			role_arn = "` + acctest.Role1ARN + `"
		}
		
		resource "sneller_elastic_proxy" "test" {
			region   = sneller_tenant_region.test.region
			log_path = "s3://` + acctest.Bucket1Name + `/log/elastic-proxy/"
			log_flags = {
				log_request			 = true
				log_query_parameters = true
				log_sql				 = true
				log_sneller_result   = true
				log_preprocessed     = true
				log_result           = true						
			}
			index = {
				ind1 = {
					database          = "test-db"
					table             = "table-x"
					ignore_total_hits = true
				}
				ind2 = {
					database          = "test-db"
					table             = "table-y"
					type_mapping = {
						timestamp = {
							type = "unix_nano_seconds"
						}
						"u_string_*" = {
							type = "contains"
							fields = {
								raw  = "text"
								test = "test"
							}
						}
					}
				}
			}
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
					data "sneller_elastic_proxy" "test" {
						region = "` + api.DefaultSnellerRegion + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", acctest.SnellerTenantID, api.DefaultSnellerRegion)),
					resource.TestCheckResourceAttr(resourceName, "location", fmt.Sprintf("s3://%s/db/elastic-proxy.json", acctest.Bucket1Name)),
					resource.TestCheckResourceAttr(resourceName, "region", api.DefaultSnellerRegion),
					resource.TestCheckResourceAttr(resourceName, "log_path", "s3://"+acctest.Bucket1Name+"/log/elastic-proxy/"),
					resource.TestCheckResourceAttr(resourceName, "log_flags.log_request", "true"),
					resource.TestCheckResourceAttr(resourceName, "log_flags.log_query_parameters", "true"),
					resource.TestCheckResourceAttr(resourceName, "log_flags.log_sql", "true"),
					resource.TestCheckResourceAttr(resourceName, "log_flags.log_sneller_result", "true"),
					resource.TestCheckResourceAttr(resourceName, "log_flags.log_preprocessed", "true"),
					resource.TestCheckResourceAttr(resourceName, "log_flags.log_result", "true"),
					resource.TestCheckResourceAttr(resourceName, "index.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "index.ind1.database", "test-db"),
					resource.TestCheckResourceAttr(resourceName, "index.ind1.table", "table-x"),
					resource.TestCheckResourceAttr(resourceName, "index.ind1.ignore_total_hits", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "index.ind1.type_mapping.#"),
					resource.TestCheckResourceAttr(resourceName, "index.ind2.database", "test-db"),
					resource.TestCheckResourceAttr(resourceName, "index.ind2.table", "table-y"),
					resource.TestCheckResourceAttr(resourceName, "index.ind2.ignore_total_hits", "false"),
					resource.TestCheckResourceAttr(resourceName, "index.ind2.type_mapping.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "index.ind2.type_mapping.timestamp.type", "unix_nano_seconds"),
					resource.TestCheckNoResourceAttr(resourceName, "index.ind2.type_mapping.timestamp.fields.%"),
					resource.TestCheckResourceAttr(resourceName, "index.ind2.type_mapping.u_string_*.type", "contains"),
					resource.TestCheckResourceAttr(resourceName, "index.ind2.type_mapping.u_string_*.fields.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "index.ind2.type_mapping.u_string_*.fields.raw", "text"),
					resource.TestCheckResourceAttr(resourceName, "index.ind2.type_mapping.u_string_*.fields.test", "test"),
				),
			},
		},
	})
}
