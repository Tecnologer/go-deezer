package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tecnologer/deezer/src/deezer"
	"github.com/tecnologer/go-secrets"
)

var port = flag.Int("port", 8088, "port")
var port2 = flag.Int("port2", 8089, "port")
var verbouse = flag.Bool("v", false, "enable verbose")

func main() {
	flag.Parse()

	if *verbouse {
		logrus.SetLevel(logrus.DebugLevel)
	}

	deezerSecrets, err := secrets.GetGroup("deezer")
	if err != nil {
		panic(err)
	}

	appID := deezerSecrets.GetString("app_id")
	secretKey := deezerSecrets.GetString("secret_key")
	redirectURL := deezerSecrets.GetString("redirect_uri")

	d := deezer.New(appID, secretKey, redirectURL)

	http.HandleFunc("/search-artist", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		params := req.URL.Query()

		if len(params) == 0 {
			log.Println("no params in the url")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid request"))
			return
		}
		query := params.Get("q")

		data, err := d.SearchArtist(query)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("fail searching: %v", err)))
			return
		}

		datares, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("fail parsing result: %v", err)))
			return
		}

		w.Write(datares)
	})

	d.Start(*port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port2), nil))
}
