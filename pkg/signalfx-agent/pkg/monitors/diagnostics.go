package monitors

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/leadership"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

const maxLineLength = 120

func endpointToDiagnosticText(endpoint services.Endpoint, isMonitored bool) string {
	var items []string

	items = append(items, "internalId: "+string(endpoint.Core().ID))
	if !isMonitored {
		items[0] += " (UNMONITORED)"
	}

	endpointMap := services.EndpointAsMap(endpoint)
	sortedKeys := utils.SortMapKeys(endpointMap)
	for _, k := range sortedKeys {
		items = append(items, fmt.Sprintf("%s: %v", k, endpointMap[k]))
	}

	out := " - " + strings.Join(items, "\n   ")

	return out
}

// EndpointsDiagnosticText returns diagnostic text about discovered endpoints
func (mm *MonitorManager) EndpointsDiagnosticText() string {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	out := "Discovered Endpoints:             "
	for _, endpoint := range mm.discoveredEndpoints {
		out += "\n" + endpointToDiagnosticText(endpoint, mm.isEndpointMonitored(endpoint)) + "\n"
	}
	if len(out) == 0 {
		out = "None"
	}
	return out
}

// SummaryDiagnosticText is a shorter version of DiagnosticText()
func (mm *MonitorManager) SummaryDiagnosticText() string {
	return fmt.Sprintf(
		"Active Monitors:                  %d\n"+
			"Configured Monitors:              %d\n"+
			"Discovered Endpoint Count:        %d\n"+
			"Bad Monitor Config:               %s",
		len(mm.activeMonitors),
		len(mm.monitorConfigs),
		len(mm.discoveredEndpoints),
		mm.BadConfigDiagnosticText(),
	)
}

func formatEnabledMetrics(metrics []string, indent int) string {
	metricList := strings.Join(metrics, ", ")
	enabledMetricsPrefix := utils.IndentLines("Enabled Metrics: ", indent)
	text := fmt.Sprintf("%s[%s]", enabledMetricsPrefix, metricList)

	if len(text) <= maxLineLength {
		return text
	}

	// Single line is too long, wrap it on multiple lines instead.
	// fmt string is equally unreadable so just join it all together.
	return strings.Join([]string{
		enabledMetricsPrefix, "[\n",
		utils.IndentLines(wordwrap.WrapString(metricList, uint(maxLineLength-indent+2)), indent+2), "\n",
		utils.IndentLines("]", indent),
	}, "")
}

// DiagnosticText returns a string to be served on the diagnostic socket
func (mm *MonitorManager) DiagnosticText() string {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	activeMonText := ""
	for i := range mm.activeMonitors {
		am := mm.activeMonitors[i]

		serviceStats := "Not using auto-discovery"
		if am.endpoint != nil {
			serviceStats = fmt.Sprintf(
				`Discovery Rule: %s
Monitored Endpoint ID: %s
Endpoint Dimensions: %s`,
				am.config.MonitorConfigCore().DiscoveryRule,
				am.endpoint.Core().ID,
				utils.FormatStringMapCompact(am.endpoint.Dimensions()),
			)
		}
		activeMonText += fmt.Sprintf(
			`%s. %s
    Reporting Interval (seconds): %d
%s
%s
    Config:
%s
`,
			am.config.MonitorConfigCore().MonitorID,
			am.config.MonitorConfigCore().Type,
			am.config.MonitorConfigCore().IntervalSeconds,
			formatEnabledMetrics(am.output.EnabledMetrics(), 4),
			utils.IndentLines(serviceStats, 4),
			utils.IndentLines(config.ToString(am.config), 6))
	}
	return "Active Monitors:\n" + activeMonText
}

// BadConfigDiagnosticText returns a text representation of any bad monitor
// config that is preventing things from being monitored.
func (mm *MonitorManager) BadConfigDiagnosticText() string {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	if len(mm.badConfigs) > 0 {
		var texts []string
		for k := range mm.badConfigs {
			conf := mm.badConfigs[k]
			texts = append(texts, fmt.Sprintf("[type: %s, error: %s]",
				conf.Type, conf.ValidationError))
		}
		return strings.Join(texts, " ")
	}
	return "None"
}

// InternalMetrics returns a list of datapoints about the internal status of
// the monitors
func (mm *MonitorManager) InternalMetrics() []*datapoint.Datapoint {
	return []*datapoint.Datapoint{
		sfxclient.Gauge("sfxagent.active_monitors", nil, int64(len(mm.activeMonitors))),
		sfxclient.Gauge("sfxagent.configured_monitors", nil, int64(len(mm.monitorConfigs))),
		sfxclient.Gauge("sfxagent.discovered_endpoints", nil, int64(len(mm.discoveredEndpoints))),
		sfxclient.Gauge("sfxagent.k8s_leader", map[string]string{"leader_node": leadership.CurrentLeader()}, 1),
	}
}
