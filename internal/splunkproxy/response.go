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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

// Response represents the response sent back to the client
type Response struct {
	Headers    http.Header `json:"headers,omitempty"`
	Body       []byte      `json:"payload,omitempty"`
	StatusCode int         `json:"status"`
	isInit     bool
}

func (resp Response) MarshalJSON() ([]byte, error) {
	bodyAsBase64 := base64.StdEncoding.EncodeToString(resp.Body)

	payload := struct {
		Headers    http.Header `json:"headers,omitempty"`
		Body       string      `json:"payload,omitempty"`
		StatusCode int         `json:"status"`
	}{StatusCode: resp.StatusCode, Body: bodyAsBase64, Headers: resp.Headers}
	return json.Marshal(payload)
}

// getRawData transforms the response to a payload that splunkd can decode
// splunkd protocol for response to init packet: "0\n" (empty byte with length 0) to indicate success
// splunkd protocol for response to data packet: <len_response_bytes>\n<response>
func (resp Response) getRawData() string {
	if resp.isInit {
		return "0\n"
	}
	respData, err := json.Marshal(&resp)
	if err != nil {
		respData = []byte("Failed to serialize response data")
	}
	rawData := fmt.Sprintf("%d\n%s", len(respData), respData)
	return rawData
}
