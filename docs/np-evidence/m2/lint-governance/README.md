# M2 静态分析存量债治理（卡 t_088ba214）

日期：2026-11-24 ｜ 分支：`codex/np-lint-governance`（基于 `codex/np-main-m2@fb73a7d`）
工具链：golangci-lint **v1.64.8**（与 CI `@latest` 实际解析版本逐字节一致，见 `snapshots/ci-run-45390618050-quality.log` L338）

## 1. 问题与归因（继承自 t_7cecb4f1）

- ci.yml quality job「📋 静态分析」步 = `go vet ./... && go mod verify` + `golangci-lint run`。
- `.golangci.yml` 自初始提交启用约 60 个 linter 且 `new: false`；gofumpt 格式步修绿（PR #53 run 45390618050）后该步首次可达即红：**CI 实测 4038 条 error 行**（该 run 日志已归档本目录），本地同二进制 A/B 实证与格式修复无因果（4067 vs 4036），job 在任一树上都不可能绿。

## 2. 治理策略：「正确性/安全门禁 + 风格债台账（ratchet）」

原则：**不掩盖真实缺陷，不用 nolint 滥用，不削弱测试**。

1. `.golangci.yml` 重写为治理版：门禁集只保留编译正确性、错误处理、安全、格式类 linter，**必须保持 0 findings**；
2. 纯风格/意见型/复杂度类 linter 移出门禁集，逐族登记债台账（§5），以 follow-up 卡分批修复后逐个恢复门禁（ratchet，只紧不松）；
3. gosec 每条排除均写明理由并留跟踪卡；G113 以代码加固代替豁免；
4. 测试文件只放宽两类有明确惯例依据的检查（gosec/noctx/gocritic、revive context-as-argument）。

## 3. 数字轨迹（同二进制 v1.64.8，快照见 `snapshots/`）

| 快照 | findings | 说明 |
|---|---|---|
| baseline-full / baseline-fresh | 4001 | 本分支治理前（gofumpt 修绿后的树） |
| round1 | 836 | errors.Is/As AST 改写、errorlint %w、goimports、usestdlibvars、misspell -fix 等 |
| round2（及 go123 对照） | 752 | 手工修复 errorlint 类型断言、gosimple/staticcheck/ineffassign/wastedassign/unconvert、noctx、sqlclosecheck、bodyclose 等 |
| round3-afterfix | 206 | 死代码清理、EnhancedError MarshalJSON、mcp_client 重试泄漏、onlyoffice ctx 链路等 |
| round4 | 131 | 门禁集收敛后剩余 |
| round7-current | **64** | 门禁内残余（rowserrcheck 15 / revive 41 / errorlint 4 / gosec G113 2 / bodyclose 1 / sqlclosecheck 1） |
| round7-afterfix | 11 | 第 7 批代码修复后，仅剩 2 类待决策项（G113×2、test context-as-argument×9） |
| **round8-final** | **0（RC=0）** | 配置决策落地后，门禁集全绿、无告警 |

注：round5(0)/round6(80) 为中间调试快照（round5 曾误带过滤参数），以 round7/round8 为准。

## 4. 本轮（round 7/8）修复与决策明细

**代码修复（64 处门禁内发现，无 nolint）：**
- rowserrcheck ×15 + sqlclosecheck ×1 + bodyclose ×1：补 `rows.Err()` 检查（仓储层返回错误、检测服务按其无错误签名记日志）、`defer rows.Close()`、mcp_client 重试闭包内失败路径立即释放响应体（同时修复重定向错误下的响应体泄漏）；
- errorlint ×4：`migrator.go` `err != migrate.ErrNoChange` → `errors.Is`；
- revive ×33：empty-block ×10（含 2 处占位测试改为显式断言/t.Logf——**加强而非削弱**）、indent-error-flow ×6、if-return ×6、redefines-builtin-id ×3（删除与 Go 1.21+ 内建 min/max 重复的本地包装）、context-keys-type ×5（`internal/tracing` 引入私有键类型 + 导出 `TraceIDKey/SpanIDKey`，logger/hook 服务同步改键，跨包语义一致）、unexported-return ×1（`NewEnhancedCaseService` 改返接口并补 `SetSubjectRecheckService` 进接口——router.go:88 已按此契约调用）、blank-imports ×1（补注册理由注释）；
- gosec G113 ×2：**代码加固**（`metadataNumbersEqual` 拒绝 >64 字符数字串，防 big.Rat 解析指数放大内存），配置侧仍留带理由的排除（stdlib ≥ Go 1.17 已修 CVE-2022-23772 + 护栏双保险，跟踪卡复核后移除）。

