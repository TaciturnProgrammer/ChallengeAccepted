package models

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

//Challenge holds the data for a challenge
type Challenge struct {
	Progress        int
	Target          int
	ProgressPercent int
	Difference      int `datastore:"-"`
	Metric          string
	Status          string
	CID             string `datastore:"-"`
	Activity        string
	StartTime       time.Time
	EndTime         time.Time
	Public          bool //if private value is flase
}

var dateFormat = os.Getenv("DATEFORMAT")

//NewChallenge creates new challenge for the user
func NewChallenge(r *http.Request, user *User) string {
	ctx := appengine.NewContext(r)
	r.ParseForm()
	userDate := r.FormValue("currentDate")
	targetString := r.FormValue("Target")
	endTimeString := r.FormValue("EndTime")
	activity := r.FormValue("Activity")
	metric := r.FormValue("Metric")
	publicBox := r.Form["Public"]

	if targetString == "" {
		return "Target should not be empty."
	}

	if endTimeString == "" {
		return "End time should not be empty."
	}

	if activity == "" {
		return "Challenge should not be empty."
	}

	if metric == "" {
		return "Metric should not be empty."
	}

	endTime, err := time.Parse(dateFormat, endTimeString)
	if err != nil {
		log.Errorf(ctx, "endTime err:%v", err)
	}

	target, err := strconv.Atoi(targetString)
	if err != nil {
		return "Target should be a number"
	}

	log.Infof(ctx, "userDate", userDate)

	startTime, err := time.Parse(dateFormat, userDate)
	if err != nil {
		log.Errorf(ctx, "Error in creating startTime", err)
	}

	log.Infof(ctx, "startTime", startTime)

	if startTime == endTime {
		log.Errorf(ctx, "End date should not be today's date", startTime, endTime)
		return "End date should not be today's date"
	}
	public := false
	if len(publicBox) != 0 {
		public = true
	}

	log.Infof(ctx, "public:", public)
	challenge := &Challenge{
		Activity:  activity,
		Progress:  0,
		Target:    target,
		EndTime:   endTime,
		StartTime: startTime,
		Metric:    metric,
		Public:    public,
	}
	challenge.Status = getCurrentStatus(challenge)
	challenge.ProgressPercent = int((float64(challenge.Progress) / float64(challenge.Target)) * 100)

	userKey := GetUserKey(r, user.Email)
	challengeKey := datastore.NewIncompleteKey(ctx, "Challenge", userKey)
	_, err = datastore.Put(ctx, challengeKey, challenge)
	if err != nil {
		log.Errorf(ctx, "Error in creating challenge")
	}
	return ""
}

//EditChallenge edits the challenge for the user
func EditChallenge(r *http.Request) string {
	ctx := appengine.NewContext(r)
	r.ParseForm()
	endTimeString := r.FormValue("editEndTime")
	progressString := r.FormValue("editProgress")
	keyID := r.FormValue("editId")
	publicBox := r.Form["Public"]

	challengeKey, err := datastore.DecodeKey(keyID)
	if err != nil {
		log.Errorf(ctx, "Challenge.go : EditChallenge: Error in decoding key", keyID)
		return "Internal error"
	}

	challenge := Challenge{}

	err = datastore.Get(ctx, challengeKey, &challenge)
	if err != nil {
		log.Errorf(ctx, "Challenge.go : EditChallenge: Error in datastore.Get", err)
		return "Internal error"
	}

	if progressString != "" {
		progress, err := strconv.Atoi(progressString)
		if err != nil {
			return "Progress should be a number"
		}
		if challenge.Target >= progress {
			challenge.Progress = progress
		} else {
			return "Progress cannot be greater than Target"
		}
	}

	//setting end date
	if endTimeString != "" {
		endTime, err := time.Parse(dateFormat, endTimeString)
		if err != nil {
			log.Infof(ctx, "endTime err:%v", err)
		}
		challenge.EndTime = endTime
	}

	//setting if public
	if len(publicBox) != 0 {
		public := true
		challenge.Public = public
	}

	challenge.Status = getCurrentStatus(&challenge)
	challenge.ProgressPercent = int((float64(challenge.Progress) / float64(challenge.Target)) * 100)

	_, err = datastore.Put(ctx, challengeKey, &challenge)
	if err != nil {
		log.Errorf(ctx, "Challenge.go : EditChallenge: Error in putting challenge", err)
		return "Internal error"
	}

	return ""
}

