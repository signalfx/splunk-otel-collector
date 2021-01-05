// Copyright 2021, OpenTelemetry Authors
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

package smartagentreceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"
)

func TestSmartAgentReceiver(t *testing.T) {
	type args struct {
		config       Config
		nextConsumer consumer.MetricsConsumer
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default",
			args: args{
				config: Config{
					ReceiverSettings: configmodels.ReceiverSettings{
						TypeVal: typeStr,
						NameVal: typeStr + "/default",
					},
				},
				nextConsumer: consumertest.NewMetricsNop(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer := tt.args.nextConsumer
			receiver := NewReceiver(zap.NewNop(), tt.args.config, consumer)
			err := receiver.Start(context.Background(), componenttest.NewNopHost())
			assert.NoError(t, err)
			err = receiver.Shutdown(context.Background())
			assert.NoError(t, err)
		})
	}
}
