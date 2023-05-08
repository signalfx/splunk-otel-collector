// Copyright OpenTelemetry Authors
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

package main

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
)

func main() {

	URL := &config.URL{
		URL: &url.URL{
			Scheme: "http",
			Host:   os.Getenv("endpoint"),
			Path:   os.Getenv("path"),
		},
	}
	cfg := &remote.ClientConfig{
		URL:              URL,
		HTTPClientConfig: config.HTTPClientConfig{},
	}
	client, err := remote.NewWriteClient("mock_prw_client", cfg)
	if err != nil {
		panic(err)
	}

	metrics := []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "fake_metric_total"},
				{Name: "instance", Value: cfg.URL.Host},
			},
			Samples: []prompb.Sample{
				{Value: 42, Timestamp: time.Now().UnixMilli()},
			},
		},
	}

	req := &prompb.WriteRequest{
		Timeseries: metrics,
	}
	compressed := encodeWriteRequest(req)

	for {
		client.Store(context.Background(), compressed)
		time.Sleep(10 * time.Second)
	}
}

func encodeWriteRequest(request *prompb.WriteRequest) []byte {
	data, err := proto.Marshal(request)
	if err != nil {
		panic(err)
	}

	return snappy.Encode(nil, data)
}
