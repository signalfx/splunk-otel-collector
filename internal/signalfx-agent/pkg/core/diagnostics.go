package core

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"time"

	// Import for side-effect of registering http handler
	_ "net/http/pprof"

	"github.com/davecgh/go-spew/spew"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/core/writer/tap"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// VersionLine should be populated by the startup logic to contain version
// information that can be reported in diagnostics.
// nolint: gochecknoglobals
var VersionLine string

// Serves the diagnostic status on the specified path
func (a *Agent) serveDiagnosticInfo(host string, port uint16) {
	if a.diagnosticServer != nil {
		a.diagnosticServer.Close()
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(a.diagnosticTextHandler))
	mux.Handle("/metrics", http.HandlerFunc(a.internalMetricsHandler))
	mux.Handle("/tap-dps", http.HandlerFunc(a.datapointTapHandler))

	a.diagnosticServer = &http.Server{
		Addr:        fmt.Sprintf("%s:%d", host, port),
		Handler:     mux,
		ReadTimeout: 5 * time.Second,
		// Set this to 0 so that streaming works.
		WriteTimeout: 0, //5 * time.Second,
	}

	go func() {
		log.Infof("Serving internal metrics at %s:%d", host, port)
		for {
			err := a.diagnosticServer.ListenAndServe()
			if err != nil {
				if err == http.ErrServerClosed {
					return
				}
				log.WithFields(log.Fields{
					"host":  host,
					"port":  port,
					"error": err,
				}).Error("Problem with diagnostic server")
			}
			// Retry after a cool down since sometimes the port can be still
			// bound up from a previously running agent that was restarted.
			time.Sleep(15 * time.Second)
		}
	}()
}

func readStatusInfo(host string, port uint16, section string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/?section=%s", host, port, section))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (a *Agent) internalMetricsHandler(rw http.ResponseWriter, req *http.Request) {
	jsonOut, err := json.Marshal(a.InternalMetrics())
	if err != nil {
		log.WithError(err).Error("Could not serialize internal metrics to JSON")
		rw.WriteHeader(500)
		_, _ = rw.Write([]byte(err.Error()))
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(200)

	_, _ = rw.Write(jsonOut)
}

// InternalMetrics aggregates internal metrics from subcomponents and returns a
// list of datapoints that represent the instaneous state of the agent
func (a *Agent) InternalMetrics() []*datapoint.Datapoint {
	out := make([]*datapoint.Datapoint, 0)

	mstat := runtime.MemStats{}
	runtime.ReadMemStats(&mstat)
	out = append(out, []*datapoint.Datapoint{
		sfxclient.Cumulative("sfxagent.go_total_alloc", nil, int64(mstat.TotalAlloc)),
		sfxclient.Gauge("sfxagent.go_sys", nil, int64(mstat.Sys)),
		sfxclient.Cumulative("sfxagent.go_mallocs", nil, int64(mstat.Mallocs)),
		sfxclient.Cumulative("sfxagent.go_frees", nil, int64(mstat.Frees)),
		sfxclient.Gauge("sfxagent.go_heap_alloc", nil, int64(mstat.HeapAlloc)),
		sfxclient.Gauge("sfxagent.go_heap_sys", nil, int64(mstat.HeapSys)),
		sfxclient.Gauge("sfxagent.go_heap_idle", nil, int64(mstat.HeapIdle)),
		sfxclient.Gauge("sfxagent.go_heap_inuse", nil, int64(mstat.HeapInuse)),
		sfxclient.Gauge("sfxagent.go_heap_released", nil, int64(mstat.HeapReleased)),
		sfxclient.Gauge("sfxagent.go_stack_inuse", nil, int64(mstat.StackInuse)),
		sfxclient.Gauge("sfxagent.go_next_gc", nil, int64(mstat.NextGC)),
		sfxclient.Cumulative("sfxagent.go_pause_total_ns", nil, int64(mstat.PauseTotalNs)),
		sfxclient.Cumulative("sfxagent.go_gc_cpu_fraction", nil, int64(mstat.GCCPUFraction)),
		sfxclient.Gauge("sfxagent.go_num_gc", nil, int64(mstat.NumGC)),
		sfxclient.Gauge("sfxagent.go_gomaxprocs", nil, int64(runtime.GOMAXPROCS(0))),
		sfxclient.Gauge("sfxagent.go_num_goroutine", nil, int64(runtime.NumGoroutine())),
	}...)

	out = append(out, a.writer.InternalMetrics()...)
	out = append(out, a.observers.InternalMetrics()...)
	out = append(out, a.monitors.InternalMetrics()...)

	for i := range out {
		if out[i].Dimensions == nil {
			out[i].Dimensions = make(map[string]string)
		}

		out[i].Dimensions["host"] = a.lastConfig.Hostname
		out[i].Timestamp = time.Now()
	}
	return out
}

func (a *Agent) ensureProfileServerRunning(host string, port int) {
	if !a.profileServerRunning {
		// We don't use that much memory so the default mem sampling rate is
		// too small to be very useful. Setting to 1 profiles ALL allocations
		runtime.MemProfileRate = 1
		// Crank up CPU profile rate too since our CPU usage tends to be pretty
		// bursty around read cycles.
		runtime.SetCPUProfileRate(-1)
		runtime.SetCPUProfileRate(2000)

		go func() {
			a.profileServerRunning = true
			// This is very difficult to access from the host on mac without
			// exposing it on all interfaces
			log.Println(http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil))
		}()
	}
}

