package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	deezer "github.com/deezer/src"
	"github.com/tecnologer/go-secrets"
)

var port = flag.Int("port", 8088, "port")

func main() {
	flag.Parse()

	deezerSecrets, err := secrets.GetGroup("deezer")
	if err != nil {
		panic(err)
	}
	appID := deezerSecrets.GetString("appId")
	secretKey := deezerSecrets.GetString("secret_key")
	redirectURL := deezerSecrets.GetString("redirect_uri")

	deezer := deezer.NewDeezer(appID, secretKey, redirectURL)

	if !deezer.IsAuth() {
		deezer.OpenOAuth()
	}

	host := fmt.Sprintf(":%d", *port)
	log.Println(host)
	log.Println(http.ListenAndServe(host, http.HandlerFunc(webhookReceiver)))
}

func webhookReceiver(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "application/json")

	params := req.URL.Query()

	if len(params) == 0 {
		log.Println("no params in the url")
		return
	}
	code := params.Get("code")
	fmt.Println(code)
}
