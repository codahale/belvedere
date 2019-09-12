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
	token := oauth2.Token{
		TokenType:    "mock",
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		Expiry:       time.Now().Add(10 * time.Hour),
	}
	// Mock out token source for default app credentials.
	gock.New("https://oauth2.googleapis/token").
		Reply(200).
		JSON(token)

	// Mock out token source for fake GCE tests.
	gock.New("http://metadata.server.fake/computeMetadata/v1/instance/service-accounts/default/token").
		Reply(200).
		JSON(token)
}
