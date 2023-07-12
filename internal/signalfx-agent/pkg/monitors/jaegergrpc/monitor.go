package jaegergrpc

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/jaegertracing/jaeger/proto-gen/api_v2"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/jaegergrpc/jaegerprotobuf"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

const gracefulShutdownTimeout = time.Second * 5

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// TLSCreds is a configuration struct for specifying a cert file and key file for tls
type TLSCreds struct {
	// The cert file to use for tls
	CertFile string `yaml:"certFile"`
	// The key file to use for tls
	KeyFile string `yaml:"keyFile"`
}

// Credentials returns a grpc credentials transport
func (tls *TLSCreds) Credentials() (credentials.TransportCredentials, error) {
	var creds credentials.TransportCredentials
	var err error
	if tls != nil {
		creds, err = credentials.NewServerTLSFromFile(tls.CertFile, tls.KeyFile)
	}

	return creds, err
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"false" singleInstance:"true"`
	// The host:port on which to listen for traces.
	ListenAddress string `yaml:"listenAddress" default:"0.0.0.0:14250"`
	// TLS are optional tls credential settings to configure the GRPC server with
	TLS *TLSCreds `yaml:"tls,omitempty"`
}

// Monitor that accepts and forwards SignalFx data
type Monitor struct {
	Output       types.Output
	grpc         *grpc.Server
	listenerLock sync.Mutex
	cancel       context.CancelFunc
	ln           net.Listener
	logger       *utils.ThrottledLogger
}

var _ api_v2.CollectorServiceServer = (*Monitor)(nil)

func extractRemoteAddressToContext(ctx context.Context) (net.IP, bool) {
	var sourceIP net.IP

	// get peer connection info from grpc context
	p, hasSource := peer.FromContext(ctx)
	if !hasSource || p.Addr.String() == "" {
		return nil, false
	}

	// separate host from port
	host, _, err := net.SplitHostPort(p.Addr.String())
	if err != nil {
		return nil, false
	}

	// parse ip from host
	sourceIP = net.ParseIP(host)
	if sourceIP == nil {
		return nil, false
	}

	return sourceIP, true
}

// PostSpans implements the jeager api_v2.CollectorServiceServer interface.  The grpc server will pass the jaeger
// batches it receives to this method.  This method will convert the jaeger batches to SignalFx spans and pass them
// on to the output writer.
func (m *Monitor) PostSpans(ctx context.Context, r *api_v2.PostSpansRequest) (*api_v2.PostSpansResponse, error) {
	if r == nil {
		return &api_v2.PostSpansResponse{}, nil
	}

	// convert the batch to SignalFx metrics
	spans := jaegerprotobuf.JaegerProtoBatchToSFX(&r.Batch)

	// tag the source on the span meta data
	source, hasSource := extractRemoteAddressToContext(ctx)
	if hasSource {
		for i := range spans {
			// monitor Output expects Meta to be non-nil
			if spans[i].Meta == nil {
				spans[i].Meta = map[interface{}]interface{}{}
			}
			spans[i].Meta[constants.DataSourceIPKey] = source
		}
	}

	// send the spans on through the agent
	m.Output.SendSpans(spans...)

	return &api_v2.PostSpansResponse{}, nil
}

func (m *Monitor) setupListener(ctx context.Context, conf *Config) (net.Listener, error) {
	for ctx.Err() == nil {
		// create a listener with the configured ListenAddress
		ln, err := net.Listen("tcp", conf.ListenAddress)

		// return if listener was successfully created
		if err == nil {
			return ln, nil
		}

		m.logger.Errorf("could not start grpc listener %v", err)

		// wait until the next interval to retry
		select {
		case <-time.After(time.Duration(conf.IntervalSeconds) * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}

	}
	return nil, ctx.Err()
}

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = utils.NewThrottledLogger(log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID}), 30*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	// parse tls configurations
	var creds credentials.TransportCredentials
	if conf.TLS != nil {
		var err error
		creds, err = conf.TLS.Credentials()
		if err != nil {
			return err
		}
	} else {
		creds = insecure.NewCredentials()
	}

	// create the grpc server
	m.grpc = grpc.NewServer(grpc.Creds(creds))

	// start the grpc server
	go func() {
		// register the monitor with the grpc server
		api_v2.RegisterCollectorServiceServer(m.grpc, m)

		// setup the listener
		ln, err := m.setupListener(ctx, conf)
		if err != nil {
			return
		}

		// save the completed listener to the monitor so tests can scrape its address
		m.listenerLock.Lock()
		m.ln = ln
		m.listenerLock.Unlock()

		// start the server
		if err := m.grpc.Serve(m.ln); err != nil {
			m.logger.Errorf("failed to start server in %s monitor", monitorType)
		}
	}()

	return nil
}

// Shutdown stops the forwarder and correlation MTSs
func (m *Monitor) Shutdown() {
	// cancel the monitor context
	if m.cancel != nil {
		m.cancel()
	}

	// stop the grpc server
	if m.grpc != nil {
		// set up a timeout function to stop the grpc server if it does not
		// gracefully stop in a reasonable time frame
		timeout := time.AfterFunc(gracefulShutdownTimeout, func() {
			m.grpc.Stop()
		})

		// gracefully shutdown the grpc server
		m.grpc.GracefulStop()

		// stop the timeout
		timeout.Stop()
	}
}
