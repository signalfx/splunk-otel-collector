package splunk

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/signalfx/golib/v3/trace"

	"github.com/signalfx/golib/v3/event"
)

func TestSplunkEventMarshal(t *testing.T) {
	myEvent := logEvent{
		Time:       time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / int64(time.Millisecond),
		Host:       "localhost",
		Source:     "sfx",
		SourceType: "sfx",
		Index:      "sfx",
		Event: eventdata{
			Category:   event.USERDEFINED,
			EventType:  "Type",
			Meta:       map[string]string{"foo": "bar"},
			Dimensions: map[string]string{"foo": "bar"},
			Properties: map[string]string{"foo": "bar"},
		},
	}
	b, err := json.Marshal(myEvent)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\"time\":631152000000,\"host\":\"localhost\",\"source\":\"sfx\",\"sourcetype\":\"sfx\",\"index\":\"sfx\",\"event\":{\"category\":1000000,\"eventType\":\"Type\",\"meta\":{\"foo\":\"bar\"},\"dimensions\":{\"foo\":\"bar\"},\"properties\":{\"foo\":\"bar\"}}}"
	if string(b) != expected {
		t.Fatalf("JSON serialization does not match, expected: %s,\n got %s", expected, string(b))
	}
}

func TestSplunkMetricMarshal(t *testing.T) {
	myMetric := logMetric{
		Time:       time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / int64(time.Millisecond),
		Host:       "localhost",
		Source:     "sfx",
		SourceType: "sfx",
		Index:      "sfx",
		Event:      "metric",
		Fields:     map[string]string{"foo": "bar"},
	}
	b, err := json.Marshal(myMetric)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\"time\":631152000000,\"host\":\"localhost\",\"source\":\"sfx\",\"sourcetype\":\"sfx\",\"index\":\"sfx\",\"event\":\"metric\",\"fields\":{\"foo\":\"bar\"}}"
	if string(b) != expected {
		t.Fatalf("JSON serialization does not match, expected: %s,\n got %s", expected, string(b))
	}
}

func TestSplunkSpanMarshal(t *testing.T) {
	name := "foo"
	duration := int64(1000)
	kind := "dashboard"
	annotations := []*trace.Annotation{{
		Timestamp: &duration,
		Value:     &name,
	}}
	serviceName := "myservice"
	localhost := "127.0.0.1"
	myTrace := logSpan{
		Time:       time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / int64(time.Millisecond),
		Host:       "localhost",
		Source:     "sfx",
		SourceType: "sfx",
		Index:      "sfx",
		Event: spandata{
			Meta:        map[string]string{"foo": "bar"},
			Debug:       nil,
			Name:        &name,
			Duration:    &duration,
			Kind:        &kind,
			ParentID:    nil,
			Shared:      nil,
			Tags:        map[string]string{"foo": "bar"},
			Annotations: annotations,
			ID:          "myID",
			TraceID:     "myTraceID",
			LocalEndpoint: &trace.Endpoint{
				ServiceName: &serviceName,
				Ipv4:        &localhost,
				Ipv6:        nil,
				Port:        nil,
			},
			RemoteEndpoint: nil,
		},
	}
	b, err := json.Marshal(myTrace)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\"time\":631152000000,\"host\":\"localhost\",\"source\":\"sfx\",\"sourcetype\":\"sfx\",\"index\":\"sfx\",\"event\":{\"meta\":{\"foo\":\"bar\"},\"name\":\"foo\",\"duration\":1000,\"kind\":\"dashboard\",\"tags\":{\"foo\":\"bar\"},\"annotations\":[{\"timestamp\":1000,\"value\":\"foo\"}],\"id\":\"myID\",\"traceID\":\"myTraceID\",\"localEndpoint\":{\"serviceName\":\"myservice\",\"ipv4\":\"127.0.0.1\"}}}"
	if string(b) != expected {
		t.Fatalf("JSON serialization does not match, expected: %s,\n got %s", expected, string(b))
	}
}
