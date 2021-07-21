package main

import (
	"fmt"
	"net/http"
	"html/template"
	"os"
)

var myTemplate *template.Template

// data for auth URL
type ViewData struct {
	ClientID string
}

func main() {
	// get environment vars
	var err error
	var clientID = os.Getenv("CLIENTID")
	var clientSecret = os.Getenv("CLIENTSECRET")

	myTemplate, err = template.ParseFiles("views/index.html")
	if err != nil {
		panic(err)
	}

	fmt.Println("clientID: " + clientID)
	fmt.Println("clientSecret: " + clientSecret)

	http.HandleFunc("/", handler)
	http.ListenAndServe(":3000", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vd := ViewData{ClientID: os.Getenv("CLIENTID")}
	err := myTemplate.Execute(w, vd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}