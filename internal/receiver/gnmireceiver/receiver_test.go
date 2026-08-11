// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gnmireceiver

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	rcvrmetadata "github.com/signalfx/splunk-otel-collector/internal/receiver/gnmireceiver/internal/metadata"
)

type mockGNMIServer struct {
	gnmipb.UnimplementedGNMIServer
	lastReq      *gnmipb.SubscribeRequest
	lastMD       metadata.MD
	updates      int
	mu           sync.Mutex
	subscribeHit atomic.Int32
	failFirst    atomic.Bool
	sendError    bool
	// silent accepts the subscription but never responds, so tests can tell
	// "request sent" apart from "target responded".
	silent bool
}

func (m *mockGNMIServer) lastRequest() *gnmipb.SubscribeRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastReq
}

func (m *mockGNMIServer) lastMetadata() metadata.MD {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastMD
}

func (m *mockGNMIServer) Subscribe(stream gnmipb.GNMI_SubscribeServer) error {
	req, err := stream.Recv()
	if err != nil {
		return err
	}
	md, _ := metadata.FromIncomingContext(stream.Context())
	m.mu.Lock()
	m.lastReq = req
	m.lastMD = md.Copy()
	m.mu.Unlock()
	m.subscribeHit.Add(1)

	if m.failFirst.CompareAndSwap(true, false) {
		return status.Error(codes.Unavailable, "simulated session failure")
	}

	if m.sendError {
		return status.Error(codes.InvalidArgument, "simulated subscription error")
	}
	if m.silent {
		<-stream.Context().Done()
		return nil
	}
	if err := stream.Send(&gnmipb.SubscribeResponse{
		Response: &gnmipb.SubscribeResponse_SyncResponse{SyncResponse: true},
	}); err != nil {
		return err
	}
	for i := 0; i < m.updates; i++ {
		notif := &gnmipb.Notification{
			Timestamp: time.Now().UnixNano(),
			Update: []*gnmipb.Update{{
				Path: &gnmipb.Path{Elem: []*gnmipb.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"name": "eth0"}},
					{Name: "state"},
					{Name: "counters"},
					{Name: "in-octets"},
				}},
				Val: &gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: uint64(i)}}, //nolint:gosec // disable G115: loop bound is small and non-negative
			}},
		}
		if err := stream.Send(&gnmipb.SubscribeResponse{
			Response: &gnmipb.SubscribeResponse_Update{Update: notif},
		}); err != nil {
			return err
		}
	}
	<-stream.Context().Done()
	return nil
}

func startMockServer(t *testing.T, srv *mockGNMIServer) (addr string, stop func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	gnmipb.RegisterGNMIServer(grpcServer, srv)
	go func() { _ = grpcServer.Serve(lis) }()

	return lis.Addr().String(), grpcServer.Stop
}

func testTarget(endpoint string) TargetConfig {
	tc := NewDefaultTargetConfig()
	tc.ClientConfig.Endpoint = endpoint
	tc.ClientConfig.TLS.Insecure = true
	tc.Redial = time.Second
	tc.Subscriptions = []SubscriptionConfig{{
		Path:           "/interfaces/interface/state/counters",
		Mode:           modeSample,
		SampleInterval: 10 * time.Millisecond,
		Default:        &MetricConfig{Type: metricTypeSum, Unit: "By"},
	}}
	return tc
}

func TestReceiverStartShutdown(t *testing.T) {
	srv := &mockGNMIServer{updates: 3}
	addr, stop := startMockServer(t, srv)
	defer stop()

	cfg := &Config{Targets: []TargetConfig{testTarget(addr)}}
	sink := new(consumertest.MetricsSink)
	r := newGNMIReceiver(cfg, receivertest.NewNopSettings(rcvrmetadata.Type), sink)

	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	require.Eventually(t, func() bool {
		return srv.subscribeHit.Load() >= 1
	}, 2*time.Second, 10*time.Millisecond)

	req := srv.lastRequest()
	require.NotNil(t, req)
	sl := req.GetSubscribe()
	require.NotNil(t, sl)
	assert.Equal(t, gnmipb.SubscriptionList_STREAM, sl.GetMode())
	require.Len(t, sl.GetSubscription(), 1)

	require.NoError(t, r.Shutdown(context.Background()))
}

