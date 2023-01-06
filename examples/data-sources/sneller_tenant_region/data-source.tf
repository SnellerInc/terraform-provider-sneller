data "sneller_tenant_region" "test" {
  region = "us-east-1"
}

output "role-arn" {
  value = data.sneller_tenant_region.test.role_arn
}

output "sqs-arn" {
  value = data.sneller_tenant_region.test.sqs_arn
}