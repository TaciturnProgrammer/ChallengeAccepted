package models

import (
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

//Notifications holds the data for the incoming requets for a given user
type Notifications struct {
	Friends    int
	Challenges int
}

//CreateNewNotifications creates a new entry for a user
func CreateNewNotifications(r *http.Request, userKey *datastore.Key) error {
	ctx := appengine.NewContext(r)
	notifications := Notifications{
		Friends:    0,
		Challenges: 0,
	}
	key := datastore.NewKey(ctx, "Notifications", "", 0, userKey)
	_, err := datastore.Put(ctx, key, &notifications)
	if err != nil {
		log.Errorf(ctx, "models.Notifications: CreateNewNotifications : Error in creating Notifications", err)
		return err
	}
	return nil
}

//UpdateNotifications updates the users notifications
func UpdateNotifications(r *http.Request, userKey *datastore.Key, friends int, challenges int) error {
	ctx := appengine.NewContext(r)
	notification := &Notifications{}
	query := datastore.NewQuery("Notifications").Ancestor(userKey).KeysOnly()
	keys, err := query.GetAll(ctx, nil)
	log.Infof(ctx, "models.Notifications: UpdateNotifications : Error in getting Notifications", userKey, keys)
	if err != nil {
		log.Errorf(ctx, "models.Notifications: UpdateNotifications : Error in getting Notifications", err)
		return err
	}
	//we should just have one notifications per user
	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		// Note: this function's argument ctx shadows the variable ctx
		//       from the surrounding function.
		err := datastore.Get(ctx, keys[0], notification)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}
		notification.Friends += friends
		notification.Challenges += challenges
		_, err = datastore.Put(ctx, keys[0], notification)
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

//GetUserNotifications returns user notifications
func GetUserNotifications(r *http.Request, userKey *datastore.Key) (Notifications, error) {
	ctx := appengine.NewContext(r)
	query := datastore.NewQuery("Notifications").Ancestor(userKey)
	notifications := []Notifications{}
	_, err := query.GetAll(ctx, &notifications)
	if err != nil {
		log.Errorf(ctx, "models.Notifications: GetUserNotifications : Error in getting Notifications", err)
	}
	return notifications[0], err
}
