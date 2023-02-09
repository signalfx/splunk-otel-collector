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
	"io"
	"regexp"

	"github.com/sirupsen/logrus"
)

type LogWriterCloser struct {
	log logrus.FieldLogger
}

func NewLogWriterCloser(log logrus.FieldLogger) *LogWriterCloser {
	return &LogWriterCloser{log: log}
}

func (lwc *LogWriterCloser) Write(p []byte) (n int, err error) {
	lwc.log.Info(string(Scrub(p)))
	return len(p), nil
}

func (lwc *LogWriterCloser) Close() error {
	return nil
}

type LogProvider struct {
	log logrus.FieldLogger
}

func (s *LogProvider) NewFile(p string) io.WriteCloser {
	return NewLogWriterCloser(s.log)
}

func (s *LogProvider) Flush() {
}

var scrubPassword = regexp.MustCompile(`<password>(.*)</password>`)

func Scrub(in []byte) []byte {
	return scrubPassword.ReplaceAll(in, []byte(`<password>********</password>`))
}
