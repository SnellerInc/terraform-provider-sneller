# Obtain a list of databases
data "sneller_databases" "test" {
  region = "us-east-1"
}

# Obtain a list of tables per database
data "sneller_database" "test" {
  region   = data.sneller_databases.test.region
  for_each = toset(data.sneller_databases.test.databases)
  database = each.key
}

locals {
  # tables contains an array of tuples that
  # hold all the database/table combinations
  # in this region
  tables = flatten([
    for k, v in data.sneller_database.test : [
      for t in v.tables : { database = k, table = t }
    ]
  ])
}

# Obtain a table information for all tables
# in this region
data "sneller_table" "test" {
  region   = data.sneller_databases.test.region
  for_each = {for t in local.tables : "${t.database}/${t.table}" => t}
  database = each.value.database
  table    = each.value.table
}

# inputs contains a list of inputs for every
# table in this region
output "inputs" {
  value = {
    for k, v in data.sneller_table.test : k => [for i in v.inputs : i.pattern]
  }
}
