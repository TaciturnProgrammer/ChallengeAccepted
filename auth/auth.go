package auth

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	uuid "github.com/satori/go.uuid"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const google = "google"
const facebook = "facebook"

var (
	//oauthStateString is some random string, random for each request
	oauthStateString = uuid.NewV4().String()
	store            = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET_KEY")))
)

//OAuthHandler handles the OAuth provider type
func OAuthHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]

	authurl := ""
	if provider == google {
		authurl = googleOAuthConfig.AuthCodeURL(oauthStateString)
	} else if provider == facebook {
		authurl = facebookOAuthConfig.AuthCodeURL(oauthStateString)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.Redirect(w, r, authurl, http.StatusTemporaryRedirect)
}

//OAuthCallbackHandler gets the user data from the OAuth provider
func OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) *User {
	ctx := appengine.NewContext(r)

	state := r.FormValue("state")
	if state != oauthStateString {
		log.Infof(ctx, "invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", 500)
		return nil
	}

	vars := mux.Vars(r)
	provider := vars["provider"]

	var user *User
	if provider == google {
		user = googleOAuthHandler(w, r)
	} else if provider == facebook {
		user = facebookOAuthHandler(w, r)
	}

	//create session with user email
	session, err := store.Get(r, "session")
	if err != nil {
		log.Errorf(ctx, "OAuthCallbackHandler : Error in getting session")
	}

	session.Values["user"] = user.Email

	err = session.Save(r, w)
	if err != nil {
		log.Errorf(ctx, "OAuthCallbackHandler : Error in creating session")
	}

	return user
}
