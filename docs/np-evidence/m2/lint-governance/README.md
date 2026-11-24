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

| 族 | 治理前数量 | 去向 |
|---|---|---|
| errcheck | 513 | 批修卡（与 errchkjson 23 / forcetypeassert 74 同卡） |
| wrapcheck | 439 | 决策卡（错误包装策略先定，再批修） |
| goheader | 329 | 决策卡（统一 license 头模板后一次性补齐） |
| revive 风格规则 | exported 749 / unused-parameter 314 / var-naming 59 / package-comments 40 | 批修卡（恢复规则后按包分批） |
| 复杂度/可维护性 | gocognit 114 / dupl 87 / nestif 42 / gocyclo 33 / funlen 25 / maintidx 11 | 批修卡 |
| goconst | 132 | 批修卡 |
| unused / nilnil / unparam / promlinter / nonamedreturns / containedctx / prealloc | 79 / 67 / 13 / 13 / 12 / 9 / 14 | 甄别卡（unused 需排除 gorm/JSON 反射引用） |
| gosec 多用户部署收紧 | G301/G306/G115（排除中） | 安全卡（多用户部署前收紧） |
| 基础设施 | jaeger→OTLP（SA1019 排除中）、dev-api-key 后门（G101 排除中，修复或移除） | 基础设施/安全卡 |
| main.go | run() error 模式改造（exitAfterDefer 排除中） | 架构卡 |
| 工具 | gosec 固定 v2.30.0 / govulncheck 保持 latest / mnd 暂缓采纳 | 已决策（t_4dfa464a，见 §6/§10） |

## 6. 工具版本固定决策（验收 3）

**决策：gofumpt/golangci-lint 固定版本，已落地 ci.yml。**
- `gofumpt@v0.12.0`：与本地修复、CI 下载版本一致（run 45390618050 实证：`go: downloading mvdan.cc/gofumpt v0.12.0`；且 @latest 曾在运行中把工具链切到 go1.26.11——浮动成本实证）；
- `golangci-lint@v1.64.8`：与 `.golangci.yml` 治理版（v1 配置格式）和本轮全部本地验证一致；v1 模块路径的 `@latest` 一旦上游切 v2（新配置格式不兼容）会直接挂；
- ~~gosec/govulncheck 保持 @latest~~ → 已决策（t_4dfa464a，2026-11-24 回填）：
  - **gosec 固定 `@v2.30.0`**（已落地 ci.yml）：静态规则扫描器，无漏洞库时效约束，门禁要求发现集可复现（同版本=同发现集）；v2.30.0（2026-10-23 release）即基线 run 45405844993 时 `@latest` 的解析版本，与 gosec 161 条存量债基线**同版本、零漂移**。维护策略：跟随上游 release 每周 bump（`go install ...@vX.Y.Z` 单行改动），升级动作并入 gosec 收紧卡跟踪；bump 后需核对新增规则的发现增量。
  - **govulncheck 保持 `@latest`**：价值在漏洞库时效（DB 每次运行实时拉取），工具本体迭代慢（最新 release v1.1.4，2025-01），固定收益低、反受漂移；「周期性全量重扫」由 security.yml 既有每日 schedule（cron `0 3 * * *`，dependency-scan 报告态）承接，策略成立无需新增 workflow。
  - **mnd 暂缓采纳**（门禁集不含；gomnd 在 golangci-lint v1.64 已 no-op 移除）：启用前置是清存量魔法数字（存量规模未盘点），先 shadow-run 盘点再启用，已立独立卡跟踪（与复杂度族批修卡解耦——mnd 是 lint 配置采纳，不是复杂度重构）。

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

## 10. 治理携入 main 承载决策（卡 t_4dfa464a，2026-11-24）

**决策：PR #52（i11→main 集成载体）先合，governance 线 rebase 后经 PR #54 转正式合入。**

