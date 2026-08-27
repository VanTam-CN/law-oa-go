/**
 * OpenTelemetry分布式追踪系统 - 基于最新最佳实践
 * 提供企业级分布式追踪、链路监控和性能分析功能
 */

package tracing

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerConfig 追踪器配置
type TracerConfig struct {
	ServiceName        string            `json:"serviceName" yaml:"serviceName"`
	ServiceVersion     string            `json:"serviceVersion" yaml:"serviceVersion"`
	Environment        string            `json:"environment" yaml:"environment"`
	Enabled            bool              `json:"enabled" yaml:"enabled"`
	SamplingRate       float64           `json:"samplingRate" yaml:"samplingRate"` // 采样率 0.0-1.0
	Exporters          []string          `json:"exporters" yaml:"exporters"`       // 导出器: jaeger, zipkin, otlp
	JaegerEndpoint     string            `json:"jaegerEndpoint" yaml:"jaegerEndpoint"`
	ZipkinEndpoint     string            `json:"zipkinEndpoint" yaml:"zipkinEndpoint"`
	OTLPEndpoint       string            `json:"otlpEndpoint" yaml:"otlpEndpoint"`
	Headers            map[string]string `json:"headers" yaml:"headers"`
	Timeout            time.Duration     `json:"timeout" yaml:"timeout"`
	BatchTimeout       time.Duration     `json:"batchTimeout" yaml:"batchTimeout"`
	MaxExportBatchSize int               `json:"maxExportBatchSize" yaml:"maxExportBatchSize"`
	MaxQueueSize       int               `json:"maxQueueSize" yaml:"maxQueueSize"`
	ResourceAttributes map[string]string `json:"resourceAttributes" yaml:"resourceAttributes"`
}

// DefaultTracerConfig 返回默认追踪器配置
func DefaultTracerConfig() *TracerConfig {
	return &TracerConfig{
		ServiceName:        "law-oa-go",
		ServiceVersion:     "2.1.0",
		Environment:        "development",
		Enabled:            true,
		SamplingRate:       1.0, // 开发环境100%采样
		Exporters:          []string{"jaeger"},
		JaegerEndpoint:     "http://localhost:14268/api/traces",
		ZipkinEndpoint:     "http://localhost:9411/api/v2/spans",
		OTLPEndpoint:       "http://localhost:4318/v1/traces",
		Headers:            make(map[string]string),
		Timeout:            time.Second * 30,
		BatchTimeout:       time.Second * 5,
		MaxExportBatchSize: 512,
		MaxQueueSize:       2048,
		ResourceAttributes: map[string]string{
			"service.namespace":   "law-oa",
			"service.instance.id": getInstanceID(),
		},
	}
}

// ProductionTracerConfig 返回生产环境追踪器配置
func ProductionTracerConfig() *TracerConfig {
	return &TracerConfig{
		ServiceName:    "law-oa-go",
		ServiceVersion: "2.1.0",
		Environment:    "production",
		Enabled:        true,
		SamplingRate:   0.1, // 生产环境10%采样
		Exporters:      []string{"otlp", "jaeger"},
		JaegerEndpoint: "http://jaeger:14268/api/traces",
		ZipkinEndpoint: "http://zipkin:9411/api/v2/spans",
		OTLPEndpoint:   "http://otel-collector:4318/v1/traces",
		Headers: map[string]string{
			"authorization": "Bearer ${OTEL_TOKEN}",
		},
		Timeout:            time.Second * 30,
		BatchTimeout:       time.Second * 10,
		MaxExportBatchSize: 1024,
		MaxQueueSize:       4096,
		ResourceAttributes: map[string]string{
			"service.namespace":      "law-oa",
			"service.instance.id":    getInstanceID(),
			"deployment.environment": "production",
			"telemetry.sdk.name":     "opentelemetry",
			"telemetry.sdk.language": "go",
			"telemetry.sdk.version":  "1.28.0",
		},
	}
}

// TracerProvider 包装器
type TracerProvider struct {
	provider trace.TracerProvider
	tracer   trace.Tracer
	config   *TracerConfig
}

// Global tracer provider
var globalTracer *TracerProvider

