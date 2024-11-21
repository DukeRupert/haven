package types

// UserPageData contains all data needed for user page rendering
type UserPageData struct {
	Title       string
	Description string
	Auth        AuthContext
	User        UserDetails
}