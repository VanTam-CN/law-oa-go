# M2 完整 I11 与真实事件复核 · 汇总裁定

本报告为任务 `t_d83a4c85` 的最终交付：汇总两个前置任务（托管侧 `t_b938e756`、本机核验 `t_92dc3e8f`）与早期 M2/M1/规划期证据，对真实事件 run 中出现的每一个失败给出 pre-existing / new 分类，并给出最终判定。结构化判定见 [verification.json](verification.json)；归档清单见 [MANIFEST.sha256](MANIFEST.sha256) 与 [inventory.json](inventory.json)；本次第一手补录的远端快照见 [snapshots/](snapshots/)（UTC 时间戳）。

## 一、最终判定：通过（PASS）

- **完整 I11 门禁 PASS**。托管侧：修复后两轮真实 push 事件 [45379241429](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379241429)（head `0ca6eb2c`）与 [45379855010](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379855010)（head `52f9d40a`）均 `I11_FULL_PASS`：go=32、tagged=10、jest=124、浏览器 9 套件 34 例、PG 保全摘要 `b54a6df9…ff402d`（sources=3 receipts=4 linked=3 manual=1）。本机侧：`t_92dc3e8f` GO——浏览器链六环全通，必跑集 Go 32 + tagged 11@`52f9d40`（含首次执行的 TestAsyncPGCompletionAfterNativeEdit）+ Jest 124 + 浏览器 34 例全部恰一次执行并通过，15 项负向用例全部兑现期望失败。
- **真实事件复核 PASS**。8 个登记的真实 push/dispatch 事件 + PR #51 事件边界（head/base/test-merge 三时点在录；六套 workflow 的 pull_request 过滤均为 main/develop，指向 m4 的 PR 零触发属记录在案的过滤事实，非跳过门禁）全部可追溯；窗口边缘两个 byproduct run（`45381532273` i11 分支、`45382638826` m2 分支）经本次第一手核验均为 success。
- **失败全部定性，unclassified = 0**。详见下表。
- **「真实事件尚未核验」状态：显式关闭**（就 I11 分支线与真实草案 PR #51 而言，真实事件已核验并留痕）。**延续项**：main 目标集成 PR 事件从未创建，属独立登记的后续验收项（`t_ddd82725` 持有；同仓 PR 门禁与 `I11_REFS` head/base 留痕机制自 `52f9d40a` 就绪）——分支 push/dispatch 事件不能替代该项。

## 二、失败分类表（真实事件 run）

### A 组 · I11 分支真实事件（PR #51 线，观察窗口 2026-11-24T10:47:35Z–11:16Z）

