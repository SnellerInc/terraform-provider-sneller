# Sneller cache bucket
resource "aws_s3_bucket" "sneller_cache" {
  # Enable this for production to avoid trashing your cache bucket
  # lifecycle { prevent_destroy = true }

  bucket = "sneller-cache-${data.aws_caller_identity.current.account_id}-${var.region}"
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
