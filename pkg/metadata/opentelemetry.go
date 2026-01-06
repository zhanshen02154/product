package metadata

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

// GetTraceIdFromSpan 从Span获取TraceID
func GetTraceIdFromSpan(ctx context.Context) string {
	span := trace.SpanContextFromContext(ctx)
	if span.HasTraceID() {
		return span.TraceID().String()
	}
	return ""
}
