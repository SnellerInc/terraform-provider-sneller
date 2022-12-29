# IAM role that allows access to the Sneller S3 buckets
resource "aws_iam_role" "sneller_s3" {
  name               = "sneller-s3"
  assume_role_policy = data.aws_iam_policy_document.sneller_s3.json
}

data "aws_iam_policy_document" "sneller_s3" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = [data.sneller_tenant.tenant.tenant_role_arn]
    }
  }
}

resource "aws_iam_role_policy" "sneller_s3_cache" {
  role   = aws_iam_role.sneller_s3.id
  name   = "s3-cache"
  policy = data.aws_iam_policy_document.sneller_s3_cache.json
}

resource "aws_iam_role_policy" "sneller_s3_source" {
  role   = aws_iam_role.sneller_s3.id
  name   = "s3-source"
  policy = data.aws_iam_policy_document.sneller_s3_source.json
}