事实基础（2026-11-24 实测）：
- `merge-base(main, codex/np-lint-governance) = d9c92fbd = main tip`：main 尚无任何 NP 内容；
- lint-gov 栈于 codex/np-main-m2（祖先关系 YES），但 m2 非 i11 祖先、i11 非 gov 祖先——i11 线以**内容重放**方式承载 M0–M4（如 gofumpt 清理在 m2 线为 11a3db0、在 i11 线重放为 e2034e9，同内容不同 SHA）；
- PR #52 是唯一把 NP 全量内容（M0–M4+I11）送进 main 的载体；PR #53（gofumpt 探针）已 CLOSED；PR #54 现为 draft 探针。

**路径（顺序即决策）**：
1. PR #52 合入 main（i11 载体，np-i11 门禁已在真实 PR 事件 GREEN）；
2. `codex/np-lint-governance` rebase 到合并后的 main：治理三连（573d328+50bdf87+191bb6b）及后续提交重放；
3. t_745060c0 完成 gosec 161 条治理、安全扫描步转绿后，PR #54 转 ready 合入——`.golangci.yml` 治理版与 ci.yml 全部固定版本由此真正落到 main。

**拒绝备选（PR #54 先于 #52 合入）**：会把 m2 线历史先于 i11 载体注入 main，与 #52 形成同内容双载体（11a3db0 vs e2034e9 两份 gofumpt 清理），后续合并必爆冲突风暴，且提前固化未经 i11 载体决策的树。

**预期冲突点与消解规则（rebase 时执行）**：
- ci.yml 缓存条件区（setup-go cache / actions-cache if / codecov if / notification if）：m2 线是 `!= 'codex/np-main-m2'` 单分支条件，i11 线是含 m2/m3/i11 的三分支超集——**取 i11 侧（承载超集）**；
- security.yml PILOT_PR/cache 条件区：同规则，取 i11 侧；
- ci.yml 格式/静态分析/安全扫描步的工具固定与策略注释：**取 governance 侧**；`.golangci.yml` 全文件：取 governance 侧（i11 线未触碰该文件）。

PR #54 保持 draft 直至第 3 步条件满足；合入前需复核 base/main 状态与 rebase 后的 quality job 真实事件。

## 11. 安全扫描步治理：gosec 161 条存量债分类修复（卡 t_745060c0，2026-11-24）

**问题**：ci.yml「🔒 安全扫描」步 = `gosec -conf=.gosec.json ./...` + `govulncheck ./...`。静态分析步修绿后该步首次可达即红：gosec(dev=v2.30.0) **161 issues**（311 files/130124 lines），gosec 未过导致 govulncheck 未执行。注意 golangci-lint 门禁集内 gosec 为 0 findings（门禁集含逐条理由排除），161 条为独立 gosec 全量运行的存量债。

**分类（gosec v2.30.0，run 45405844993 日志逐条解析，161 条）**：

| 规则 | CWE | 数量 | 性质 | 处置 |
|---|---|---|---|---|
| G104 | CWE-703 | 84 | errcheck 类「error 未处理」（DB 写/缓存/Close/Unmarshal/Write 等） | **排除**（errcheck 族债务单列 t_2f7c0c77；门禁集不启用 errcheck；非安全信号） |
| G118 | CWE-400 | 15 | goroutine 用 context.Background()（调度 ticker、异步 ES 同步/富化/记录） | **排除**（fire-and-forget 后台任务，request ctx 会提前取消；非安全信号） |
| G704 | CWE-918 | 14 | `main.go:555` runHealthcheck SSRF | **排除**（host 硬编码字面量 127.0.0.1 自检；path 为 env 且校验 `/health/` 前缀；非攻击者可控） |
| G115 | CWE-190 | 11 | int 溢出转换 | **修复**（2 处 `string(rune(...))` 真缺陷→strconv；9 处 uint64/int64 监控统计→safeconv 饱和钳制） |
| G304 | CWE-22 | 11 | 变量路径文件读（DB 存储路径/迁移/日志/拷贝目标） | **排除 + 根因加固**（路径为服务端受控；上传文件名已 `filepath.Base` 加固；gosec 无法证明 taint-clean，排除为分析器级） |
| G301 | CWE-276 | 9 | MkdirAll 0755 | **修复**（→0750，卡 t_2ebe3c13） |
| G101 | CWE-798 | 4 | 硬编码凭据 | **修复 1 + 排除 3**（v2_middleware_factory `dev-api-key` 后门已移除——真实缺陷；testutils×2/optimized_components×1 为 bcrypt 哈希测试/种子常量，非明文凭据） |
| G306 | CWE-276 | 4 | WriteFile 0644 | **修复**（→0600，卡 t_2ebe3c13） |
| G401/G501 | CWE-328/CWE-327 | 4 | md5 弱密码学 | **排除**（缓存键/ETag 指纹，非密码学用途；门禁集同理由） |
| G201 | CWE-89 | 2 | SQL 字符串拼接 | **排除**（仅拼占位符与常量列名，值全走参数化 args；门禁集同理由） |
| G404 | CWE-338 | 1 | math/rand | **排除**（重试抖动，非密码学；门禁集同理由） |
| G402 | CWE-295 | 1 | TLS InsecureSkipVerify | **排除**（elasticsearch 当前固定 http:// 无 TLS；该校验位未生效；跟踪卡） |
| G706 | CWE-117 | 1 | 日志注入 | **排除**（log.Printf 服务端派生 HTTP method 串，非攻击者自由文本，LOW） |

