# gofumpt 遗留漂移清理（t_7cecb4f1）

来源：`docs/np-evidence/m2/full-i11/final-review/README.md` 第二节 F4/F5。
CI 步骤：`.github/workflows/ci.yml` quality job「🔍 代码格式检查」= `gofumpt -w -s . && git diff --exit-code`（`go install mvdan.cc/gofumpt@latest`，浮动版本）。

## 事实与处置

- 全仓漂移实测（gofumpt v0.12.0，`gofumpt -l .`）：**147 个文件** = internal/ 139 + tests/ 5 + **任务书未列的 3 个**（cmd/gen_password、cmd/migrate、frontend/migrate）。CI 步骤作用于全仓 `.`，只修 144 个探针 PR 仍会红，故全部修复。
- 修复提交：**11a3db0**（codex/np-main-m2，`style(m2): clear gofumpt drift across repo`），纯格式（import 分组排序、长字面量折行；39400+/39410-，净 -10 行），147 文件改动集与漂移清单逐一精确一致，无任何非 Go 文件。
- 探针 PR：**#53**（draft，`codex/gofumpt-probe-main`（= m2 线 tip）→ main），与 #49/#50 同构的临时探针。

## 验收标准逐条

1. **`gofumpt -l internal/ tests/` 为空** — 达成（实测全仓 `gofumpt -l .` 与 `gofmt -l .` 均空，见 snapshots/drift-and-fix-boundary.txt）。
2. **build 与受影响测试组复跑** — 达成，未削弱任何测试：
   - `go build ./...`、`go vet ./...`、`go mod verify` 全过（snapshots/build-vet-and-internal-rerun.txt）；
   - 受影响 10 个 internal 包（24 个 _test.go 所在）`-race -count=1` 复跑全绿，与修复前基线通过集一致；
   - tests/integration/m1 受影响 5 文件中 4 个 `np_m1_pg` 文件的 11 个测试：一次性 postgres:16.14（`np_first_slice_dev:np_m1:160014`）上恰一次执行全过，`check-go-events.py` = `M2_PG_PASS required=5 passed=64`（snapshots/m1-tagged11-rerun.txt）；
   - 第 5 个文件（workbench_approval_browser_test.go，标签 `np_m1_pg,np_m1_browser_pg`）需前端 Playwright 栈，本机未装、未复跑；该文件改动仅 2 行 import 位置移动，且同代码语义的完整浏览器链（vite 代理→Go 夹具→Playwright）已由 t_92dc3e8f 于 2026-11-24 全链验证通过；
   - 复跑中 internal/health 出现一次瞬时失败，已实证排除（文件不在改动集、SQLite 隔离、两棵树稳定复现通过），见 snapshots/health-flake-excluded.txt。
3. **指向 main 的探针 PR 格式检查步转绿** — 达成：PR #53 / run [45390618050](https://github.com/VanTam-CN/law-oa-go/actions/runs/45390618050)「🔍 代码格式检查」步 **success**（#49 run 45334857440、#50 run 45334868179 同步红灯的即此步）。CI 实际下载 `mvdan.cc/gofumpt v0.12.0`，与本机修复工具版本/工具链（go1.26.11）完全一致（snapshots/pr53-run-45390618050-format-step.log）。**注**：quality job 随后的「📋 静态分析」步红灯，归因见下。
4. **np-i11 解耦保持** — 达成：np-i11 为托管侧独立门禁，不含本仓 ci.yml quality job；本次改动 147 文件全部为 `.go`，未触碰任何 workflow/脚本（`git diff --name-only 4e83cc0 11a3db0 | grep -cv '\.go$'` = 0），不影响 I11 判定及其证据。

## 「静态分析」步红灯归因（非本次引入）

- run 45390618050 该步失败于 golangci-lint（CI 下载 v1.64.8）海量发现：uniq 后约 4400 条（revive 1202 / errcheck 513 / wrapcheck 439 / goheader 329 等）。
- A/B 实测（同目录 detach checkout，同版本 golangci-lint，同配置）：修复前树 4067 条 vs 修复后树 4036 条（净 -31；少量增减为 import 块重排引起的 goheader/gci/goimports 消息位移抖动）。两棵树都是千级存量债，job 在两棵树上都不可能绿。
- 此前每次 PR 都死在格式步，「静态分析」步从未可达——属**存量债首次暴露**，非本次引入。`.golangci.yml` 自初始提交启用约 60 个 linter 且 `new: false`。
- 其余 4 个触发的工作流（PGO、fuzzing、Test Pipeline、Law OA Go CI/CD Pipeline）同样红灯，均为历史已知慢性类别（见 final-review 历史分类），与格式无关；安全测试工作流 success。

## 证据清单（snapshots/）

| 文件 | 内容 |
| --- | --- |
| drift-and-fix-boundary.txt | 漂移/改动边界、目录分布、修复后 gofumpt/gofumpt -l 全空实证 |
| build-vet-and-internal-rerun.txt | build/vet/mod verify + 10 internal 包复跑 |
| m1-tagged11-rerun.txt | m1 11 测试逐条恰一次通过 + 校验器判定 |
| health-flake-excluded.txt | health 瞬时失败定位与跨树排除 |
| lint-ab-summary.txt / lint-ab-pre-4e83cc0-full.txt / lint-ab-post-11a3db0-full.txt | golangci-lint A/B 归因与全量原始输出 |
| pr53-run-45390618050-format-step.log | CI 格式步完整日志（含 gofumpt v0.12.0 下载行） |
| pr53-run-45390618050-quality-jobs.json | quality job 全步骤结论 |
| pr53-all-workflow-runs.txt / pr53-other-workflow-failures.txt | 探针分支 6 工作流终态与失败一线归因 |

## 残余风险与后续建议（不在本卡范围）

1. **静态分析存量债（~4 千条）**：格式步已绿后每次 PR 都会撞上它红灯。建议另立卡片专项治理（调 linter 集合或分批修复）。
2. **浮动版本风险**：ci.yml 用 `gofumpt@latest` / `golangci-lint@latest`，上游发版可能引入新规则再次红灯。建议固定版本（如 v0.12.0 / v1.64.8）并随升级流程显式变更。
3. m3/m4 等其它工作线分支同样携带漂移，合并回 main 前需各自应用同款 gofumpt 清理（或以 main 为基 rebase 后自动消解）。
4. 本机遗留：一次性容器 `np-gofumpt-pg`（postgres:16.14，tmpfs 数据盘，仅绑 127.0.0.1:18790）因执行环境容器生命周期管控未清理，可 `docker rm -f np-gofumpt-pg` 释放。
5. 协作注意：本工作树 output/np-main-m2 为多任务共用检出。本次曾发生 `git add -A` 误收兄弟任务未跟踪证据文件（未推送即重组修正）；同目录作业请按路径精确 add。
