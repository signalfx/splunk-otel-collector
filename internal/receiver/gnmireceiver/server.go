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
	"time"

	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type gnmiServer struct {
	gnmipb.UnimplementedGNMIServer
	consumer consumer.Metrics
	parser   *metricParser
	logger   *zap.Logger
}

func (s *gnmiServer) Set(ctx context.Context, req *gnmipb.SetRequest) (*gnmipb.SetResponse, error) {
	updates := make([]*gnmipb.Update, 0, len(req.GetUpdate())+len(req.GetReplace()))
	updates = append(updates, req.GetUpdate()...)
	updates = append(updates, req.GetReplace()...)
	if len(updates) > 0 {
		metrics, err := s.parser.parse(&gnmipb.SubscribeResponse{
			Response: &gnmipb.SubscribeResponse_Update{
				Update: &gnmipb.Notification{
					Timestamp: time.Now().UnixNano(),
					Prefix:    req.GetPrefix(),
					Update:    updates,
				},
			},
		})
		if err != nil {
			s.logger.Error("failed to parse incoming gNMI Set request", zap.Error(err))
		}
		if metrics.DataPointCount() > 0 {
			if err := s.consumer.ConsumeMetrics(ctx, metrics); err != nil {
				s.logger.Error("failed to forward metrics from incoming gNMI Set request", zap.Error(err))
				return nil, err
			}
		}
	}
	return &gnmipb.SetResponse{Timestamp: time.Now().UnixNano()}, nil
}
