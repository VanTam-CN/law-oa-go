# main 目标集成 PR 事件核验（PR #52）

任务 t_ddd82725 遗留项收口：建立 main 目标集成 PR 并登记真实 pull_request 事件的 head/base/test-merge 三时点 SHA、逐 job 结果、缓存/artifact/通知等外部动作边界。完整 I11 分支线已由 t_d83a4c85 裁定 PASS，本报告只覆盖 main 目标 PR 线。

## 结论

**PASS。** PR #52（`codex/np-main-i11` → `main`，Draft）于 2026-11-24T11:59:49Z 建立，6 个 workflow 的 pull_request 事件 run 全部触发并被完整记录；np-i11 门禁 job 在 PR 事件上 **GREEN**（go=32、tagged=11、jest=124、`I11_FULL_PASS`）；三时点 SHA 经 job 留痕与 PR API 交叉一致；副作用边界全零；全部失败完成分类（pre-existing 9、new 1、by-design 红 2 类，unclassified=0）。

## PR 与三时点 SHA

| 项 | 值 | 来源交叉 |
|---|---|---|
| PR | [#52](https://github.com/VanTam-CN/law-oa-go/pull/52)（Draft、OPEN、MERGEABLE、0 评论） | API |
| head | `309d1dd4559a434ae9f65b82d3a6154499b3b95f`（codex/np-main-i11，I11 通过候选） | PR API = np-i11 job `I11_REFS head=` |
| base | `main` @ `d9c92fbd583a0cbbf7f9b4c071e1f2329981f4d7` | PR API = `I11_REFS base=`；与集成计划固定基线一致，main 未前进 |
| test-merge | `4c5db35c62c3ebc49345ce0fd3691a00128e30db` | PR API `merge_commit_sha` = np-i11 `I11_EVENT pull_request checkout=`（即 GITHUB_SHA） |

计划 §5 合同兑现：PR 事件的 GITHUB_SHA 是合并测试提交（4c5db35c），与分支 head（309d1dd4）分别记录，未混称。52f9d40a 预备的同仓 PR 事件门禁与 I11_REFS 留痕机制首次在真实 PR 事件上兑现。

## 事件面（6 run，11:59:53Z 注册）

| run | workflow | 结论 | 关键 job |
|---|---|---|---|
| 45389286869 | tests.yml | failure | **np-i11 (full) success**；Go Tests/Frontend Tests/E2E failure（分类见下）；np-m2-early、coverage-check skip；test-summary success |
| 45389287095 | ci.yml | failure | Code Quality=gofumpt(pre)；build×2、deploy×2、notify 全 skip |
| 45389286968 | ci-cd.yml | failure | 集成静态分析 success；unit=failure(mysql unit 缺失)、integration=cancelled(fail-fast)、benchmark/fuzzing failure(见分类)；docker/deploy/report skip |
| 45389286887 | pgo.yml | failure | 检查PGO支持 failure(go build --help 探针损坏，main schedule 连续 5 次同签名)；pilot-pgo failure(tests/performance 失败包传染)；collect 链 skip |
| 45389287032 | security.yml | **success** | smoke+依赖扫描+报告全绿（upload 全部被条件禁用） |
| 45389286873 | fuzzing.yml | failure | preflight(默认 all 路由)+check success；4×fuzz job failure(runner 缺 redis-server/mysql unit)；**fuzzing-summary failure = by-design 正确红**（选中 job 失败汇总不得假绿——I11 汇总语义在真实事件上验证） |

## 失败分类（unclassified=0）

- **pre-existing（9）**：TestPerformanceMonitoring（tests/performance 未被 i11 改动，本地 darwin/go1.25.6 复现，main 09-21 同 SHA run 已红）；E2E Playwright 不支持 ubuntu26.04-x64（runner 镜像漂移）；gofumpt 漂移（F4/F5 同签名，t_7cecb4f1 清理中）；mysql/redis-server unit 缺失（runner 漂移，fuzzing 修复版路由本身正常）；pgo 探针与 Fuzzing 版本探针脆弱性（main 原样）；benchmark 与 pilot-pgo 为失败包传染。
- **new（1）**：**Frontend Tests 集成适配缺口**——m3/m4 线放入 `frontend/np-m4/`、`frontend/np-entry/` 的 Playwright spec 被 main 时代 Jest testMatch 拾取（`Class extends value undefined` 套件失败），且 `tsconfig.eslint.json` 不含这些目录（ESLint 解析错误）。np-i11 自身 jest group(124) 不受影响；此缺口合并后会让 main 原始 Frontend job 持续红。已开 followup 卡。
- **by-design 红（2 类）**：fuzzing-summary（选中失败→汇总必须红）；各门禁 skip（np-m2-early/coverage-check/build/deploy/notify/docker/report）。

## 副作用边界（终查）

- artifact：6 run 全 **0**（upload 步骤全部被 52f9d40a 条件禁用）
- 缓存：仓库共 3 个全部在 `refs/heads/main`（历史遗留）；`refs/pull/52/merge` 为 **0**；PR 事件零读零写（setup-go cache:false + actions/cache 条件 skip）
- Codecov：0 上传；通知 job：skip；docker push：skip（仅 push/release）；deployments：0；releases：0；PR 自动评论：0

## 与验收矩阵的关系

本次兑现的是 I11 的「真实 PR 事件载体」缺口（分支 push/dispatch 不能替代的事件级核对）：三时点 SHA、逐 job、缓存/artifact/外部动作边界全部在录。M5（真实 main+Redis 联调，I10 真实链路）与 M6（I12 演练）仍为后续独立工作项，本 PR 不构成合并授权；保持 Draft。

## 证据

- `verification.json`：结构化结果（本目录）
- `raw/`：第一手快照（PR API 三时点、I11_REFS/EVENT log、逐 job 快照、各失败 log 签名、main 历史 run 对比、caches/artifacts 终查）
- 关联：docs/np-evidence/m2/full-i11/（托管+本机完整 I11）、docs/plans/np-main-integration/（计划与固定基线）
