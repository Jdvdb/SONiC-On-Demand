package main

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var stateString = "random-string"

var (
	config = oauth2.Config{
		ClientID:     os.Getenv("CLIENTID"),
		ClientSecret: os.Getenv("CLIENTSECRET"),
		Scopes:       []string{"playlist-modify-public", "playlist-modify-private", "playlist-read-private", "playlist-read-collaborative"},
		RedirectURL:  "http://localhost:3000/callback",
		Endpoint:     spotify.Endpoint,
	}
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.ListenAndServe(":3000", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	var htmlIndex = `<html>
	<body>
		<a href="/login">Click here to login!</a>
	</body>
	</html>`
	fmt.Fprintf(w, htmlIndex)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := config.AuthCodeURL(stateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	token, err := printAuthToken(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
		return
	}
	fmt.Println(token)
}

func printAuthToken(state string, code string) (*oauth2.Token, error) {
	if state != stateString {
		return nil, fmt.Errorf("invalid state")
	}
	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	return token, nil
}
