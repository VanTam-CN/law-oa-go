package tracing

import (
	"context"
	"strings"
	"testing"
)

// clearOTLPEnv 清空 OTLP 相关环境变量，避免宿主环境影响断言。
func clearOTLPEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_PROTOCOL",
		"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL",
	} {
		t.Setenv(key, "")
	}
}

func TestDefaultTracerConfigUsesOTLP(t *testing.T) {
	config := DefaultTracerConfig()

	if len(config.Exporters) != 1 || config.Exporters[0] != "otlp" {
		t.Fatalf("DefaultTracerConfig().Exporters = %v, want [otlp]", config.Exporters)
	}
	if config.OTLPProtocol != "http" {
		t.Fatalf("DefaultTracerConfig().OTLPProtocol = %q, want http", config.OTLPProtocol)
	}
	if !strings.HasPrefix(config.OTLPEndpoint, "http://") {
		t.Fatalf("DefaultTracerConfig().OTLPEndpoint = %q, want http:// prefix", config.OTLPEndpoint)
	}
}

func TestProductionTracerConfigUsesOTLPOnly(t *testing.T) {
	config := ProductionTracerConfig()

	if len(config.Exporters) != 1 || config.Exporters[0] != "otlp" {
		t.Fatalf("ProductionTracerConfig().Exporters = %v, want [otlp]", config.Exporters)
	}
}

func TestResolveOTLPEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   string
		tracesEnv  string
		genericEnv string
		legacy     string
		want       string
		wantErr    bool
	}{
		{
			name:     "显式配置完整 URL",
			endpoint: "http://collector.example.com:4318/v1/traces",
			want:     "http://collector.example.com:4318/v1/traces",
		},
		{
			name:     "显式配置 host:port 补全协议",
			endpoint: "collector.example.com:4318",
			want:     "http://collector.example.com:4318",
		},
		{
			name:       "traces 专用环境变量优先于通用环境变量",
			tracesEnv:  "https://traces.example.com:4318/v1/traces",
			genericEnv: "http://generic.example.com:4318/v1/traces",
			want:       "https://traces.example.com:4318/v1/traces",
		},
		{
			name:       "通用环境变量兜底",
			genericEnv: "http://generic.example.com:4318/v1/traces",
			want:       "http://generic.example.com:4318/v1/traces",
		},
		{
			name:   "旧 jaeger 端点兜底",
			legacy: "http://jaeger:4318",
			want:   "http://jaeger:4318",
		},
		{
			name:    "无任何配置报错",
			wantErr: true,
		},
		{
			name:     "非 http(s) 协议报错",
			endpoint: "ftp://collector.example.com:4318",
			wantErr:  true,
		},
		{
			name:      "显式配置优先于环境变量",
			endpoint:  "http://explicit.example.com:4318/v1/traces",
			tracesEnv: "http://env.example.com:4318/v1/traces",
			want:      "http://explicit.example.com:4318/v1/traces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearOTLPEnv(t)
			t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", tt.tracesEnv)
			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", tt.genericEnv)

			config := DefaultTracerConfig()
			config.OTLPEndpoint = tt.endpoint

			got, err := resolveOTLPEndpoint(config, tt.legacy)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveOTLPEndpoint() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveOTLPEndpoint() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("resolveOTLPEndpoint() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveOTLPProtocol(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		env      string
		want     string
		wantErr  bool
	}{
		{name: "默认 http", want: "http"},
		{name: "显式 http", protocol: "http", want: "http"},
		{name: "显式 http/protobuf", protocol: "http/protobuf", want: "http"},
		{name: "显式 grpc", protocol: "grpc", want: "grpc"},
		{name: "环境变量 grpc", env: "grpc", want: "grpc"},
		{name: "显式配置优先于环境变量", protocol: "http", env: "grpc", want: "http"},
		{name: "非法协议报错", protocol: "thrift", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearOTLPEnv(t)
			t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", tt.env)

			config := DefaultTracerConfig()
			config.OTLPProtocol = tt.protocol

			got, err := resolveOTLPProtocol(config)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveOTLPProtocol() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveOTLPProtocol() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("resolveOTLPProtocol() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLegacyJaegerOTLPEndpoint(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "collector 14268 映射到 4318", in: "http://jaeger:14268/api/traces", want: "http://jaeger:4318"},
		{name: "无端口默认 4318", in: "http://jaeger/api/traces", want: "http://jaeger:4318"},
		{name: "https 保留协议", in: "https://jaeger.example.com:14268/api/traces", want: "https://jaeger.example.com:4318"},
		{name: "自定义端口保留", in: "http://jaeger:4319/custom", want: "http://jaeger:4319"},
		{name: "空值返回空", in: "", want: ""},
		{name: "非法端点返回空", in: "ftp://jaeger:14268/api/traces", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultTracerConfig()
			config.JaegerEndpoint = tt.in

			if got := legacyJaegerOTLPEndpoint(config); got != tt.want {
				t.Fatalf("legacyJaegerOTLPEndpoint(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCreateExportersSupportedNames(t *testing.T) {
	for _, name := range []string{"otlp", "otel", "jaeger", "zipkin", "noop", "none"} {
		t.Run(name, func(t *testing.T) {
			clearOTLPEnv(t)

			config := DefaultTracerConfig()
			config.Exporters = []string{name}
			config.OTLPEndpoint = "http://localhost:4318/v1/traces"
			config.ZipkinEndpoint = "http://localhost:9411/api/v2/spans"

			exporters, err := createExporters(config)
			if err != nil {
				t.Fatalf("createExporters() error = %v", err)
			}
			if len(exporters) != 1 {
				t.Fatalf("createExporters() returned %d exporters, want 1", len(exporters))
			}
			if err := exporters[0].Shutdown(context.Background()); err != nil {
				t.Fatalf("exporter Shutdown() error = %v", err)
			}
		})
	}
}

func TestCreateExportersJaegerAliasUsesLegacyEndpoint(t *testing.T) {
	clearOTLPEnv(t)

	config := DefaultTracerConfig()
	config.Exporters = []string{"jaeger"}
	config.OTLPEndpoint = ""
	config.JaegerEndpoint = "http://jaeger:14268/api/traces"

	exporters, err := createExporters(config)
	if err != nil {
		t.Fatalf("createExporters() error = %v", err)
	}
	if len(exporters) != 1 {
		t.Fatalf("createExporters() returned %d exporters, want 1", len(exporters))
	}
	if err := exporters[0].Shutdown(context.Background()); err != nil {
		t.Fatalf("exporter Shutdown() error = %v", err)
	}
}

func TestCreateOTLPExporterProtocols(t *testing.T) {
	clearOTLPEnv(t)

	for _, protocol := range []string{"", "http", "http/protobuf", "grpc"} {
		t.Run("protocol="+protocol, func(t *testing.T) {
			config := DefaultTracerConfig()
			config.OTLPProtocol = protocol

			exporter, err := createOTLPExporter(config, "")
			if err != nil {
				t.Fatalf("createOTLPExporter() error = %v", err)
			}
			if err := exporter.Shutdown(context.Background()); err != nil {
				t.Fatalf("exporter Shutdown() error = %v", err)
			}
		})
	}

	config := DefaultTracerConfig()
	config.OTLPProtocol = "thrift"
	if _, err := createOTLPExporter(config, ""); err == nil {
		t.Fatal("createOTLPExporter() error = nil, want error for unsupported protocol")
	}
}

func TestCreateExportersRejectsUnsupported(t *testing.T) {
	config := DefaultTracerConfig()
	config.Exporters = []string{"datadog"}

	if _, err := createExporters(config); err == nil {
		t.Fatal("createExporters() error = nil, want error for unsupported exporter")
	}
}

func TestCreateExportersRejectsEmpty(t *testing.T) {
	config := DefaultTracerConfig()
	config.Exporters = nil

	if _, err := createExporters(config); err == nil {
		t.Fatal("createExporters() error = nil, want error for empty exporter list")
	}
}
