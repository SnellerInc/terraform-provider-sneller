package sneller

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUser(t *testing.T) {
	initVars(t)

	email := fmt.Sprintf("john.doe+%s@example.com", acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum))
	resourceName := "data.sneller_user.test"
	baseConfig := providerConfig + `
		resource "sneller_user" "test" {
			email       = "` + email + `"
			is_admin    = true
			locale      = "nl-NL"
			given_name  = "John"
			family_name = "Doe"
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
					data "sneller_users" "test" {}
						
					data "sneller_user" "test" {
						// get the 'user_id' for the given e-mail address
						user_id = one([for u in data.sneller_users.test.users : u.user_id if u.email == "` + email + `"])
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "is_admin", "true"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "is_federated", "false"),
					resource.TestCheckResourceAttr(resourceName, "locale", "nl-NL"),
					resource.TestCheckResourceAttr(resourceName, "given_name", "John"),
					resource.TestCheckResourceAttr(resourceName, "family_name", "Doe"),
				),
			},
		},
	})
}
