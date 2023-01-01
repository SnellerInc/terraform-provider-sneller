---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sneller_user Resource - sneller"
subcategory: ""
description: |-
  Configure a Sneller user.
---

# sneller_user (Resource)

Configure a Sneller user.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email` (String) Email address.

### Optional

- `family_name` (String) User's family name.
- `given_name` (String) User's given name.
- `is_admin` (Boolean) Administrator.
- `is_enabled` (Boolean) User enabled.
- `locale` (String) User's locale.

### Read-Only

- `id` (String) Terraform identifier.
- `is_federated` (Boolean) User is using a federated identity provider.
- `last_updated` (String) Timestamp of the last Terraform update.
- `user_id` (String) User identifier.

