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

package splunkproxy

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

// Handler is a handler function that takes a persistentconn request and returns a response or error
// if error is returned, the server will return a 500 (Internal Server Error) response with the returned
// error's message as the response body
type Handler func(Request) (Response, error)

// noMatchingHandler is the default handler returned when no matching path
// is found from a request
func noMatchingHandler(_ Request) (Response, error) {
	return Response{
		StatusCode: http.StatusNotFound,
		Body:       []byte("The requested path is not found."),
	}, nil
}

// route represents a registered route that has a corresponding handler
type route struct {
	Pattern *regexp.Regexp
	Handler Handler
	Methods []string
}

// newRoute creates a new route object
func newRoute(pathPattern string, handler Handler, allowedMethods []string) *route {
	re := translatePatternToRegexp(pathPattern)
	return &route{
		Pattern: re,
		Handler: handler,
		Methods: allowedMethods,
	}
}

// translatePatternToRegexp translates a path pattern to a regexp
func translatePatternToRegexp(pathPattern string) *regexp.Regexp {
	parts := strings.Split(pathPattern, "/")
	regexpStrParts := make([]string, len(parts))
	for idx, p := range parts {
		if strings.HasPrefix(p, ":") {
			p = fmt.Sprintf(`(?P<%s>[\S|^\/]+)`, p[1:])
		}
		regexpStrParts[idx] = p
	}
	regexpStr := strings.Join(regexpStrParts, "/")
	re := regexp.MustCompile(regexpStr)
	return re
}

// handlerRegistry is where all routes are stored
type handlerRegistry struct {
	routes []*route
}

// getHandler gets the handler based on the input request's path info
func (rg *handlerRegistry) getHandler(req Request) Handler {
	handler := noMatchingHandler
	for _, rt := range rg.routes {
		if matches := rt.Pattern.FindStringSubmatch(req.Path); len(matches) > 0 && slices.Contains(rt.Methods, req.Method) {
			matchGroupNames := rt.Pattern.SubexpNames()
			for idx, name := range matchGroupNames {
				// Since the Regexp as a whole cannot be named, first matched name is always the empty string
				if name != "" {
					req.Params[name] = matches[idx]
				}
			}
			return rt.Handler
		}
	}
	return handler
}

// register func registers a path with a handler
func (rg *handlerRegistry) register(path string, handler Handler, allowedMethods []string) {
	route := newRoute(path, handler, allowedMethods)
	rg.routes = append(rg.routes, route)
}
