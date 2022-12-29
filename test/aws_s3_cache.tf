# Sneller cache bucket
resource "aws_s3_bucket" "sneller_cache" {
  # Enable this for production to avoid trashing your cache bucket
  # lifecycle { prevent_destroy = true }

  bucket = "sneller-cache-${lower(data.sneller_tenant.tenant.tenant_id)}-${var.region}"
}

# Disable all public access to the cache bucket
resource "aws_s3_bucket_public_access_block" "sneller_cache" {
  bucket = aws_s3_bucket.sneller_cache.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "sneller_s3_cache" {
  # R/W access to the sneller-cache (prefix: /db/)
  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.sneller_cache.arn]
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = ["db/*"]
    }
  }
  statement {
    actions   = ["s3:PutObject", "s3:GetObject", "s3:DeleteObject"]
    resources = ["${aws_s3_bucket.sneller_cache.arn}/db/*"]
  }
}