func TestSubscriptionEstablishedLoggedOnlyAfterResponse(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	settings := receivertest.NewNopSettings(rcvrmetadata.Type)
	settings.Logger = zap.New(core)

	silentSrv := &mockGNMIServer{silent: true}
	addr, stop := startMockServer(t, silentSrv)
	defer stop()

	target := testTarget(addr)
	cfg := &Config{Targets: []TargetConfig{target}}
	r := newGNMIReceiver(cfg, settings, new(consumertest.MetricsSink))
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	require.Eventually(t, func() bool {
		return silentSrv.subscribeHit.Load() >= 1
	}, 2*time.Second, 10*time.Millisecond)
	assert.Empty(t, logs.FilterMessage("gNMI subscription established").All(),
		"established must not be logged before the target responds")
	require.NoError(t, r.Shutdown(context.Background()))

	core2, logs2 := observer.New(zap.InfoLevel)
	settings2 := receivertest.NewNopSettings(rcvrmetadata.Type)
	settings2.Logger = zap.New(core2)

	srv := &mockGNMIServer{updates: 1}
	addr2, stop2 := startMockServer(t, srv)
	defer stop2()

	cfg2 := &Config{Targets: []TargetConfig{testTarget(addr2)}}
	r2 := newGNMIReceiver(cfg2, settings2, new(consumertest.MetricsSink))
	require.NoError(t, r2.Start(context.Background(), componenttest.NewNopHost()))
	require.Eventually(t, func() bool {
		return len(logs2.FilterMessage("gNMI subscription established").All()) == 1
	}, 3*time.Second, 10*time.Millisecond)
	require.NoError(t, r2.Shutdown(context.Background()))
}

func TestCredentialMetadata(t *testing.T) {
	tests := []struct {
		name             string
		username         string
		password         string
		expectedUser     []string
		expectedPassword []string
	}{
		{
			name:             "username and password",
			username:         "admin",
			password:         "secret",
			expectedUser:     []string{"admin"},
			expectedPassword: []string{"secret"},
		},
		{
			name:             "username only omits password",
			username:         "admin",
			expectedUser:     []string{"admin"},
			expectedPassword: nil,
		},
		{
			name:             "no credentials",
			expectedUser:     nil,
			expectedPassword: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &mockGNMIServer{updates: 1}
			addr, stop := startMockServer(t, srv)
			defer stop()

			target := testTarget(addr)
			target.Username = configopaque.String(tt.username)
			target.Password = configopaque.String(tt.password)
			cfg := &Config{Targets: []TargetConfig{target}}
			r := newGNMIReceiver(cfg, receivertest.NewNopSettings(rcvrmetadata.Type), new(consumertest.MetricsSink))

			require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
			require.Eventually(t, func() bool {
				return srv.subscribeHit.Load() >= 1
			}, 3*time.Second, 10*time.Millisecond)
			md := srv.lastMetadata()
			assert.Equal(t, tt.expectedUser, md.Get("username"))
			assert.Equal(t, tt.expectedPassword, md.Get("password"))
			require.NoError(t, r.Shutdown(context.Background()))
		})
	}
}

func TestReceiverShutdownWithoutStart(t *testing.T) {
	cfg := &Config{Targets: []TargetConfig{testTarget("localhost:1")}}
	r := newGNMIReceiver(cfg, receivertest.NewNopSettings(rcvrmetadata.Type), new(consumertest.MetricsSink))

	require.NoError(t, r.Shutdown(context.Background()))
}

func TestReceiverRedialsAfterSessionFailure(t *testing.T) {
	srv := &mockGNMIServer{updates: 1}
	srv.failFirst.Store(true)
	addr, stop := startMockServer(t, srv)
	defer stop()

	target := testTarget(addr)
	target.Redial = 100 * time.Millisecond
	cfg := &Config{Targets: []TargetConfig{target}}
	r := newGNMIReceiver(cfg, receivertest.NewNopSettings(rcvrmetadata.Type), new(consumertest.MetricsSink))

	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	require.Eventually(t, func() bool {
		return srv.subscribeHit.Load() >= 2
	}, 5*time.Second, 20*time.Millisecond, "expected the client to redial and re-subscribe")
	require.NoError(t, r.Shutdown(context.Background()))
}

func TestReceiverRedialDisabledStopsAfterFailure(t *testing.T) {
	srv := &mockGNMIServer{}
	srv.failFirst.Store(true)
	addr, stop := startMockServer(t, srv)
	defer stop()

	target := testTarget(addr)
	target.Redial = 0
	cfg := &Config{Targets: []TargetConfig{target}}
	r := newGNMIReceiver(cfg, receivertest.NewNopSettings(rcvrmetadata.Type), new(consumertest.MetricsSink))

	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	require.Eventually(t, func() bool {
		return srv.subscribeHit.Load() == 1
	}, 2*time.Second, 10*time.Millisecond)
	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, int32(1), srv.subscribeHit.Load(), "redial is disabled, session must not be retried")
	require.NoError(t, r.Shutdown(context.Background()))
}

