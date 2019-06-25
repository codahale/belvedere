package it

import (
	"context"
	"testing"

	"golang.org/x/oauth2/google"
)

func TestMockTokenSource(t *testing.T) {
	MockTokenSource()

	source, err := google.DefaultTokenSource(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	_, err = source.Token()
	if err != nil {
		t.Fatal(err)
	}
}
