package tracerz

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/deferz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/ioz"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	runtimeTrace "runtime/trace"
)

var lock sync.Mutex
var appTracer trace.Tracer
var appName string

func StartDump(ctx context.Context) io.Closer {
	traceFile := os.Getenv("I6_GO_RUNTIME_TRACE_DUMP")
	if traceFile == "" {
		return ioz.CloserV(func() {})
	}

	filez.CreateParentDirs(traceFile)

	dfz := deferz.New(ctx)
	defer dfz.Close()

	f, err := os.OpenFile(traceFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	errorz.Check(err)
	dfz.AddCloserS(f)

	gz := gzip.NewWriter(f)
	dfz.AddCloserS(gz)

	err = runtimeTrace.Start(gz)
	errorz.Check(err)
	dfz.Add(runtimeTrace.Stop)

	return dfz.Detach()
}

func Init(name string) {
	if appTracer != nil {
		panic(fmt.Sprintf("tracer already initialized: %s", appName))
	}
	lock.Lock()
	defer lock.Unlock()
	if appTracer != nil {
		panic(fmt.Sprintf("tracer already initialized: %s", appName))
	}
	appName = name
	appTracer = otel.Tracer(name)
}

type keyType int

const key keyType = 1

func Span(ctx context.Context, firstName string, names ...string) (context.Context, io.Closer) {
	if appTracer == nil {
		return ctx, ioz.CloserV(func() {})
	}
	fullName := firstName
	for _, name := range names {
		name = blobz.New(name).Limited(64).Ascii()
		fullName = fmt.Sprintf("%s %s", fullName, name)
	}
	ctx, span := appTracer.Start(ctx, fullName)
	ctx = context.WithValue(ctx, key, span)
	return ctx, ioz.CloserV(func() {
		span.End()
	})
}

func AttrString(ctx context.Context, name string, value string) {
	AttrAscii(ctx, name, []byte(value))
}

func get(ctx context.Context) trace.Span {
	v := ctx.Value(key)
	if v == nil {
		return nil
	}
	return v.(trace.Span)
}

func AttrAscii(ctx context.Context, name string, value []byte) {
	span := get(ctx)
	if span != nil {
		span.SetAttributes(attribute.String(fix(name), blobz.New(value).Limited(4096).Ascii()))
	}
}

func Error(ctx context.Context, msg string) {
	span := get(ctx)
	if span != nil {
		span.SetStatus(codes.Error, msg)
	}
}

func Ok(ctx context.Context, msg string) {
	span := get(ctx)
	if span != nil {
		span.SetStatus(codes.Ok, msg)
	}
}

func AttrInt(ctx context.Context, name string, value int) {
	span := get(ctx)
	if span != nil {
		span.SetAttributes(attribute.Int(fix(name), value))
	}
}

func fix(name string) string {
	return fmt.Sprintf("i6.%s", name)
}

func AttrInt64(ctx context.Context, name string, value int64) {
	span := get(ctx)
	if span != nil {
		span.SetAttributes(attribute.Int64(fix(name), value))
	}
}

func AttrBool(ctx context.Context, name string, value bool) {
	span := get(ctx)
	if span != nil {
		span.SetAttributes(attribute.Bool(fix(name), value))
	}
}

func Inject(ctx context.Context, headers http.Header) {
	tracerPropagator := otel.GetTextMapPropagator()
	tracerPropagator.Inject(ctx, propagation.HeaderCarrier(headers))
}
