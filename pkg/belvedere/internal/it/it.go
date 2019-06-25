package it

import (
	"time"

	"golang.org/x/oauth2"
	"gopkg.in/h2non/gock.v1"
)

func MockTokenSource() {
	gock.New("https://oauth2.googleapis/token").
		Reply(200).
		JSON(oauth2.Token{
			TokenType:    "mock",
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			Expiry:       time.Now().Add(10 * time.Hour),
		})
}
