/**
 * OpenTelemetry分布式追踪系统 - 基于最新最佳实践
 * 提供企业级分布式追踪、链路监控和性能分析功能
 */

package tracing

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// TracerConfig 追踪器配置
type TracerConfig struct {
	ServiceName    string   `json:"serviceName" yaml:"serviceName"`
	ServiceVersion string   `json:"serviceVersion" yaml:"serviceVersion"`
	Environment    string   `json:"environment" yaml:"environment"`
	Enabled        bool     `json:"enabled" yaml:"enabled"`
	SamplingRate   float64  `json:"samplingRate" yaml:"samplingRate"` // 采样率 0.0-1.0
	Exporters      []string `json:"exporters" yaml:"exporters"`       // 导出器: otlp(默认), otel(同 otlp), jaeger(兼容别名→OTLP), zipkin, noop
	// JaegerEndpoint 旧版 Jaeger collector 端点，仅用于兼容旧配置：
	// exporters 配置为 "jaeger" 时映射为 OTLP 端点（collector 14268 端口 → OTLP 4318）。
	// Deprecated: 请改用 OTLPEndpoint 或标准 OTEL_EXPORTER_OTLP_* 环境变量。
	JaegerEndpoint     string            `json:"jaegerEndpoint" yaml:"jaegerEndpoint"`
	ZipkinEndpoint     string            `json:"zipkinEndpoint" yaml:"zipkinEndpoint"`
	OTLPEndpoint       string            `json:"otlpEndpoint" yaml:"otlpEndpoint"`
	OTLPProtocol       string            `json:"otlpProtocol" yaml:"otlpProtocol"` // OTLP 传输协议: http(默认) 或 grpc
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
		Exporters:          []string{"otlp"},
		JaegerEndpoint:     "http://localhost:14268/api/traces", // Deprecated: 仅兼容旧配置
		ZipkinEndpoint:     "http://localhost:9411/api/v2/spans",
		OTLPEndpoint:       "http://localhost:4318/v1/traces",
		OTLPProtocol:       "http",
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
		Exporters:      []string{"otlp"},
		JaegerEndpoint: "http://jaeger:14268/api/traces", // Deprecated: 仅兼容旧配置
		ZipkinEndpoint: "http://zipkin:9411/api/v2/spans",
		OTLPEndpoint:   "http://otel-collector:4318/v1/traces",
		OTLPProtocol:   "http",
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

// noopSpanExporter 显式空导出器（exporters 配置 noop/none 时使用）
type noopSpanExporter struct{}

// Global tracer provider
var globalTracer *TracerProvider

// InitTracer 初始化分布式追踪器
func InitTracer(config *TracerConfig) error {
	if config == nil {
		config = DefaultTracerConfig()
	}

	if !config.Enabled {
		otel.SetTracerProvider(noop.NewTracerProvider())
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

// createExporters 创建导出器。
// 配置兼容：otlp/otel 走 OTLP；jaeger 为已停止维护的旧导出器，映射到 OTLP
// （Jaeger >=1.35 原生接收 OTLP）；zipkin 保持不变；noop/none 为显式空导出。
func createExporters(config *TracerConfig) ([]sdktrace.SpanExporter, error) {
	var exporters []sdktrace.SpanExporter

	for _, exporterName := range config.Exporters {
		switch strings.ToLower(strings.TrimSpace(exporterName)) {
		case "otlp", "otel":
			otlpExporter, err := createOTLPExporter(config, "")
			if err != nil {
				return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
			}
			exporters = append(exporters, otlpExporter)

		case "jaeger":
			log.Printf("tracing: exporter %q is deprecated and now exports via OTLP; set exporters: [otlp] to silence this warning", exporterName)
			otlpExporter, err := createOTLPExporter(config, legacyJaegerOTLPEndpoint(config))
			if err != nil {
				return nil, fmt.Errorf("failed to create OTLP exporter for legacy jaeger alias: %w", err)
			}
			exporters = append(exporters, otlpExporter)

		case "zipkin":
			zipkinExporter, err := createZipkinExporter(config)
			if err != nil {
				return nil, fmt.Errorf("failed to create Zipkin exporter: %w", err)
			}
			exporters = append(exporters, zipkinExporter)

		case "noop", "none":
			exporters = append(exporters, noopSpanExporter{})

		case "":
			// 空白项忽略，便于配置中留空
			continue

		default:
			return nil, fmt.Errorf("unsupported exporter: %s", exporterName)
		}
	}

	if len(exporters) == 0 {
		return nil, fmt.Errorf("no valid exporters configured")
	}

	return exporters, nil
}

// createZipkinExporter 创建Zipkin导出器
func createZipkinExporter(config *TracerConfig) (sdktrace.SpanExporter, error) {
	if config.ZipkinEndpoint == "" {
		return nil, fmt.Errorf("zipkin endpoint not configured")
	}

	return zipkin.New(config.ZipkinEndpoint)
}

// createOTLPExporter 创建OTLP导出器（otlptracehttp 或 otlptracegrpc）。
// legacyEndpoint 仅在旧 jaeger 配置兼容路径使用：OTLPEndpoint 与标准环境变量均未配置时兜底。
func createOTLPExporter(config *TracerConfig, legacyEndpoint string) (sdktrace.SpanExporter, error) {
	endpoint, err := resolveOTLPEndpoint(config, legacyEndpoint)
	if err != nil {
		return nil, err
	}

	protocol, err := resolveOTLPProtocol(config)
	if err != nil {
		return nil, err
	}

	headers := expandHeaders(config.Headers)

	switch protocol {
	case "grpc":
		clientOptions := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpointURL(endpoint),
		}
		if config.Timeout > 0 {
			clientOptions = append(clientOptions, otlptracegrpc.WithTimeout(config.Timeout))
		}
		if len(headers) > 0 {
			clientOptions = append(clientOptions, otlptracegrpc.WithHeaders(headers))
		}
		return otlptrace.New(context.Background(), otlptracegrpc.NewClient(clientOptions...))

	case "http":
		clientOptions := []otlptracehttp.Option{
			otlptracehttp.WithEndpointURL(endpoint),
		}
		if config.Timeout > 0 {
			clientOptions = append(clientOptions, otlptracehttp.WithTimeout(config.Timeout))
		}
		if len(headers) > 0 {
			clientOptions = append(clientOptions, otlptracehttp.WithHeaders(headers))
		}
		return otlptrace.New(context.Background(), otlptracehttp.NewClient(clientOptions...))

	default:
		return nil, fmt.Errorf("unsupported OTLP protocol: %s", protocol)
	}
}

// resolveOTLPEndpoint 解析 OTLP 端点：显式配置 > 标准环境变量 > 旧 jaeger 端点兜底。
func resolveOTLPEndpoint(config *TracerConfig, legacyEndpoint string) (string, error) {
	if endpoint := strings.TrimSpace(config.OTLPEndpoint); endpoint != "" {
		return normalizeOTLPEndpoint(endpoint)
	}
	if endpoint := firstEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		return normalizeOTLPEndpoint(endpoint)
	}
	if endpoint := strings.TrimSpace(legacyEndpoint); endpoint != "" {
		return normalizeOTLPEndpoint(endpoint)
	}
	return "", fmt.Errorf("OTLP endpoint not configured: set otlpEndpoint or OTEL_EXPORTER_OTLP_TRACES_ENDPOINT/OTEL_EXPORTER_OTLP_ENDPOINT")
}

// resolveOTLPProtocol 解析 OTLP 传输协议：显式配置 > 标准环境变量 > http。
func resolveOTLPProtocol(config *TracerConfig) (string, error) {
	protocol := strings.ToLower(strings.TrimSpace(config.OTLPProtocol))
	if protocol == "" {
		protocol = strings.ToLower(firstEnv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "OTEL_EXPORTER_OTLP_PROTOCOL"))
	}

	switch protocol {
	case "":
		return "http", nil
	case "http", "http/protobuf":
		return "http", nil
	case "grpc":
		return "grpc", nil
	default:
		return "", fmt.Errorf("unsupported OTLP protocol %q: want http or grpc", protocol)
	}
}

