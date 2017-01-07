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
	s0 = iota // 0 - Request sent,
	s1 = iota // 1 - Incoming Request,
	s2 = iota // 2 - we are friends yay!!,
	s3 = iota // 3 - Get off my lawn!!,
	s4 = iota // 4 - We are not friends Anymore
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

		return err
	}, trasnactionoptions)
	if err != nil {
		log.Errorf(ctx, "Transaction failed: %v", err)
		var w http.ResponseWriter
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = UpdateNotifications(r, recipientUserKey, 1, 0)
	log.Infof(ctx, "models.NewFriendRequest: UpdateNotifications", recipientUserKey)
	if err != nil {
		log.Errorf(ctx, "Error models.NewFriendRequest: UpdateNotifications", err)
	}

}

//AcceptFriendRequest accept the request - Status - 2
func AcceptFriendRequest(r *http.Request, user *User) {
	ctx := appengine.NewContext(r)
	recipientEmail := r.FormValue("recipientEmail")
	senderUserKey := datastore.NewKey(ctx, "User", user.Email, 0, nil)
	recipientUserKey := datastore.NewKey(ctx, "User", recipientEmail, 0, nil)

	trasnactionoptions := &datastore.TransactionOptions{
		XG:       true,
		Attempts: 1,
	}
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		//get sender friend
		//change Status
		//put/update the friend

		query := datastore.NewQuery("Friend").Filter("UserKey = ", recipientUserKey).Ancestor(senderUserKey)
		log.Infof(ctx, "1 models.AcceptFriendRequest: senderUserKey: ", senderUserKey)
		senderFriend := []Friend{}

		keys, err := query.GetAll(ctx, &senderFriend)
		if err != nil {
			log.Errorf(ctx, "2 AcceptFriendRequest: Error getting query.GetAll(ctx, &senderFriend) in datastore", err)
			return err
		}
		senderFriend[0].Status = s2
		_, err = datastore.Put(ctx, keys[0], &senderFriend[0])
		if err != nil {
			log.Errorf(ctx, "3 AcceptFriendRequest: Error putting senderUserKey in datastore", err, senderUserKey)
			return err
		}

		//add friend to reciever
		query = datastore.NewQuery("Friend").Filter("UserKey = ", senderUserKey).Ancestor(recipientUserKey)
		log.Infof(ctx, "4 models.AcceptFriendRequest: friendKindKey: ", recipientUserKey)
		reciepientFriend := []Friend{}

		keys, err = query.GetAll(ctx, &reciepientFriend)
		if err != nil {
			log.Errorf(ctx, "5 AcceptFriendRequest: Error getting senderFriendKindKey in datastore", err)
			return err
		}
		reciepientFriend[0].Status = s2
		_, err = datastore.Put(ctx, keys[0], &reciepientFriend[0])
		if err != nil {
			log.Errorf(ctx, "6 AcceptFriendRequest: Error putting friendKindKey in datastore", err)
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

	err = UpdateNotifications(r, senderUserKey, -1, 0)
	log.Infof(ctx, "7 models.AcceptFriendRequest: : UpdateNotifications", senderUserKey)
	if err != nil {
		log.Errorf(ctx, "8 Error models.AcceptFriendRequest: : UpdateNotifications", err)
	}

}

//RejectFriendRequest rejects the request - Status - 3
func RejectFriendRequest(r *http.Request, useremail string) {
}

//DeleteFriend deletes a friend of the user- Status - 4
func DeleteFriend(r *http.Request, useremail string) {
}
