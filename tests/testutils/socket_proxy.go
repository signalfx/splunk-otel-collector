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
	"sync"
	"sync/atomic"
	"testing"
)

type SocketProxy struct {
	Path              string
	Endpoint          string
	ContainerEndpoint string
	listener          net.Listener
	wg                sync.WaitGroup
	active            atomic.Bool
	t                 testing.TB
}

func CreateDockerSocketProxy(t testing.TB) *SocketProxy {
	port := GetAvailablePort(t)

	dockerHost := "172.17.0.1"
	if runtime.GOOS == "darwin" {
		dockerHost = "host.docker.internal"
	}

	return &SocketProxy{
		Path:              "/var/run/docker.sock",
		Endpoint:          fmt.Sprintf("0.0.0.0:%d", port),
		ContainerEndpoint: fmt.Sprintf("%s:%d", dockerHost, port),
		t:                 t,
	}
}

func (s *SocketProxy) Start() error {
	l, err := net.Listen("tcp", s.Endpoint)
	if err != nil {
		return err
	}
	s.listener = l
	s.wg.Add(1)
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
		s.wg.Done()
	}()
	return nil
}

func (s *SocketProxy) Stop() {
	s.active.Store(false)
}
