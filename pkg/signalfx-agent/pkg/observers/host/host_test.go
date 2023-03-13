package host

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/observers"
	"github.com/stretchr/testify/require"
)

var (
	exe, _ = os.Executable()
)

func openTestTCPPorts(t *testing.T) []*net.TCPListener {
	tcpLocalhost, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	require.Nil(t, err)

	tcpV6Localhost, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("[::]"),
		Port: 0,
	})
	require.Nil(t, err)

	tcpAllPorts, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 0,
	})
	require.Nil(t, err)

	return []*net.TCPListener{
		tcpLocalhost,
		tcpV6Localhost,
		tcpAllPorts,
	}
}

func openTestUDPPorts(t *testing.T) []*net.UDPConn {
	udpLocalhost, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	require.Nil(t, err)

	udpV6Localhost, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("[::]"),
		Port: 0,
	})
	require.Nil(t, err)

	udpAllPorts, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 0,
	})
	require.Nil(t, err)

	return []*net.UDPConn{
		udpLocalhost,
		udpV6Localhost,
		udpAllPorts,
	}
}

var selfPid = os.Getpid()

func Test_HostObserver(t *testing.T) {
	config := &Config{
		PollIntervalSeconds: 1,
	}

	var o *Observer
	var obsLock sync.Mutex
	var endpointLock sync.Mutex
	var endpoints map[services.ID]services.Endpoint

	startObserver := func() (unlock func()) {
		endpoints = make(map[services.ID]services.Endpoint)
		obsLock.Lock()

		o = &Observer{
			serviceCallbacks: &observers.ServiceCallbacks{
				Added: func(se services.Endpoint) {
					endpointLock.Lock()
					defer endpointLock.Unlock()
					endpoints[se.Core().ID] = se
				},
				Removed: func(se services.Endpoint) {
					endpointLock.Lock()
					defer endpointLock.Unlock()
					delete(endpoints, se.Core().ID)
				},
			},
		}
		err := o.Configure(config)
		if err != nil {
			panic("could not setup observer")
		}
		return obsLock.Unlock
	}

	t.Run("Omit PID", func(t *testing.T) {
		config.OmitPIDDimension = true
		defer func() { config.OmitPIDDimension = false }()
		tcpConns := openTestTCPPorts(t)
		defer startObserver()()

		require.Eventually(t, func() bool {
			endpointLock.Lock()
			defer endpointLock.Unlock()
			return len(endpoints) >= len(tcpConns)
		}, 2*time.Second, time.Millisecond)

		for _, e := range endpoints {
			_, ok := e.Dimensions()["pid"]
			require.False(t, ok)
		}

	})

	t.Run("Basic connections", func(t *testing.T) {
		tcpConns := openTestTCPPorts(t)
		udpConns := openTestUDPPorts(t)

		defer func() {
			for _, conn := range tcpConns {
				conn.Close()
			}
			for _, conn := range udpConns {
				conn.Close()
			}
		}()
		defer startObserver()()

		require.Eventually(t, func() bool {
			endpointLock.Lock()
			defer endpointLock.Unlock()
			return len(endpoints) >= len(tcpConns)+len(udpConns)
		}, 2*time.Second, time.Millisecond)

		t.Run("TCP ports", func(t *testing.T) {
			for _, conn := range tcpConns {
				host, port, _ := net.SplitHostPort(conn.Addr().String())
				expectedID := fmt.Sprintf("%s-%s-TCP-%d", host, port, selfPid)
				endpointLock.Lock()
				e := endpoints[services.ID(expectedID)]
				endpointLock.Unlock()
				require.NotNil(t, e)
				endpoint := e.(*services.EndpointCore)

				portNum, _ := strconv.Atoi(port)
				require.EqualValues(t, endpoint.Port, portNum)
				require.Equal(t, filepath.Base(exe), endpoint.Name)
				require.Equal(t, endpoint.PortType, services.TCP)
				if host[0] == ':' {
					require.Equal(t, endpoint.DerivedFields()["is_ipv6"], true)
				} else {
					require.Equal(t, endpoint.DerivedFields()["is_ipv6"], false)
				}

				_, ok := e.Dimensions()["pid"]
				require.True(t, ok)
			}
		})

		t.Run("UDP Ports", func(t *testing.T) {
			if runtime.GOOS == "windows" {
				t.Skip("skipping test on windows")
			}
			for _, conn := range udpConns {
				host, port, _ := net.SplitHostPort(conn.LocalAddr().String())
				expectedID := fmt.Sprintf("%s-%s-UDP-%d", host, port, selfPid)
				endpointLock.Lock()
				e := endpoints[services.ID(expectedID)]
				endpointLock.Unlock()
				require.NotNil(t, e)
				endpoint := e.(*services.EndpointCore)
				portNum, _ := strconv.Atoi(port)
				require.EqualValues(t, endpoint.Port, portNum)
				require.Equal(t, filepath.Base(exe), endpoint.Name)
				require.Equal(t, services.UDP, endpoint.PortType)
				if host[0] == ':' {
					require.Equal(t, endpoint.DerivedFields()["is_ipv6"], true)
				} else {
					require.Equal(t, endpoint.DerivedFields()["is_ipv6"], false)
				}

				_, ok := e.Dimensions()["pid"]
				require.True(t, ok)
			}
		})
	})
}
