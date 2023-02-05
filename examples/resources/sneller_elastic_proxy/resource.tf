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