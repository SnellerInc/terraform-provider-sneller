# Sneller cache bucket
resource "aws_s3_bucket" "sneller_source" {
  # Enable this for production to avoid trashing your source bucket
  # lifecycle { prevent_destroy = true }

  bucket = "sneller-source-${lower(data.sneller_tenant.tenant.tenant_id)}-${var.region}"
}

# Disable all public access to the cache bucket
resource "aws_s3_bucket_public_access_block" "sneller_source" {
  bucket = aws_s3_bucket.sneller_source.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "sneller_s3_source" {
  # Read-only access to the sneller-cache (prefix set via `source_prefix` variable)
  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.sneller_source.arn]
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = ["${var.source_prefix}*"]
    }
  }
  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.sneller_source.arn}/${var.source_prefix}*"]
  }
}

# Bucket notification the the region's SQS queue
resource "aws_s3_bucket_notification" "sneller_source" {
  bucket = aws_s3_bucket.sneller_source.id

  queue {
    queue_arn     = sneller_tenant_region.region.sqs_arn
    events        = ["s3:ObjectCreated:*"]
    filter_prefix = var.source_prefix
    # filter_suffix = ".ndjson"
  }
}

# Store an object in the source bucket
resource "aws_s3_object" "object" {
  depends_on = [
    # Store the object in the source bucket after
    # the S3 event notification has been set up. This
    # ensures that the object will be ingested. 
    aws_s3_bucket_notification.sneller_source,

    # Make sure the table definition has been created
    sneller_table.test
  ]

  bucket = aws_s3_bucket.sneller_source.bucket
  key    = "${var.source_prefix}test.ndjson"
  source = "data/test.ndjson"
}
