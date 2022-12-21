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
