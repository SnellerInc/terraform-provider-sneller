locals {
  roles = toset([
    "role1-${lower(var.snellerTenantId)}",
    "role2-${lower(var.snellerTenantId)}",
  ]),
  roles_and_cache_buckets = [
    for rb in flatten([
      for r in keys(local.roles) : [
        for b in keys(local.cache_buckets) : {
          role         = r
          cache_bucket = b
        }
      ]
    ])
  ]
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

resource "aws_iam_role_policy" "sneller_cache" {
  for_each = local.roles_and_cache_buckets

  role   = aws_iam_role.role[each.value.role].id
  name   = "s3-cache-${each.value.cache_bucket}"
  policy = data.aws_iam_policy_document.sneller_cache[each.value.cache_bucket].json
}
