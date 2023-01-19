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

Then, navigate to the `test` directory. 

```shell
$ cd test
```

Run the following command to initialize the workspace and apply the sample configuration (make sure you set `SNELLER_TOKEN`).

```shell
$ export TF_VAR_SNELLER_TOKEN=<...>
$ terraform init && terraform apply
```

This will automatically:
1. Create the IAM role that can be used by Sneller to access the S3 bucket.
1. Create the Sneller source bucket (Sneller can only read).
1. Create the Sneller cache bucket (Sneller can read/write).
1. Assign the IAM role to the Sneller configuration for this region (default `us-east-1`).
1. Setup the S3 event notification that signals Sneller when new data arrives in the source bucket.
1. Create a table definition (db: `test-db`, table: `test-table`).
1. Upload a test JSON file.

## Debugging
When you need to debug a resource/data-source, then debugging the integration tests is typically
the most straightforward option. Testing the provider requires some environment variables:

 * `TF_ACC` should be set to `1` to enable running the acceptance tests.
 * `SNELLER_TOKEN` should be set to the bearer token of your tenant.
 * `TENANT_ACCOUNT_ID` is optional and should be set to the AWS account identifier. If it's
   not set, then it defaults to the AWS account identifier of the AWS variables in the environment.
 * `SNELLER_API_ENDPOINT` is optional and defaults to the default API endpoint of the production
   environment. This can be changed when testing against another end-point.

The testing process is quite slow, due to the amount of API calls. Therefore, it's recommended to
set the time-out to `10m`.

You can use the following `.vscode/settings.json` file in the root of the repository to enable the
settings for running unit tests from within Visual Studio Code:
```json
{
    "go.testEnvVars": {
        "TF_ACC": "1",
        "SNELLER_TOKEN": "SA0xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
        //"TENANT_ACCOUNT_ID": "093008718846",
        //"SNELLER_API_ENDPOINT": "http://localhost:8080",
    },
    "go.testTimeout": "10m",
}
```

## Running examples
The `examples` folder contain the Terraform samples for the documentation, but can also run
independent. Make sure to set the `TF_VAR_sneller_token`