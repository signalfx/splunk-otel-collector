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
	"fmt"
	"io"
	"time"

	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type gnmiClient struct {
	host     component.Host
	consumer consumer.Metrics
	target   *TargetConfig
	parser   *metricParser
	logger   *zap.Logger
	conn     *grpc.ClientConn
	settings component.TelemetrySettings
}

func newGNMIClient(
	target *TargetConfig,
	host component.Host,
	settings component.TelemetrySettings,
	nextConsumer consumer.Metrics,
	parser *metricParser,
) *gnmiClient {
	return &gnmiClient{
		target:   target,
		host:     host,
		settings: settings,
		consumer: nextConsumer,
		parser:   parser,
		logger:   settings.Logger,
	}
}

func (c *gnmiClient) run(ctx context.Context) {
	defer func() {
		if c.conn != nil {
			_ = c.conn.Close()
		}
	}()

	endpoint := c.target.ClientConfig.Endpoint
	for ctx.Err() == nil {
		if err := c.subscribe(ctx); err != nil && ctx.Err() == nil {
			if c.target.Redial > 0 {
				c.logger.Warn("gNMI session ended, will retry",
					zap.String("endpoint", endpoint),
					zap.Duration("redial", c.target.Redial),
					zap.Error(err))
			} else {
				c.logger.Error("gNMI session ended and reconnection is disabled (redial: 0)",
					zap.String("endpoint", endpoint),
					zap.Error(err))
			}
		}

		if c.target.Redial <= 0 {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(c.target.Redial):
		}
	}
}

func (c *gnmiClient) connect(ctx context.Context) error {
	if _, err := c.buildSubscribeRequest(); err != nil {
		return err
	}
	conn, err := c.target.ClientConfig.ToClientConn(ctx, c.host.GetExtensions(), c.settings)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *gnmiClient) subscribe(ctx context.Context) error {
	streamCtx := ctx
	if c.target.Username != "" {
		kv := []string{"username", string(c.target.Username)}
		if c.target.Password != "" {
			kv = append(kv, "password", string(c.target.Password))
		}
		streamCtx = metadata.AppendToOutgoingContext(ctx, kv...)
	}

	client := gnmipb.NewGNMIClient(c.conn)
	stream, err := client.Subscribe(streamCtx)
	if err != nil {
		return err
	}

	req, err := c.buildSubscribeRequest()
	if err != nil {
		return err
	}
	if err := stream.Send(req); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
		c.logger.Debug("gNMI subscribe request send returned EOF, reading stream status",
			zap.String("endpoint", c.target.ClientConfig.Endpoint))
	}

	established := false
	for {
		resp, recvErr := stream.Recv()
		if recvErr != nil {
			if errors.Is(recvErr, io.EOF) || ctx.Err() != nil {
				return nil
			}
			return recvErr
		}

		if !established {
			established = true
			c.logger.Info("gNMI subscription established",
				zap.String("endpoint", c.target.ClientConfig.Endpoint),
				zap.Int("subscriptions", len(c.target.Subscriptions)))
		}

		metrics, parseErr := c.parser.parse(resp)
		if parseErr != nil {
			c.logger.Error("failed to parse gNMI response",
				zap.String("endpoint", c.target.ClientConfig.Endpoint),
				zap.Error(parseErr))
		}
		if metrics.DataPointCount() == 0 {
			continue
		}
		if consumeErr := c.consumer.ConsumeMetrics(ctx, metrics); consumeErr != nil {
			c.logger.Error("failed to forward metrics",
				zap.String("endpoint", c.target.ClientConfig.Endpoint),
				zap.Error(consumeErr))
		}
	}
}

func (c *gnmiClient) buildSubscribeRequest() (*gnmipb.SubscribeRequest, error) {
	subs := make([]*gnmipb.Subscription, 0, len(c.target.Subscriptions))
	for i := range c.target.Subscriptions {
		s := &c.target.Subscriptions[i]
		path, err := parseGNMIPath(s.Origin, s.Path)
		if err != nil {
			return nil, err
		}
		subs = append(subs, &gnmipb.Subscription{
			Path:              path,
			Mode:              subscriptionMode(s.Mode),
			SampleInterval:    uint64(s.SampleInterval.Nanoseconds()),    //nolint:gosec // disable G115: validated non-negative in Config.Validate
			HeartbeatInterval: uint64(s.HeartbeatInterval.Nanoseconds()), //nolint:gosec // disable G115: validated non-negative in Config.Validate
			SuppressRedundant: s.SuppressRedundant,
		})
	}
	return &gnmipb.SubscribeRequest{
		Request: &gnmipb.SubscribeRequest_Subscribe{
			Subscribe: &gnmipb.SubscriptionList{
				Subscription: subs,
				Mode:         gnmipb.SubscriptionList_STREAM,
				Encoding:     gnmiEncoding(c.target.Encoding),
			},
		},
	}, nil
}

func subscriptionMode(mode string) gnmipb.SubscriptionMode {
	switch mode {
	case modeOnChange:
		return gnmipb.SubscriptionMode_ON_CHANGE
	case modeTargetDefined:
		return gnmipb.SubscriptionMode_TARGET_DEFINED
	default:
		return gnmipb.SubscriptionMode_SAMPLE
	}
}

func gnmiEncoding(encoding string) gnmipb.Encoding {
	switch encoding {
	case encodingJSON:
		return gnmipb.Encoding_JSON
	case encodingJSONIETF:
		return gnmipb.Encoding_JSON_IETF
	default:
		return gnmipb.Encoding_PROTO
	}
}

func parseGNMIPath(origin, path string) (*gnmipb.Path, error) {
	parsed, err := ygot.StringToStructuredPath(path)
	if err != nil {
		return nil, fmt.Errorf("invalid gNMI path %q: %w", path, err)
	}
	parsed.Origin = origin
	return parsed, nil
}
