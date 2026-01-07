package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// NewOTLPExporter は OTLP (OpenTelemetry Protocol) over HTTP を使用するExporterを作成します。
// Jaeger をはじめとする多くのバックエンドは現在この形式を標準としてサポートしています。
//
// 使用例 (JaegerをDockerで起動している場合):
//
//	exporter, err := telemetry.NewOTLPExporter(ctx, "localhost:4318", true)
//
// 引数:
//
//	endpoint: コレクタのエンドポイント (例: "localhost:4318")
//	insecure: trueの場合、TLS(SSL)を使用せずに接続します (ローカル開発用)
func NewOTLPExporter(ctx context.Context, endpoint string, insecure bool) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
	}

	if insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	// 開発環境などでは圧縮を無効化したい場合があるかもしれません
	// opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.NoCompression))

	return otlptracehttp.New(ctx, opts...)
}
