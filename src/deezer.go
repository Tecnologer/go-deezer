package deezer

import (
	"cmd/internal/browser"
	"fmt"

	"github.com/deezer/src/models"
)

const (
	deezerURL = "https://connect.deezer.com/"
	oauth     = "oauth/auth.php"
)

type Deezer struct {
	RedirectURL string
	appID       string
	secretKey   string
	auth        *models.AcessToken
}

//NewDeezer creates a new instance of deezer
func NewDeezer(appID, secretKey, redirect string) *Deezer {
	return &Deezer{
		appID:     appID,
		secretKey: secretKey,
	}
}

//OpenOAuth opens the url to authenticate in deezer
func (d *Deezer) OpenOAuth() bool {
	perms := "basic_access,email"
	url := fmt.Sprintf("%s/%s?app_id=%s&redirect_uri=%s&perms=%s", deezerURL, oauth, d.appID, d.RedirectURL, perms)
	return browser.Open(url)
}

//IsAuth returns if exists a token
func (d *Deezer) IsAuth() bool {
	return d.auth != nil
}
