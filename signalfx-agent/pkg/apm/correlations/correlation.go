// Copyright  Splunk, Inc.
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

package correlations

import (
	"github.com/signalfx/signalfx-agent/pkg/apm/log"
)

// Type is the type of correlation
type Type string

const (
	// Service is for correlating services
	Service Type = "service"
	// Environment is for correlating environments
	Environment Type = "environment"
)

// Correlation is a struct referencing
type Correlation struct {
	// Type is the type of correlation
	Type Type
	// DimName is the dimension name
	DimName string
	// DimValue is the dimension value
	DimValue string
	// Value is the value to makeRequest with the DimName and DimValue
	Value string
}

func (c *Correlation) Logger(l log.Logger) log.Logger {
	return l.WithFields(log.Fields{
		"correlation.type":     c.Type,
		"correlation.dimName":  c.DimName,
		"correlation.dimValue": c.DimValue,
		"correlation.value":    c.Value,
	})
}