// InitTracer 初始化分布式追踪器
func InitTracer(config *TracerConfig) error {
	if config == nil {
		config = DefaultTracerConfig()
	}

	if !config.Enabled {
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		return nil
	}

	// 创建资源
	res, err := createResource(config)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建导出器
	exporters, err := createExporters(config)
	if err != nil {
		return fmt.Errorf("failed to create exporters: %w", err)
	}

	// 创建采样器
	sampler := createSampler(config)

	// 创建追踪器提供者
	var providerOpts []sdktrace.TracerProviderOption

	// 为每个导出器添加批处理器选项
	for _, exporter := range exporters {
		batchOpts := []sdktrace.BatchSpanProcessorOption{
			sdktrace.WithBatchTimeout(config.BatchTimeout),
			sdktrace.WithMaxExportBatchSize(config.MaxExportBatchSize),
			sdktrace.WithMaxQueueSize(config.MaxQueueSize),
		}
		providerOpts = append(providerOpts, sdktrace.WithBatcher(exporter, batchOpts...))
	}

	// 添加资源和采样器
	providerOpts = append(providerOpts,
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	provider := sdktrace.NewTracerProvider(providerOpts...)

	// 设置全局追踪器提供者
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 创建全局追踪器
	globalTracer = &TracerProvider{
		provider: provider,
		tracer:   provider.Tracer(config.ServiceName),
		config:   config,
	}

	return nil
}

// createResource 创建资源
func createResource(config *TracerConfig) (*resource.Resource, error) {
	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String(config.ServiceName),
		semconv.ServiceVersionKey.String(config.ServiceVersion),
		semconv.DeploymentEnvironmentKey.String(config.Environment),
	}

	// 添加自定义资源属性
	for k, v := range config.ResourceAttributes {
		attributes = append(attributes, attribute.String(k, v))
	}

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attributes...),
	)
}

// createExporters 创建导出器
func createExporters(config *TracerConfig) ([]sdktrace.SpanExporter, error) {
	var exporters []sdktrace.SpanExporter

	for _, exporterName := range config.Exporters {
		switch exporterName {
		case "jaeger":
			jaegerExporter, err := createJaegerExporter(config)
			if err != nil {
				return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
			}
			exporters = append(exporters, jaegerExporter)

		case "zipkin":
			zipkinExporter, err := createZipkinExporter(config)
			if err != nil {
				return nil, fmt.Errorf("failed to create Zipkin exporter: %w", err)
			}
			exporters = append(exporters, zipkinExporter)

		case "otlp":
			otlpExporter, err := createOTLPExporter(config)
			if err != nil {
				return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
			}
			exporters = append(exporters, otlpExporter)

		default:
			return nil, fmt.Errorf("unsupported exporter: %s", exporterName)
		}
	}

	if len(exporters) == 0 {
		return nil, fmt.Errorf("no valid exporters configured")
	}

	return exporters, nil
}

// createJaegerExporter 创建Jaeger导出器
func createJaegerExporter(config *TracerConfig) (sdktrace.SpanExporter, error) {
	if config.JaegerEndpoint == "" {
		return nil, fmt.Errorf("Jaeger endpoint not configured")
	}

	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
}

// createZipkinExporter 创建Zipkin导出器
func createZipkinExporter(config *TracerConfig) (sdktrace.SpanExporter, error) {
	if config.ZipkinEndpoint == "" {
		return nil, fmt.Errorf("Zipkin endpoint not configured")
	}

	return zipkin.New(config.ZipkinEndpoint)
}

// createOTLPExporter 创建OTLP导出器
func createOTLPExporter(config *TracerConfig) (sdktrace.SpanExporter, error) {
	if config.OTLPEndpoint == "" {
		return nil, fmt.Errorf("OTLP endpoint not configured")
	}

	clientOptions := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.OTLPEndpoint),
		otlptracehttp.WithTimeout(config.Timeout),
	}

	// 添加头部信息
	if len(config.Headers) > 0 {
		headers := make(map[string]string)
		for k, v := range config.Headers {
			headers[k] = os.ExpandEnv(v) // 展开环境变量
		}
		clientOptions = append(clientOptions, otlptracehttp.WithHeaders(headers))
	}

	client := otlptracehttp.NewClient(clientOptions...)
	return otlptrace.New(context.Background(), client)
}

// createSampler 创建采样器
func createSampler(config *TracerConfig) sdktrace.Sampler {
	if config.SamplingRate <= 0 {
		return sdktrace.AlwaysSample()
	}
	if config.SamplingRate >= 1 {
		return sdktrace.AlwaysSample()
	}
	return sdktrace.TraceIDRatioBased(config.SamplingRate)
}

// GetGlobalTracer 获取全局追踪器
func GetGlobalTracer() trace.Tracer {
	if globalTracer != nil {
		return globalTracer.tracer
	}
	return otel.Tracer("default")
}

// GetGlobalTracerProvider 获取全局追踪器提供者
func GetGlobalTracerProvider() trace.TracerProvider {
	if globalTracer != nil {
		return globalTracer.provider
	}
	return otel.GetTracerProvider()
}

// ==============================
// 便捷的追踪函数
// ==============================

// StartSpan 开始一个新的span
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return GetGlobalTracer().Start(ctx, name, opts...)
}

// StartSpanWithAttributes 开始带属性的span
func StartSpanWithAttributes(ctx context.Context, name string, attrs []attribute.KeyValue, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	opts = append(opts, trace.WithAttributes(attrs...))
	return GetGlobalTracer().Start(ctx, name, opts...)
}

