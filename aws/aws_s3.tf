locals {
  cache_buckets = toset([
    "sneller-cache1-${lower(var.snellerTenantId)}",
    "sneller-cache2-${lower(var.snellerTenantId)}",
  ])
}

# Sneller cache bucket
resource "aws_s3_bucket" "sneller_cache" {
  for_each = local.cache_buckets
  bucket   = each.key
}

# Disable all public access to the cache bucket
resource "aws_s3_bucket_public_access_block" "sneller_cache" {
  for_each = local.cache_buckets
  bucket   = aws_s3_bucket.sneller_cache[each.key].id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Grant Sneller S3 role access to the bucket
resource "aws_s3_bucket_policy" "sneller_cache" {
  for_each = local.cache_buckets
  bucket   = aws_s3_bucket.sneller_cache[each.key].id
  policy   = data.aws_iam_policy_document.sneller_cache[each.key].json
}

data "aws_iam_policy_document" "sneller_cache" {
  for_each = local.cache_buckets

  # R/W access to the sneller-cache (prefix: /db/)
  statement {
    principals {
      type        = "AWS"
      identifiers = [for r in local.roles : aws_iam_role.role[r].arn]
    }
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.sneller_cache[each.key].arn]
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = ["db/*"]
    }
  }
  statement {
    principals {
      type        = "AWS"
      identifiers = [for r in local.roles : aws_iam_role.role[r].arn]
    }
    actions   = ["s3:PutObject", "s3:GetObject", "s3:DeleteObject"]
    resources = ["${aws_s3_bucket.sneller_cache[each.key].arn}/db/*"]
  }
}
