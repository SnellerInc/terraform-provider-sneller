package datasource_test

import (
	"terraform-provider-sneller/sneller/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUsers(t *testing.T) {
	resourceName := "data.sneller_users.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: acctest.ProviderConfig + `
					data "sneller_users" "test" {}`,

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "users.#"),
					resource.TestCheckResourceAttrSet(resourceName, "users.0.user_id"),
					resource.TestCheckResourceAttrSet(resourceName, "users.0.email"),
					resource.TestCheckResourceAttrSet(resourceName, "users.0.is_enabled"),
				),
			},
		},
	})
}
