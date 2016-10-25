package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"golang.org/x/oauth2"
	fb "golang.org/x/oauth2/facebook"
)

var (
	facebookOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("FACEBOOK_CLIENTID"),
		ClientSecret: os.Getenv("FACEBOOK_CLIENTSECRET"),
		RedirectURL:  "http://localhost:8080/OAuth/Callback/facebook",
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     fb.Endpoint,
	}
)

//FaceBookUser is used to unmarshall fb oauth data in auth package
type faceBookUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Link      string `json:"link"`
}

func facebookOAuthHandler(w http.ResponseWriter, r *http.Request) *User {
	ctx := appengine.NewContext(r)
	code := r.FormValue("code")
	token, err := facebookOAuthConfig.Exchange(ctx, code)
	if err != nil {
		log.Infof(ctx, "facebookOAuthConfig.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", 500)
		return nil
	}

	client := urlfetch.Client(ctx)
	response, err := client.Get("https://graph.facebook.com/me?fields=id,name,first_name,last_name,email,gender&access_token=" + token.AccessToken)
	defer response.Body.Close()
	if err != nil {
		log.Infof(ctx, "facebookOAuthHandler client.Get() failed with '%s'\n", err)
		http.Redirect(w, r, "/", 500)
		return nil
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Infof(ctx, "facebookOAuthHandler ioutil.ReadAll(response.Body) failed with '%s'\n", err)
		http.Redirect(w, r, "/", 500)
		return nil
	}

	var fbuser faceBookUser
	err = json.Unmarshal(contents, &fbuser)
	if err != nil {
		log.Infof(ctx, "facebookOAuthHandler json.Unmarshal failed with '%s'\n", err)
		http.Redirect(w, r, "/", 500)
		return nil
	}

	user := &User{
		Provider:  facebook,
		Name:      fbuser.Name,
		Email:     fbuser.Email,
		FirstName: fbuser.FirstName,
		LastName:  fbuser.LastName,
		UserID:    fbuser.ID,
		AvatarURL: "http://graph.facebook.com/" + fbuser.ID + "/picture",
	}

	return user
}
