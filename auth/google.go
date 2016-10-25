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
	goog "golang.org/x/oauth2/google"
)

var googleOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENTID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENTSECRET"),
	RedirectURL:  "http://localhost:8080/OAuth/Callback/google",
	Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
		"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint: goog.Endpoint,
}

//GoogleUser is used to unmarshall google oauth data in auth package
type googleUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
	Link      string `json:"link"`
	Picture   string `json:"picture"`
}

func googleOAuthHandler(w http.ResponseWriter, r *http.Request) *User {
	ctx := appengine.NewContext(r)
	code := r.FormValue("code")
	token, err := googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		log.Infof(ctx, "googleOAuthConfig.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", 500)
		return nil
	}

	client := urlfetch.Client(ctx)
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	defer response.Body.Close()
	if err != nil {
		log.Infof(ctx, "googleOAuthHandler client.Get() failed with '%s'\n", err)
		http.Redirect(w, r, "/", 500)
		return nil
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Infof(ctx, "googleOAuthHandler ioutil.ReadAll(response.Body) failed with '%s'\n", err)
		http.Redirect(w, r, "/", 500)
		return nil
	}

	var googleuser googleUser
	err = json.Unmarshal(contents, &googleuser)
	if err != nil {
		log.Infof(ctx, "googleOAuthHandler json.Unmarshal failed with '%s'\n", err)
		http.Redirect(w, r, "/", 500)
		return nil
	}

	user := &User{
		Provider:  google,
		Name:      googleuser.Name,
		Email:     googleuser.Email,
		FirstName: googleuser.FirstName,
		LastName:  googleuser.LastName,
		UserID:    googleuser.ID,
		AvatarURL: googleuser.Picture,
	}

	return user

}