| # | Run / 事件 | 失败 | 分类 | 依据与链接 |
| --- | --- | --- | --- | --- |
| F1 | [45378121147](https://github.com/VanTam-CN/law-oa-go/actions/runs/45378121147) push@`076f7e7` full | `np-i11 (full)` job 红灯：`json.decoder.JSONDecodeError: Expecting value … at scripts/np-ci/check-i11-tests.py:48` | **new** | 校验器由同一提交 `076f7e7` 引入，`full-i11.sh` 将 stderr 工具诊断并入 go test -json 结构化流；产品 go 套件本身 rc=0。全库证据 grep：该签名仅存在于本窗口托管证据，与全部历史记录（H1–H6）无同签名。raw/diag-157-full.log 第 13–15 行。修复 `0ca6eb2c`（stderr 分离）后两轮 full GREEN |
| F2 | [45378568403](https://github.com/VanTam-CN/law-oa-go/actions/runs/45378568403) push@`374146c` full | 与 F1 完全相同的 traceback（确定性复现） | **new（同一缺陷）** | 同签名同行号；raw/diag-159-full.log；同一修复闭环 |
| F3 | 负向 dispatch [45378128854](https://github.com/VanTam-CN/law-oa-go/actions/runs/45378128854) / [45378578245](https://github.com/VanTam-CN/law-oa-go/actions/runs/45378578245) / [45379251327](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379251327) / [45379862576](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379862576)，4×8 探针全红 | 每个 job 命中指定独立诊断码后退出 1：`M2_PG_UNAVAILABLE`、`M2_DSN_MISSING`、`M2_REQUIRED_TEST_SKIPPED`、`M2_REQUIRED_TEST_MISSING`、`M2_NO_TESTS`、`I11_FUZZ_PREFLIGHT_REJECTED`、`I11_FUZZ_SUMMARY_REJECTED`×2 | **by-design 预期红灯**（验证手段，非缺陷；不入 pre-existing/new 缺陷分类） | 32/32 探针恰好命中各自要求的诊断码且错误可读，无附带失败；hosted/README.md 事件表 + verification.json hosted_events |
| — | [45381532273](https://github.com/VanTam-CN/law-oa-go/actions/runs/45381532273) push@`309d1dd4`（docs binding 触发的 byproduct，窗口边缘未登记） | 无失败（success） | — | 本次第一手补录：snapshots/byproduct-run-45381532273-i11.json、snapshots/runs-i11-branch.json |

### B 组 · 临时隔离探针 PR 的 pull_request 事件（窗口内发生、前置任务未登记，本次补录并分类）

| # | Run / 事件 | 失败 | 分类 | 依据与链接 |
| --- | --- | --- | --- | --- |
| F4 | [45334857440](https://github.com/VanTam-CN/law-oa-go/actions/runs/45334857440) PR #49（m2→main 探针，head `448047e3`，现已 closed）pull_request | ci.yml「Code Quality / 代码格式检查」`gofumpt -w -s . && git diff --exit-code` 退出 1；diff 为 `tests/integration/m1/workbench_approval_browser_test.go`（`0600`→`0o600`）及 M1 期测试文件长 `[]string` 折行 | **pre-existing（历史遗留）** | `448047e3` 为纯 docs 提交未触碰 Go；本地第一手复现（工作树 `bfcb7de1`，Go 内容与 `448047e3` 一致）：`gofumpt -l` 命中 tests/ 5 文件 + internal/ 139 个遗留文件，裸 `gofmt -l tests/` 零告警（漂移为 gofumpt 特有规则）；失败文件最后改动为 M1 期提交（f72b1aa、54f826c）。该 job 是试点 PR 隔离设计中**保留**的既有质量检查，不属 np-i11 门禁。snapshots/pr49-run-45334857440.json、pr49-…-failed.tail220.txt、gofumpt-local-reproduction.txt |
| F5 | [45334868179](https://github.com/VanTam-CN/law-oa-go/actions/runs/45334868179) PR #50（m3→main 探针，head `c9415b81`，现已 closed）pull_request | 同一步骤、同一工具、同签名失败（日志含 `0o600` diff 与 exit code 1） | **pre-existing（与 F4 同源）** | 同上；snapshots/pr50-run-45334868179.json、pr50-…-failed.tail220.txt |
| — | 同两事件的 3+5 个 cancelled run（ci-cd/fuzzing/security 等） | cancelled，非失败 | 不分类（记录在案） | runs-pull-request.json |

### C 组 · 比对基准：历史 PGO/CI 故障记录

| # | 记录 | 签名 | 位置 | 与 A/B 组比对 |
| --- | --- | --- | --- | --- |
| H1 | fuzzing.yml workflow-file 失败 | jobs=0 / check-runs=0，"likely workflow file issue" | run [44508335288](https://github.com/VanTam-CN/law-oa-go/actions/runs/44508335288)（主干 `26355fc`）、[45159753188](https://github.com/VanTam-CN/law-oa-go/actions/runs/45159753188)（dev `9e94e0a`，0 job）；docs/plans/np-main-integration/remote-observation.json、ci-review.md；m2/early-ci/verification.json | 无同签名（M2 early-ci 内已修复 `fa4a790`） |
| H2 | 审批编号唯一键碰撞 DEFECT-NP-M1-REQNUM-001（卡片 t_462633c0） | duplicate key `idx_approval_requests_request_number`；workbench/TestM1R3ApprovalPartySnapshot HTTP 500 | m1/review-r5/README.md 门禁段；看板卡 t_462633c0 | 无同签名（托管 diag 日志 grep `duplicate key`/`idx_*` 0 命中）；登记缺陷，不在本分支处置 |
| H3 | M1 R3 快照失败轮 | 主体快照/审批清单错位、HTTP 500 失败轮 | m1/review-r3…r5/README.md | 无同签名 |
| H4 | idx_users_username 空库演示初始化碰撞（PGO 隔离试点发现） | duplicate key `idx_users_username` | m2/pr-isolation/README.md | 无同签名（PGO 基线用不可登录合成用户绕开；另计） |
| H5 | 历史密文与摘要错配（卡片 t_f7991d61） | 摘要-密文对漂移 | m1/review-r5/README.md 门禁段 | 无同签名；登记缺陷，不在本分支处置 |
| H6 | 全仓 gofumpt/gofmt 遗留漂移（F4/F5 的直接基准） | gofumpt 特有规则 diff | snapshots/gofumpt-local-reproduction.txt（internal/ 139 + tests/ 5） | pre-existing，先于窗口全部提交 |

**Unclassified：无。** 所有失败均可定性：1 个 new（已修复并两轮 GREEN 验证）、32 个 by-design 预期红灯、2 个 pre-existing（临时探针 PR 上的遗留格式漂移，不属 np-i11 门禁）。

## 三、早期证据归档清单

归档范围、来源与校验：[MANIFEST.sha256](MANIFEST.sha256)（SHA-256，118 文件）、[inventory.json](inventory.json)（逐文件哈希/字节数/路径/来源分组）。

| 组 | 来源 | 文件数 | 说明 |
| --- | --- | --- | --- |
| m2_hosted_full_i11 | docs/np-evidence/m2/full-i11/hosted/（t_b938e756，证据提交 `bfcb7de`） | 44 | 8 事件逐 run/jobs/artifacts 快照、diag 日志、workflow 源码、PR/caches/webhooks、verification.json |
| m2_local_browser_chain_verify | docs/np-evidence/m2/i11-browser-chain-verify/（t_92dc3e8f） | 38 | 本机 GO 报告 + run1/run2 快照 + my-window 第一手（归档前未入 git，本任务随本次证据提交固化） |
| m2_early_ci | docs/np-evidence/m2/early-ci/ | 2 | 早期 CI 验收（含 fuzzing 45159753188 记录、负向预期红、复跑边界） |
| m2_pr_isolation | docs/np-evidence/m2/pr-isolation/（t_cc50f4c9） | 3 | 试点 PR 隔离验收（含 idx_users_username 发现、actionlint 既有提示） |
| m1_review_rounds | docs/np-evidence/m1/（README + review-r3…r10） | 9 | 历史失败轮与两处登记缺陷（t_462633c0、t_f7991d61）的载体 |
| plans_np_main_integration | docs/plans/np-main-integration/ | 4 | fuzzing 44508335288 历史记录、CI 评审、验收计划 |
| final_review | docs/np-evidence/m2/full-i11/final-review/（本任务） | 18（另含本清单自身） | 本报告、verification.json、清单、snapshots/（15 份 UTC 快照） |

不可变性：全部文件以 SHA-256 校验清单固化；`i11-browser-chain-verify` 与 `final-review` 两目录由本任务证据提交绑入 git 历史；远端快照带 UTC 时间戳（snapshots/snapshot-time.txt，2026-11-24T11:37Z 起）。日志固定位说明：仓库 .gitignore 忽略 `*.log`（全仓零跟踪先例），hosted raw 的 diag 日志与 run2/my-window 的原始 .log 留在工作区并由 MANIFEST.sha256 记录哈希；本裁定依赖的关键失败签名均已转为可跟踪文件入库（defect-signature-diag157.txt、pr49/pr50 失败日志尾部 .txt、gofumpt-local-reproduction.txt）。

## 四、遗留风险与后续跟进

1. **main 目标集成 PR 事件未发生**（`t_ddd82725` 持有）：head/base/test-merge 实际值与事件级副作用（含浏览器链的 PR 运行）待该 PR 创建时核对；`52f9d40a` 已备好同仓 PR 门禁与 `I11_REFS` 留痕。
2. **gofumpt 遗留漂移未清理**（F4/F5）：internal/ 139 + tests/ 5 文件；任何命中 main/develop 过滤的 PR 事件都会在「代码格式检查」红灯。已登记后续清理卡，不阻塞 I11 判定。
3. **全量 PGO 性能结果与 M5 真入口验收未执行**：pr-isolation 明确试点不据此推断性能提升或 M5 验收。
4. **run#2 native 套件 JSON 快照竞态缺口**：已由 `t_92dc3e8f` 自有窗口 native 10 例第一手覆盖；托管侧 native 证据以该窗口与 45379855010 绿链为准。
5. **运维残留**：4 个退出态容器 `np-i11-verify-db{,2,3,4}` 待 `docker rm`；宿主 `/usr/local/bin python3` 损坏（Killed:9），夹具链以 `/usr/bin/python3`（3.9.6）运行正常，建议修复。
6. **登记在案、不属本分支的两专项缺陷**：审批编号唯一键碰撞（t_462633c0）、历史密文与摘要错配（t_f7991d61）——保持登记状态，处置另卡。
7. **不可知项（按设计记录）**：GitHub 用户侧邮件/watch 通知状态；PR #49/#50 上 8 个 cancelled 兄弟 workflow 的并发取消确切触发者（cancelled 非失败，无失败步可归因）。

## 五、证据时间线（UTC）

| 时点 | 事件 |
| --- | --- |
| 2026-11-23 10:50Z | 早期 CI：dev 分支 fuzzing 0-job 失败 45159753188（H1 第二实例）→ 同日修复 `fa4a790` |
| 2026-11-24 03:12Z | 试点 PR 隔离 push/dispatch 验收（45317903182 正常 / 45317907862 五负向 / 45318673293 tip） |
| 2026-11-24 05:39Z | 临时探针 PR #49/#50 pull_request 事件：F4/F5 gofumpt 失败 + 8 cancelled（本报告补录） |
| 2026-11-24 10:47–11:16Z | I11 托管窗口：两轮红（F1/F2）→ 修复 `0ca6eb2c` → 两轮 GREEN；负向 32/32 预期红；PR #51 三时点在录 |
| 2026-11-24 11:09Z | `309d1dd4` docs binding 触发 byproduct 45381532273（success，本次补录） |
| 2026-11-24 11:16Z | `bfcb7de` 证据推送触发 byproduct 45382638826（success） |
| 2026-11-24 11:31–11:40Z | 本任务第一手快照与裁定（snapshots/snapshot-time.txt） |
