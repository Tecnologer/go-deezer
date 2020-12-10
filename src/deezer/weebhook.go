package deezer

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tecnologer/deezer/src/models"
)

var chUpdates chan interface{}

func (d *Deezer) setWebhook(port int) (chan interface{}, error) {
	host := fmt.Sprintf(":%d", port)
	logrus.Debug(host)
	chUpdates = make(chan interface{})
	go func() {
		http.HandleFunc("/auth", d.authWebhook)
		http.ListenAndServe(host, nil)
		//http.HandlerFunc(d.webhookReceiver)
	}()

	return chUpdates, nil
}

func (d *Deezer) authWebhook(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "application/json")

	params := req.URL.Query()

	if len(params) == 0 {
		log.Println("no params in the url")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("invalid request"))
		return
	}
	code := params.Get("code")
	logrus.Debug(code)
	chUpdates <- &models.AuthCode{Code: code}
	res.Write([]byte("redirecting"))
}
