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

// splunkdRequest represents the request sent from splunkd
type splunkdRequest struct {
	Restmap struct {
		Name string `json:"name"`
		Conf struct {
			Handler         string `json:"handler"`
			Match           string `json:"match"`
			OutputModes     string `json:"output_modes"`
			PassHTTPHeaders string `json:"passHttpHeaders"`
			PassPayload     string `json:"passPayload"`
			Script          string `json:"script"`
			Scripttype      string `json:"scripttype"`
		} `json:"conf"`
	} `json:"restmap"`
	Server struct {
		RestURI    string `json:"rest_uri"`
		Hostname   string `json:"hostname"`
		Servername string `json:"servername"`
		GUID       string `json:"guid"`
	} `json:"server"`
	Session struct {
		User      string `json:"user"`
		Authtoken string `json:"authtoken"`
	} `json:"session"`
	Ns struct {
		App  string `json:"app"`
		User string `json:"user"`
	} `json:"ns"`
	Method     string     `json:"method"`
	PathInfo   string     `json:"path_info"`
	RestPath   string     `json:"rest_path"`
	OutputMode string     `json:"output_mode"`
	Payload    string     `json:"payload,omitempty"`
	Query      [][]string `json:"query"`
	Headers    [][]string `json:"headers"`
	Form       [][]string `json:"form,omitempty"`
	Connection struct {
		SrcIP         string `json:"src_ip"`
		Ssl           bool   `json:"ssl"`
		ListeningPort int    `json:"listening_port"`
	} `json:"connection"`
	OutputModeExplicit bool `json:"output_mode_explicit"`
}

// Request contains information of an incoming request
type Request struct {
	Headers    http.Header
	Query      map[string]string
	Form       map[string]string
	Params     map[string]string
	OutputMode string
	Method     string
	Body       []byte
	Path       string
	isInit     bool
}

// parseRequests creates a Request object by parsing information from a request packet.
func (s *Server) parseRequest(p *requestPacket) error {
	if p.isFirst() {
		request := Request{isInit: true}
		s.requestChan <- request
	}
	if p.hasBlock() {
		block := p.block
		var splunkdReq splunkdRequest
		if err := json.Unmarshal([]byte(block), &splunkdReq); err != nil {
			return fmt.Errorf("decode persistent connection request: %w", err)
		}
		body, err := base64.StdEncoding.DecodeString(splunkdReq.Payload)
		if err != nil {
			return fmt.Errorf("decode base64 request payload: %w", err)
		}
		request := Request{
			OutputMode: splunkdReq.OutputMode,
			Headers:    tupleListToHeader(splunkdReq.Headers),
			Method:     splunkdReq.Method,
			Query:      tupleListToMap(splunkdReq.Query),
			Form:       tupleListToMap(splunkdReq.Form),
			Body:       body,
			Path:       splunkdReq.Restmap.Conf.Match,
			Params:     make(map[string]string),
		}
		s.requestChan <- request
	}
	return nil
}
