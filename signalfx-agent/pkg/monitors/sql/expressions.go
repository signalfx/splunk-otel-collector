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

package sql

import (
	"database/sql"
	"reflect"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

type ExprEnv map[string]interface{}

func newExprEnv(rowSlice []interface{}, columnNames []string) ExprEnv {
	out := make(map[string]interface{}, len(columnNames))

	values := scannedToValues(rowSlice)
	for i, name := range columnNames {
		// It is expected that the caller has the lengths matched properly or
		// else this will panic.
		out[name] = values[i]
	}

	return ExprEnv(out)
}

func scannedToValues(scanners []interface{}) []interface{} {
	vals := make([]interface{}, len(scanners))

	for i := range scanners {
		switch s := scanners[i].(type) {
		case *sql.NullString:
			if s.Valid {
				vals[i] = s.String
			}
		case *sql.NullFloat64:
			if s.Valid {
				vals[i] = s.Float64
			}
		case *sql.NullTime:
			if s.Valid {
				vals[i] = s.Time
			}
		case *sql.NullBool:
			if s.Valid {
				vals[i] = s.Bool
			}
		case *interface{}:
			if s == nil {
				continue
			}
			switch s2 := (*s).(type) {
			case []uint8:
				vals[i] = string(s2)
			case *[]uint8:
				vals[i] = string(*s2)
			}
		default:
			vals[i] = scanners[i]
		}
	}

	return vals
}

var floatType = reflect.TypeOf(float64(0))

func convertToFloatOrPanic(val interface{}) float64 {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	if !rVal.IsValid() {
		panic("value is null")
	}

	if !rVal.Type().ConvertibleTo(floatType) {
		// expr will recover from panics and return them as errors when
		// evaluating the expression.
		panic("value must be float")
	}

	return rVal.Convert(floatType).Float()
}

func (e ExprEnv) GAUGE(metric string, dims map[string]interface{}, val interface{}) *datapoint.Datapoint {

	return sfxclient.GaugeF(metric, utils.StringInterfaceMapToStringMap(dims), convertToFloatOrPanic(val))
}

func (e ExprEnv) CUMULATIVE(metric string, dims map[string]interface{}, val interface{}) *datapoint.Datapoint {
	return sfxclient.CumulativeF(metric, utils.StringInterfaceMapToStringMap(dims), convertToFloatOrPanic(val))
}
