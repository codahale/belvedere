package belvedere

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/logging/v2"
	"gopkg.in/h2non/gock.v1"
)

func TestLogs(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://logging.googleapis.com/v2/entries:list?alt=json&prettyPrint=false").
		JSON(logging.ListLogEntriesRequest{
			Filter: `timestamp>="2019-06-25T13:18:33.000000043Z" ` +
				`jsonPayload.container.metadata.app="my-app" ` +
				`health`,
			OrderBy:       "timestamp desc",
			PageSize:      1000,
			ResourceNames: []string{"projects/my-project"},
		}).
		Reply(200).
		JSON(logging.ListLogEntriesResponse{
			Entries: []*logging.LogEntry{
				{
					Timestamp:   "2019-06-25T14:55:01.000000000Z",
					JsonPayload: googleapi.RawMessage(`{"message": "woo","instance":{"name":"example-v2-abcd"},"container":{"metadata":{"release":"v2"}}}`),
				},
			},
		})

	ts := time.Date(2019, 6, 25, 13, 18, 33, 43, time.UTC)
	actual, err := Logs(context.TODO(), "my-project", "my-app", "", "", ts, []string{"health"})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"2019-06-25T14:55:01.000000000Z (v2/example-v2-abcd) woo",
	}

	if !cmp.Equal(actual, expected) {
		t.Error(cmp.Diff(actual, expected))
	}
}
