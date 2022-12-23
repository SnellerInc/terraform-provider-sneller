# Terraform Provider Sneller

This repo implements the [Terraform](https://www.terraform.io/) provider for Sneller.

## Build provider

Run the following command to build the provider

```shell
$ go build -o terraform-provider-sneller
```

## Local release build

```shell
$ go install github.com/goreleaser/goreleaser@latest
```

```shell
$ make release
```

You will find the releases in the `/dist` directory. You will need to rename the provider binary to `terraform-provider-sneller` and move the binary into [the appropriate subdirectory within the user plugins directory](https://learn.hashicorp.com/tutorials/terraform/provider-use?in=terraform/providers#install-sneller-provider).

## Test sample configuration

First, build and install the provider.

```shell
$ make install
```

Then, navigate to the `examples` directory. 

```shell
$ cd examples
```

Run the following command to initialize the workspace and apply the sample configuration (make sure you set `SNELLER_TOKEN`).

```shell
$ export TF_VAR_SNELLER_TOKEN=<...>
$ terraform init && terraform apply
```

This will automatically:
1. Create the IAM role that can be used by Sneller to access the S3 bucket.
1. Create the Sneller source bucket (Sneller can only read).
1. Create the Sneller cache bucket (Snelelr can read/write).
1. Assign the IAM role to the Sneller configuration for this region (default `us-east-1`).
1. Setup the S3 event notification that signals Sneller when new data arrives in the source bucket.
1. Create a table definition (db: `test-db`, table: `test-table`).
1. Upload a test JSON file.