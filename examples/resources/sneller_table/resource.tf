resource "sneller_table" "test" {
  # Enable this for production to avoid trashing your table
  # lifecycle { prevent_destroy = true }

  database = "test-db"
  table    = "test-table"

  inputs = [
    {
      pattern = "s3://sneller-source-data/data/{tenant}/{yyyy}/{mm}/{dd}/*.ndjson"
      format  = "json"

      json_hints =  [
        {
          field = "timestamp", 
          hints = ["unix_nano_seconds"]
        }
      ]
    }
  ]

  partitions = [
    {
      field = "tenant"
    },{
      field = "date"
      type  = "datetime"
      value = "{yyyy}-{mm}-{dd}T00:00:00Z"
    }
  ]

  retention_policy = {
    field     = "timestamp"
    valid_for = "100d"
  }
}
