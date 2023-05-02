//go:build linux
// +build linux

package collectd

import (
	"net"
	"net/http"
	"time"

	"github.com/mailru/easyjson"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/ingest-protocols/protocol/collectd"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/utils/collectdutil"
)

// WriteHTTPServer is a reimplementation of what the metricproxy collectd
// endpoint does.  The main difference from metric proxy is that we propagate
// the meta field from collectd datapoints onto the resulting datapoints so
// that we can correlate metrics from collectd to specific monitors in the
// agent.  The server will run on the configured localhost port.
type WriteHTTPServer struct {
	dpCallback    func([]*datapoint.Datapoint)
	eventCallback func([]*event.Event)
	ipAddr        string
	port          uint16
	// Port can be 0, which lets the kernel choose a free port.  activePort
	// will be the chosen port once the server is running.
	activePort int
	server     *http.Server
}

// NewWriteHTTPServer creates but does not start a new write server
func NewWriteHTTPServer(ipAddr string, port uint16,
	dpCallback func([]*datapoint.Datapoint), eventCallback func([]*event.Event)) (*WriteHTTPServer, error) {

	inst := &WriteHTTPServer{
		ipAddr:        ipAddr,
		port:          port,
		dpCallback:    dpCallback,
		eventCallback: eventCallback,
		server: &http.Server{
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 1 * time.Second,
		},
	}
	inst.server.Handler = inst

	return inst, nil
}

// Start begins accepting connections on the write server.  Will return an
// error if it cannot bind to the configured port.
func (s *WriteHTTPServer) Start() error {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(s.ipAddr),
		Port: int(s.port),
	})
	if err != nil {
		return err
	}

	s.activePort = listener.Addr().(*net.TCPAddr).Port

	go func() { _ = s.server.Serve(listener) }()
	return nil
}

// RunningPort returns the TCP port that the server is running on. Should not
// be called before the Start method is called.
func (s *WriteHTTPServer) RunningPort() int {
	return s.activePort
}

// Shutdown stops the write server immediately
func (s *WriteHTTPServer) Shutdown() error {
	return s.server.Close()
}

// ServeHTTP accepts collectd write_http requests and sends the resulting
// datapoint/events to the configured callback functions.
func (s *WriteHTTPServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var writeBody JSONWriteBody
	if err := easyjson.UnmarshalFromReader(req.Body, &writeBody); err != nil {
		log.WithError(err).Error("Could not decode body of write_http request")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// This is yet another way that collectd plugins can tell the agent what
	// the monitorID is.  This is specifically useful for the notifications
	// emitted by the metadata plugin.
	monitorID := req.URL.Query().Get("monitorID")

	var events []*event.Event
	// Preallocate space for dps
	dps := make([]*datapoint.Datapoint, 0, len(writeBody)*2)
	for _, f := range writeBody {
		collectdutil.ConvertWriteFormat((*collectd.JSONWriteFormat)(f), &dps, &events)
	}

	for i := range events {
		if monitorID != "" {
			events[i].Properties["monitorID"] = monitorID
		}
	}
	for i := range dps {
		if monitorID != "" && dps[i].Meta["monitorID"] == nil {
			dps[i].Meta["monitorID"] = monitorID
		}
	}

	if len(events) > 0 {
		s.eventCallback(events)
	}
	if len(dps) > 0 {
		s.dpCallback(dps)
	}

	// Ingest returns this response but write_http doesn't care if it's there
	// or not and seems to dump the responses every couple of minutes for some
	// reason.
	//rw.Write([]byte(`"OK"`))
}
