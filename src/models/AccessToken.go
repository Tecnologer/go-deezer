package models

import "time"

//AcessToken for response of access token auth
type AcessToken struct {
	Token     string `json:"access_token"`
	Expires   int    `json:"expires"`
	CreatedAt time.Time
}

//IsTokenExpired returns true if the Expires time has completed
func (t *AcessToken) IsTokenExpired() bool {
	duration := time.Since(t.CreatedAt)
	return int(duration*time.Second) > t.Expires-60
}
