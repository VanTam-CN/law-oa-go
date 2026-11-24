# M2 完整 I11 托管侧执行与真实 PR 事件记录

本目录记录完整 I11 的**托管侧**执行：真实分支事件（push 与 workflow_dispatch）上的完整门禁运行，以及一次真实 PR 事件（#51）的事件与执行边界。全部数据取自 GitHub API 与 job 日志的实抓快照，带时间戳与 run id；本机 full-gate 执行属 t_92dc3e8f，最终汇总裁定属 t_d83a4c85。

结构化结论见 [verification.json](verification.json)；原始 API 快照、逐 job 诊断与 workflow 源码快照见 [raw/](raw/)。

## 事件清单（观察窗口 2026-11-24T10:47:35Z 起）

| Run | 事件 | head | kind | 结论 | 判定 |
| --- | --- | --- | --- | --- | --- |
| [45378121147](https://github.com/VanTam-CN/law-oa-go/actions/runs/45378121147) | push | `076f7e7` | full | failure | 门禁校验器缺陷红（各前置阶段均 rc=0，产品测试未失败） |
| [45378128854](https://github.com/VanTam-CN/law-oa-go/actions/runs/45378128854) | dispatch negative | `076f7e7` | 负向 | failure | 8/8 探针按预期红灯，逐 job 命中独立诊断码 |
| [45378568403](https://github.com/VanTam-CN/law-oa-go/actions/runs/45378568403) | push | `374146c` | full | failure | 同上缺陷复现（确定性） |
| [45378578245](https://github.com/VanTam-CN/law-oa-go/actions/runs/45378578245) | dispatch negative | `374146c` | 负向 | failure | 8/8 预期红灯 |
| [45379241429](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379241429) | push | `0ca6eb2c` | full | **success** | 首次托管完整门禁 GREEN（I11_FULL_PASS） |
| [45379251327](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379251327) | dispatch negative | `0ca6eb2c` | 负向 | failure | 8/8 预期红灯 |
| [45379855010](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379855010) | push | `52f9d40a` | full | **success** | GREEN（PR head 提交上的完整门禁） |
| [45379862576](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379862576) | dispatch negative | `52f9d40a` | 负向 | failure | 8/8 预期红灯 |

负向矩阵累计 4×8=32 个探针执行，32/32 命中各自要求的独立诊断码并退出 1：`M2_PG_UNAVAILABLE`、`M2_DSN_MISSING`、`M2_REQUIRED_TEST_SKIPPED`、`M2_REQUIRED_TEST_MISSING`、`M2_NO_TESTS`、`I11_FUZZ_PREFLIGHT_REJECTED`、`I11_FUZZ_SUMMARY_REJECTED fuzz-selected-failed`、`I11_FUZZ_SUMMARY_REJECTED fuzz-missing-proof`。每个 job 的 checkout 身份由 `I11_EVENT … checkout=<GITHUB_SHA>` 行留痕。

## 正常链（full）绿证据

[diag-161-full.log](raw/diag-161-full.log) 与 [diag-163-full.log](raw/diag-163-full.log) 逐阶段记录：fuzz-normal、build、redis（`TestI11ResponseCacheRealRedis` 有 PASS 凭证后放行）、go（required=32）、tagged（required=10）、typecheck、jest（required=124）、frontend-build、九个浏览器套件 route-b/selection/storage/navigation/editing/tabs/legacy/prod/native 共 34 例，route-b 另产出 PG 保全摘要 `b54a6df9…ff402d`（sources=3 receipts=4 linked=3 manual=1），最终 `I11_FULL_PASS`。

## PR 事件边界

真实草案 PR [#51](https://github.com/VanTam-CN/law-oa-go/pull/51)（`codex/np-main-i11` → `codex/np-main-m4`，镜像 #48 里程碑间草案模式）：

- head / base / test-merge 实际值（三个读取时点）见 verification.json 的 `pr_event`；最终读取 head=`309d1dd4`、base=`2989b285`、test-merge=`cbf21682`。
- `pull_request` 事件**零触发**：六套 workflow 的 pull_request 过滤均为 main/develop（pgo 为 main），指向 m4 的 PR 不产生任何 check-run、评论或上传；这是被记录的边界事实，不是被跳过的门禁。head commit 上的 check-run 全部来自本表 push/dispatch 事件。
- 从 `52f9d40a` 起 np-i11 亦支持同仓 `pull_request` 事件（`I11_REFS head=… base=…` 留痕、checkout test-merge 合并提交，缓存/codecov/artifact 过滤同步扩展到 i11 PR head）——该机制供**另行登记的 main 目标集成 PR 事件**使用，本次未创建该 PR（属协调者单独登记项）。
- PR 评论数 0；`mergeable_state=unstable` 来自负向探针的预期红灯 check-run，非产品缺陷信号。

## 缺陷生命周期（new，非历史故障）

`076f7e7` 引入的 `check-i11-tests.py` 在解析 go test -json 日志时于第 48 行抛 `JSONDecodeError`（`full-i11.sh` 把 stderr 工具诊断与结构化事件并入同一流）；两轮 full 红灯均为该缺陷，产品 go 套件本身 rc=0。`0ca6eb2c` 将 stderr 分离后，后续两轮 full 均 GREEN。与历史 PGO/CI 故障记录（fuzzing workflow-file 44508335288、审批编号唯一键碰撞、M1 R3 快照失败）无同签名，分类为 new。

## 边界（8 个事件全部成立）

- **artifact**：0 产出（np-i11 无 upload-artifact 步骤；legacy 报告步骤被 ref 过滤）；runner 临时目录随 job 销毁，符合白名单空集规则。
- **缓存**：`setup-go cache:false` 显式关闭、无 actions/cache 步骤；分支 ref 缓存 0 条；全仓 3 条缓存均在 main（最后访问 2026-11-23），本窗口零读写、零跨 run 复用。
- **通知**：仓库 webhook 0 个；np-i11 无任何评论/邮件/第三方上传/部署步骤；PR 评论 job 被过滤未运行；PR #51 评论 0。

## 边界外事项

观察窗口止于记录生成时刻；`309d1dd4` 之后的分支迭代、以及 main 目标集成 PR 事件（含浏览器链的 pull_request 运行）由协调者继续登记，不在本记录范围内。本机 full-gate 与必跑测试矩阵证据由 t_92dc3e8f 产出；汇总裁定与失败分类表由 t_d83a4c85 汇总。
