package it

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"gopkg.in/h2non/gock.v1"
)

// MockTokenSource mocks the GCP SDK's default token source endpoint with a stub token.
func MockTokenSource() {
	f, err := ioutil.TempFile(os.TempDir(), "belvedere-test-creds-*")
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	_, err = f.Write([]byte(`{
  "client_id": "fake.apps.googleusercontent.com",
  "client_secret": "fake",
  "refresh_token": "fake",
  "type": "authorized_user"
}`))
	if err != nil {
		panic(err)
	}

	_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", f.Name())
	gock.New("https://oauth2.googleapis/token").
		Reply(http.StatusOK).
		JSON(oauth2.Token{
			TokenType:    "mock",
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			Expiry:       time.Now().Add(10 * time.Hour),
		})
}
