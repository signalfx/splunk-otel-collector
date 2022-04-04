package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type MyErrorHandler struct {
}

func (m *MyErrorHandler) Handle(err error) {
	log.Fatal(err)
}

func main() {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("test-service"),
		),
	)
	if err != nil {
		log.Fatalf("Error creating resource: %v", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("otelcollector:4317"),
	)
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetErrorHandler(&MyErrorHandler{})
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	log.Println("Start tracing app")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		select {
		case <-c:
			os.Exit(0)
		}
	}()
	tracer := otel.Tracer("test-tracer")

	commonLabels := []attribute.KeyValue{
		attribute.String("request", "get"),
		attribute.String("response", "query"),
		attribute.String("company", "splunk"),
	}
	numberOfRuns := 0
	for {
		// work begins
		numberOfRuns++
		_, span := tracer.Start(
			context.Background(),
			"work",
			trace.WithAttributes(commonLabels...))
		log.Printf("Executing run %d\n", numberOfRuns)
		i := rand.Intn(3) * numberOfRuns % 30
		time.Sleep(time.Duration(i) * time.Second)
		span.End()
		tracerProvider.ForceFlush(context.Background())
	}
}
