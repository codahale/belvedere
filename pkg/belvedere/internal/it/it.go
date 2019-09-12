package it

import (
	"os"
	"time"

	"golang.org/x/oauth2"
	"gopkg.in/h2non/gock.v1"
)

// MockTokenSource mocks the GCP SDK's default token source endpoint with a stub token.
func MockTokenSource() {
	_ = os.Setenv("GCE_METADATA_HOST", "metadata.server.fake")
	gock.New("https://oauth2.googleapis/token").
		Reply(200).
		JSON(oauth2.Token{
			TokenType:    "mock",
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			Expiry:       time.Now().Add(10 * time.Hour),
		})
}
