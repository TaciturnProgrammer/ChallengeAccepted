package models

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/taciturnprogrammer/challengeaccepted/auth"
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

//CreateNewUser using email and returns a datastore key and adds an entry in the notifications
func CreateNewUser(r *http.Request, user *auth.User) error {
	ctx := appengine.NewContext(r)

	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		// Note: this function's argument ctx shadows the variable ctx
		//       from the surrounding function.
		newuser := &User{
			Provider:    user.Provider,
			Email:       user.Email,
			Name:        user.Name,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			NickName:    user.NickName,
			Description: user.Description,
			UserID:      user.UserID,
			AvatarURL:   user.AvatarURL,
		}
		userKey := GetUserKey(r, newuser.Email)
		_, err := datastore.Put(ctx, userKey, newuser)
		if err != nil {
			log.Errorf(ctx, "models.User: CreateNewUser : Error in creating user", err, userKey, newuser, *newuser, &newuser)
			return err
		}

		err = CreateNewNotifications(r, userKey)
		if err != nil {
			log.Errorf(ctx, "models.User: CreateNewUser : Error in creating CreateNewNotifications", err, userKey)
			return err
		}
		return err
	}, nil)
	if err != nil {
		log.Errorf(ctx, "Transaction failed: %v", err)
		var w http.ResponseWriter
		http.Error(w, "Internal Server Error", 500)
		return err
	}

	return nil
}

//GetUsersByName returns users by name
func GetUsersByName(r *http.Request) []User {
	ctx := appengine.NewContext(r)
	searchString := r.FormValue("searchString")
	query := datastore.NewQuery("User").Filter("LastName =", searchString)
	users := []User{}
	_, err := query.GetAll(ctx, &users)
	if err != nil {
		log.Errorf(ctx, "models.user GetUsersByName: Error fetching the Users: ", err)
	}
	return users
}

//GetUsersByKeys retrieves users from keys
func GetUsersByKeys(r *http.Request, keys []*datastore.Key) []User {
	ctx := appengine.NewContext(r)
	users := make([]User, len(keys))
	err := datastore.GetMulti(ctx, keys, &users)
	if err != nil {
		log.Errorf(ctx, "models.user GetUsersByKeys: Error fetching the Users: ", err, keys, users)
	}
	return users
}

//GetUserKey returns userkey
func GetUserKey(r *http.Request, email string) *datastore.Key {
	ctx := appengine.NewContext(r)
	return datastore.NewKey(ctx, "User", email, 0, nil)
}
