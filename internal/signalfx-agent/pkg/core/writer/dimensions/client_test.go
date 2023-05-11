package dimensions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/stretchr/testify/require"
)

var putPathRegexp = regexp.MustCompile(`/v2/dimension/([^/]+)/([^/]+)`)
var patchPathRegexp = regexp.MustCompile(`/v2/dimension/([^/]+)/([^/]+)/_/sfxagent`)

type dim struct {
	Key          string            `json:"key"`
	Value        string            `json:"value"`
	Properties   map[string]string `json:"customProperties"`
	Tags         []string          `json:"tags"`
	TagsToRemove []string          `json:"tagsToRemove"`
	WasPatch     bool              `json:"-"`
}

func waitForDims(dimCh <-chan dim, count, waitSeconds int) []dim { // nolint: unparam
	var dims []dim
	timeout := time.After(time.Duration(waitSeconds) * time.Second)

loop:
	for {
		select {
		case dim := <-dimCh:
			dims = append(dims, dim)
			if len(dims) >= count {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	return dims
}

func makeHandler(dimCh chan<- dim, forcedResp *atomic.Value) http.HandlerFunc {
	forcedResp.Store(200)

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		forcedRespInt := forcedResp.Load().(int)
		if forcedRespInt != 200 {
			rw.WriteHeader(forcedRespInt)
			return
		}

		log.Printf("Test server got request: %s", r.URL.Path)
		var re *regexp.Regexp
		switch r.Method {
		case "PUT":
			re = putPathRegexp
		case "PATCH":
			re = patchPathRegexp
		default:
			rw.WriteHeader(404)
			return
		}
		match := re.FindStringSubmatch(r.URL.Path)
		if match == nil {
			rw.WriteHeader(404)
			return
		}

		var bodyDim dim
		if err := json.NewDecoder(r.Body).Decode(&bodyDim); err != nil {
			rw.WriteHeader(400)
			return
		}
		bodyDim.WasPatch = r.Method == "PATCH"
		bodyDim.Key = match[1]
		bodyDim.Value = match[2]

		dimCh <- bodyDim

		rw.WriteHeader(200)
	})
}

func setup() (*DimensionClient, chan dim, *atomic.Value, context.CancelFunc) {
	dimCh := make(chan dim)

	var forcedResp atomic.Value
	server := httptest.NewServer(makeHandler(dimCh, &forcedResp))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
		server.Close()
	}()

	client, err := NewDimensionClient(ctx, &config.WriterConfig{
		PropertiesMaxBuffered:      10,
		PropertiesMaxRequests:      10,
		PropertiesSendDelaySeconds: 1,
		PropertiesHistorySize:      1000,
		LogDimensionUpdates:        true,
		APIURL:                     server.URL,
	})
	if err != nil {
		panic("could not make dim client: " + err.Error())
	}
	client.Start()

	return client, dimCh, &forcedResp, cancel
}