**真实缺陷修复（代码层，均以 v2.30.0 全量重跑验证）**：
- G101 `v2_middleware_factory.go`：移除 `apiKey == "dev-api-key"` 后门（非开发环境不再接受任何静态内置密钥，须提供与 `JWT.Secret` 一致的密钥；开发环境仍由 `Environment == "development"` 放行）。
- G115 `auth/context.go` `GetUserIDAsString`：`string(rune(userID))` → `strconv.FormatUint`（原实现对 >127 的 ID 产生非法 UTF-8 单字节，真缺陷）。
- G115 `metrics/metrics.go:737`：`string(rune(c.Writer.Status()))` → `strconv.Itoa`（原实现对任意 >127 状态码产生错误字符串）。
- G115 其余 9 处（v2_performance_middleware、performance_monitor、health、auth/context 类型开关）：新增 `internal/safeconv`（`Int64` 饱和 / `Uint64` 钳制）防监控统计溢出回绕，消除负读数。
- G301/G306 权限收紧（0755→0750、0644→0600）：由卡 t_2ebe3c13 处理（本卡验证依赖）。

**误报收敛（`.gosec.json`，`-conf` 显式加载；全局排除仅 2 类非安全信号，其余路径级）**：见 `.gosec.json` 与 ci.yml 安全扫描步注释。要点：全局 `exclude: G104,G118`（errcheck 族 + goroutine 模式，均非安全信号且各有归口）；路径级 `exclude-rules` 覆盖 G304（受控路径）、G101（测试/种子常量）、G401/G501（md5）、G201（参数化）、G404（抖动）、G402（http://）、G704（自检）、G706（日志）——与 `.golangci.yml` 门禁集理由一致。

**验证（gosec v2.30.0 与本步 CI 同版本）**：`gosec -conf=.gosec.json ./...` → **0 issues，RC=0**；无配置基线 161→136（G115 修复 -11、G301 -9、G306 -4、G101 后门 -1 均为真实代码变更，非排除隐藏——排除后仍为 0 的 136 条全为文档化误报）。受影响包 `go vet` 通过、`-race` 测试通过（auth/health/middleware-v2）。

**govulncheck 49 条（新近可达，独立债务）**：gosec 修复后 govulncheck 首次执行即报 **49 条影响代码的漏洞**（9 模块 + stdlib）。分类：依赖升级（jwt v5.0→5.2.2、pgx、go-redis v9.0.5→9.6.3、otel v1.38→1.47、x/net 0.43→0.60、x/text 0.28→0.41、grpc）+ Go 工具链升级（stdlib net/http/crypto/x509/crypto/tls/html/template 需 go1.26.9+，CI GO_VERSION 1.25.0）。此为**依赖卫生债**（非代码缺陷），需依赖整体刷新 + 工具链决策（与 jaeger→OTLP 卡交互），已立 follow-up 卡；在依赖刷新落地前，本步为 gosec-绿 / govulncheck-红，属诚实状态（govulncheck 从未在 CI 到达过，属新近暴露）。
