data "sneller_elastic_proxy" "test" {
  region = "us-east-1"
}

# Generate a map that contains the
# database and table for each index
output "indexes" {
  value = {
    for index, conf in data.sneller_elastic_proxy.test.index : index => {
      database = conf.database,
      table    = conf.table,
    }
  }
}
