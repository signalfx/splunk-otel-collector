// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package signalfxgatewayprometheusremotewritereceiver

import (
	"context"
	"errors"
	"net"
	"net/url"
	"syscall"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
)

type MockPrwClient struct {
	Client  remote.WriteClient
	Timeout time.Duration
}

func NewMockPrwClient(addr string, path string, timeout time.Duration) (MockPrwClient, error) {
	URL := &config.URL{
		URL: &url.URL{
			Scheme: "http",
			Host:   addr,
			Path:   path,
		},
	}
	cfg := &remote.ClientConfig{
		URL:              URL,
		Timeout:          model.Duration(timeout),
		HTTPClientConfig: config.HTTPClientConfig{},
	}
	client, err := remote.NewWriteClient("mock_prw_client", cfg)
	return MockPrwClient{
		Client:  client,
		Timeout: timeout,
	}, err
}

func (prwc *MockPrwClient) SendWriteRequest(wr *prompb.WriteRequest) error {

	data, err := proto.Marshal(wr)
	if err != nil {
		return err
	}

	compressed := snappy.Encode(nil, data)

	ctx, cancel := context.WithTimeout(context.Background(), prwc.Timeout)
	defer cancel()
	retry := 10
	for retry > 0 {
		err = prwc.Client.Store(ctx, compressed)
		if nil == err {
			return nil
		}
		if errors.Is(err, syscall.ECONNREFUSED) {
			retry--
			time.Sleep(2 * time.Second)
		} else {
			return err
		}
	}
	return errors.New("failed to send prometheus remote write requests to server")
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
