---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sneller_databases Data Source - sneller"
subcategory: ""
description: |-
  Provides configuration of the tenant.
---

# sneller_databases (Data Source)

Provides configuration of the tenant.



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `region` (String) Region from which to fetch the databases. When not set, then it default's to the tenant's home region.

### Read-Only

- `databases` (Set of String) Set of databases in the specified region.
- `id` (String) Terraform identifier.
- `location` (String) S3 url where the databases are stored (i.e. `s3://sneller-cache-bucket/db/`).

