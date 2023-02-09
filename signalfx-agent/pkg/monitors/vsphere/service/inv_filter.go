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

package service

import (
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type InvFilter interface {
	keep(dims pairs) (bool, error)
}

func NewFilter(expression string) (InvFilter, error) {
	if expression == "" {
		return nopFilter{}, nil
	}
	f := filter{v: vm.VM{}}
	var err error
	f.program, err = expr.Compile(expression)
	if err != nil {
		return nopFilter{}, err
	}
	return f, nil
}

type nopFilter struct{}

func (n nopFilter) keep(pairs) (bool, error) {
	return true, nil
}

type filter struct {
	program *vm.Program
	v       vm.VM
}

type env struct {
	Datacenter string
	Cluster    string
}

func (f filter) keep(dims pairs) (bool, error) {
	e := env{}
	for _, dim := range dims {
		if dim[0] == dimDatacenter {
			e.Datacenter = dim[1]
		} else if dim[0] == dimCluster {
			e.Cluster = dim[1]
		}
	}
	result, err := f.v.Run(f.program, e)
	if err != nil {
		// default to keep if there's a failure
		return true, err
	}
	return result.(bool), nil
}
