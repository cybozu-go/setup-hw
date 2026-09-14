package main

import (
	"context"
	"io"
	"os"
	"sort"
	"sync"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// memoryExporter keeps finished spans in memory for the SVG renderer.
type memoryExporter struct {
	mu    sync.Mutex
	spans []sdktrace.ReadOnlySpan
}

func (e *memoryExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.spans = append(e.spans, spans...)
	return nil
}

func (e *memoryExporter) Shutdown(context.Context) error { return nil }

func (e *memoryExporter) sorted() []sdktrace.ReadOnlySpan {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := append([]sdktrace.ReadOnlySpan(nil), e.spans...)
	sort.Slice(out, func(i, j int) bool { return out[i].StartTime().Before(out[j].StartTime()) })
	return out
}

// setupTracing returns a tracer, the in-memory exporter, and a shutdown function.
// If jsonPath is non-empty, spans are additionally written there as OTLP-style JSON (stdouttrace).
func setupTracing(jsonPath string) (trace.Tracer, *memoryExporter, func(context.Context) error, error) {
	mem := &memoryExporter{}
	res, _ := resource.Merge(resource.Default(), resource.NewWithAttributes(semconv.SchemaURL,
		semconv.ServiceName("redfish-trace")))
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(mem)),
	}
	var closer io.Closer
	if jsonPath != "" {
		f, err := os.Create(jsonPath)
		if err != nil {
			return nil, nil, nil, err
		}
		closer = f
		exp, err := stdouttrace.New(stdouttrace.WithWriter(f))
		if err != nil {
			return nil, nil, nil, err
		}
		opts = append(opts, sdktrace.WithBatcher(exp))
	}
	tp := sdktrace.NewTracerProvider(opts...)
	shutdown := func(ctx context.Context) error {
		err := tp.Shutdown(ctx)
		if closer != nil {
			closer.Close()
		}
		return err
	}
	return tp.Tracer("redfish-trace"), mem, shutdown, nil
}
