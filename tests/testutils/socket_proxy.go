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

package testutils

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync/atomic"
	"testing"
)

type SocketProxy struct {
	listener          net.Listener
	t                 testing.TB
	Path              string
	Endpoint          string
	ContainerEndpoint string
	active            atomic.Bool
}

func CreateDockerSocketProxy(tb testing.TB) (*SocketProxy, error) {
	port := GetAvailablePort(tb)

	dockerHost := "172.17.0.1"
	if runtime.GOOS == "darwin" {
		dockerHost = "host.docker.internal"
	}

	s := &SocketProxy{
		Path:              "/var/run/docker.sock",
		Endpoint:          fmt.Sprintf("0.0.0.0:%d", port),
		ContainerEndpoint: fmt.Sprintf("%s:%d", dockerHost, port),
		t:                 tb,
	}
	err := s.Start()
	if err != nil {
		return nil, err
	}
	tb.Cleanup(s.Stop)
	return s, nil
}

func (s *SocketProxy) Start() error {
	l, err := net.Listen("tcp", s.Endpoint)
	if err != nil {
		return err
	}
	s.listener = l
	s.active.Store(true)
	go func() {
		for s.active.Load() {
			conn, err := l.Accept()
			if err != nil {
				break
			}
			go func(c net.Conn) {
				socketConn, err := net.Dial("unix", s.Path)
				if err != nil {
					s.t.Log("Error dialing", err)
					return
				}
				// Echo all incoming data.
				go func() {
					_, _ = io.Copy(c, socketConn)
					_ = socketConn.Close()
					_ = c.Close()
				}()
				_, _ = io.Copy(socketConn, c)
			}(conn)
		}
	}()
	return nil
}

func (s *SocketProxy) Stop() {
	s.active.Store(false)
}
