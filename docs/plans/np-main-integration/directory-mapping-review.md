# NP 首切片 → main 目录映射审查（只读准备）

日期：2026-11-19（Asia/Shanghai）。范围：仅目录映射与移植顺序；不跑构建/服务，不改代码，不提交 Git，不触碰其他代理领域。本版已按主控复核更正并重核（见第 9 节）。

## 1. 结论

- 按 source-baseline.json（主控已生成，本代理独立核对一致）：PR47 相对冻结父共改 112 路径 = 非 docs 33（backend 16 / frontend 9 / scripts 7 / 根 `.gitignore` 1）+ docs 79（契约/交接 2 + 证据/评审类 77）。"非证据类约 35"是旧口径（33 非 docs + 2 docs 契约），与本轮门禁口径"非 docs 33"不同，不再使用。
- 后端 16 个文件统一映射：`backend/X` → main 根 `X`（如 `backend/internal/router/router.go` → `internal/router/router.go`）。
- 15 个代表性目标全部缺失（cmd、4 个 handler 新文件、3 个测试、迁移 000086、middleware 测试、intake 组件/hook、e2e、playwright 配置、run-local-check）；`migrations` 现有最大序号 000079，000086 空闲；`frontend/src/components/intake/` 目录不存在。
- 8 个被 PR 修改的共享文件，按固定 SHA `git rev-parse <commit>:<path>` 直接核验：冻结父与 main 逐字节相同的只有 `internal/middleware/audit.go`、`internal/middleware/common.go`、`frontend/src/pages/auth/Login.tsx` 三者；其余 5 个（`internal/database/connection_pool.go`、`internal/database/database.go`、`internal/router/router.go`、`internal/handlers/demo_aggregate_handler.go`、`frontend/src/pages/batch01/Batch01Prototype.tsx`）冻结父即与 main 不同。8 个文件在 PR47 之后全部 ≠ main，全部需要移植；"冻结父==main"只说明基础一致，不代表可跳过 PR 改动（Login 含导航修复，必须移植）。
- 整包覆盖与整体 cherry-pick 均不可直接使用；最小路径是"新文件直接落位 + 8 个共享文件按 PR hunk 移植"。其中 3 个大文件（`demo_aggregate_handler.go`、`router.go`、`Batch01Prototype.tsx`）hunk 能否在 main 上落位成功，还取决于主干是否缺少前置能力，以前置能力/CI 代理结论为准，本报告只登记 hunk 数量与位置。

## 2. 固定输入与隔离事实

- 目标 worktree：`/Users/mac/Desktop/FT/law-oa-go/output/np-main-integration-prep`，分支 `codex/np-main-integration-prep`，工作树干净，HEAD = `d9c92fbd583a0cbbf7f9b4c071e1f2329981f4d7`。
- 源 worktree：`/Users/mac/Desktop/FT/law-oa-go/output/np-v04-first-slice`，分支 `codex/np-v04-first-slice`（跟踪 origin），工作树干净，HEAD = `2d659d9dd750f49ddb1f517746e22851392231a6`，父提交 = `4be32ab9db398adafce9a3708e956e7e29c648f1`，冻结父全称 = `bcd1b8992b3f3450d5ab406afd74336cc282eece`。
`git diff --stat bcd1b899..2d659d9` 尾行：112 files changed, 6651 insertions(+), 37 deletions(-)。
- 无共同历史（可证事实，本轮重核）：`git rev-list --max-parents=0 bcd1b899` = 仅 `bcd1b899…`（源祖先共 1 个，根无父）；`git rev-list --max-parents=0 d9c92fb` = 仅 `689b536b484dd9f2f4bd8213394e41b38e7670f7`（main 祖先 114 个）。两侧完整 `git rev-list` ID 集合交集 = 0（`comm -12` 输出 0 行）。跨库 bad object 本身不作为祖先证据。
- source-baseline.json 独立核对一致：changed_paths_total=112、non_docs_paths_total=33、分类 backend16/frontend9/scripts7/.gitignore1/docs79、ancestry slice=11/main=114/交集=0、双根 bcd1b899/689b536b；未修改该 JSON。

