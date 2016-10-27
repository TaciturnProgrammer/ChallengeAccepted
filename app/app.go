package app

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/taciturnprogrammer/challengeaccepted/auth"
	"github.com/taciturnprogrammer/challengeaccepted/middleware"
	"github.com/taciturnprogrammer/challengeaccepted/models"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

var dateFormat = os.Getenv("DATEFORMAT")

var templatesDir = "../templates/"

var templates map[string]*template.Template

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET_KEY")))

func init() {
	//init templates
	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	layouts, _ := filepath.Glob(templatesDir + "layouts/*.gohtml")
	includes, _ := filepath.Glob(templatesDir + "includes/*.gohtml")

	// Generate our templates map from our layouts/ and includes/ directories
	for _, layout := range layouts {
		files := append(includes, layout)
		templates[filepath.Base(layout)] = template.Must(template.ParseFiles(files...))
	}

	//routing stuff
	router := mux.NewRouter()

	//OAuth routes
	authsubrouter := router.PathPrefix("/OAuth/").Subrouter()
	authsubrouter.HandleFunc("/Authorize/{provider}", auth.OAuthHandler)
	authsubrouter.HandleFunc("/Callback/{provider}", authenticationHandler)

	//App routes
	approuter := router.PathPrefix("/app/").Subrouter()
	approuter.HandleFunc("/home", homePageHandler)
	approuter.HandleFunc("/logout", logoutHandler)
	approuter.HandleFunc("/addChallenge", createChallenge).Methods("POST")
	approuter.HandleFunc("/editChallenge", editChallenge).Methods("POST")
	approuter.HandleFunc("/deleteChallenge", deleteChallenge).Methods("POST")
	http.Handle("/app/", middleware.AuthMiddleware(approuter))

	router.HandleFunc("/", landingPageHandler)
	http.Handle("/", router)

}

func renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	// Ensure the template exists in the map.
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("The template %s does not exist.", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", data)
}

func landingPageHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	if session.IsNew || session.Values["user"] == nil {
		//not logged in, show login page
		renderTemplate(w, "landing.gohtml", nil)
		return
	}
	http.Redirect(w, r, "/app/home", 302)
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	session, _ := store.Get(r, "session")

	useremail := session.Values["user"]
	if useremail == nil {
		//if logged in redirect to home
		renderTemplate(w, "landing.gohtml", nil)
		log.Infof(ctx, "No user in session")
		return
	}

	//Get user data from datastore
	var user models.User
	userKey := datastore.NewKey(ctx, "User", useremail.(string), 0, nil)
	err := datastore.Get(ctx, userKey, &user)
	if err != nil {
		log.Errorf(ctx, "authenticationHandler : Error in retrieving user from datastore")
		renderTemplate(w, "landing.gohtml", nil)
		return
	}

	challenges := models.GetAllChallenges(r, userKey)

	var tmplData = make(map[string]interface{})
	tmplData["User"] = user
	tmplData["Challenges"] = challenges

	renderTemplate(w, "home.gohtml", tmplData)
}

func authenticationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	user := auth.OAuthCallbackHandler(w, r)

	//Add user to datastore
	_, err := datastore.Put(ctx, datastore.NewKey(ctx, "User", user.Email, 0, nil), user)
	if err != nil {
		log.Errorf(ctx, "authenticationHandler : Error in creating user")
		http.Redirect(w, r, "/", 500)
		return
	}

	http.Redirect(w, r, "/app/home", http.StatusTemporaryRedirect)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func createChallenge(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	session, _ := store.Get(r, "session")
	useremail := session.Values["user"]
	if useremail == nil {
		//if logged in redirect to home
		renderTemplate(w, "landing.gohtml", nil)
		log.Infof(ctx, "No user in session")
		return
	}

	err := models.NewChallenge(r, useremail.(string))
	if err != "" {
		fmt.Fprintf(w, err)
	}

	http.Redirect(w, r, "/app/home", http.StatusSeeOther)
}

func editChallenge(w http.ResponseWriter, r *http.Request) {

}

func deleteChallenge(w http.ResponseWriter, r *http.Request) {
	models.DeleteChallenge(r)
	http.Redirect(w, r, "/app/home", http.StatusSeeOther)
}
