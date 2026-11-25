// Copyright Splunk Inc.
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

// This file parses the XML rendered by the Windows Event Log API. It has no
// build constraint so its parsing logic can be unit tested on any platform.

package supportbundle

import (
	"encoding/xml"
	"errors"
	"fmt"
	"time"
)

type eventLogEntry struct {
	TimeGenerated time.Time
	Source        string
	Strings       []string
	EventID       uint32
	Level         uint16
}

type renderedEvent struct {
	System    renderedEventSystem `xml:"System"`
	EventData renderedEventData   `xml:"EventData"`
}

type renderedEventSystem struct {
	Provider    renderedEventProvider    `xml:"Provider"`
	TimeCreated renderedEventTimeCreated `xml:"TimeCreated"`
	EventID     uint32                   `xml:"EventID"`
	Level       uint16                   `xml:"Level"`
}

type renderedEventProvider struct {
	Name            string `xml:"Name,attr"`
	EventSourceName string `xml:"EventSourceName,attr"`
}

type renderedEventTimeCreated struct {
	SystemTime string `xml:"SystemTime,attr"`
}

type renderedEventData struct {
	Data []string `xml:"Data"`
}

func parseEventLogEvent(data []byte) (eventLogEntry, error) {
	var event renderedEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return eventLogEntry{}, fmt.Errorf("decode event XML: %w", err)
	}
	source := event.System.Provider.Name
	if source == "" {
		source = event.System.Provider.EventSourceName
	}
	if source == "" {
		return eventLogEntry{}, errors.New("event XML is missing the provider name or event source name")
	}

	timeGenerated, err := time.Parse(time.RFC3339Nano, event.System.TimeCreated.SystemTime)
	if err != nil {
		return eventLogEntry{}, fmt.Errorf("parse event time %q: %w", event.System.TimeCreated.SystemTime, err)
	}

	return eventLogEntry{
		TimeGenerated: timeGenerated,
		Source:        source,
		Strings:       event.EventData.Data,
		EventID:       event.System.EventID,
		Level:         event.System.Level,
	}, nil
}
