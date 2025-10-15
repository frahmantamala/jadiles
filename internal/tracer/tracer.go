package tracer

import (
	"context"

	"github.com/frahmantamala/jadiles/internal"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const ApplicationTraceIDKey = "application-trace-id"

func StartSpan(ctx context.Context, resourceName, operationName string) (ddtrace.Span, context.Context) {
	return tracer.StartSpanFromContext(
		ctx,
		operationName,
		tracer.ResourceName(resourceName),
		tracer.Tag(ApplicationTraceIDKey, internal.ExtractTraceID(ctx)),
	)
}

func FinishSpan(span ddtrace.Span, err error) {
	span.Finish(tracer.WithError(err))
}
