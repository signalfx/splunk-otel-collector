package selfdescribe

import (
	"go/doc"
	"reflect"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
)

type monitorDoc struct {
	MonitorMetadata
	Config           structMetadata `json:"config"`
	AcceptsEndpoints bool           `json:"acceptsEndpoints"`
	SingleInstance   bool           `json:"singleInstance"`
}

func monitorsStructMetadata() []monitorDoc {
	sms := []monitorDoc{}
	// Set to track undocumented monitors
	monTypesSeen := map[string]bool{}

	if packages, err := CollectMetadata("pkg/monitors"); err != nil {
		log.Fatal(err)
	} else {
		for _, pkg := range packages {
			for _, monitor := range pkg.Monitors {
				monType := monitor.MonitorType

				if _, ok := monitors.ConfigTemplates[monType]; !ok {
					log.Errorf("Found metadata for %s monitor in %s but it doesn't appear to be registered",
						monType, pkg.Path)
					continue
				}
				t := reflect.TypeOf(monitors.ConfigTemplates[monType]).Elem()
				monTypesSeen[monType] = true

				checkSendAllLogic(monType, monitor.Metrics, monitor.SendAll)
				checkDuplicateMetrics(pkg.Path, monitor.Metrics)
				checkMetricTypes(pkg.Path, monitor.Metrics)

				if monitor.Groups == nil {
					monitor.Groups = make(map[string]*GroupMetadata)
				}

				for name, metric := range monitor.Metrics {
					group := ""
					if metric.Group != nil {
						group = *metric.Group
					}

					if monitor.Groups[group] == nil {
						monitor.Groups[group] = &GroupMetadata{}
					}

					groupMeta := monitor.Groups[group]
					groupMeta.Metrics = append(groupMeta.Metrics, name)
					sort.Strings(groupMeta.Metrics)
				}

				mc, _ := t.FieldByName("MonitorConfig")
				mmd := monitorDoc{
					Config: getStructMetadata(t),
					MonitorMetadata: MonitorMetadata{
						SendAll:      monitor.SendAll,
						SendUnknown:  monitor.SendUnknown,
						NoneIncluded: monitor.NoneIncluded,
						MonitorType:  monType,
						Dimensions:   monitor.Dimensions,
						Groups:       monitor.Groups,
						Metrics:      monitor.Metrics,
						Properties:   monitor.Properties,
						Doc:          monitor.Doc,
					},
					AcceptsEndpoints: mc.Tag.Get("acceptsEndpoints") == strconv.FormatBool(true),
					SingleInstance:   mc.Tag.Get("singleInstance") == strconv.FormatBool(true),
				}
				mmd.Config.Package = pkg.PackagePath

				sms = append(sms, mmd)
			}
		}
	}

	sort.Slice(sms, func(i, j int) bool {
		return sms[i].MonitorType < sms[j].MonitorType
	})

	for k := range monitors.ConfigTemplates {
		if !monTypesSeen[k] {
			log.Warnf("Monitor Type %s is registered but does not appear to have documentation", k)
		}
	}

	return sms
}

func dimensionsFromNotes(allDocs []*doc.Package) map[string]DimMetadata {
	dm := map[string]DimMetadata{}
	for _, note := range notesFromDocs(allDocs, "DIMENSION") {
		dm[note.UID] = DimMetadata{
			Description: commentTextToParagraphs(note.Body),
		}
	}
	return dm
}

func checkDuplicateMetrics(path string, metrics map[string]MetricMetadata) {
	seen := map[string]bool{}

	for metric := range metrics {
		if seen[metric] {
			log.Errorf("duplicate metric '%s' found in %s", metric, path)
		}
		seen[metric] = true
	}
}

func checkMetricTypes(path string, metrics map[string]MetricMetadata) {
	for metric, info := range metrics {
		t := info.Type
		if t != "gauge" && t != "counter" && t != "cumulative" {
			log.Errorf("Bad metric type '%s' for metric %s in %s", t, metric, path)
		}
	}
}

func checkSendAllLogic(monType string, metrics map[string]MetricMetadata, sendAll bool) {
	if len(metrics) == 0 {
		return
	}

	hasDefault := false
	for _, metricInfo := range metrics {
		hasDefault = hasDefault || metricInfo.Default
	}
	if hasDefault && sendAll {
		log.Warnf("sendAll was specified on monitor type '%s' but some metrics were also marked as 'default'", monType)
	} else if !hasDefault && !sendAll {
		log.Warnf("sendAll was not specified on monitor type '%s' and no metrics are marked as 'default'", monType)
	}
}
