# AWS terraform
The integration tests require a fixed set of AWS resources to enable execution.
This Terraform definition creates the required AWS resources. This script
creates the following resources:

 - Two IAM roles that can be used to access this tenant's S3 buckets.
 - Two S3 buckets that are used as Sneller cache buckets

Make sure to set the following two variables:

 - `TF_VAR_SNELLERTENANTID` contains the Sneller tenant ID (i.e.
   `TA0M1K9FPAKR`) that is created in Sneller. If you don't have created a
   tenant yet, then register a new tenant on the Sneller website.
 - `TF_VAR_SNELLERAWSACCOUNTID` contains the AWS account ID (i.e.
   `701831592002`) of the Sneller AWS account that holds the tenant. The
   default value points to the production AWS account and should be used for
   normal testing.

Run `terraform init; terraform apply` to create the AWS resources.