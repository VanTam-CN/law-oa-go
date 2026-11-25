# i11/m3/m4 线 gofumpt 漂移清理（t_b453c53f）

来源：t_7cecb4f1（codex/np-main-m2@11a3db0 全仓清理）的残余风险第 3 条（`docs/np-evidence/m2/gofumpt-cleanup/README.md`）。事实：主干集成载体 PR #52（codex/np-main-i11 → main，Draft）在 head 309d1dd 上「🔍 Code Quality → 代码格式检查」步红（run [45389287095](https://github.com/VanTam-CN/law-oa-go/actions/runs/45389287095)）。

## 处置决策：选项 B（各线直接重跑 gofumpt），不 cherry-pick 11a3db0

- 修复前实测 i11/m3/m4 三线各有 **151 个**漂移文件（internal/143 + tests/5 + cmd/2 + frontend/1），比 m2 线 11a3db0 覆盖的 147 个多 4 个——cherry-pick 无法覆盖各线新增漂移文件。
- CI 判定即 `gofumpt -w -s . && git diff --exit-code`（`go install mvdan.cc/gofumpt@latest`），对 tip 树重跑工具产出的就是 CI 的精确不动点。
- 工具与 CI 完全同版本：本地 `gofumpt v0.12.0 (go1.26.11)`；本次 CI 日志实证下载 `mvdan.cc/gofumpt v0.12.0` 并切 go1.26.11 工具链（snapshots/pr52-run45395793290-format-step.log）。

## 三线修复提交（纯 .go，151 文件/线）

| 线 | 提交 | 规模 | 备注 |
| --- | --- | --- | --- |
| codex/np-main-i11 | e2034e9 | 39417+/39410- | base 为 7c34bd0（兄弟任务 20:34 的 jest/eslint scoping，仅 frontend 两配置文件，与本修复零交集，先于本修复推送） |
| codex/np-main-m3 | 3cb6737 | 39418+/39411- | 无独立 PR |
| codex/np-main-m4 | c4dbc3d | 39417+/39410- | 为 PR #51（i11→m4）的 base，已随本修复前进 |

提交纪律：共用检出热点（见 m2 README 第 5 条）——每次提交前重查 HEAD/status，`git add -u` + `grep -cv '\.go$'`=0 双校验，未收编任何外来文件。

## 验收标准逐条

1. **三线 tip `gofumpt -l .` 与 `gofmt -l .` 均为空** — 达成（snapshots/drift-and-fix-boundary.txt，修复后新鲜实测）。
2. **build 与受影响测试组复跑（同 t_7cecb4f1 口径）** — 达成，在 i11 修复树（7c34bd0+e2034e9）上：
   - `go build ./...`、`go vet ./...`、`go mod verify` 全过；
   - 受影响 10 个 internal 包（errors/handlers/health/middleware/models/performance/repositories/security/services/services-test）`-race -count=1` 全绿（snapshots/internal-10pkgs-rerun.txt；父任务遇到的 internal/health 瞬时失败本轮未复现）；
   - m1 tagged 11 测试于一次性 PG（np-gofumpt-pg，postgres:16.14 tmpfs，127.0.0.1:18790）恰一次执行全过，`check-go-events.py` = `M2_PG_PASS required=5 passed=64`（snapshots/m1-tagged11-rerun.txt）；
   - m3/m4 tip 各跑 `go build ./...` 通过（snapshots/m3-m4-build.txt）。
3. **PR #52「代码格式检查」步转绿** — 达成：head 已前进至 e2034e9，run [45395793290](https://github.com/VanTam-CN/law-oa-go/actions/runs/45395793290)（quality job 138451088849）步骤 5「🔍 代码格式检查」= **success**（snapshots/pr52-run45395793290-quality-steps.json / pr52-run45395793290-format-step.log）。
4. **np-i11 门禁不动、解耦保持** — 达成：本卡全部提交仅含 .go 与本证据目录，未触碰任何 workflow；push 事件上 np-i11 (full) 于 e2034e9 **success**（run [45395793181](https://github.com/VanTam-CN/law-oa-go/actions/runs/45395793181)）。

## 伴随红灯定性（非本卡范围、与格式无关、修复前后不变）

- quality job「📋 静态分析」步失败：~4 千条存量 lint 债（golangci-lint v1.64.8），父任务 A/B 已归因，专项卡 t_088ba214。
- Test Pipeline「🐹 Go Tests」失败：新旧 head 失败签名逐字一致（`--- FAIL: TestPerformanceMonitoring`，tests/performance，snapshots/gotest-ab-signature.txt）；「🎭 E2E Tests」同为双红——均为历史慢性类别（m2 README 第 24–29 行同源归因）。

## 边界记录

- 作业期间 i11 线被兄弟任务插入 7c34bd0（jest/eslint scoping），本地修复/测试均在该含 7c34bd0 的树上完成，验证口径即为实际推送状态。
- m3 push 遇 4 次 SSL 抖动重试后成功；disposable 容器 np-gofumpt-pg 沿用父任务遗留实例（口令鉴权经 PGPASSWORD 注入，首次无口令全连接失败、未污染 DB）。

## 残余

- ci.yml `gofumpt@latest` 浮动版本仍可能因上游发版再漂移（与 t_088ba214 固定版本建议合并处理）。
- 后续任何线在合并前的增量提交仍需保持 gofumpt 干净（可复用本卡命令）。
