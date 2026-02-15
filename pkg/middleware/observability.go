package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "datalens_http_request_duration_seconds",
		Help:    "Duration of HTTP requests.",
		Buckets: prometheus.DefBuckets,
	}, []string{"path", "method", "status"})

	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "datalens_http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"path", "method", "status"})
)

// ObservabilityMiddleware combines Prometheus metrics and OpenTelemetry tracing
func ObservabilityMiddleware(serviceName string) func(next http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 1. Tracing
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			opts := []oteltrace.SpanStartOption{
				oteltrace.WithAttributes(semconv.HTTPMethod(r.Method)),
				oteltrace.WithAttributes(semconv.HTTPRoute(r.URL.Path)),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			// Wrap ResponseWriter to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// 2. Serve Request
			next.ServeHTTP(ww, r.WithContext(ctx))

			// 3. Record Metrics
			duration := time.Since(start).Seconds()
			status := fmt.Sprintf("%d", ww.Status())
			path := r.URL.Path // Note: High cardinality if paths have IDs, but manageable for now or use chi route context if available

			// Attempt to use Chi's RouteContext to get the route pattern instead of the raw path to avoid high cardinality
			if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
				if routeContext.RoutePattern() != "" {
					path = routeContext.RoutePattern()
				}
			}

			httpDuration.WithLabelValues(path, r.Method, status).Observe(duration)
			httpRequestsTotal.WithLabelValues(path, r.Method, status).Inc()

			// Update Span
			span.SetAttributes(semconv.HTTPStatusCode(ww.Status()))
		})
	}
}
