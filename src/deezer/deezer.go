package deezer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tecnologer/deezer/src/deezer/settings"
	"github.com/tecnologer/deezer/src/models"
	"github.com/tecnologer/deezer/src/tools/browser"
)

const (
	deezerAPIURL    = "https://api.deezer.com"
	deezerURL       = "https://connect.deezer.com/"
	oauthURL        = "oauth/auth.php"
	accessTokenURL  = "oauth/access_token.php"
	searchArtistURL = "search/artist"
)

//Deezer struct
type Deezer struct {
	RedirectURL string
	appID       string
	secretKey   string
	auth        *models.AcessToken
}

//New creates a new instance of deezer
func New(appID, secretKey, redirect string) *Deezer {
	token, _ := settings.GetToken()
	if token == nil {
		token = &models.AcessToken{}
	}
	return &Deezer{
		appID:       appID,
		secretKey:   secretKey,
		RedirectURL: redirect,
		auth:        token,
	}
}

//OpenOAuth opens the url to authenticate in deezer
func (d *Deezer) OpenOAuth() bool {
	perms := "basic_access,email"
	endpoint := d.getEndpointURL(oauthURL)
	url := fmt.Sprintf("%s&app_id=%s&redirect_uri=%s/auth&perms=%s", endpoint, d.appID, d.RedirectURL, perms)
	return browser.Open(url)
}

func (d *Deezer) getEndpointURL(action string) string {
	return fmt.Sprintf("%s/%s?output=json", deezerURL, action)
}
func (d *Deezer) getAPIEndpointURL(action string) string {
	return fmt.Sprintf("%s/%s?output=json", deezerAPIURL, action)
}

//IsAuth returns if exists a token
func (d *Deezer) IsAuth() bool {
	return d.auth != nil && d.auth.Token != "" && !d.auth.IsTokenExpired()
}

func (d *Deezer) Start(port int) {
	updates, err := d.setWebhook(port)
	if err != nil {
		logrus.WithError(err).Error("register for updates")
		return
	}

	if !d.IsAuth() {
		d.OpenOAuth()
	}

	go func() {
		for update := range updates {
			switch update.(type) {
			case *models.AuthCode:
				d.getToken(update.(*models.AuthCode))
			}
		}
	}()
}

func (d *Deezer) getToken(code *models.AuthCode) {
	endpoint := d.getEndpointURL(accessTokenURL)
	v := url.Values{}
	v.Add("app_id", d.appID)
	v.Add("secret", d.secretKey)
	v.Add("code", code.Code)

	res, err := http.PostForm(endpoint+"?output=json", v)

	if err != nil {
		logrus.WithError(err).Error("getting token")
		return
	}

	resBody, err := ioutil.ReadAll(res.Body)
	logrus.Debug(string(resBody))

	err = json.Unmarshal(resBody, d.auth)
	if err != nil {
		logrus.WithError(err).WithField("body", string(resBody)).Error("parsing token")
	}
	err = settings.SetToken(d.auth)
	if err != nil {
		logrus.WithError(err).Warn("get token: saving token in settings")
	}
}

func (d *Deezer) SearchArtist(name string) ([]*models.Artist, error) {
	endpoint := d.getAPIEndpointURL(searchArtistURL)
	v := url.Values{}
	v.Add("q", name)
	v.Add("access_token", d.auth.Token)

	endpoint += "&" + v.Encode()
	logrus.Debug(endpoint)
	res, err := http.Get(endpoint)

	if err != nil {
		return nil, errors.Wrap(err, "getting request from server")
	}

	obRes := &models.SearchArtistResult{}

	//err = json.NewDecoder(res.Body).Decode(obRes)

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading body artist response")
	}
	logrus.Debug(string(resBody))

	err = json.Unmarshal(resBody, obRes)
	if err != nil {
		return nil, errors.Wrap(err, "parsing search artist response")
	}
	return obRes.Data, nil
}
