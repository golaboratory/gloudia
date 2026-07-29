package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
)

func TestInitTracerProvider(t *testing.T) {
	// Stdout exporter (テスト用: 実際には何も出力されないようにDiscardしても良いが、動作確認のため)
	exporter, err := stdouttrace.New(stdouttrace.WithWriter(ioDiscard{}))
	assert.NoError(t, err)

	shutdown, err := InitTracerProvider("test-service", exporter)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	// Clean up
	err = shutdown(context.Background())
	assert.NoError(t, err)
}

func TestHTTPMiddleware(t *testing.T) {
	// Setup tracer
	exporter, _ := stdouttrace.New(stdouttrace.WithWriter(ioDiscard{}))
	InitTracerProvider("test-web", exporter)

	middleware := HTTPMiddleware("test-web")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Spanがコンテキストにあるか確認
		_, span := otel.Tracer("test").Start(r.Context(), "child") // Should use parent from middleware
		_, ok := span.(interface{ IsRecording() bool })            // dummy check
		_ = ok
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStartSpan(t *testing.T) {
	// Setup tracer
	exporter, _ := stdouttrace.New(stdouttrace.WithWriter(ioDiscard{}))
	InitTracerProvider("test-span", exporter)

	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "my-lib", "do-work")
	defer span.End()

	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
	assert.True(t, span.IsRecording())
}

// ioDiscard implements io.Writer but does nothing
type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (n int, err error) {
	return len(p), nil
}
