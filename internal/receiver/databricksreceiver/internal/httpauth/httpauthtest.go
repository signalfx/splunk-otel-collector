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

package httpauth

import "net/http"

// FakeHandler implements http.Handler, and handles fake requests for testing,
// appending requests to an array member for later inspection
type FakeHandler struct {
	Reqs []*http.Request
}

func (h *FakeHandler) ServeHTTP(_ http.ResponseWriter, req *http.Request) {
	h.Reqs = append(h.Reqs, req)
}
