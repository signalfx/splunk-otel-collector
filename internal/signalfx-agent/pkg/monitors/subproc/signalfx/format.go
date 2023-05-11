package signalfx

import (
	"fmt"

	"github.com/signalfx/golib/v3/trace"
)

// JSONDatapointV1 is the JSON API format for /v1/datapoint
//
//easyjson:json
type JSONDatapointV1 struct {
	//easyjson:json
	Source string  `json:"source"`
	Metric string  `json:"metric"`
	Value  float64 `json:"value"`
}

// JSONDatapointV2 is the V2 json datapoint sending format
//
//easyjson:json
type JSONDatapointV2 map[string][]*BodySendFormatV2

// BodySendFormatV2 is the JSON format signalfx datapoints are expected to be in
//
//easyjson:json
type BodySendFormatV2 struct {
	Metric     string            `json:"metric"`
	Timestamp  int64             `json:"timestamp"`
	Value      ValueToSend       `json:"value"`
	Dimensions map[string]string `json:"dimensions"`
}

func (bodySendFormat *BodySendFormatV2) String() string {
	return fmt.Sprintf("DP[metric=%s|time=%d|val=%s|dimensions=%s]", bodySendFormat.Metric, bodySendFormat.Timestamp, bodySendFormat.Value, bodySendFormat.Dimensions)
}

// ValueToSend are values are sent from the gateway to a receiver for the datapoint
type ValueToSend interface{}

// JSONEventV2 is the V2 json event sending format
//
//easyjson:json
type JSONEventV2 []*EventSendFormatV2

// EventSendFormatV2 is the JSON format signalfx datapoints are expected to be in
//
//easyjson:json
type EventSendFormatV2 struct {
	EventType  string                 `json:"eventType"`
	Category   *string                `json:"category"`
	Dimensions map[string]string      `json:"dimensions"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  *int64                 `json:"timestamp"`
}

// InputAnnotation associates an event that explains latency with a timestamp.
// Unlike log statements, annotations are often codes. Ex. “ws” for WireSend
//
//easyjson:json
type InputAnnotation struct {
	Endpoint  *trace.Endpoint `json:"endpoint"`
	Timestamp *float64        `json:"timestamp"`
	Value     *string         `json:"value"`
}

// ToV2 converts an InputAnnotation to a V2 InputAnnotation, which basically
// means dropping the endpoint.  The endpoint must be considered in other
// logic to know which span to associate the endpoint with.
func (a *InputAnnotation) ToV2() *trace.Annotation {
	return &trace.Annotation{
		Timestamp: GetPointerToInt64(a.Timestamp),
		Value:     a.Value,
	}
}

// GetpointerToInt64 does that
func GetPointerToInt64(p *float64) *int64 {
	if p == nil {
		return nil
	}
	i := int64(*p)
	return &i
}

// BinaryAnnotation associates an event that explains latency with a timestamp.
//
//easyjson:json
type BinaryAnnotation struct {
	Endpoint *trace.Endpoint `json:"endpoint"`
	Key      *string         `json:"key"`
	Value    *interface{}    `json:"value"`
}

// InputSpan defines a span that is the union of v1 and v2 spans
//
//easyjson:json
type InputSpan struct {
	trace.Span
	Timestamp         *float64            `json:"timestamp"`
	Duration          *float64            `json:"duration"`
	Annotations       []*InputAnnotation  `json:"annotations"`
	BinaryAnnotations []*BinaryAnnotation `json:"binaryAnnotations"`
}

// InputSpanList is an array of InputSpan pointers
//
//easyjson:json
type InputSpanList []*InputSpan
