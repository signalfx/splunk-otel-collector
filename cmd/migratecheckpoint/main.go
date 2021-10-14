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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

type Fingerprint struct {
	FirstBytes []byte
}

type Reader struct {
	Fingerprint *Fingerprint
	Offset      int64
}

type opType int

const (
	Get opType = iota
	Set
	Delete
)

type operation struct {
	// Key specifies key which is going to be get/set/deleted
	Key string
	// Value specifies value that is going to be set or holds result of get operation
	Value []byte
	// Type describes the operation type
	Type opType
}

type Operation *operation

func SetOperation(key string, value []byte) Operation {
	return &operation{
		Key:   key,
		Value: value,
		Type:  Set,
	}
}

var defaultBucket = []byte(`default`)

type fileStorageClient struct {
	db *bbolt.DB
}

func newClient(filePath string, timeout time.Duration) (*fileStorageClient, error) {
	options := &bbolt.Options{
		Timeout: timeout,
		NoSync:  true,
	}
	db, err := bbolt.Open(filePath, 0600, options)
	if err != nil {
		return nil, err
	}

	initBucket := func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(defaultBucket)
		return err
	}
	if err := db.Update(initBucket); err != nil {
		return nil, err
	}

	return &fileStorageClient{db}, nil
}

// Set will store data. The data can be retrieved using the same key
func (c *fileStorageClient) Set(key string, value []byte) error {
	return c.Batch(SetOperation(key, value))
}

// Batch executes the specified operations in order. Get operation results are updated in place
func (c *fileStorageClient) Batch(ops ...Operation) error {
	batch := func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(defaultBucket)
		if bucket == nil {
			return errors.New("storage not initialized")
		}

		var err error
		for _, op := range ops {
			switch op.Type {
			case Get:
				op.Value = bucket.Get([]byte(op.Key))
			case Set:
				err = bucket.Put([]byte(op.Key), op.Value)
			case Delete:
				err = bucket.Delete([]byte(op.Key))
			default:
				return errors.New("wrong operation type")
			}

			if err != nil {
				return err
			}
		}

		return nil
	}

	return c.db.Update(batch)
}

// Close will close the database
func (c *fileStorageClient) Close() error {
	return c.db.Close()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	containerLogPathFluentd := getEnv("CONTAINER_LOG_PATH_FLUENTD", "/var/log/splunk-fluentd-containers.log.pos")
	containerLogPathOtel := getEnv("CONTAINER_LOG_PATH_OTEL", "/var/lib/otel_pos/receiver_filelog_")

	customLogPathFluentd := getEnv("CUSTOM_LOG_PATH_FLUENTD", "/var/log/splunk-fluentd-*.pos")
	customLogPathOtel := getEnv("CUSTOM_LOG_PATH_OTEL", "/var/lib/otel_pos/receiver_filelog_")

	// Check whether it has already Otel's checkpoints
	_, err := os.Stat(containerLogPathOtel)
	if !os.IsNotExist(err) {
		// Otel checkpoint already present. no need to migrate.
		return
	}

	// Check whether there are fluentd's position file exist
	_, err = os.Stat(containerLogPathFluentd)
	if os.IsNotExist(err) {
		// Fluentd position file does not exist. no need to migrate.
		return
	}

	// Container File Logs
	lines, err := readLines(containerLogPathFluentd)
	if err != nil {
		fmt.Println("Error reading container fluentd's log path")
		panic(err)
	}

	var readers []*Reader
	var reader *Reader

	client, err := newClient(containerLogPathOtel, 100)
	if err != nil {
		fmt.Println(err)
	}

	for _, line := range lines {
		data := strings.Fields(line)
		filePath, hexPos := data[0], data[1]
		reader, err = convertToOtel(filePath, hexPos)
		if err != nil {
			continue
		}
		readers = append(readers, reader)

	}
	buf := syncLastPollFiles(readers)

	err = client.Set("$.file_input.knownFiles", buf.Bytes())
	if err != nil {
		fmt.Println(err)
	}
	client.Close()

	// Custom File Logs
	var myExp = regexp.MustCompile(`\/var\/log\/splunk\-fluentd\-(?P<name>[\w0-9-_]+)\.pos`)
	matches, _ := filepath.Glob(customLogPathFluentd)
	for _, match := range matches {
		captured := myExp.FindStringSubmatch(match)
		if len(captured) > 0 {
			lines, err = readLines(match)
			if err != nil {
				panic(err)
			}

			readers = []*Reader{}
			for _, line := range lines {
				data := strings.Fields(line)
				filePath, hexPos := data[0], data[1]
				reader, err = convertToOtel(filePath, hexPos)
				if err != nil {
					continue
				}
				readers = append(readers, reader)
			}
			if len(readers) > 0 {
				buf = syncLastPollFiles(readers)
				client, err = newClient(customLogPathOtel+captured[1], 100)
				if err != nil {
					fmt.Println(err)
				}
				err = client.Set("$.file_input.knownFiles", buf.Bytes())
				if err != nil {
					fmt.Println(err)
				}
				client.Close()
			}
		}
	}

	migrateJournaldPos()
	fmt.Println("Checkpoint migration completed")
}

func readLines(path string) ([]string, error) {
	File, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer File.Close()

	var lines []string
	scanner := bufio.NewScanner(File)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func convertToOtel(path string, hexPos string) (*Reader, error) {
	fp, err := getFingerPrint(path)
	if err != nil {
		return nil, err
	}

	offset, err := strconv.ParseInt(hexPos, 16, 64)
	if err != nil {
		return nil, err
	}

	reader := &Reader{
		Fingerprint: fp,
		Offset:      offset,
	}
	return reader, nil
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

type journaldCursor struct {
	Cursor string `json:"journal"`
}

func migrateJournaldPos() {
	journaldLogPathFluentd := getEnv("JOURNALD_LOG_PATH_FLUENTD", "splunkd-fluentd-journald-*.pos.json")
	journaldLogPathOtel := getEnv("JOURNALD_LOG_PATH_OTEL", "/var/lib/otel_pos/receiver_journald_")

	var myExp = regexp.MustCompile(`\/var\/log\/splunkd\-fluentd\-journald\-(?P<name>[\w0-9-_]+)\.pos\.json`)
	matches, _ := filepath.Glob(journaldLogPathFluentd)
	for _, match := range matches {
		captured := myExp.FindStringSubmatch(match)
		if len(captured) > 0 {
			jsonFile, err := os.Open(match)
			if err != nil {
				continue
			}
			byteValue, _ := ioutil.ReadAll(jsonFile)
			var cursor journaldCursor
			err = json.Unmarshal(byteValue, &cursor)
			if err != nil {
				continue
			}

			client, err := newClient(journaldLogPathOtel+captured[1], 100)
			if err != nil {
				fmt.Println(err)
			}
			err = client.Set("$.journald_input.lastReadCursor", []byte(cursor.Cursor))
			if err != nil {
				fmt.Println(err)
			}
			client.Close()
		}
	}
}
