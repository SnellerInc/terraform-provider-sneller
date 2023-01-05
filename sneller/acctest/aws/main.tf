terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.15"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

variable "snellerTenantId" {
  type        = string
  description = "Sneller tenant ID"
  default    = "TA0M1UQNTM7H" # development account
  #default     = "TA0M1K9FPAKR" # production account
}

variable "snellerAwsAccountID" {
  type        = string
  description = "Sneller AWS account ID"
  default    = "671229366946" # development account
  #default     = "701831592002"  # production account
}