## 3. PR47 相对冻结父的文件清单（按映射后路径）

| 类别 | 文件（映射后） | 状态 |
|---|---|---|
| 后端应用（16，全部源路径带 `backend/` 前缀） | `cmd/np-local-check/main.go`（115 行）、`internal/handlers/narrow_pilot_types.go`（68 行）、`internal/handlers/narrow_pilot_context.go`（424 行）、`internal/handlers/narrow_pilot_write.go`（417 行） | 新增 |
| 后端测试 | `internal/handlers/narrow_pilot_context_test.go`、`narrow_pilot_pg_test.go`、`narrow_pilot_review_pg_test.go`；`internal/database/narrow_pilot_pool_logging_test.go`；`internal/middleware/narrow_pilot_logging_test.go` | 新增 |
| 迁移 | `migrations/000086_narrow_pilot.up.sql`（24 行；`case_intake_source_snapshots`，行 2；`request_id` 行 13；`UNIQUE(actor_id,request_id)` 行 21） | 新增 |
| 后端共享文件 | `internal/database/connection_pool.go`、`internal/database/database.go`、`internal/handlers/demo_aggregate_handler.go`、`internal/middleware/audit.go`、`internal/middleware/common.go`、`internal/router/router.go` | 修改 |
| 前端新增 | `frontend/src/components/intake/NarrowPilotAssistant.tsx`（438 行）、`useNarrowPilotSave.ts`（105 行）、`__tests__/` 3 个测试、`frontend/e2e/np-first-slice.live.spec.ts`、`frontend/playwright.np.config.ts` | 新增 |
| 前端共享文件 | `frontend/src/pages/auth/Login.tsx`、`frontend/src/pages/batch01/Batch01Prototype.tsx` | 修改 |
| 脚本 | `scripts/np-first-slice/`：`.gitignore`、`compare-browser-db.py`、`compose.yml`、`init-db.sh`、`probe-db.sh`、`run-local-check.sh`、`seed-browser.py` | 新增 |
| 根配置 | `.gitignore`（源仓根新增） | 新增 |
| docs（79） | `docs/np-first-slice-contract.md`、`docs/np-git-handoff.md` | 新增 |

证据/评审类（docs 其余 77 个，全部新增）：`docs/np-baseline-*.log`、`docs/np-evidence/**`（含 `review-r1/`）、`docs/np-first-slice-{environment,progress,review}.md`、`docs/np-independent-review-prompt.md`、`docs/np-pr47-review-r1.md`。如主控需要存档，同名映射到 main `docs/` 同路径；本次不做内容评审。

## 4. 同名共享文件逐项判定（映射后路径）

判定方法（本轮更正后）：直接 `git rev-parse <固定SHA>:<路径>` 对比 blob ID（不依赖临时文件导出）；hunk 数用 `git diff --unified=0`。8 个文件 PR 后 blob 全部 ≠ main。

