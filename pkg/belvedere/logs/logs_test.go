package logs

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/logging/v2"
	"gopkg.in/h2non/gock.v1"
)

func TestList(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://logging.googleapis.com/v2/entries:list?alt=json&prettyPrint=false").
		JSON(logging.ListLogEntriesRequest{
			Filter: `resource.type="gce_instance" ` +
				`timestamp>="2019-06-25T13:18:33.000000043Z" ` +
				`jsonPayload.container.metadata.app="my-app" ` +
				`health`,
			OrderBy:       "timestamp desc",
			PageSize:      1000,
			ResourceNames: []string{"projects/my-project"},
		}).
		Reply(http.StatusOK).
		JSON(logging.ListLogEntriesResponse{
			Entries: []*logging.LogEntry{
				{
					Timestamp:   "2019-06-25T14:55:01.000000000Z",
					JsonPayload: googleapi.RawMessage(`{"message": "woo","instance":{"name":"example-v2-abcd"},"container":{"name":"/nginx","metadata":{"release":"v2"}}}`),
				},
			},
		})

	s, err := NewService(context.TODO(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	// mock out the clock
	(s.(*service)).clock = func() time.Time {
		return time.Date(2019, 6, 25, 13, 18+15, 33, 43, time.UTC)
	}

	actual, err := s.List(context.TODO(), "my-app", "", "", 15*time.Minute, []string{"health"})
	if err != nil {
		t.Fatal(err)
	}

	expected := []Entry{
		{
			Timestamp: time.Date(2019, 6, 25, 14, 55, 1, 0, time.UTC),
			Release:   "v2",
			Instance:  "example-v2-abcd",
			Container: "nginx",
			Message:   "woo",
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Error(cmp.Diff(expected, actual))
	}
}
