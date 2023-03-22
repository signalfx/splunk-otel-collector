// Package selfdescribe contains code that knows how to pull metadata from
// various agent component out into a well structured format that can be fed
// into other workflows to generate documentation or other resources such as
// chef recipies.  The main interface is the JSON() function, that returns
// everything encoded as JSON.
package selfdescribe

import (
	"encoding/json"
	"io"
	"reflect"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/config/sources"
)

// JSON returns a json encoded string of all of the documentation for the
// various components in the agent.  It is meant to be used as an intermediate
// form which serves as a data source for automatically generating docs about
// the agent.
func JSON(writer io.Writer) {
	osm, err := observersStructMetadata()
	if err != nil {
		panic(err)
	}

	out, err := json.MarshalIndent(map[string]interface{}{
		"TopConfig":             getStructMetadata(reflect.TypeOf(config.Config{})),
		"GenericMonitorConfig":  getStructMetadata(reflect.TypeOf(config.MonitorConfig{})),
		"GenericObserverConfig": getStructMetadata(reflect.TypeOf(config.ObserverConfig{})),
		"Monitors":              monitorsStructMetadata(),
		"Observers":             osm,
		"SourceConfig":          getStructMetadata(reflect.TypeOf(sources.SourceConfig{})),
	}, "", "  ")
	if err != nil {
		panic(err)
	}
	_, _ = writer.Write(out)
}
