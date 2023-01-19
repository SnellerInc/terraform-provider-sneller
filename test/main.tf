terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.15"
    }

    sneller = {
      source  = "snellerinc/sneller"
    }
  }
}

provider "aws" {
  region = var.region
}

provider "sneller" {
  #api_endpoint   = "http://localhost:8080"
  api_endpoint   = "https://api-production.__REGION__.sneller.io"
  default_region = var.region
  token          = var.sneller_token
}

variable "region" {
  type        = string
  description = "AWS region"
  default     = "us-east-1"
}

variable "sneller_token" {
  type        = string
  description = "Sneller token"
}

variable "source_prefix" {
  type        = string
  description = "Source prefix"
  default     = ""
}
