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
	"errors"
	"io"

	"github.com/aristanetworks/goarista/gnmireverse"
	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type gnmiServer struct {
	gnmireverse.UnimplementedGNMIReverseServer
	consumer consumer.Metrics
	parser   *metricParser
	logger   *zap.Logger
}

func (s *gnmiServer) Publish(stream gnmireverse.GNMIReverse_PublishServer) error {
	for {
		resp, err := stream.Recv()
		response := resp.GetResponse()
		switch response.(type) {
		case *gnmipb.SubscribeResponse_Error:
			s.logger.Error(resp.GetError().Message) // SA1019: deprecated yet still in use
			continue
		case *gnmipb.SubscribeResponse_SyncResponse:
			s.logger.Debug("received a sync request")
			continue
		}
		if errors.Is(err, io.EOF) {
			s.logger.Debug("Received an EOF error")
			return stream.SendAndClose(&emptypb.Empty{})
		}

		if err != nil {
			s.logger.Debug("Received an error", zap.Error(err))
			return err
		}
		if err := s.consumeSubscribeResponse(stream.Context(), resp); err != nil {
			return err
		}
	}
}

func (s *gnmiServer) PublishGet(stream gnmireverse.GNMIReverse_PublishGetServer) error {
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			return err
		}
		for _, notification := range resp.GetNotification() {
			if err := s.consumeSubscribeResponse(stream.Context(), &gnmipb.SubscribeResponse{
				Response: &gnmipb.SubscribeResponse_Update{Update: notification},
			}); err != nil {
				return err
			}
		}
	}
}

func (s *gnmiServer) consumeSubscribeResponse(ctx context.Context, resp *gnmipb.SubscribeResponse) error {
	metrics, parseErr := s.parser.parse(resp)
	if parseErr != nil {
		s.logger.Error("failed to parse incoming gNMI reverse response", zap.Error(parseErr))
	}
	if metrics.DataPointCount() == 0 {
		s.logger.Debug("No data points reported")
		return nil
	}
	s.logger.Debug("got metrics!", zap.Int("m", metrics.DataPointCount()))
	if err := s.consumer.ConsumeMetrics(ctx, metrics); err != nil {
		s.logger.Error("failed to forward metrics from incoming gNMI reverse response", zap.Error(err))
		return err
	}
	return nil
}
