resource "sneller_user" "test" {
  email       = "john.doe@example.com"
  is_admin    = true
  locale      = "en-US"
  given_name  = "John"
  family_name = "Doe"
}