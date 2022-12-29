locals {
  roles = toset([
    "role1-${lower(var.snellerTenantId)}",
    "role2-${lower(var.snellerTenantId)}",
  ])
}

resource "aws_iam_role" "role" {
  for_each           = local.roles
  name               = each.key
  assume_role_policy = data.aws_iam_policy_document.sneller.json
}

data "aws_iam_policy_document" "sneller" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${var.snellerAwsAccountID}:role/tenant-${var.snellerTenantId}"]
    }
  }
}
