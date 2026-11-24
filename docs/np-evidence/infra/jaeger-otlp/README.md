# jaeger→OTLP exporter 迁移 — 证据（t_72ad1a27）

## 1. 结论

- jaeger exporter（2023-07 起官方停止维护）已从导出链路移除，改为 `otlptracehttp`/`otlptracegrpc`。
- `.golangci.yml` 的 SA1019 逐文件排除与 tracer.go import 注释已删除；本地全仓 golangci-lint v1.64.8 **0 findings，RC=0**。
- 真实 PR 事件（PR #56，head e931f17，run 45411804610）：「代码格式检查」「静态分析」两步 **success**；「安全扫描」仍为已知 gosec 存量（t_745060c0），quality job 整体 failure 与本次迁移无关。

## 2. 迁移设计（配置兼容）

| `exporters` 值 | 行为 |
|---|---|
| `otlp`（默认）/ `otel` | OTLP 导出（`otlpProtocol`：`http` 默认 4318 / `grpc` 4317） |
| `jaeger`（弃用别名） | 仍接受；按 OTLP 导出（Jaeger >=1.35 原生接收 OTLP），`jaegerEndpoint` 14268 → 4318；启动输出弃用告警 |
| `zipkin` | 保持不变 |
| `noop` / `none` | 显式空导出器（丢弃 span） |

端点解析优先级：`otlpEndpoint` > `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` > `OTEL_EXPORTER_OTLP_ENDPOINT`；`jaeger` 别名再回退 `jaegerEndpoint`（同主机、14268→4318、去路径）。
协议解析优先级：`otlpProtocol` > `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL` > `OTEL_EXPORTER_OTLP_PROTOCOL` > `http`。

## 3. 变更清单

- `internal/tracing/tracer.go`：删除 jaeger 导入与 SA1019 注释；新增 otlptracegrpc；`createJaegerExporter` → OTLP 别名映射；端点/协议解析、header 展开、noop exporter。
- `internal/tracing/tracer_test.go`（新增；`.gitignore` 增补例外）：默认/生产配置、端点解析 8 例、协议 7 例、legacy 映射 6 例、导出器名 6 例、协议构造 4 例、异常路径 2 例。
- `.golangci.yml`：删除 `internal/tracing/tracer\.go` 的 SA1019 排除。
- `go.mod` / `go.sum`：移除 `go.opentelemetry.io/otel/exporters/jaeger v1.17.0`，新增 `otlptracegrpc v1.38.0`；未做无关的 tidy 清理。
- 部署侧文档：`.env.example`、`docs/operations-manual.md` §5.6、`docs/DEPLOYMENT_ENTERPRISE.md` §5.4、`docker-compose.yml`（OTLP env + jaeger 4317/4318 端口）。

## 4. 本地验证（快照见 `snapshots/`）

| 命令 | 结果 |
|---|---|
| `golangci-lint run --timeout=10m`（v1.64.8） | 0 findings，RC=0（`lint-final.txt`） |
| `go vet ./...` | RC=0（`go-vet.txt`） |
| `go mod verify` | all modules verified（`go-mod-verify.txt`） |
| `go build ./...` | 通过 |
| `gofumpt -l -s .` | 0 文件漂移（`gofumpt.txt`） |
| `go test -race -count=1 ./internal/tracing/...` | ok，新增测试全绿（`go-test-tracing.txt`） |
| `go test -race -count=1 ./internal/services/...` | ok（`go-test-services.txt`） |

## 5. 探针 PR 与托管核验

- PR **#56**（draft）：`codex/np-jaeger-otlp` → `main`，head **e931f17**（基于 `codex/np-lint-governance@191bb6b`，含父卡提交；合并载体决策见 t_4dfa464a）。
- run **45411804610**（🚀 CI/CD Pipeline，pull_request 事件）：
  - 🔍 代码格式检查：**success**
  - 📋 静态分析：**success**
  - 🔒 安全扫描：failure（gosec 161，已知存量 t_745060c0；后续 test 步骤因此 skipped）
- 同一 head 其他 workflow：安全测试 45411804628 **success**；Test Pipeline 45411804595 failure；Law OA Go CI/CD 45411804602 集成静态分析 **success**；Fuzzing/PGO 为已知红灯（t_46d9c62d 等）。见 `ci-runs-summary.txt`。
- Test Pipeline（45411804595）归因：Go Tests 失败点为 `internal/health` 的 `TestConflictP0ReadinessCheck_ProductionBadCanonicalHashFailsClosed`，属 **pre-existing 抖动**——测试用 `replace(首字符, '0')` 篡改哈希，首字符本来就是 `'0'` 时 UPDATE 无效，约 1/16 概率误通过。本地 `-count=100` 复现 5 次 FAIL（`flaky-health-repro.txt`），已立卡 **t_b14396c0**。CI 日志中 `law-oa-go/internal/tracing ok 1.041s`（见 `ci-test-pipeline-attribution.txt`）。
- Law OA Go CI/CD integration 失败为 workflow 基础设施问题（MySQL `Access denied for user 'root'@'localhost'`，设置数据库步骤），与代码无关。

## 6. 残留与说明

- `jaeger` 别名保留弃用告警；建议部署配置改为 `exporters: [otlp]`。
- 同时配置 `otlp` 与 `jaeger` 且解析到同一端点会产生重复 span；建议只保留 `otlp`。
- `internal/tracing` 当前无外部 `InitTracer` 调用方（main 启动链路未接入）；本卡只迁移导出器实现与配置契约。