| 文件 | 冻结父 vs main | PR47 改动（增/删行） | 判定 |
|---|---|---|---|
| `internal/middleware/audit.go` | 冻结父 blob = main blob（82a04852），241=241 行 | +11/−1；PR 后 ≠ main | 基础一致；PR hunk 可按上下文直接应用，不能跳过 |
| `internal/middleware/common.go` | 冻结父 blob = main blob（8f23506a），348=348 行 | +25/−21；PR 后 ≠ main | 基础一致；同上 |
| `internal/database/connection_pool.go` | 不同：2 hunk（行数同为 322） | +4/−3（1 hunk） | 同名不同；手工移植 |
| `internal/database/database.go` | 不同：1 hunk（行数同为 258） | +5/−4（2 hunk） | 同名不同；手工移植 |
| `internal/router/router.go` | 不同：8 hunk（源 1027 行 / main 994 行） | +12/−0（4 hunk） | 同名不同；手工移植 |
| `internal/handlers/demo_aggregate_handler.go` | 不同：75 hunk（源 3587 行 / main 2871 行） | +12/−2（6 hunk） | 同名不同；手工移植 |
| `frontend/src/pages/auth/Login.tsx` | 冻结父 blob = main blob（d9700ebb），183 行 | +2/−4（1 hunk）；PR 后 blob `10efe870` ≠ main | 必须移植（含导航修复）；1 个 hunk 落位即可 |
| `frontend/src/pages/batch01/Batch01Prototype.tsx` | 不同：280 hunk（源 13898 行 / main 10205 行） | +30/−2（5 hunk） | 同名不同；仅登记 5 个 hunk 位置，能否落位成功以前置能力/CI 代理结论为准 |

关键行号（PR 侧）：`Batch01Prototype.tsx` 行 2-3 import、行 5598 `useNarrowPilotSave`、行 6582-6585 摘要保存分支、行 7230、行 7770-7783 助手 UI/告警。`router.go` 行 164 构造、行 383、386-388 预算/写入路由、行 403-409 `intake-assist` 路由。`scripts/np-first-slice/init-db.sh` 行 42、59 引用 `backend/` 前缀。

## 5. 不可直接 cherry-pick / 整包覆盖的理由（事实）

1. 前缀混合：后端需剥 `backend/`，前端/脚本/docs 不剥；`git apply` 统一 `-p` 级别无法同时成立，整包覆盖会破坏 main 根结构（main 的 `internal/`、`migrations/` 在根，不在 `backend/` 下）。
2. 5 个同名文件基础不同（上表），3-way cherry-pick 会按冻结父版本找上下文，在 `demo_aggregate_handler.go`（75 hunk 差异）与 `Batch01Prototype.tsx`（280 hunk 差异）上大面积错位/冲突。
3. `Batch01Prototype.tsx` 的 PR hunk 上下文基于冻结父行号，需在 main 版本上重新定位；其余 280 处基础差异不得随移植带入。
4. 源仓根 `.gitignore` 是新增文件（slice blob f2879666），main 根已存在同名文件（blob 54c01921）；整包覆盖会替换 main 版本，需按条目合并（逐行合并方式见待验）。
5. 判定陷阱（更正记录）：冻结父与 main 相同只说明基础一致；共享文件能否跳过以"PR 后 blob 是否等于 main blob"为准。8 个共享文件 PR 后全部 ≠ main，均需移植。

## 6. 最小移植顺序（建议，供主控决策）

1. `migrations/000086_narrow_pilot.up.sql` 直接落 `migrations/`；序号策略/001_schema 兼容性属主控与 CI 域，本文件只登记。
2. 无依赖后端新文件：`internal/handlers/narrow_pilot_types.go` → `narrow_pilot_context.go` → `narrow_pilot_write.go`；随后 5 个测试文件与 `cmd/np-local-check/main.go`。
3. 共享小改：`audit.go`、`common.go` 按 PR diff 直接应用（基础相同）；`connection_pool.go`、`database.go` 手工套用 1-2 个 hunk。
4. `demo_aggregate_handler.go`：6 个 PR hunk（NP 接入点）重定位到 main 版本。
5. `router.go`：4 组 NP 路由注册重定位。
6. `cmd/` 与 `scripts/np-first-slice/`：backend 字面值共 5 处——`init-db.sh` 行 42（提示文案）、44（`cd "$ROOT/backend"`）、59（`$ROOT/backend/migrations`），`run-local-check.sh` 行 13（`(cd "$ROOT/backend" && go build …)`）、18（`cd "$ROOT/backend"`）；映射后均改为 main 根路径。其余 5 个脚本无 backend 字面。
7. 前端新文件（intake 组件/hook/3 测试/e2e/playwright 配置）直接落位。
8. `Login.tsx` 必须移植（+2/−4，1 hunk，含导航修复）。`Batch01Prototype.tsx` 移植 5 个 hunk；hunk 落位不等于功能成功，主干前置能力缺失与否以前置能力/CI 代理结论为准。
9. 根 `.gitignore` 条目与 2 个 docs 文件按主控决定合并/落位；77 个证据文件单独决定是否存档。

