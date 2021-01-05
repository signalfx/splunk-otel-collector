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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type Receiver struct {
	logger       *zap.Logger
	config       *Config
	nextConsumer consumer.MetricsConsumer
}

var _ component.MetricsReceiver = (*Receiver)(nil)

func NewReceiver(logger *zap.Logger, config Config, nextConsumer consumer.MetricsConsumer) *Receiver {
	return &Receiver{
		logger:       logger,
		config:       &config,
		nextConsumer: nextConsumer,
	}
}

func (r *Receiver) Start(_ context.Context, host component.Host) error {
	return nil
}

func (r *Receiver) Shutdown(context.Context) error {
	return nil
}
