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
// Copyright (c) 2019 Junyu Wang
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package splunkproxy implements the persistent script protocol that splunk core uses
// to communicate with app's persistent REST endpoint. This package handles basic routing and request/response
// handling.
package splunkproxy

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// Server represents the persistentconn server that handles request
// and writes response back to the client
type Server struct {
	requestChan       chan Request
	responseChan      chan Response
	responseQueue     *list.List
	registry          *handlerRegistry
	responseQueueLock *sync.Mutex
}

// NewServer creates a persistentconn server
func NewServer() *Server {
	return &Server{
		requestChan:       make(chan Request),
		responseChan:      make(chan Response),
		responseQueue:     list.New(),
		registry:          &handlerRegistry{},
		responseQueueLock: new(sync.Mutex),
	}
}

// Handle registers a handler function for a given path (or path pattern).
// A path pattern is in the format of "<component>/:<param_1>/<component>/..." and a path component
// starting with ":" indicates it's a parameter which will be inferred from the actual path in the request
// E.g. if the registered path pattern is "entity/:name/data" and the
// path in the request is "entity/hello/data", then the key-value pair {"name": "hello"} will be stored
// in the request's params which can be later referenced inside of the handler.
func (s *Server) Handle(path string, handler Handler, allowedMethods ...string) {
	s.registry.register(path, handler, allowedMethods)
}

// Run starts a persistentconn server and starts handling request sent from
// client (with splunkd as the middle layer)
func (s *Server) Run() error {
	go s.handleRequest()
	go s.processResponse()
	return s.startProcessingInputPackets(os.Stdin)
}

// startProcessingInputPackets starts a separate goroutine that reads request sent from client
// and is the entrypoint of a server process
func (s *Server) startProcessingInputPackets(input io.Reader) error {
	for {
		inPacket, err := readPacket(input)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read persistent connection packet: %w", err)
		}
		if err := s.parseRequest(inPacket); err != nil {
			return err
		}
	}
}

// handleRequest takes request that comes in and find the corresponding handler
func (s *Server) handleRequest() {
	for req := range s.requestChan {
		s.responseQueueLock.Lock()
		elem := s.responseQueue.PushBack(struct{}{})
		s.responseQueueLock.Unlock()

		// handle request in a goroutine
		go func(req Request, slot *list.Element) {
			var resp Response
			if req.isInit {
				resp = Response{isInit: true}
			} else {
				handler := s.registry.getHandler(req)
				handlerResponse, err := handler(req)
				if err != nil {
					handlerResponse.StatusCode = http.StatusInternalServerError
					handlerResponse.Body = []byte(err.Error())
				}
				resp = handlerResponse
			}
			slot.Value = resp
			s.responseChan <- resp
		}(req, elem)
	}
}

// processResponse proccesses response from handler and sent the response back to the client
func (s *Server) processResponse() {
	for range s.responseChan {
		_, _ = s.flushResponses(os.Stdout)
	}
}

// flushResponses go through responses in the response queue of the server, and it flushes consecutive
// responses starting from the front of the queue in batch to ensure that responses are synchronized in the same
// order as the corresponding requests.
func (s *Server) flushResponses(output io.Writer) (int, error) {
	s.responseQueueLock.Lock()
	defer s.responseQueueLock.Unlock()
	// prepare response data to flush to stdout
	elem := s.responseQueue.Front()
	flushedElList := make([]*list.Element, 0)

	writer := bufio.NewWriter(output)
	for ; elem != nil; elem = elem.Next() {
		resp, ok := elem.Value.(Response)
		if !ok {
			break
		}
		data := resp.getRawData()
		_, err := writer.WriteString(data)
		if err != nil {
			return 0, err
		}
		flushedEl := elem
		flushedElList = append(flushedElList, flushedEl)
	}
	err := writer.Flush()
	if err != nil {
		return 0, err
	}
	// clean up flushed element from the queue
	for _, flushedEl := range flushedElList {
		s.responseQueue.Remove(flushedEl)
	}
	return len(flushedElList), nil
}