func TestSubscriptionErrorEndsSession(t *testing.T) {
	srv := &mockGNMIServer{sendError: true}
	addr, stop := startMockServer(t, srv)
	defer stop()

	target := testTarget(addr)
	target.Redial = 100 * time.Millisecond
	cfg := &Config{Targets: []TargetConfig{target}}
	r := newGNMIReceiver(cfg, receivertest.NewNopSettings(rcvrmetadata.Type), new(consumertest.MetricsSink))

	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	require.Eventually(t, func() bool {
		return srv.subscribeHit.Load() >= 2
	}, 5*time.Second, 20*time.Millisecond, "error response should end the session and trigger redial")
	require.NoError(t, r.Shutdown(context.Background()))
}

func TestReceiverStartFailsOnInvalidPath(t *testing.T) {
	target := testTarget("localhost:57400")
	target.Subscriptions[0].Path = "/interfaces/interface[]/state"
	cfg := &Config{Targets: []TargetConfig{target}}
	r := newGNMIReceiver(cfg, receivertest.NewNopSettings(rcvrmetadata.Type), new(consumertest.MetricsSink))

	err := r.Start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid gNMI path")
	require.NoError(t, r.Shutdown(context.Background()))
}

func TestKeyedPathIsParsed(t *testing.T) {
	p, err := parseGNMIPath("openconfig", "/interfaces/interface[name=eth0]/state")
	require.NoError(t, err)
	require.Len(t, p.GetElem(), 3)
	assert.Equal(t, "interface", p.GetElem()[1].GetName())
	assert.Equal(t, map[string]string{"name": "eth0"}, p.GetElem()[1].GetKey())
	assert.Equal(t, "openconfig", p.GetOrigin())
}

func TestBuildSubscribeRequest(t *testing.T) {
	tc := testTarget("localhost:57400")
	tc.Encoding = encodingJSONIETF
	tc.Subscriptions = []SubscriptionConfig{
		{
			Path: "/a/b", Origin: "openconfig", Mode: modeSample, SampleInterval: time.Second,
			Default: &MetricConfig{Type: metricTypeSum},
		},
		{Path: "/c", Mode: modeOnChange, Default: &MetricConfig{Type: metricTypeGauge}},
	}
	c := &gnmiClient{target: &tc}

	req, err := c.buildSubscribeRequest()
	require.NoError(t, err)
	sl := req.GetSubscribe()
	require.NotNil(t, sl)
	assert.Equal(t, gnmipb.SubscriptionList_STREAM, sl.GetMode())
	assert.Equal(t, gnmipb.Encoding_JSON_IETF, sl.GetEncoding())
	require.Len(t, sl.GetSubscription(), 2)

	assert.Equal(t, gnmipb.SubscriptionMode_SAMPLE, sl.GetSubscription()[0].GetMode())
	assert.Equal(t, "openconfig", sl.GetSubscription()[0].GetPath().GetOrigin())
	assert.Equal(t, uint64(time.Second.Nanoseconds()), sl.GetSubscription()[0].GetSampleInterval()) //nolint:gosec // disable G115: constant is positive
	assert.Equal(t, gnmipb.SubscriptionMode_ON_CHANGE, sl.GetSubscription()[1].GetMode())
}

func TestSubscriptionModeMapping(t *testing.T) {
	assert.Equal(t, gnmipb.SubscriptionMode_SAMPLE, subscriptionMode(modeSample))
	assert.Equal(t, gnmipb.SubscriptionMode_ON_CHANGE, subscriptionMode(modeOnChange))
	assert.Equal(t, gnmipb.SubscriptionMode_TARGET_DEFINED, subscriptionMode(modeTargetDefined))
	assert.Equal(t, gnmipb.SubscriptionMode_SAMPLE, subscriptionMode("unknown"))
}

func TestEncodingMapping(t *testing.T) {
	assert.Equal(t, gnmipb.Encoding_PROTO, gnmiEncoding(encodingProto))
	assert.Equal(t, gnmipb.Encoding_JSON, gnmiEncoding(encodingJSON))
	assert.Equal(t, gnmipb.Encoding_JSON_IETF, gnmiEncoding(encodingJSONIETF))
	assert.Equal(t, gnmipb.Encoding_PROTO, gnmiEncoding(""))
}

func TestParseGNMIPath(t *testing.T) {
	p, err := parseGNMIPath("openconfig", "/interfaces/interface/state")
	require.NoError(t, err)
	assert.Equal(t, "openconfig", p.GetOrigin())
	require.Len(t, p.GetElem(), 3)
	assert.Equal(t, "interfaces", p.GetElem()[0].GetName())
	assert.Equal(t, "state", p.GetElem()[2].GetName())

	empty, err := parseGNMIPath("", "/")
	require.NoError(t, err)
	assert.Empty(t, empty.GetElem())
}

func TestParseGNMIPathRejectsMalformed(t *testing.T) {
	_, err := parseGNMIPath("", "/interfaces/interface[]/state")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid gNMI path")
}
