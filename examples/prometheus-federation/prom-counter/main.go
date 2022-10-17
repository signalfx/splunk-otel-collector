// Copyright OpenTelemetry Authors
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

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.uber.org/zap"
)

func initMeter() {
	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{})
	if err != nil {
		log.Panicf("failed to initialize prometheus exporter %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", exporter.ServeHTTP)
	server := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           mux,
		Addr:              ":8080",
	}
	go server.ListenAndServe()
}

func main() {
	// set up prometheus
	initMeter()
	// logging
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger.Info("Start Prometheus metrics app")
	meter := global.Meter("counter")
	valueRecorder := metric.Must(meter).NewInt64ValueRecorder("prom_counter")
	ctx := context.Background()
	valueRecorder.Measurement(0)
	commonLabels := []attribute.KeyValue{attribute.String("A", "1"), attribute.String("B", "2"), attribute.String("C", "3")}
	counter := int64(0)
	meter.RecordBatch(ctx,
		commonLabels,
		valueRecorder.Measurement(counter))
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			counter++
			meter.RecordBatch(ctx,
				commonLabels,
				valueRecorder.Measurement(counter))
			break
		case <-c:
			ticker.Stop()
			logger.Info("Stop Prometheus metrics app")
			return
		}
	}
}
