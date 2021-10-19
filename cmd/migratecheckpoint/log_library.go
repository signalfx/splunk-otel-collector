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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// This code is from opentelemetry-log-collection/operator

type Fingerprint struct {
	FirstBytes []byte
}

type Reader struct {
	Fingerprint *Fingerprint
	Offset      int64
}

type journaldCursor struct {
	Cursor string `json:"journal"`
}

func getFingerPrint(path string) (*Fingerprint, error) {
	File, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer File.Close()

	buf := make([]byte, 1000)

	n, err := File.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("reading fingerprint bytes: %s", err)
	}

	fp := &Fingerprint{
		FirstBytes: buf[:n],
	}

	return fp, nil

}

func syncLastPollFiles(readers []*Reader) bytes.Buffer {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	// Encode the number of known files
	if err := enc.Encode(len(readers)); err != nil {
		fmt.Println(err)
	}

	// Encode each known file
	for _, fileReader := range readers {
		if err := enc.Encode(fileReader); err != nil {
			fmt.Println(err)
		}
	}
	return buf
}