func TestDimensionClient(t *testing.T) {
	client, dimCh, forcedResp, cancel := setup()
	defer cancel()

	require.NoError(t, client.AcceptDimension(&types.Dimension{
		Name:  "host",
		Value: "test-box",
		Properties: map[string]string{
			"a": "b",
			"c": "d",
		},
		Tags: map[string]bool{
			"active": true,
		},
		MergeIntoExisting: true,
	}))

	dims := waitForDims(dimCh, 1, 3)
	require.Equal(t, dims, []dim{
		{
			Key:   "host",
			Value: "test-box",
			Properties: map[string]string{
				"a": "b",
				"c": "d",
			},
			Tags:         []string{"active"},
			TagsToRemove: []string{},
			WasPatch:     true,
		},
	})

	t.Run("same dimension with different values", func(t *testing.T) {
		require.NoError(t, client.AcceptDimension(&types.Dimension{
			Name:  "host",
			Value: "test-box",
			Properties: map[string]string{
				"e": "f",
			},
			Tags: map[string]bool{
				"active": false,
			},
			MergeIntoExisting: true,
		}))

		dims = waitForDims(dimCh, 1, 3)
		require.Equal(t, dims, []dim{
			{
				Key:   "host",
				Value: "test-box",
				Properties: map[string]string{
					"e": "f",
				},
				Tags:         []string{},
				TagsToRemove: []string{"active"},
				WasPatch:     true,
			},
		})
	})

	t.Run("property with empty value", func(t *testing.T) {
		require.NoError(t, client.AcceptDimension(&types.Dimension{
			Name:  "host",
			Value: "empty-box",
			Properties: map[string]string{
				"a": "",
			},
			Tags: map[string]bool{
				"active": false,
			},
			MergeIntoExisting: true,
		}))

		dims = waitForDims(dimCh, 1, 3)
		require.Equal(t, dims, []dim{
			{
				Key:   "host",
				Value: "empty-box",
				Properties: map[string]string{
					"a": "",
				},
				Tags:         []string{},
				TagsToRemove: []string{"active"},
				WasPatch:     true,
			},
		})
	})

	require.NoError(t, client.AcceptDimension(&types.Dimension{
		Name:  "AWSUniqueID",
		Value: "abcd",
		Properties: map[string]string{
			"a": "b",
		},
		Tags: map[string]bool{
			"is_on": true,
		},
	}))

	dims = waitForDims(dimCh, 1, 3)
	require.Equal(t, dims, []dim{
		{
			Key:   "AWSUniqueID",
			Value: "abcd",
			Properties: map[string]string{
				"a": "b",
			},
			Tags:     []string{"is_on"},
			WasPatch: false,
		},
	})

	t.Run("send a distinct prop/tag set for existing dim with server error", func(t *testing.T) {
		forcedResp.Store(500)

		// send a distinct prop/tag set for same dim with an error
		require.NoError(t, client.AcceptDimension(&types.Dimension{
			Name:  "AWSUniqueID",
			Value: "abcd",
			Properties: map[string]string{
				"a": "b",
				"c": "d",
			},
			Tags: map[string]bool{
				"running": true,
			},
		}))
		dims = waitForDims(dimCh, 1, 3)
		require.Len(t, dims, 0)

		forcedResp.Store(200)
		dims = waitForDims(dimCh, 1, 3)

		// After the server recovers the dim should be resent.
		require.Equal(t, dims, []dim{
			{
				Key:   "AWSUniqueID",
				Value: "abcd",
				Properties: map[string]string{
					"a": "b",
					"c": "d",
				},
				Tags:     []string{"running"},
				WasPatch: false,
			},
		})
	})

	t.Run("does not retry 4xx responses", func(t *testing.T) {
		forcedResp.Store(400)

		// send a distinct prop/tag set for same dim with an error
		require.NoError(t, client.AcceptDimension(&types.Dimension{
			Name:  "AWSUniqueID",
			Value: "aslfkj",
			Properties: map[string]string{
				"z": "y",
			},
		}))
		dims = waitForDims(dimCh, 1, 3)
		require.Len(t, dims, 0)

		forcedResp.Store(200)
		dims = waitForDims(dimCh, 1, 3)
		require.Len(t, dims, 0)
	})

	t.Run("does retry 404 responses", func(t *testing.T) {
		forcedResp.Store(404)

		// send a distinct prop/tag set for same dim with an error
		require.NoError(t, client.AcceptDimension(&types.Dimension{
			Name:  "AWSUniqueID",
			Value: "id404",
			Properties: map[string]string{
				"z": "x",
			},
		}))
		dims = waitForDims(dimCh, 1, 3)
		require.Len(t, dims, 0)

		forcedResp.Store(200)
		dims = waitForDims(dimCh, 1, 3)
		require.Equal(t, dims, []dim{
			{
				Key:   "AWSUniqueID",
				Value: "id404",
				Properties: map[string]string{
					"z": "x",
				},
				WasPatch: false,
			},
		})
	})

	t.Run("send a duplicate", func(t *testing.T) {
		require.NoError(t, client.AcceptDimension(&types.Dimension{
			Name:  "AWSUniqueID",
			Value: "abcd",
			Properties: map[string]string{
				"a": "b",
				"c": "d",
			},
			Tags: map[string]bool{
				"running": true,
			},
		}))

		dims = waitForDims(dimCh, 1, 3)
		require.Len(t, dims, 0)
	})

	// send something unique again
	t.Run("send something unique to same dim", func(t *testing.T) {
		require.NoError(t, client.AcceptDimension(&types.Dimension{
			Name:  "AWSUniqueID",
			Value: "abcd",
			Properties: map[string]string{
				"c": "d",
			},
			Tags: map[string]bool{
				"running": true,
			},
		}))

		dims = waitForDims(dimCh, 1, 3)

		require.Equal(t, dims, []dim{
			{
				Key:   "AWSUniqueID",
				Value: "abcd",
				Properties: map[string]string{
					"c": "d",
				},
				Tags:     []string{"running"},
				WasPatch: false,
			},
		})
	})

	t.Run("send a distinct patch that covers the same prop keys", func(t *testing.T) {
		require.NoError(t, client.AcceptDimension(&types.Dimension{
			Name:  "host",
			Value: "test-box",
			Properties: map[string]string{
				"a": "z",
			},
			MergeIntoExisting: true,
		}))

		dims = waitForDims(dimCh, 1, 3)
		require.Equal(t, dims, []dim{
			{
				Key:   "host",
				Value: "test-box",
				Properties: map[string]string{
					"a": "z",
				},
				Tags:         []string{},
				TagsToRemove: []string{},
				WasPatch:     true,
			},
		})
	})

	t.Run("send a distinct patch that covers the same tags", func(t *testing.T) {
		require.NoError(t, client.AcceptDimension(&types.Dimension{
			Name:  "host",
			Value: "test-box",
			Tags: map[string]bool{
				"active": true,
			},
			MergeIntoExisting: true,
		}))

		dims = waitForDims(dimCh, 1, 3)
		require.Equal(t, dims, []dim{
			{
				Key:          "host",
				Value:        "test-box",
				Properties:   map[string]string{},
				Tags:         []string{"active"},
				TagsToRemove: []string{},
				WasPatch:     true,
			},
		})
	})
}

