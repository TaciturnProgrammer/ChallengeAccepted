package auth

// User contains the information common amongst most OAuth and OAuth2 providers.
type User struct {
	Provider    string
	Email       string
	Name        string
	FirstName   string
	LastName    string
	NickName    string
	Description string
	UserID      string
	AvatarURL   string
}
