# 最终闭环版交接

结果：**34/34 PASS，0 FAIL，0 SKIP**；PG 30 项，旧 handler 守卫 4 项。`-race` 和 services/handlers `go vet` 均退出 0。此为隔离本地专项证据，不代表全 MVP 验收。

已覆盖主线程最后两处后端改动：

- `FreezeIntakeSubmissionInTx` 使用 `ResolveApprovalIntakeID` 检查持久关联。新增真实 Submit 回归证明 legacy draft 删除 metadata.intake_id 后不能通过早退 no-op 绕过：返回 `APPROVAL_SNAPSHOT_UNRESOLVED`、409，12 张业务表完全无变化；恢复原关联后同一 draft 可提交、生成 1 轮快照并持有唯一接案槽。
- 成案及 completed replay 在审批成案关联写入后、事务 commit 前执行待办交接。新增首次成案/已完成重试 × inbox update/create 共 4 项。GORM 故障回调必须在真实事务中命中；失败时 12 表整行内容恢复原状，恢复后唯一完整案件与 2 个批准主体，结果待办保持原 ID/key，未重复或改写用户待办状态。

本次代码增量只有 `internal/services/approval_collaboration_postgres_test.go` 文件末尾两个新测试函数、5 个叶子用例。`47-final-append-only-test-audit.json` 证明原 42,059 字节前缀保持不变；原 HTTP 测试与两个 handler fixture 文件也保持不变。没有替换、删除或削弱任何旧测试，没有修改业务实现。

最终证据：

- `49-final-transition-race.jsonl`：完整机器可读日志；`.log` 为文本副本。
- `50-final-transition-vet.log`：go vet 日志，退出码 0。
- `52-final-transition-results-summary.json`：34 个叶子逐项结果、命令与退出码、原 29 项全部保留的校验。
- `48-final-transition-source-before.json` / `51-final-transition-source-after.json`：同一执行进程在测试前后紧邻采集 447 个 Go 源码、模块、迁移文件的 SHA-256，差异为空。
- `46-case-completion-inbox-regression.log`：新增 5 项单独运行的真实 PG 故障命中及恢复证据。
- 原编号碰撞红灯 `09`、`21` 和原结果摘要 `24` 未覆盖，内容哈希仍保持不变。

环境仍为 `LAW_OA_COLLAB_PG_DSN` 指向指定本地 socket、55437、lawoa_intake_mvp_test/mac，每项独立随机 schema；`GOPROXY=off`，沿用指定 GOCACHE。没有真实 API、依赖安装、commit/push 或子代理调用。新增 services 测试文件仍受 `*_test.go` 忽略规则影响，主线程交接时需要显式纳入。

本任务到此交接，停止写入。