func TestFlappyUpdates(t *testing.T) {
	client, dimCh, _, cancel := setup()
	defer cancel()

	// Do some flappy updates
	for i := 0; i < 5; i++ {
		require.NoError(t, client.AcceptDimension(&types.Dimension{
			Name:  "pod_uid",
			Value: "abcd",
			Properties: map[string]string{
				"index": fmt.Sprintf("%d", i),
			},
		}))
		require.NoError(t, client.AcceptDimension(&types.Dimension{
			Name:  "pod_uid",
			Value: "efgh",
			Properties: map[string]string{
				"index": fmt.Sprintf("%d", i),
			},
			MergeIntoExisting: true,
		}))
	}

	dims := waitForDims(dimCh, 2, 3)
	require.ElementsMatch(t, []dim{
		{
			Key:        "pod_uid",
			Value:      "abcd",
			Properties: map[string]string{"index": "4"},
			WasPatch:   false,
		},
		{
			Key:          "pod_uid",
			Value:        "efgh",
			Properties:   map[string]string{"index": "4"},
			Tags:         []string{},
			TagsToRemove: []string{},
			WasPatch:     true,
		},
	}, dims)

	// Give it enough time to run the counter updates which happen after the
	// request is completed.
	time.Sleep(1 * time.Second)

	require.Equal(t, int64(8), atomic.LoadInt64(&client.TotalFlappyUpdates))
	require.Equal(t, int64(0), atomic.LoadInt64(&client.DimensionsCurrentlyDelayed))
	require.Equal(t, int64(2), atomic.LoadInt64(&client.requestSender.TotalRequestsStarted))
	require.Equal(t, int64(2), atomic.LoadInt64(&client.requestSender.TotalRequestsCompleted))
	require.Equal(t, int64(0), atomic.LoadInt64(&client.requestSender.TotalRequestsFailed))
}

func TestInvalidUpdatesNotSent(t *testing.T) {
	client, dimCh, _, cancel := setup()
	defer cancel()

	require.NoError(t, client.AcceptDimension(&types.Dimension{
		Name:  "host",
		Value: "",
		Properties: map[string]string{
			"a": "b",
			"c": "d",
		},
		Tags: map[string]bool{
			"active": true,
		},
		MergeIntoExisting: true,
	}))

	require.NoError(t, client.AcceptDimension(&types.Dimension{
		Name:  "",
		Value: "asdf",
		Properties: map[string]string{
			"a": "b",
			"c": "d",
		},
		Tags: map[string]bool{
			"active": true,
		},
	}))

	dims := waitForDims(dimCh, 2, 3)
	require.Len(t, dims, 0)
}
