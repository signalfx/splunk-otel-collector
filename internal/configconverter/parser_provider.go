// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configconverter

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmapprovider"
)

// converterProvider wraps a configmapprovider.Provider and accepts a list of functions that
// convert ConfigMaps. The idea is for this type to conform to the open-closed
// principle.
type converterProvider struct {
	wrapped     configmapprovider.Provider
	retrieved   configmapprovider.Retrieved
	cfgMapFuncs []CfgMapFunc
}

type CfgMapFunc func(*config.Map) *config.Map

var _ configmapprovider.Provider = (*converterProvider)(nil)

func ParserProvider(wrapped configmapprovider.Provider, funcs ...CfgMapFunc) configmapprovider.Provider {
	return &converterProvider{wrapped: wrapped, cfgMapFuncs: funcs}
}

func (p converterProvider) Retrieve(ctx context.Context, onChange func(*configmapprovider.ChangeEvent)) (configmapprovider.Retrieved, error) {
	var err error
	p.retrieved, err = p.wrapped.Retrieve(ctx, onChange)
	return &p, err
}

func (p converterProvider) Shutdown(context.Context) error {
	return nil
}

func (p *converterProvider) Get(ctx context.Context) (*config.Map, error) {
	if p.retrieved == nil {
		return nil, fmt.Errorf("must Retrieve() before attempting Get()")
	}
	cfgMap, err := p.retrieved.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("converterProvider.Get(): %w", err)
	}

	for _, cfgMapFunc := range p.cfgMapFuncs {
		cfgMap = cfgMapFunc(cfgMap)
	}

	out := config.NewMap()
	for _, k := range cfgMap.AllKeys() {
		out.Set(k, cfgMap.Get(k))
	}
	return out, nil
}

func (p *converterProvider) Close(ctx context.Context) error {
	return nil
}
