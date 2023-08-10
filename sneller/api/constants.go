package api

const (
	AdminGroup            = "admin"
	EnvSnellerApiEndpoint = "SNELLER_API_ENDPOINT"
	EnvSnellerToken       = "SNELLER_TOKEN"
	EnvSnellerRegion      = "SNELLER_REGION"
	DefaultApiEndPoint    = "https://api-production.__REGION__.sneller.ai"
	DefaultSnellerRegion  = "us-east-1"
	DefaultDbPrefix       = "db/"
)

var (
	Formats = []string{`json`, `json.gz`, `json.zst`, `cloudtrail.json.gz`, `csv`, `csv.gz`, `csv.zst`, `tsv`, `tsv.gz`, `tsv.zst`}
)
