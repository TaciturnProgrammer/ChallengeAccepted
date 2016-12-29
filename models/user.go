package models

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

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

//GetUsersByName returns users by name
func GetUsersByName(r *http.Request) []User {
	ctx := appengine.NewContext(r)
	searchString := r.FormValue("searchString")
	query := datastore.NewQuery("User").Filter("LastName =", searchString)
	users := []User{}
	_, err := query.GetAll(ctx, &users)
	if err != nil {
		log.Errorf(ctx, "Error fetching the Users: GetUsersByName: ", err)
	}
	return users
}

//GetUsersByKeys retrieves users from keys
func GetUsersByKeys(r *http.Request, keys []*datastore.Key) (users []User) {
	ctx := appengine.NewContext(r)
	err := datastore.GetMulti(ctx, keys, users)
	if err != nil {
		log.Errorf(ctx, "Error fetching the Users: GetAGetUsersByKeysll: ", err)
	}
	return users
}
