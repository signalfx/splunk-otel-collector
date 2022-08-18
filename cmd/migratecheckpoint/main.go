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
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

type Migrator struct {
	ContainerLogPathFluentd string
	ContainerLogPathOtel    string
	CustomLogPathFluentd    string
	CustomLogPathOtel       string
	CustomLogCaptureRegex   string
	JournaldLogPathFluentd  string
	JournaldLogPathOtel     string
	JournaldLogCaptureRegex string
}

func (m *Migrator) Run() {
	lines, err := readLines(m.ContainerLogPathFluentd)
	if err != nil {
		log.Println("Error reading container fluentd's log path")
		panic(err)
	}
	m.MigrateContainerPos(lines)

	matches, _ := filepath.Glob(m.CustomLogPathFluentd)
	m.MigrateCustomPos(matches)

	matches, _ = filepath.Glob(m.JournaldLogPathFluentd)
	m.MigrateJournaldPos(matches)
}

func (m *Migrator) ConvertFilePos(lines []string) ([]*Reader, bytes.Buffer) {
	var readers []*Reader
	var reader *Reader
	var err error

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
	return readers, buf
}

func (m *Migrator) MigrateContainerPos(lines []string) {
	client, err := newClient(m.ContainerLogPathOtel, 100)
	if err != nil {
		log.Fatalf("error creating a new DB client for container checkpoints: %v", err)
	}
	defer client.Close()
	_, buf := m.ConvertFilePos(lines)

	err = client.Set("$.file_input.knownFiles", buf.Bytes())
	if err != nil {
		log.Printf("error storing container checkpoints: %v", err)
	}
}

func (m *Migrator) MigrateCustomPos(matches []string) {
	var myExp = regexp.MustCompile(m.CustomLogCaptureRegex)

	for _, match := range matches {
		captured := myExp.FindStringSubmatch(match)
		if len(captured) > 0 {
			lines, err := readLines(match)
			if err != nil {
				panic(err)
			}

			readers, buf := m.ConvertFilePos(lines)

			if len(readers) > 0 {
				client, err := newClient(m.CustomLogPathOtel+captured[1], 100)
				if err != nil {
					log.Printf("error creating a new DB client for host file checkpoints: %v", err)
				}
				err = client.Set("$.file_input.knownFiles", buf.Bytes())
				if err != nil {
					log.Printf("error storing host file checkpoints: %v", err)
				}
				client.Close()
			}
		}
	}
}

func (m *Migrator) MigrateJournaldPos(matches []string) {
	var myExp = regexp.MustCompile(m.JournaldLogCaptureRegex)

	for _, match := range matches {
		captured := myExp.FindStringSubmatch(match)
		if len(captured) > 0 {
			jsonFile, err := os.Open(match)
			if err != nil {
				continue
			}
			byteValue, _ := io.ReadAll(jsonFile)
			var cursor journaldCursor
			err = json.Unmarshal(byteValue, &cursor)
			if err != nil {
				continue
			}

			client, err := newClient(m.JournaldLogPathOtel+captured[1], 100)
			if err != nil {
				log.Printf("error creating a new DB client for journald checkpoints: %v", err)
			}
			err = client.Set("$.journald_input.lastReadCursor", []byte(cursor.Cursor))
			if err != nil {
				log.Printf("error storing journald checkpoints: %v", err)
			}
			client.Close()
		}
	}
}

func main() {
	containerLogPathFluentd := getEnv("CONTAINER_LOG_PATH_FLUENTD", "/var/log/splunk-fluentd-containers.log.pos")
	containerLogPathOtel := getEnv("CONTAINER_LOG_PATH_OTEL", "/var/lib/otel_pos/receiver_filelog_")

	customLogPathFluentd := getEnv("CUSTOM_LOG_PATH_FLUENTD", "/var/log/splunk-fluentd-*.pos")
	customLogPathOtel := getEnv("CUSTOM_LOG_PATH_OTEL", "/var/lib/otel_pos/receiver_filelog_")
	customLogCaputreRegex := getEnv("CUSTOM_LOG_CAPTURE_REGEX", "\\/splunk\\-fluentd\\-(?P<name>[\\w0-9-_]+)\\.pos")

	journaldLogPathFluentd := getEnv("JOURNALD_LOG_PATH_FLUENTD", "splunkd-fluentd-journald-*.pos.json")
	journaldLogPathOtel := getEnv("JOURNALD_LOG_PATH_OTEL", "/var/lib/otel_pos/receiver_journald_")
	journaldLogCaptureRegex := getEnv("JOURNALD_LOG_CAPTURE_REGEX", "\\/splunkd\\-fluentd\\-journald\\-(?P<name>[\\w0-9-_]+)\\.pos\\.json")

	// Check whether it has already Otel's checkpoints
	_, err := os.Stat(containerLogPathOtel)
	if !os.IsNotExist(err) {
		log.Println("Otel checkpoint already present. no need to migrate.")
		return
	}

	// Check whether there are fluentd's position file exist
	_, err = os.Stat(containerLogPathFluentd)
	if os.IsNotExist(err) {
		log.Println("Fluentd position file does not exist. no need to perform migration.")
		return
	}

	migrator := &Migrator{
		ContainerLogPathFluentd: containerLogPathFluentd,
		ContainerLogPathOtel:    containerLogPathOtel,
		CustomLogPathFluentd:    customLogPathFluentd,
		CustomLogPathOtel:       customLogPathOtel,
		CustomLogCaptureRegex:   customLogCaputreRegex,
		JournaldLogPathFluentd:  journaldLogPathFluentd,
		JournaldLogPathOtel:     journaldLogPathOtel,
		JournaldLogCaptureRegex: journaldLogCaptureRegex,
	}

	migrator.Run()
	log.Println("Checkpoint migration completed")
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
