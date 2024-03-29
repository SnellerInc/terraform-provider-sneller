---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sneller_elastic_proxy Resource - sneller"
subcategory: ""
description: |-
  Configure the Elastic proxy.
---

# sneller_elastic_proxy (Resource)

Configure the Elastic proxy.

## Example Usage

```terraform
resource "sneller_elastic_proxy" "test" {
  log_path = "s3://sneller-ta0m19q7vjkd-0ce1/db/log-elastic-proxy/"
  log_flags = {
        log_request          = true
        log_query_parameters = true
        log_result           = true // may contain sensitive info
  }
  index = {
    test = {
      database                   = "test-db"
      table                      = "test-table"
      ignore_total_hits          = true
      ignore_sum_other_doc_count = true

      type_mapping = {
        timestamp = {
          type = "unix_nano_seconds"
        }
        description = {
          type = "contains"
          fields = {
            raw = "text"
          }
        }
        "*_time": {
          type = "datetime"
        }
      }
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `index` (Attributes Map) Configures an Elastic index that maps to a Sneller table. (see [below for nested schema](#nestedatt--index))
- `log_flags` (Attributes) Logging flags. (see [below for nested schema](#nestedatt--log_flags))
- `log_path` (String) Location where Elastic Proxy logging is stored (i.e. `s3://logging-bucket/elastic-proxy/`). Make sure Sneller is allowed to write to this S3 bucket.
- `region` (String) Region for which to configure the Elastic Proxy. If not set, then the configuration is assumed to be located in the tenant's home region.

### Read-Only

- `id` (String) Terraform identifier.
- `location` (String) Location of the Elastic proxy configuration file (i.e. `s3://sneller-cache-bucket/db/elastic-proxy.json`).

<a id="nestedatt--index"></a>
### Nested Schema for `index`

Required:

- `database` (String) Sneller database.
- `table` (String) Sneller table.

Optional:

- `ignore_total_hits` (Boolean) Ignore 'total_hits' in Elastic response (more efficient).
- `ignore_sum_other_doc_count` (Boolean) Ignore 'sum_other_doc_count' in Elastic response (more efficient).
- `type_mapping` (Attributes Map) Custom type mappings. (see [below for nested schema](#nestedatt--index--type_mapping))

<a id="nestedatt--index--type_mapping"></a>
### Nested Schema for `index.type_mapping`

Required:

- `type` (String) Type.

Optional:

- `fields` (Map of String) Field mappings.



<a id="nestedatt--log_flags"></a>
### Nested Schema for `log_flags`

Optional:

- `log_preprocessed` (Boolean) Log preprocessed Sneller results (may be verbose and contain sensitive data).
- `log_query_parameters` (Boolean) Log query parameters.
- `log_request` (Boolean) Log all requests.
- `log_result` (Boolean) Log result (may be verbose and contain sensitive data).
- `log_sneller_result` (Boolean) Log Sneller query result (may be verbose and contain sensitive data).
- `log_sql` (Boolean) Log generated SQL query.

## Import

Import is supported using the following syntax:

```shell
# Elastic proxy configuration can be imported by specifying the tenant and region
terraform import sneller_elastic_proxy.test TA0XXXXXXXXX/us-east-1
```
