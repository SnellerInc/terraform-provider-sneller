data "sneller_users" "test" {}

# Generate a list of sorted e-mail
# addresses of all enabled users
output "emails" {
  value = sort([for u in data.sneller_users.test.users : u.email if u.is_enabled])
}
