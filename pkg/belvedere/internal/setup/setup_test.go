package setup

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/codahale/belvedere/internal/httpmock"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

func TestService_ManagedZone(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/dns/v1/projects/my-project/managedZones/belvedere?alt=json&prettyPrint=false`,
		httpmock.RespJSON(&dns.ManagedZone{
			DnsName: "my-dns",
		}))

	s, err := NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.ManagedZone(context.Background(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	// Empty out server response since we don't care.
	got.ServerResponse = googleapi.ServerResponse{}

	want := &dns.ManagedZone{
		DnsName: "my-dns",
	}

	assert.Equal(t, "ManagedZone()", want, got)
}
