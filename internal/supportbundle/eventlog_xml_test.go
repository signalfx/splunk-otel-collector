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

package supportbundle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEventLogEvent(t *testing.T) {
	for _, tc := range []struct {
		name    string
		event   string
		wantErr string
		want    eventLogEntry
	}{
		{
			name: "event data",
			event: `<Event xmlns="http://schemas.microsoft.com/win/2004/08/events/event">
				<System>
					<Provider Name="splunk-otel-collector" />
					<EventID>70000</EventID>
					<Level>2</Level>
					<TimeCreated SystemTime="2023-11-14T22:13:20.1234567Z" />
				</System>
				<EventData>
					<Data>foo</Data>
					<Data Name="detail">bar &amp; 🧭</Data>
				</EventData>
			</Event>`,
			want: eventLogEntry{
				Source:        "splunk-otel-collector",
				EventID:       70000,
				Level:         2,
				TimeGenerated: time.Date(2023, time.November, 14, 22, 13, 20, 123456700, time.UTC),
				Strings:       []string{"foo", "bar & 🧭"},
			},
		},
		{
			name:  "without event data",
			event: `<Event><System><Provider Name="example"/><EventID>1</EventID><Level>4</Level><TimeCreated SystemTime="2023-11-14T22:13:20Z"/></System></Event>`,
			want: eventLogEntry{
				Source:        "example",
				EventID:       1,
				Level:         4,
				TimeGenerated: time.Date(2023, time.November, 14, 22, 13, 20, 0, time.UTC),
			},
		},
		{
			name:    "invalid XML",
			event:   `<Event><System>`,
			wantErr: "decode event XML",
		},
		{
			name:    "missing provider",
			event:   `<Event><System><EventID>1</EventID><TimeCreated SystemTime="2023-11-14T22:13:20Z"/></System></Event>`,
			wantErr: "missing the provider name or event source name",
		},
		{
			name:    "invalid time",
			event:   `<Event><System><Provider Name="example"/><TimeCreated SystemTime="not-a-time"/></System></Event>`,
			wantErr: "parse event time",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := parseEventLogEvent([]byte(tc.event))
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, entry)
		})
	}
}
