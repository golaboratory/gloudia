package telemetry

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// InitTracerProvider はOpenTelemetryのTracerProviderを初期化します。
// exporter を指定することで、JaegerやStdoutなど出力先を切り替えられます。
// 返り値の shutdown 関数はアプリケーション終了時に呼び出してください。
func InitTracerProvider(serviceName string, exporter sdktrace.SpanExporter) (func(context.Context) error, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// グローバルプロバイダとして登録
	otel.SetTracerProvider(tp)

	// W3C Trace Context 伝播フォーマットを設定 (分散トレーシング標準)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}

// HTTPMiddleware は受け取ったリクエストのTraceContextを抽出し、
// 新しいSpanを開始するHTTPミドルウェアです。
func HTTPMiddleware(serviceName string) func(http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ヘッダーからコンテキスト（TraceParentなど）を抽出
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Span開始
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ctx, span := tracer.Start(ctx, spanName)
			defer span.End()

			// Contextを更新して次へ
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// StartSpan はユーティリティとして、コンテキストから新しいSpanを開始します。
// tracerName は通常パッケージ名などを指定します。
func StartSpan(ctx context.Context, tracerName string, spanName string) (context.Context, trace.Span) {
	tracer := otel.Tracer(tracerName)
	return tracer.Start(ctx, spanName)
}
