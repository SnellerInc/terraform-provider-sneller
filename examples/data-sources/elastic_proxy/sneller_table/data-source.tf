data "sneller_table" "test" {
  database = "test-db"
  table    = "test-table"
}

output "inputs" {
  value = [for i in data.sneller_table.test.inputs : i.pattern]
}

output "partition_fields" {
  value = [for p in data.sneller_table.test.partitions : p.field]  
}
