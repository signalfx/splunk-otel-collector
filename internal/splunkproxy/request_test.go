// Copyright Splunk, Inc.
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

// This code is copied from original work under this license:
// MIT License
//
//Copyright (c) 2019 Junyu Wang
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package splunkproxy

import (
	"encoding/json"
	"testing"
)

func TestParseRequestPreservesMatchedRouteAndBase64PayloadBytes(t *testing.T) {
	server := &Server{requestChan: make(chan Request, 1)}
	if err := server.parseRequest(&requestPacket{
		opcode: OPCODE_REQUEST_BLOCK,
		block:  `{"method":"POST","restmap":{"conf":{"match":"/v1/traces"}},"path_info":"","payload":"AAEC/w==","headers":[["Content-Type","application/x-protobuf"]]}`,
	}); err != nil {
		t.Fatal(err)
	}

	request := <-server.requestChan
	if request.Path != "/v1/traces" {
		t.Fatalf("request path = %q, want /v1/traces", request.Path)
	}
	want := []byte{0, 1, 2, 255}
	if string(request.Body) != string(want) {
		t.Fatalf("request body = %v, want %v", request.Body, want)
	}
	if got := request.Headers.Get("Content-Type"); got != "application/x-protobuf" {
		t.Fatalf("Content-Type = %q, want application/x-protobuf", got)
	}
}

func TestParseRequestRejectsInvalidBase64Payload(t *testing.T) {
	server := &Server{requestChan: make(chan Request, 1)}
	err := server.parseRequest(&requestPacket{
		opcode: OPCODE_REQUEST_BLOCK,
		block:  `{"payload":"not base64"}`,
	})
	if err == nil {
		t.Fatal("expected invalid base64 payload to fail")
	}
}

func TestResponseJSONPreservesBase64BodyBytesAndHeaders(t *testing.T) {
	response := Response{
		StatusCode: 429,
		Body:       []byte{0x0a, 0x03, 0x08, 0x96, 0x01},
		Headers:    map[string][]string{"Content-Type": {"application/x-protobuf"}, "Retry-After": {"2"}},
	}
	encoded, err := response.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Status        int                 `json:"status"`
		PayloadBase64 string              `json:"payload"`
		Headers       map[string][]string `json:"headers"`
	}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("decode response JSON: %v", err)
	}
	if got.Status != 429 || got.PayloadBase64 != "CgMIlgE=" || got.Headers["Content-Type"][0] != "application/x-protobuf" || got.Headers["Retry-After"][0] != "2" {
		t.Fatalf("response JSON = %s, want status, base64 payload, and headers preserved", encoded)
	}
}
