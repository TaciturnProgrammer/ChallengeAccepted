package models

//User is for datastore
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
