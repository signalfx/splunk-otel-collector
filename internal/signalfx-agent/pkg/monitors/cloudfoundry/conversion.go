package cloudfoundry

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/davecgh/go-spew/spew"
	"github.com/signalfx/golib/v3/datapoint"
)

var hexIDRegexp = regexp.MustCompile(`^[a-fA-F0-9]+-[a-fA-F0-9-]+$`)

func envelopeToDatapoints(env *loggregator_v2.Envelope) ([]*datapoint.Datapoint, error) {
	// We intentionally modify the Tags map on the envelope, assuming that the
	// loggregator code that generated it is not going to reuse envelope
	// instances or tag maps.
	dims := env.GetTags()

	prefix := ""

	if env.GetSourceId() != "" {
		dims["source_id"] = env.GetSourceId()
		if hexIDRegexp.MatchString(env.GetSourceId()) {
			prefix = env.GetTags()["origin"] + "."
		} else {
			prefix = env.GetSourceId() + "."
		}
	}

	if env.GetInstanceId() != "" {
		dims["instance_id"] = env.GetInstanceId()
	}

	var metricType datapoint.MetricType

	namesToValues := make(map[string]float64)

	switch m := env.GetMessage().(type) {
	case *loggregator_v2.Envelope_Log:
	case *loggregator_v2.Envelope_Counter:
		metricType = datapoint.Counter
		namesToValues[m.Counter.GetName()] = float64(m.Counter.GetTotal())
	case *loggregator_v2.Envelope_Gauge:
		metricType = datapoint.Gauge
		for name, gauge := range m.Gauge.GetMetrics() {
			namesToValues[name] = gauge.GetValue()
		}
	case *loggregator_v2.Envelope_Timer:
	case *loggregator_v2.Envelope_Event:
	default:
		return nil, fmt.Errorf("cannot convert envelope %v to SignalFx datapoints", spew.Sdump(env))
	}

	var dps []*datapoint.Datapoint
	for name, val := range namesToValues {
		dps = append(dps, datapoint.New(prefix+cleanupName(name), dims, datapoint.NewFloatValue(val), metricType, time.Unix(0, env.GetTimestamp())))
	}

	return dps, nil
}

func cleanupName(name string) string {
	return strings.ReplaceAll(strings.TrimPrefix(name, "/"), "/", ".")
}
