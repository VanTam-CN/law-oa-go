# 实际 CaseCreationConfig 双 tracking 形状专项

仅修改 `internal/services/approval_multistage_postgres_http_test.go`。生产 Go 源码保持 `3f970040f5e40b8016c18a070b5aa1345d2b2eec`：321 个已跟踪非 test Go 文件测试前后 SHA-256 完全一致，与该基线的生产 Go diff 为空。未修改浏览器 runtime 数据库，未提交或推送。

## 实际执行

保留原 4 叶子，给普通和 integrated 两个决定入口各新增 `config_direct_two_stages`，合计 **6 PASS / 0 FAIL / 0 SKIP**，`go test -race -count=1` exit 0。只运行该 multistage 专项，未扩全库。

新增场景从空的独立 PostgreSQL schema 调用真实 v16 production bootstrap，seed 合法接案、检测、独立复核、两级工作流，然后从真实 HTTP 首提。请求同时携带 top-level `CaseCreationConfig` 与 metadata 中的配置，字段匹配浏览器输入。测试未手工创建、修改或重排 tracking，未退回、未重提，最终仅一份提交 snapshot。

验证结果：

- 首提真实产生两行：较早的非 NULL `INTAKE_BIND` 与较晚的 NULL-intake `CASE_TRACK`，均 pending。直接调用真实 repository getter，确认其按时间选择 generic 行，而非假设 canonical 优先。
- 初始和中间节点的真实 GET integration status 均为 pending、case ID 为空。第一节点 HTTP 200、under_review/compliance，数据库中两条 tracking 全部列与初始相同，无新行/failed，case INSERT 未触达；下一审批人只有一个开放决定待办。
- 最终 HTTP 200、approved、一个案件；generic tracking 变 completed 并绑定该案，较早的 canonical 槽保留原值。精确验证案件 client/lawyer/check，以及两个当事人的身份 digest 和角色，而不只比较数量。
- 最终及 same-case replay 后真实 GET integration status 都返回 completed 和相同 case ID，读取前后 12 张业务表完全一致。重放不增加案件、tracking、决定或 parties；仍为一个 case、两个准确 parties、两条决定、两行 tracking、一份 snapshot。
- 原有无配置单行、最终 INSERT 故障与恢复重试场景全部保留并复测通过。

## 证据边界

先前结论是源码路径分析结合主线程 R5 SQL；本次补足的是**作者执行的隔离本地真实 PostgreSQL + HTTP runtime 证据**，覆盖“canonical 比 generic 更早、未退回直接两级批准”的缺口。没有对浏览器 R5 的 case3/4 发起新请求，没有把本次专项升级为独立浏览器或完整 MVP 验收。

首次运行即通过，未确证新的产品失败，因此没有更改产品。getter 当前仍按 created_at DESC，不应描述为显式 canonical 优先。

`02-pg-race.jsonl` / `.log` 保留完整运行；`04-results.json` 包含六叶子结果、test SHA 和生产哈希结论；`01-production-before.json` / `03-production-after.json` 是逐文件证据；`05-test-only.diff` 是本轮唯一代码增量。测试及生产源码均已停止写入。
