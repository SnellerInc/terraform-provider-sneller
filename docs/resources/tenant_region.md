---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sneller_tenant_region Resource - sneller"
subcategory: ""
description: |-
  Configure a Sneller table
---

# sneller_tenant_region (Resource)

Configure a Sneller table

## Example Usage

```terraform
# Global tenant information
data "sneller_tenant" "tenant" {}

# IAM role that allows access to the Sneller S3 buckets
resource "aws_iam_role" "sneller_s3" {
  name               = "sneller-s3"
  assume_role_policy = data.aws_iam_policy_document.sneller_s3.json
}

# Assume role policy
data "aws_iam_policy_document" "sneller_s3" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = [data.sneller_tenant.tenant.tenant_role_arn]
    }
  }
}

# Sneller cache bucket
resource "aws_s3_bucket" "sneller_cache" {
  # Enable this for production to avoid trashing your cache bucket
  # lifecycle { prevent_destroy = true }

  bucket = "sneller-cache"
}

# Disable all public access to the cache bucket
resource "aws_s3_bucket_public_access_block" "sneller_cache" {
  bucket = aws_s3_bucket.sneller_cache.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Grant Sneller S3 role access to the bucket
resource "aws_s3_bucket_policy" "sneller_cache" {
  bucket = aws_s3_bucket.sneller_cache.id
  policy = data.aws_iam_policy_document.sneller_cache.json
}

data "aws_iam_policy_document" "sneller_cache" {
  # R/W access to the sneller-cache (prefix: /db/)
  statement {
    principals {
      type        = "AWS"
      identifiers = [aws_iam_role.sneller_s3.arn]
    }
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.sneller_cache.arn]
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = ["db/*"]
    }
  }
  statement {
    principals {
      type        = "AWS"
      identifiers = [aws_iam_role.sneller_s3.arn]
    }
    actions   = ["s3:PutObject", "s3:GetObject", "s3:DeleteObject"]
    resources = ["${aws_s3_bucket.sneller_cache.arn}/db/*"]
  }
}

resource "sneller_tenant_region" "test" {
  bucket   = aws_s3_bucket.sneller_cache.bucket
  role_arn = aws_iam_role.sneller_s3.arn
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `bucket` (String) Sneller cache bucket name.
- `role_arn` (String) ARN of the role that is used to access the S3 data in this region's cache bucket. It is also used by the ingestion process to read the source data.

### Optional

- `region` (String) Region from which to fetch the tenant configuration. When not set, then it default's to the tenant's home region.

### Read-Only

- `external_id` (String) External ID (typically the same as the tenant ID) that is passed when assuming the IAM role
- `id` (String) Terraform identifier.
- `last_updated` (String) Timestamp of the last Terraform update.
- `prefix` (String) Prefix of the files in the Sneller cache bucket (always 'db/').
- `sqs_arn` (String) ARN of the SQS resource that is used to signal the ingestion process when new data arrives.

## Import

Import is supported using the following syntax:

```shell
# Region configuration can be imported by specifying the tenant and region
terraform import sneller_tenant_region.test TA0M16BTT6Z4/us-east-1
```