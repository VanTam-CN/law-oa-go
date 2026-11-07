# B5 两级审批提前成案修复交接

源码和测试已冻结，基线 `7c5618f26ed5600de76e4ed94a682252d10a8aef`。仅改一个生产文件，新增一个专项测试文件；未提交、未推送、未改浏览器数据库或旧失败样本。供主线程提交候选、重建两链及 Boyle 独立复查。

## 最小修复

`internal/services/approval_conflict_integration_service.go` 的 ProcessApprovalWithConflict 在 ProcessApproval 返回后，只有 `updatedApproval.Status == approved` 才触发成案，并将 `updatedApproval` 传给 ensureApprovedApprovalHasCase。中间节点返回已更新的 under_review/compliance，不进入成案或 failed-tracking 分支。已成案 replay 分支、失败重试、授权/版本/唯一槽门禁均保留。

前置检查保留在每次 approve 之前：独立复核/风险、被冻结的成案材料、可用承办律师等检查必须在节点推进前完成。真实测试停用承办律师得到 409，12 张业务表完全无写入、case INSERT 未触达；恢复律师后同一节点可继续。

## 首红与真实行为验证

- `01-first-red.log`：两个 integrated 场景复现首节点 HTTP 409，数据库已 under_review、next=compliance/44、case=0，tracking 错标 failed。该轮普通入口先暴露了测试对其既有错误 envelope 的假设（普通入口不输出业务 code 字符串），没有将它冒充 B5 业务复现。
- 浏览器原始首红保留在 `08-browser-first-red-excerpt.log`，主线程另持有 `browser/final-a-stage-db.json` 的 two-tracking 证据；旧样本未修写。
- `03-http-recheck.log`：真实 PG、实际 production bootstrap（每个空独立 schema 创建 119 表）、真实 HTTP 初次提交与两个决定入口，4 叶子 `-race` PASS，0 FAIL/0 SKIP。
- 每个入口均覆盖正常两级成案、最终 case INSERT 故障后恢复重试。中间节点 HTTP 200，under_review→compliance/44；tracking 仍只有原来一行且全部列不变，无额外 NULL-intake failed 行；case INSERT=0，案件=0，决定=1，下一节点仅一个开放待办。
- 旧节点重复请求 409，12 张业务表不变。最终节点才触达真实 case INSERT；故障在真实 SQL 事务内注入，案件与 parties 不留残留，审批保持 approved/两条决定，tracking 明确 failed；恢复后真实 HTTP retry 仅生成一个案件、两个 parties，再次决定 replay 复用同案、决定仍两条。
- 普通决定入口对最终成案失败的现有响应为 409，integrated 入口为 500；本次不改变既有映射。`02-targeted-recheck.log` 保留了普通入口期望 500 的测试错误，03 已按两个明确契约重测全部四项，不以宽泛状态范围放行。
- `02-targeted-recheck.log` 中原有成案 inbox create/update 故障、completed replay create/update 故障四叶子均 PASS，真实事务回滚断言未替换或削弱。该组合命令整体为 FAIL（上述新测试 envelope 预期），不能将其整体记为通过；四项已有叶子结果与 03 的四项通过分开报告。
- `04-build.log`：`go build ./...` exit 0；`05-vet.log`：services/handlers vet exit 0；diff check exit 0。未扩全库测试。

## 冻结文件

| 文件 | SHA-256 |
|---|---|
| internal/services/approval_conflict_integration_service.go | 7690825d8accd27d47132f78cd003332293f544727acd10ed00185f6f2e5373e |
| internal/services/approval_multistage_postgres_http_test.go | 40d8de57765cb8030288061130f639010eef72de98251fe0bc322e6f1f837f84 |

结果和最终 hash 再核对见 `06-results-and-freeze.json`；生产增量见 `07-production-diff.patch`。新增 `_test.go` 可能被仓库 ignore 隐藏，候选整理需显式纳入。当前证据是本地作者专项回归，独立 review 和全新浏览器两链由主线程继续。