// legacyJaegerOTLPEndpoint 将旧 Jaeger collector 端点映射为同主机的 OTLP/HTTP 端点。
// Jaeger all-in-one/collector 的 14268 端口只提供 Jaeger 协议，OTLP/HTTP 默认端口为 4318。
func legacyJaegerOTLPEndpoint(config *TracerConfig) string {
	raw := strings.TrimSpace(config.JaegerEndpoint)
	if raw == "" {
		return ""
	}

	u, err := parseEndpointURL(raw)
	if err != nil {
		log.Printf("tracing: ignore invalid jaegerEndpoint %q: %v", raw, err)
		return ""
	}

	if u.Port() == "" || u.Port() == "14268" {
		u.Host = net.JoinHostPort(u.Hostname(), "4318")
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""

	return u.String()
}

// normalizeOTLPEndpoint 校验并规范化 OTLP 端点 URL（缺省协议补 http://）。
func normalizeOTLPEndpoint(raw string) (string, error) {
	u, err := parseEndpointURL(raw)
	if err != nil {
		return "", err
	}
	if u.Path == "/" {
		u.Path = ""
	}
	return u.String(), nil
}

// parseEndpointURL 解析 http/https 端点，允许省略协议（默认 http://）。
func parseEndpointURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty endpoint")
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint %q: %w", raw, err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid endpoint %q: missing host", raw)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid endpoint %q: unsupported scheme %q", raw, u.Scheme)
	}

	return u, nil
}

// firstEnv 返回第一个非空环境变量值。
func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

// expandHeaders 展开 header 值中的环境变量引用。
func expandHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	expanded := make(map[string]string, len(headers))
	for k, v := range headers {
		expanded[k] = os.ExpandEnv(v)
	}
	return expanded
}

// ExportSpans 丢弃所有 span
func (noopSpanExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error { return nil }

// Shutdown 无资源需要释放
func (noopSpanExporter) Shutdown(context.Context) error { return nil }

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

// contextKey 私有键类型：避免以基础类型作 context key（revive context-keys-type）
type contextKey string

const (
	// TraceIDKey 上下文中的追踪 ID 键
	TraceIDKey contextKey = "trace_id"
	// SpanIDKey 上下文中的 span ID 键
	SpanIDKey contextKey = "span_id"
)

// ContextWithTraceID 向上下文添加追踪ID
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// ContextWithSpanID 向上下文添加span ID
func ContextWithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, SpanIDKey, spanID)
}

// GetTraceIDFromContext 从上下文获取追踪ID
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	if span := trace.SpanFromContext(ctx); span != nil {
		return GetTraceID(span)
	}
	return ""
}

// GetSpanIDFromContext 从上下文获取span ID
func GetSpanIDFromContext(ctx context.Context) string {
	if spanID, ok := ctx.Value(SpanIDKey).(string); ok {
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