//GetAllInProgressChallenges returns all the challenges for the user
func GetAllInProgressChallenges(r *http.Request, user *User) []Challenge {
	ctx := appengine.NewContext(r)
	userKey := GetUserKey(r, user.Email)
	query := datastore.NewQuery("Challenge").Ancestor(userKey).Filter("ProgressPercent <", 100).Order("ProgressPercent").Order("EndTime")

	challenges := []Challenge{}

	keys, err := query.GetAll(ctx, &challenges)
	if err != nil {
		log.Errorf(ctx, "Error fetching the challenges: GetAll: ", err)
	}

	for i, key := range keys {
		challenges[i].CID = key.Encode()
		challenges[i].Status = getCurrentStatus(&challenges[i])
	}
	return challenges
}

//GetAllPublicChallenges returns all the challenges for the user
func GetAllPublicChallenges(r *http.Request, user *User) []Challenge {
	ctx := appengine.NewContext(r)
	userKey := GetUserKey(r, user.Email)
	query := datastore.NewQuery("Challenge").Ancestor(userKey).Filter("Public =", true).Order("ProgressPercent").Order("EndTime")

	challenges := []Challenge{}

	keys, err := query.GetAll(ctx, &challenges)
	if err != nil {
		log.Errorf(ctx, "Error fetching the challenges: GetAll: ", err)
	}

	for i, key := range keys {
		challenges[i].CID = key.Encode()
		challenges[i].Status = getCurrentStatus(&challenges[i])
	}
	return challenges
}

//GetAllCompletedChallenges returns all the completed challenges for the user
func GetAllCompletedChallenges(r *http.Request, user *User) []Challenge {
	ctx := appengine.NewContext(r)
	userKey := GetUserKey(r, user.Email)
	query := datastore.NewQuery("Challenge").Filter("ProgressPercent =", 100).Ancestor(userKey)

	challenges := []Challenge{}

	keys, err := query.GetAll(ctx, &challenges)
	if err != nil {
		log.Errorf(ctx, "Error fetching the challenges: GetAll: ", err)
	}

	for i, key := range keys {
		challenges[i].CID = key.Encode()
		challenges[i].Status = getCurrentStatus(&challenges[i])
	}
	return challenges
}

func getCurrentStatus(c *Challenge) string {
	if c.Target == c.Progress {
		return "You are done."
	}

	timeElapsedPercent := time.Since(c.StartTime).Hours() / c.EndTime.Sub(c.StartTime).Hours() * 100
	c.ProgressPercent = int((float64(c.Progress) / float64(c.Target)) * 100)

	onpar := int(timeElapsedPercent / 100 * float64(c.Target))
	c.Difference = onpar - c.Progress

	if c.Difference > 0 {
		return "You are " + fmt.Sprintf("%v", c.Difference) + " " + c.Metric + " behind schedule."
	} else if c.Difference < 0 {
		return "You are " + fmt.Sprintf("%v", -c.Difference) + " " + c.Metric + " ahead of schedule."
	}
	return "You are on schedule."
}

//DeleteChallenge deletes chalenge by given id
func DeleteChallenge(r *http.Request) {
	ctx := appengine.NewContext(r)
	id := r.FormValue("deleteId")
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "DeleteChallenge", err, id, key)
	}

	err = datastore.Delete(ctx, key)
	if err != nil {
		log.Errorf(ctx, "datastore.Delete(ctx, key)", key, err)
	}
}