// AddSpanAttributes 添加span属性
func AddSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// AddSpanEvents 添加span事件
func AddSpanEvents(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetSpanStatus 设置span状态
func SetSpanStatus(span trace.Span, code codes.Code, message string) {
	span.SetStatus(code, message)
}

// RecordError 记录错误到span
func RecordError(span trace.Span, err error, opts ...trace.EventOption) {
	if err != nil {
		span.RecordError(err, opts...)
	}
}

// IsSampled 检查span是否被采样
func IsSampled(span trace.Span) bool {
	return span.SpanContext().IsSampled()
}

// GetTraceID 获取追踪ID
func GetTraceID(span trace.Span) string {
	return span.SpanContext().TraceID().String()
}

// GetSpanID 获取span ID
func GetSpanID(span trace.Span) string {
	return span.SpanContext().SpanID().String()
}

// ==============================
// 上下文工具函数
// ==============================

// ContextWithTraceID 向上下文添加追踪ID
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, "trace_id", traceID)
}

// ContextWithSpanID 向上下文添加span ID
func ContextWithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, "span_id", spanID)
}

// GetTraceIDFromContext 从上下文获取追踪ID
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	if span := trace.SpanFromContext(ctx); span != nil {
		return GetTraceID(span)
	}
	return ""
}

// GetSpanIDFromContext 从上下文获取span ID
func GetSpanIDFromContext(ctx context.Context) string {
	if spanID, ok := ctx.Value("span_id").(string); ok {
		return spanID
	}
	if span := trace.SpanFromContext(ctx); span != nil {
		return GetSpanID(span)
	}
	return ""
}

// ==============================
// 常用属性Key常量
// ==============================

const (
	// HTTP属性Key
	HTTPMethodKey     = "http.method"
	HTTPURLKey        = "http.url"
	HTTPStatusCodeKey = "http.status_code"
	HTTPTargetKey     = "http.target"
	HTTPHostKey       = "http.host"
	HTTPSchemeKey     = "http.scheme"
	HTTPUserAgentKey  = "http.user_agent"
	HTTPRemoteIPKey   = "http.remote_ip"

	// 数据库属性Key
	DBSystemKey              = "db.system"
	DBNameKey                = "db.name"
	DBStatementKey           = "db.statement"
	DBOperationKey           = "db.operation"
	DBRowsAffectedKey        = "db.rows_affected"
	DBConnectionStringKeyKey = "db.connection_string"

	// 业务属性Key
	BusinessOperationKey = "business.operation"
	BusinessModuleKey    = "business.module"
	BusinessUserIDKey    = "business.user_id"
	BusinessResourceKey  = "business.resource"

	// 系统属性Key
	SystemComponentKey = "system.component"
	SystemVersionKey   = "system.version"
	SystemInstanceKey  = "system.instance"
)

// 便捷的属性创建函数
var (
	// HTTP属性 - 已废弃，请直接使用 attribute.String(HTTPMethodKey, value)
	HTTPMethod     = HTTPMethodKey
	HTTPURL        = HTTPURLKey
	HTTPStatusCode = HTTPStatusCodeKey
	HTTPTarget     = HTTPTargetKey
	HTTPHost       = HTTPHostKey
	HTTPScheme     = HTTPSchemeKey
	HTTPUserAgent  = HTTPUserAgentKey
	HTTPRemoteIP   = HTTPRemoteIPKey

	// 数据库属性 - 已废弃，请直接使用 attribute.String(DBSystemKey, value)
	DBSystem              = DBSystemKey
	DBName                = DBNameKey
	DBStatement           = DBStatementKey
	DBOperation           = DBOperationKey
	DBRowsAffected        = DBRowsAffectedKey
	DBConnectionStringKey = DBConnectionStringKeyKey

	// 业务属性 - 已废弃，请直接使用 attribute.String(BusinessOperationKey, value)
	BusinessOperation = BusinessOperationKey
	BusinessModule    = BusinessModuleKey
	BusinessUserID    = BusinessUserIDKey
	BusinessResource  = BusinessResourceKey

	// 系统属性 - 已废弃，请直接使用 attribute.String(SystemComponentKey, value)
	SystemComponent = SystemComponentKey
	SystemVersion   = SystemVersionKey
	SystemInstance  = SystemInstanceKey
)

// ==============================
// 工具函数
// ==============================

// getInstanceID 获取实例ID
func getInstanceID() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		return "unknown"
	}
	return hostname
}

// Flush 刷新追踪器
func Flush(ctx context.Context) error {
	if globalTracer != nil {
		if provider, ok := globalTracer.provider.(*sdktrace.TracerProvider); ok {
			return provider.ForceFlush(ctx)
		}
	}
	return nil
}

// Shutdown 关闭追踪器
func Shutdown(ctx context.Context) error {
	if globalTracer != nil {
		if provider, ok := globalTracer.provider.(*sdktrace.TracerProvider); ok {
			return provider.Shutdown(ctx)
		}
	}
	return nil
}

// GetTracerConfig 获取当前追踪器配置
func GetTracerConfig() *TracerConfig {
	if globalTracer != nil {
		return globalTracer.config
	}
	return nil
}
