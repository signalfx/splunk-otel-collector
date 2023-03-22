package metrics

import (
	"fmt"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
)

func makeReplicaDPs(resource string, dimensions map[string]string, desired, available int32) []*datapoint.Datapoint {
	return []*datapoint.Datapoint{
		datapoint.New(
			fmt.Sprintf("kubernetes.%s.desired", resource),
			dimensions,
			datapoint.NewIntValue(int64(desired)),
			datapoint.Gauge,
			time.Time{}),
		datapoint.New(
			fmt.Sprintf("kubernetes.%s.available", resource),
			dimensions,
			datapoint.NewIntValue(int64(available)),
			datapoint.Gauge,
			time.Time{}),
	}
}
