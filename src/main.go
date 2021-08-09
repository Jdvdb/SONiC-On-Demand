package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var stateString = "random-string"
var sonicNowPlaying = "https://player.rogersradio.ca/chdi/widget/now_playing"

var getPlaylistsURL = "https://api.spotify.com/v1/me/playlists?limit=50"

var (
	config = oauth2.Config{
		ClientID:     os.Getenv("CLIENTID"),
		ClientSecret: os.Getenv("CLIENTSECRET"),
		Scopes:       []string{"playlist-modify-public", "playlist-modify-private", "playlist-read-private", "playlist-read-collaborative"},
		RedirectURL:  "http://localhost:3000/callback",
		Endpoint:     spotify.Endpoint,
	}
)

type SonicInfo struct {
	Song_title string `json:"song_title"`
	Started_at string `json:"started_at"`
	Length     string `json:"length"`
	Spotify    string `json:"spotify"`
}

type AllPlaylists struct {
	Items  []PlaylistInfo `json:"items"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type PlaylistInfo struct {
	Name string `json:"name"`
}

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
	nowPlaying := getNowPlaying()
	fmt.Println(nowPlaying)
	makePlaylist(token)

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

func getNowPlaying() SonicInfo {
	res, err := http.Get(sonicNowPlaying)
	if err != nil {
		fmt.Println("Error getting now playing", err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading body")
	}

	data := SonicInfo{}
	json.Unmarshal(body, &data)

	return data
}

func makePlaylist(token *oauth2.Token) {
	client := http.Client{}
	req, err := http.NewRequest("GET", getPlaylistsURL, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	authorization := "Bearer " + token.AccessToken
	req.Header.Set("Authorization", authorization)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading body")
	}

	data := AllPlaylists{}
	json.Unmarshal(body, &data)
	fmt.Println(data)
}
