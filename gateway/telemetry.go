package gateway

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const serviceName = "dh-gateway"

func InitTelemetry() func() {
	if os.Getenv("OTEL_EXPORTER_JAEGER_AGENT_HOST") == "" {
		os.Setenv("OTEL_EXPORTER_JAEGER_AGENT_HOST", "localhost")
	}

	// Setting default exporter to jaeger
	exp, _ := jaeger.New(jaeger.WithAgentEndpoint())

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("environment", "prod"),
		)),
	)

	b3 := b3.New()

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		b3,
	))

	// Setting default trace provider
	otel.SetTracerProvider(tp)

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tp.ForceFlush(ctx)
	}
}
