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
	dims := env.Tags

	prefix := ""

	if env.SourceId != "" {
		dims["source_id"] = env.SourceId
		if hexIDRegexp.MatchString(env.SourceId) {
			prefix = env.Tags["origin"] + "."
		} else {
			prefix = env.SourceId + "."
		}
	}

	if env.InstanceId != "" {
		dims["instance_id"] = env.InstanceId
	}

	var metricType datapoint.MetricType

	namesToValues := make(map[string]float64)

	switch m := env.Message.(type) {
	case *loggregator_v2.Envelope_Log:
	case *loggregator_v2.Envelope_Counter:
		metricType = datapoint.Counter
		namesToValues[m.Counter.GetName()] = float64(m.Counter.GetTotal())
	case *loggregator_v2.Envelope_Gauge:
		metricType = datapoint.Gauge
		for name, gauge := range m.Gauge.GetMetrics() {
			namesToValues[name] = gauge.Value
		}
	case *loggregator_v2.Envelope_Timer:
	case *loggregator_v2.Envelope_Event:
	default:
		return nil, fmt.Errorf("cannot convert envelope %v to SignalFx datapoints", spew.Sdump(env))
	}

	var dps []*datapoint.Datapoint
	for name, val := range namesToValues {
		dps = append(dps, datapoint.New(prefix+cleanupName(name), dims, datapoint.NewFloatValue(val), metricType, time.Unix(0, env.Timestamp)))
	}

	return dps, nil
}

func cleanupName(name string) string {
	return strings.ReplaceAll(strings.TrimPrefix(name, "/"), "/", ".")
}