func (a *Agent) datapointTapHandler(rw http.ResponseWriter, req *http.Request) {
	metricQuery := utils.DecodeValueGenerically(req.URL.Query().Get("metric"))
	dimQuery := utils.DecodeValueGenerically(req.URL.Query().Get("dims"))

	var metricFilter []string
	if metricQuery != "" {
		switch v := metricQuery.(type) {
		case string:
			metricFilter = []string{v}
		case []string:
			metricFilter = v
		default:
			rw.WriteHeader(400)
			_, _ = rw.Write([]byte("bad metric query: " + spew.Sdump(v)))
			return
		}
	}

	var dimFilter map[string][]string
	if dimQuery != "" {
		switch v := dimQuery.(type) {
		case yaml.MapSlice:
			dimFilter = make(map[string][]string)
			for i := range v {
				var vals []string
				switch v2 := v[i].Value.(type) {
				case []interface{}:
					vals = utils.InterfaceSliceToStringSlice(v2)
				default:
					vals = []string{fmt.Sprintf("%v", v2)}
				}

				dimFilter[fmt.Sprintf("%v", v[i].Key)] = vals
			}
		default:
			rw.WriteHeader(400)
			_, _ = rw.Write([]byte("bad dims query: " + spew.Sdump(v)))
			return
		}
	}

	var filter dpfilters.DatapointFilter
	if metricFilter == nil && dimFilter == nil {
		filter = &dpfilters.AlwaysMatchFilter{}
	} else {
		var err error
		filter, err = dpfilters.NewOverridable(metricFilter, dimFilter)
		if err != nil {
			rw.WriteHeader(400)
			_, _ = rw.Write([]byte("could not make filter: " + err.Error()))
			return
		}
	}

	rw.WriteHeader(200)
	dpTap := tap.New(filter, rw)

	log.Infof("Datapoint tap started")
	a.writer.SetTap(dpTap)

	dpTap.Run(req.Context())

	a.writer.SetTap(nil)
	log.Infof("Datapoint tap cleared")
}

func streamDatapoints(host string, port uint16, metric string, dims string) (io.ReadCloser, error) {
	c := http.Client{
		Timeout: 0,
	}
	qs := url.Values{}
	qs.Set("metric", metric)
	qs.Set("dims", dims)
	resp, err := c.Get(fmt.Sprintf("http://%s:%d/tap-dps?%s", host, port, qs.Encode())) // nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