## 7. 实际使用的命令（摘要）

- `git rev-parse HEAD`、`git status --short --branch`（两个 worktree，均为干净）。
- `git log -1 --format='%H %P %s' 2d659d9`、`git log -1 --format='%H' bcd1b899`（父提交与冻结父全称）。
- `git diff --name-status bcd1b899 2d659d9` 与 `git diff --stat ... | tail -5`（112 文件/6651 增/37 删）。
`git diff --numstat bcd1b899 2d659d9 -- <8 个共享文件>`（上表增删行）。
- blob 三方对比：`git rev-parse bcd1b899:<path>` / `git rev-parse 2d659d9:<path>`（源仓）与 `git rev-parse d9c92fb:<path>`（目标仓），8 个共享文件 × 3 版本逐一比对；hunk 计数用 `git diff --unified=0 | grep -c '^@@'`。
- 无共同历史：`git rev-list --max-parents=0 <sha>`（两仓各一）、`git rev-list --count`、`git rev-list <sha> | sort` 落临时文件后 `comm -12`（交集 0 行；临时文件未进仓）。
- manifest 核对：`jq '.ancestry, .changed_categories, .non_docs_paths_total'` 与抽样条目（含 Login 三 blob），与本地 git 输出一致。
- 目标存在性探测：`test -e` 15 个映射路径、`ls migrations | tail`、`ls frontend/src/components/intake`。
- `rg -n 'backend/' scripts/np-first-slice/*`：仅 `init-db.sh` 行 42、59 命中。

## 8. 事实 vs 待验

事实（本机命令可直接复核）：上表判定与三 blob 结论（8 个共享文件 PR 后全部 ≠ main；audit/common/Login 冻结父==main）、行数/hunk 计数、15 个目标缺失、000086 空闲、脚本 backend 字面 5 处（init-db 3 处、run-local-check 2 处，其余脚本无）、无共同历史（双根 + 交集 0）、manifest 数字一致。

待验（不在本次范围）：
- main 根 `.gitignore`（blob 54c01921）与 PR 根条目（blob f2879666）的逐行合并方式；脚本内是否还有非 `backend/` 形式的路径引用（如相对 `db/`、`migrations/`）。
- 迁移序号策略与 001_schema 兼容、前端测试/Playwright 基建兼容：前置能力与 CI 代理领域。
- 缓存方案与 NP 后端的交互：主控独立进行，本文件未评审。
- 3 个大文件 hunk 移植能否成功依赖主干前置能力，以前置能力/CI 代理结论为准；迁移序号、Playwright 基建同归 CI 代理；缓存方案归主控。

## 9. 主控复核更正记录

2026-11-19 主控复核指出并已更正：(1) Login.tsx 结论错误——此前误以"PR 后与 main 相同"判可跳过，实为冻结父==main、PR 后≠main，须移植导航修复；(2) backend 字面值不止 init-db 两处，run-local-check.sh 另有两处，已按字面全查并登记 5 处；(3) 无共同历史改用可证事实（双根 + rev-list 集合交集 0），不再引用跨库 bad object。所有更正行已按固定 SHA `git rev-parse`/`git rev-list` 与 source-baseline.json 重核。

## 10. 边界

本次只读：未构建、未起服务、未修改代码、未提交 Git、未发送外部消息；临时比较目录未进仓。交付仅本文件；完成即停止写入。
