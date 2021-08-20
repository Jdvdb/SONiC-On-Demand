package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var stateString = "random-string"

var currentUser = ""

var currentSongs []string

var sonicNowPlayingURL = "https://player.rogersradio.ca/chdi/widget/now_playing"

var getUserIdURL = "https://api.spotify.com/v1/me"
var getPlaylistsURL = "https://api.spotify.com/v1/me/playlists?limit=50"
var makePlaylistURL = "https://api.spotify.com/v1/users/{user_id}/playlists"
var addSongURL = "https://api.spotify.com/v1/playlists/{playlist_id}/tracks"
var getSongsUrl = "https://api.spotify.com/v1/playlists/{playlist_id}/tracks?market=CA&fields=items(track.name,track.id),total&limit=100"

var authFinished = false

var (
	config = oauth2.Config{
		ClientID:     os.Getenv("CLIENTID"),
		ClientSecret: os.Getenv("CLIENTSECRET"),
		Scopes:       []string{"playlist-modify-public", "playlist-modify-private", "playlist-read-private", "playlist-read-collaborative"},
		RedirectURL:  "http://localhost:3000/callback",
		Endpoint:     spotify.Endpoint,
	}
)

var ctx = context.Background()

type SonicInfo struct {
	Song_title string `json:"song_title"`
	Started_at string `json:"started_at"`
	Length     string `json:"length"`
	Spotify    string `json:"spotify"`
}

type PlaylistList struct {
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

type SONiCPlaylist struct {
	Items []Track `json:"items"`
	Total int     `json:"total"`
}

type Track struct {
	Track PlaylistInfo `json:"track"`
}

func main() {
	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/run", runHandler)
	http.ListenAndServe(":3000", nil)
	for !authFinished {

	}
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	var htmlIndex = `<html>
	<body>
		<p>Running...</p>
	</body>
	</html>`
	fmt.Fprintf(w, htmlIndex)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := config.AuthCodeURL(stateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	// get the auth token
	token, err := getAuthToken(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
		return
	}

	client := config.Client(ctx, token)

	// get the user's id
	currentUser, err = getUserId(client)
	if err != nil {
		fmt.Println(err.Error())
	}

	// fix the URLs requiring a user ID
	fixUserURLs()

	// get exisitng playlist or create new one if needed with the ID
	playlistId, err := handlePlaylist(client)

	// fix all URLs needing the playlist ID
	fixPlaylistURLs(playlistId)

	// get a list of all songs in the playlist
	getAllSongs(client)

	http.Redirect(w, r, "/run", http.StatusTemporaryRedirect)

	// loop for songs
	go MainTask(client)
}

func getAuthToken(state string, code string) (*oauth2.Token, error) {
	if state != stateString {
		return nil, fmt.Errorf("invalid state")
	}
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	return token, nil
}

func fixUserURLs() {
	makePlaylistURL = strings.Replace(makePlaylistURL, "{user_id}", currentUser, 1)
}

func fixPlaylistURLs(playlistId string) {
	addSongURL = strings.Replace(addSongURL, "{playlist_id}", playlistId, 1)
	getSongsUrl = strings.Replace(getSongsUrl, "{playlist_id}", playlistId, 1)
}

func getUserId(client *http.Client) (string, error) {
	res, err := client.Get(getUserIdURL)
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
func handlePlaylist(client *http.Client) (string, error) {
	var playlistId, err = checkForPlaylist(client)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	} else if playlistId == "" {
		fmt.Println("Making Playlist")
		playlistId, err = makePlaylist(client)
	}
	return playlistId, nil

}

// will get playlist ID for Sonic On Demand if it exists
func checkForPlaylist(client *http.Client) (string, error) {
	res, err := client.Get(getPlaylistsURL)
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

	data := PlaylistList{}
	json.Unmarshal(body, &data)
	for _, value := range data.Items {
		if value.Name == "SONiC On Demand" {
			return value.Id, nil
		}
	}
	return "", nil
}

func makePlaylist(client *http.Client) (string, error) {
	requestBody, err := json.Marshal(map[string]string{
		"name":        "SONiC On Demand",
		"description": "Playlsit made from SONiC 102.9",
	})

	req, err := http.NewRequest("POST", makePlaylistURL, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println(err.Error())
	}

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

	data := PlaylistInfo{}
	json.Unmarshal(body, &data)

	return data.Id, nil
}

func getAllSongs(client *http.Client) error {
	totalSongs := 100

	for currentOffest := 0; currentOffest < totalSongs; currentOffest += 100 {

		currentURL := getSongsUrl + "&offset=" + strconv.Itoa(currentOffest)

		res, err := client.Get(currentURL)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		data := SONiCPlaylist{}
		json.Unmarshal(body, &data)
		for _, value := range data.Items {
			currentSongs = append(currentSongs, value.Track.Id)
		}
		totalSongs = data.Total
	}

	return nil
}

func checkForSong(songId string) bool {
	for i := 0; i < len(currentSongs); i++ {
		if currentSongs[i] == songId {
			return true
		}
	}
	return false
}

func addSong(client *http.Client, songId string) error {
	songURI := "spotify:track:" + songId

	requestBody, err := json.Marshal(map[string][]string{
		"uris": []string{songURI},
	})

	req, err := http.NewRequest("POST", addSongURL, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println(err.Error())
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	defer res.Body.Close()

	// also ensure song is added to app
	currentSongs = append(currentSongs, songId)

	return nil
}

func MainTask(client *http.Client) {
	ticker := time.NewTicker(15 * time.Second)

	for _ = range ticker.C {

		nowPlaying := getNowPlaying()
		fmt.Println(nowPlaying)

		if nowPlaying.Spotify == "" {
			fmt.Println("Song not on spotify")
		} else if checkForSong(nowPlaying.Spotify) {
			fmt.Println("Song already in playlist")
		} else {
			addSong(client, nowPlaying.Spotify)
		}
	}
}
