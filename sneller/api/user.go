package api

type User struct {
	UserID      string
	Email       string
	IsEnabled   bool
	IsFederated bool
	Locale      string
	GivenName   string
	FamilyName  string
	Picture     string
	Groups      []string // only returned when fetching user details
}
