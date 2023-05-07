package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

func main() {
	// Adjust the metrics as needed
	metrics := []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "fake_metric_total"},
				{Name: "instance", Value: "localhost:54090"},
			},
			Samples: []prompb.Sample{
				{Value: 42, Timestamp: time.Now().UnixNano() / int64(time.Millisecond)},
			},
		},
	}

	// Create a WriteRequest with the sample metrics
	req := &prompb.WriteRequest{
		Timeseries: metrics,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		panic(err)
	}

	compressed := snappy.Encode(nil, data)

	targetURL := os.Getenv("TARGET_URL")
	if targetURL == "" {
		targetURL = "http://otelcollector:54090/metrics"
	}

	// Continuously send fake metrics
	for {
		_, err := http.Post(targetURL, "application/x-protobuf", bytes.NewReader(compressed))
		if err != nil {
			fmt.Printf("Error sending metrics: %v\n", err)
		} else {
			fmt.Println("Metrics sent successfully")
		}

		// Adjust the interval between metric sends as needed
		time.Sleep(10 * time.Second)
	}
}
