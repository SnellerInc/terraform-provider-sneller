# The sneller_tenant data-source holds all the global
# tenant information
data "sneller_tenant" "tenant" {    
}

# The sneller_tenant resource allows the tenant's
# regional configuration
resource "sneller_tenant_region" "region" {
  depends_on = [
    # Wait until Sneller has access to the cache bucket
    aws_s3_bucket_policy.sneller_cache
  ]
  
  region   = var.region
  bucket   = aws_s3_bucket.sneller_cache.bucket
  role_arn = aws_iam_role.sneller_s3.arn
}

# The sneller_table resource configures the
# table definition
resource "sneller_table" "test" {
  # Enable this for production to avoid trashing your table
  # lifecycle { prevent_destroy = true }

  region   = sneller_tenant_region.region.region
  database = "test-db"
  table    = "test-table"

  input {
    pattern = "s3://${aws_s3_bucket.sneller_source.bucket}/${var.source_prefix}*.ndjson"
    format  = "json"
  }
}

# Output of all relevant Sneller resources
output tenant {
  value = data.sneller_tenant.tenant
}

output tenant_region {
  value = sneller_tenant_region.region
}

output tenant_table {
  value = sneller_table.test
}
