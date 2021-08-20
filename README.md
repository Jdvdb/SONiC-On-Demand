# SONiC-On-Demand
Lightweight golang application put into multi stage docker build to constantly update a playlist with music from the radio station SONiC 102.9.

## Requirements To Run:
- Ensure you have docker installed on your machine.
- Create a spotify developer account along with a project for a client ID and secret.
- Machine must be connected to the internet.
- A spotify account from which the playlist will be created.

## Quick-Start Guide
1. Get your client ID and secret and put them into a '.env' file like the example (env-example.txt) in the project in your current directory.
2. Use `docker pull jordanvdb/sonic-on-demand` to grab the image from Docker Hub. (note you must have compatible version).
3. Run the image as follows: `docker run -p 3000:3000 --env-file .env jordanvdb/sonic-on-demand`.
4. Open a web browser on your device and navigate to `localhost:3000` and you should be redirected to a spotify login page.
5. Once logged in, leave the device alone and it will continue to update a playlist called 'SONiC On Demand' on your account.

### Supported Archs:
- amd64
- arm64
- arm/v6
- arm/v7

### Sources:
Basic auth flow: Alex Pliutau, Getting Started with OAuth2 in Go
https://itnext.io/getting-started-with-oauth2-in-go-1c692420e03

Help with requests: Abu Ashraf Masnun, Making HTTP Requests In Golang
https://medium.com/@masnun/making-http-requests-in-golang-dd123379efe7

Using a ticker: Elliot Forbes, Go Tickers Tutorial
https://tutorialedge.net/golang/go-ticker-tutorial/
