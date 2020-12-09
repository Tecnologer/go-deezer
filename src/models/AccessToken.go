package models

//AcessToken for response of access token auth
type AcessToken struct {
	Token   string `json:"access_token"`
	Expires int    `json:"expires"`
}