**配置决策（均有注释与台账）：**
- 测试文件放行 revive `context-as-argument`（`t *testing.T` 首参是 Go 测试辅助函数约定，规则面向生产签名；非测试代码仍强制）；
- gosec 排除族 G204/G304/G104/G301/G306/G115 逐条理由见 `.golangci.yml` 注释；G113 见上；
- importas `no-unaliased: false`（代码库从未全面使用别名，原配置自初始提交即不可满足）；
- goimports `local-prefixes` 修正为真实模块路径 `law-oa-go`（原 `github.com/your-org` 占位从未生效）；移除与 goimports 重复的 gci；
- gocritic：enabled-checks 精简为非默认项 `emptyFallthrough`（其余约 23 项本就在默认启用集，行为不变）；显式 disable `ifElseChain`/`exitAfterDefer`（理由见配置注释）；
- revive 收敛为正确性信号规则集（19 条）；stylecheck 关闭 ST1000/ST1003/ST1016/ST1020-22 纯风格项；
- **gomnd 移除**：v1.64 中该 linter 已 fully inactivated（no-op），保留只会产生告警；mnd 采纳记入债台账；
- v1.64 弃用配置项迁移：`run.skip-dirs/skip-files → issues.exclude-dirs/exclude-files`、`output.format → output.formats`、`output.uniq-by-line → issues.uniq-by-line`（迁移后 0 告警）。

## 5. 存量债台账（移出门禁集，follow-up 卡逐批修复后恢复）

> 更新（t_2f7c0c77，2026-11-24）：errcheck 族批修完成并恢复门禁，见下表「已恢复」行。

| 族 | 治理前数量 | 去向 |
|---|---|---|
| ~~errcheck / errchkjson / forcetypeassert~~ | 实测基线 216 / 31 / 116（=363；卡面 513/23/74 为治理前更早快照口径） | **✅ 已批修恢复门禁（t_2f7c0c77，2026-11-24）：存量 363 发现全部清零，门禁集 0 findings，0 新增 nolint，无测试削弱** |
| wrapcheck | 439 | 决策卡（错误包装策略先定，再批修） |
| goheader | 329 | 决策卡（统一 license 头模板后一次性补齐） |
| revive 风格规则 | exported 749 / unused-parameter 314 / var-naming 59 / package-comments 40 | 批修卡（恢复规则后按包分批） |
| 复杂度/可维护性 | gocognit 114 / dupl 87 / nestif 42 / gocyclo 33 / funlen 25 / maintidx 11 | 批修卡 |
| goconst | 132 | 批修卡 |
| unused / nilnil / unparam / promlinter / nonamedreturns / containedctx / prealloc | 79 / 67 / 13 / 13 / 12 / 9 / 14 | 甄别卡（unused 需排除 gorm/JSON 反射引用） |
| gosec 多用户部署收紧 | G301/G306/G115（排除中） | 安全卡（多用户部署前收紧） |
| 基础设施 | jaeger→OTLP（SA1019 排除中）、dev-api-key 后门（G101 排除中，修复或移除） | 基础设施/安全卡 |
| main.go | run() error 模式改造（exitAfterDefer 排除中） | 架构卡 |
| 工具 | mnd 采纳、gosec/govulncheck 是否固定版本 | 工具卡 |

## 6. 工具版本固定决策（验收 3）

**决策：gofumpt/golangci-lint 固定版本，已落地 ci.yml。**
- `gofumpt@v0.12.0`：与本地修复、CI 下载版本一致（run 45390618050 实证：`go: downloading mvdan.cc/gofumpt v0.12.0`；且 @latest 曾在运行中把工具链切到 go1.26.11——浮动成本实证）；
- `golangci-lint@v1.64.8`：与 `.golangci.yml` 治理版（v1 配置格式）和本轮全部本地验证一致；v1 模块路径的 `@latest` 一旦上游切 v2（新配置格式不兼容）会直接挂；
- gosec/govulncheck 保持 @latest：漏洞库时效优先，是否固定列入工具卡。

## 7. 验证（全部真实执行）

- `golangci-lint run --timeout=10m`（v1.64.8）：**0 findings，RC=0，0 告警**（`snapshots/round8-final.txt`）；
- `go vet ./...`：通过；`go mod verify`：all modules verified；
- `gofumpt -l internal/ cmd/ main.go tests/`：0 漂移（格式步不会回红）；
- 受影响包测试 `-race -count=1`：services、services/test、repositories、database、handlers（-skip WorkbenchApprovalBrowser，本地缺 playwright 栈，沿用 t_7cecb4f1 先例且本轮未触碰该文件）、concurrency、security、validators、middleware、models **全部 ok**；无测试削弱（empty-block 占位两处改为显式断言/t.Log 为加强）。

## 8. 探针 PR 与托管核验（真实事件）

- 探针 PR：**#54**（draft，`codex/np-lint-governance` → main，head 50bdf87）
- CI/CD Pipeline run：**45405844993**（2026-11-24T13:40Z，PR 事件触发）
- quality job「🔍 Code Quality」step 结论（真实托管事件）：
  | step | 结论 |
  |---|---|
  | 🔍 代码格式检查（gofumpt@v0.12.0 固定版） | **success** |
  | 📋 静态分析（go vet + go mod verify + golangci-lint@v1.64.8） | **success** ← 本卡验收目标转绿 |
  | 🔒 安全扫描 | failure：gosec@latest(dev) **161 issues**（311 files/130k lines，gosec 通过后才会跑 govulncheck，未达） |
