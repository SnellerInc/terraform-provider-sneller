---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sneller_elastic_proxy Data Source - sneller"
subcategory: ""
description: |-
  Elastic proxy configuration.
---

# sneller_elastic_proxy (Data Source)

Elastic proxy configuration.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `region` (String) Region for which to obtain the Elastic Proxy configuration.

### Read-Only

- `id` (String) Terraform identifier.
- `index` (Attributes Map) Configures an Elastic index that maps to a Sneller table. (see [below for nested schema](#nestedatt--index))
- `location` (String) Location of the Elastic proxy configuration file (i.e. `s3://sneller-cache-bucket/db/elastic-proxy.json`).
- `log_flags` (Attributes) Logging flags (see [below for nested schema](#nestedatt--log_flags))
- `log_path` (String) Location where Elastic Proxy logging is stored (i.e. `s3://logging-bucket/elastic-proxy/`).

<a id="nestedatt--index"></a>
### Nested Schema for `index`

Read-Only:

- `database` (String) Sneller database.
- `ignore_total_hits` (Boolean) Ignore 'total_hits' in Elastic response (more efficient).
- `ignore_sum_other_doc_count` (Boolean) Ignore 'sum_other_doc_count' in Elastic response (more efficient).
- `table` (String) Sneller table.
- `type_mapping` (Attributes Map) Custom type mappings. (see [below for nested schema](#nestedatt--index--type_mapping))

<a id="nestedatt--index--type_mapping"></a>
### Nested Schema for `index.type_mapping`

Read-Only:

- `fields` (Map of String) Field mappings.
- `type` (String) Type.



<a id="nestedatt--log_flags"></a>
### Nested Schema for `log_flags`

Read-Only:

- `log_preprocessed` (Boolean) Log preprocessed Sneller results (may be verbose and contain sensitive data).
- `log_query_parameters` (Boolean) Log query parameters.
- `log_request` (Boolean) Log all requests.
- `log_result` (Boolean) Log result (may be verbose and contain sensitive data).
- `log_sneller_result` (Boolean) Log Sneller query result (may be verbose and contain sensitive data).
- `log_sql` (Boolean) Log generated SQL query.


