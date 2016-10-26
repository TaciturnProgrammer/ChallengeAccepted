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
	CID             string `datastore:"-"`
	Activity        string
	Progress        int
	Target          int
	Metric          string
	StartTime       time.Time
	EndTime         time.Time
	Status          string
	ProgressPercent int `datastore:"-"`
	Difference      int `datastore:"-"`
}

var dateFormat = os.Getenv("DATEFORMAT")

//NewChallenge creates new challenge for the user
func NewChallenge(r *http.Request, useremail string) string {
	ctx := appengine.NewContext(r)

	targetString := r.FormValue("Target")
	endTimeString := r.FormValue("EndTime")
	activity := r.FormValue("Activity")
	metric := r.FormValue("Metric")

	if targetString == "" || endTimeString == "" || activity == "" || metric == "" {

		return "Inputs should not be empty"
	}

	endTime, err := time.Parse(dateFormat, endTimeString)
	if err != nil {
		log.Infof(ctx, "endTime err:%v", err)
	}

	target, err := strconv.Atoi(targetString)
	if err != nil {
		return "Target should be a number"
	}

	timeNow := time.Now().Format(dateFormat)
	startTime, _ := time.Parse(dateFormat, timeNow)

	if startTime == endTime {
		return "End date should not be today's date"
	}

	challenge := &Challenge{
		Activity:  activity,
		Progress:  0,
		Target:    target,
		EndTime:   endTime,
		StartTime: startTime,
		Metric:    metric,
	}

	userKey := datastore.NewKey(ctx, "User", useremail, 0, nil)
	challengeKey := datastore.NewIncompleteKey(ctx, "Challenge", userKey)
	_, err = datastore.Put(ctx, challengeKey, challenge)
	if err != nil {
		log.Errorf(ctx, "Error in creating challenge")
	}
	return ""
}

//GetAllChallenges returns all the challenges for the user
func GetAllChallenges(r *http.Request, userKey *datastore.Key) []Challenge {
	ctx := appengine.NewContext(r)

	query := datastore.NewQuery("Challenge").Ancestor(userKey).Order("EndTime")

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
	id := r.FormValue("Id")
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Infof(ctx, "Requested URL: %q", id)
	}

	err = datastore.Delete(ctx, key)
}
