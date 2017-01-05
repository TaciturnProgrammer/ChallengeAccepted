package models

import (
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

//Friend of the ancestor user
type Friend struct {
	Status  int //  0 - Request sent, 1 - Incoming Request, 2 - we are friends yay!!, 3 - Get off my lawn!!, 4 - We are not friends Anymore
	UserKey *datastore.Key
}

const (
	s0 = iota
	s1 = iota
	s2 = iota
	s3 = iota
	s4 = iota
)

//GetAllFriends for the user
func GetAllFriends(r *http.Request, user *User) []Friend {
	ctx := appengine.NewContext(r)
	userKey := datastore.NewKey(ctx, "User", user.Email, 0, nil)
	query := datastore.NewQuery("Friend").Ancestor(userKey).Order("-Status")
	friends := []Friend{}
	_, err := query.GetAll(ctx, &friends)
	if err != nil {
		log.Errorf(ctx, "Error fetching the fiends: GetAll: ", err)
	}
	return friends
}

//NewFriendRequest new friend request - Status - 0
func NewFriendRequest(r *http.Request, user *User) {
	ctx := appengine.NewContext(r)
	keyID := r.FormValue("recipientID")

	senderUserKey := datastore.NewKey(ctx, "User", user.Email, 0, nil)
	recipientUserKey, err := datastore.DecodeKey(keyID)
	if err != nil {
		log.Errorf(ctx, "Error Decoding recipientUserKey from request : NewFriendRequest()", err, keyID)
	}
	trasnactionoptions := &datastore.TransactionOptions{
		XG:       true,
		Attempts: 3,
	}
	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		//add friend to sender
		senderFriendKindKey := datastore.NewKey(ctx, "Friend", keyID, 0, senderUserKey)
		senderFriend := Friend{
			Status:  s0,
			UserKey: recipientUserKey,
		}

		_, err = datastore.Put(ctx, senderFriendKindKey, &senderFriend)
		if err != nil {
			log.Errorf(ctx, "Error putting senderFriendKindKey in datastore", err, senderFriendKindKey)
			return err
		}

		//add friend to reciever
		friendKindKey := datastore.NewKey(ctx, "Friend", senderUserKey.Encode(), 0, recipientUserKey)
		reciepientFriend := Friend{
			Status:  s1,
			UserKey: senderUserKey,
		}

		_, err = datastore.Put(ctx, friendKindKey, &reciepientFriend)
		if err != nil {
			log.Errorf(ctx, "Error putting friendKindKey in datastore", err)
			return err
		}

		err = UpdateNotifications(r, recipientUserKey, 1, 0)
		log.Infof(ctx, "models.NewFriendRequest: UpdateNotifications", recipientUserKey)
		if err != nil {
			log.Errorf(ctx, "Error models.NewFriendRequest: UpdateNotifications", err)
			return err
		}

		return err
	}, trasnactionoptions)
	if err != nil {
		log.Errorf(ctx, "Transaction failed: %v", err)
		var w http.ResponseWriter
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

//AcceptFriendRequest accept the request - Status - 2
func AcceptFriendRequest(r *http.Request, user *User) {
}

//RejectFriendRequest rejects the request - Status - 3
func RejectFriendRequest(r *http.Request, useremail string) {
}

//DeleteFriend deletes a friend of the user- Status - 4
func DeleteFriend(r *http.Request, useremail string) {
}
