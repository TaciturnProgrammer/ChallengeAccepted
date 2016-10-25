package middleware

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET_KEY")))

// AuthMiddleware is run before all the handler to make sure user is logged in
func AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		if session.IsNew || session.Values["user"] == nil {
			//not logged in, show login page
			http.Redirect(w, r, "/", 302)
		}
		h.ServeHTTP(w, r)
	})

}
