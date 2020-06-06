package setup

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"gopkg.in/h2non/gock.v1"
)

func TestService_ManagedZone(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	s, err := NewService(context.Background(), option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	gock.New("https://dns.googleapis.com/dns/v1/projects/my-project/managedZones/belvedere?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(&dns.ManagedZone{
			DnsName: "my-dns",
		})

	actual, err := s.ManagedZone(context.Background(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	expected := &dns.ManagedZone{
		DnsName: "my-dns",
		ServerResponse: googleapi.ServerResponse{
			HTTPStatusCode: http.StatusOK,
			Header:         http.Header{"Content-Type": {"application/json"}},
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
