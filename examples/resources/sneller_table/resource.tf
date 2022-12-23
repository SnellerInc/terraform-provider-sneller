resource "sneller_table" "test" {
  # Enable this for production to avoid trashing your table
  # lifecycle { prevent_destroy = true }

  database = "test-db"
  table    = "test-table"

  input {
    pattern = "s3://sneller-source-data/data/*.ndjson"
    format  = "json"
  }
}
