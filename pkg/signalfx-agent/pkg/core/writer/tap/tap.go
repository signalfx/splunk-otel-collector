package tap

import (
	"context"
	"io"
	"net/http"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/sirupsen/logrus"
)

// DatapointTap accepts datapoints and asynchronouly writes a string
// representation of them to the output, filtering as requested.
type DatapointTap struct {
	filter dpfilters.DatapointFilter
	out    io.Writer
	buffer chan []*datapoint.Datapoint
}

// New makes a new tap
func New(filter dpfilters.DatapointFilter, out io.Writer) *DatapointTap {
	return &DatapointTap{
		filter: filter,
		out:    out,
		buffer: make(chan []*datapoint.Datapoint, 100),
	}
}

// Run the tap and write out datapoints
func (t *DatapointTap) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case dps := <-t.buffer:
			for _, dp := range dps {
				if t.filter != nil && !t.filter.Matches(dp) {
					continue
				}
				_, _ = t.out.Write([]byte(utils.DatapointToString(dp)))
				if f, ok := t.out.(http.Flusher); ok {
					f.Flush()
				}
			}
		}
	}
}

// Accept should be called by the writer with every datapoint
func (t *DatapointTap) Accept(dps []*datapoint.Datapoint) {
	if t == nil {
		return
	}

	select {
	case t.buffer <- dps:
		break
	default:
		logrus.Error("Could not process datapoint in tap due to full buffer")
	}
}
