package belvedere

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/codahale/belvedere/internal/assert"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/logging/v2"
	"google.golang.org/api/option"
	"gopkg.in/h2non/gock.v1"
)

func TestLogService_List(t *testing.T) {
	defer gock.Off()

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
					Timestamp: "2019-06-25T14:55:01.000000000Z",
					JsonPayload: googleapi.RawMessage(`{"message": "woo","instance":` +
						`{"name":"example-v2-abcd"},"container":{"name":"/nginx","metadata":{"release":"v2"}}}`),
				},
			},
		})

	logs, err := logging.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := &logService{
		project: "my-project",
		logs:    logs,
		clock: func() time.Time {
			return time.Date(2019, 6, 25, 13, 18+15, 33, 43, time.UTC)
		},
	}

	got, err := s.List(context.Background(), "my-app", "", "", 15*time.Minute, []string{"health"})
	if err != nil {
		t.Fatal(err)
	}

	want := []LogEntry{
		{
			Timestamp: time.Date(2019, 6, 25, 14, 55, 1, 0, time.UTC),
			Release:   "v2",
			Instance:  "example-v2-abcd",
			Container: "nginx",
			Message:   "woo",
		},
	}

	assert.Equal(t, "Logs()", want, got)
}
