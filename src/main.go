package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var stateString = "random-string"

var currentUser = ""

var sonicNowPlayingURL = "https://player.rogersradio.ca/chdi/widget/now_playing"

var getUserIdURL = "https://api.spotify.com/v1/me"
var getPlaylistsURL = "https://api.spotify.com/v1/me/playlists?limit=50"
var makePlaylistURL = "https://api.spotify.com/v1/users/{user_id}/playlists"

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
	Id   string `json:"id"`
}

type UserId struct {
	Id string `json:"id"`
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
	token, err := getAuthToken(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
		return
	}

	currentUser, err = getUserId(token)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(currentUser)

	fixAllURLs()

	nowPlaying := getNowPlaying()
	fmt.Println(nowPlaying)

	id, err := handlePlaylist(token)
	fmt.Println(id)

}

func getAuthToken(state string, code string) (*oauth2.Token, error) {
	if state != stateString {
		return nil, fmt.Errorf("invalid state")
	}
	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	return token, nil
}

func fixAllURLs() {
	makePlaylistURL = strings.Replace(makePlaylistURL, "{user_id}", currentUser, 1)
}

func getUserId(token *oauth2.Token) (string, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", getUserIdURL, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	authorization := "Bearer " + token.AccessToken
	req.Header.Set("Authorization", authorization)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	data := UserId{}
	json.Unmarshal(body, &data)
	return data.Id, nil
}

func getNowPlaying() SonicInfo {
	res, err := http.Get(sonicNowPlayingURL)
	if err != nil {
		fmt.Println(err.Error())
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	data := SonicInfo{}
	json.Unmarshal(body, &data)

	return data
}

// will either find or create Sonic Playlist and return ID
func handlePlaylist(token *oauth2.Token) (string, error) {
	var playlistId, err = checkForPlaylist(token)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	} else if playlistId == "" {
		fmt.Println("Making Playlist")
		playlistId, err = makePlaylist(token)
	}
	return playlistId, nil

}

// will get playlist ID for Sonic On Demand if it exists
func checkForPlaylist(token *oauth2.Token) (string, error) {
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
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading body")
		return "", err
	}

	data := AllPlaylists{}
	json.Unmarshal(body, &data)
	for _, value := range data.Items {
		if value.Name == "SONiC On Demand" {
			return value.Id, nil
		}
	}
	return "", nil
}

func makePlaylist(token *oauth2.Token) (string, error) {
	client := http.Client{}

	requestBody, err := json.Marshal(map[string]string{
		"name":        "SONiC On Demand",
		"description": "Playlsit made from SONiC 102.9",
	})
	req, err := http.NewRequest("POST", makePlaylistURL, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println(err.Error())
	}

	authorization := "Bearer " + token.AccessToken
	req.Header.Set("Authorization", authorization)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading body")
		return "", err
	}

	data := PlaylistInfo{}
	json.Unmarshal(body, &data)

	return data.Id, nil
}