- 归因：静态分析/格式两步为 **pre-existing 红灯修复**（本卡目标）；安全扫描步是静态分析修绿后**新近可达**的同款存量债（旧树从未执行到该步），gosec 独立运行的 161 条与 golangci-lint 门禁集内的 gosec 0 findings 不冲突（门禁集含逐条理由排除 + G104 重复项策略），已立卡治理。
- 同 run 其他 workflow：fuzzing.yml 5 个 fuzz 目标全红——与 i11 分支当日 5 连红（11:59–12:54，早于本分支存在）同款，且 fuzz 输出重定向至 runner temp 从未归档（观测缺陷），判为 pre-existing 流水线问题，已立卡；PGO「检查PGO支持」红为 m2 线已知分类。

## 9. 快照清单

见 `MANIFEST.sha256`（16 个文件：治理前基线、round1-8 全过程、PR #53 run 45390618050 与 PR #54 run 45405844993 quality job 原始日志、fuzz job 日志）。

> 卡 t_2f7c0c77（2026-11-24）追加 2 个快照：`snapshots/t2f7c0c77-baseline-363.txt`（errcheck 族基线 363 发现）与 `snapshots/t2f7c0c77-final-lint.txt`（门禁恢复后 0 findings，RC=0）。MANIFEST 更新至 18 条目（含 3 个 gitignored 的 CI 原始日志引用，沿用既有模式）。

## 10. errcheck 族批修与门禁恢复（卡 t_2f7c0c77，2026-11-24）

**范围**：errcheck / errchkjson / forcetypeassert 三族存量发现批修，恢复进门禁集（ratchet）。

### 10.1 基线（191bb6b 实测，同二进制 v1.64.8）

| 族 | 实测基线 | 卡面计数 | 说明 |
|---|---|---|---|
| errcheck | 216 | 513 | 卡面 513 为治理前更早快照口径；实测以 191bb6b 为准 |
| errchkjson | 31 | 23 | 同上（23 unsafe-type + 8 not-checked，源码确认两类均由「检查错误」修复） |
| forcetypeassert | 116 | 74 | 86「type assertion must be checked」+ 30「right hand must be only type assertion」 |
| **合计** | **363** | — | 基线快照 `snapshots/t2f7c0c77-baseline-363.txt` |

### 10.2 修复策略（无 nolint，无测试削弱）

- **errchkjson**：读 `errchkjson@v0.4.0` 源码确认「unsafe type」仅在错误被丢弃到 `_`（blankIdentifier）时触发 → 统一改为 `data, err := json.Marshal(...); if err != nil { return ... }`，同时清除 unsafe-type 与 not-checked 两类，不削弱错误处理；
- **defer Close 统一模式**：72 处 `defer <expr>.Close()` → `defer func() { _ = <expr>.Close() }()`（延迟清理阶段错误无后续处理路径，显式丢弃并附理由注释）；`frontend/migrate/main.go` 4 处 `m.Close()` 返回 2 值 → `_, _ = m.Close()`；
- **forcetypeassert**：`x, ok := v.(T); if !ok { ... }`；生产代码失败返回 error/安全默认，内部不变量（对象池/L1 缓存/gin.H 响应）用带诊断消息的 panic，日志字段类型不符时**跳过字段**避免记录时 panic；
- **errcheck 其他**：`os.Setenv`/`os.WriteFile`/`w.Write`/`RegisterValidation`/`gorm` 调用等全部改为检查错误（错误传播或带理由注释的显式丢弃），103 处 `_ =` 均附中文理由注释，16 处 panic 均为内部不变量保护。

### 10.3 结果与验证（全部真实执行）

- **存量 363 发现全部清零**（216 errcheck + 31 errchkjson + 116 forcetypeassert），`go build ./...` 通过，`go vet ./...` 通过；
- **门禁恢复后全仓 lint**：`golangci-lint run -c .golangci.yml`（v1.64.8，errcheck 族已 enable）→ **0 findings，RC=0**（`snapshots/t2f7c0c77-final-lint.txt`）；0 新增 nolint（nolintlint 门禁内）；
- **gofumpt -l 变更文件**：0 漂移（CI 格式步不会回红）；
- **受影响包 `-race -count=1` 复跑**：handlers、services、repositories、database、health、common、api、auth、cache、concurrency、security、validation、middleware、errors、logger、models、elasticsearch、rbac、metrics、cmd、frontend/migrate、test、tests/integration **全部 ok**；
- **tests/performance**：`TestPerformanceMonitoring` 的 `duplicate metrics collector registration` panic 为 **pre-existing**（base 191bb6b 同款复现，与本卡无关，未触碰该测试函数）；非本卡范围，沿用 pre-existing 记录惯例。

### 10.4 门禁恢复决策

- `.golangci.yml` `linters.enable` 加回 `errcheck`、`errchkjson`、`forcetypeassert`（错误处理组）；errcheck 保持默认 `check-blank: false`（103 处有理由注释的 `_ =` 为治理约定，由评审而非 linter 强制）；G104 排除注释更新（errcheck 已在门禁，不双份）；
- 债台账 §5 已将 errcheck 族标记「✅ 已批修恢复门禁」。
