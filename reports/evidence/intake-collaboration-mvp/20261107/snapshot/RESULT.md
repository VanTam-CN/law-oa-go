# N1/N3 快照与补正 helper 交付

- 任务日期：2026-11-07；工作目录：`/private/tmp/lawoa-intake-collaboration-20261107`。
- 基线 HEAD：`a806831a7cef839c2b971cd4ab260cce77f42b8b`；本轮为多人共享工作区，未提交、未推送。只新增列明的7个源文件及本证据目录。
- 判定：**helper 作者自测 PASS**；独立审查、handler/API权限链、完整成案/待办交接及生产均不在本报告 PASS 范围。
- 源文件摘要见 `source-sha256.json`；共享主代理/其他代理后续改动可能影响集成证据。

## 已实现

1. typed ApprovalSnapshot 映射原表；SupplementRequest/Response、ResubmitApprovalRequest（JSON `supplement_responses`）。
2. active pointer 权威读取；无 pointer 的 intake 校验内嵌/持久/metadata一致；拒绝缺来源与多来源冲突，不按创建时间猜测。旧非intake唯一持久快照可读取，通用无快照审批保持兼容。
3. Apply仅内存投影；Freeze只写不可变新快照并设置审批对象pointer/conflict字段，所有事实来自锁定的intake、正式检查、服务器主体与材料清单。旧审批metadata不复制为新事实。
4. 正式检测绑定、客户与多方身份/原业务角色集合、事实确认版本、复核结论/证据哈希/有效期、当前指派及回避独立性校验。独立复核快照保存绑定字段与哈希，不复制复核自由文本。
5. 结构化清单白名单与服务端item_id；旧文字记录只读投影legacy_text。逐项回复保存只改受控协作键与draft_revision，保留lawyer_facts_version/status；重提要求必填回复并支持同请求重放/异请求冲突。
6. 000082部分唯一索引及PG级快照/版本决定不可变触发器；无新轮次up/down/up，有新轮次或版本决定拒绝down，不删历史数据。

## 实际验证

- `helper-final.log`：13个顶层测试、26个子场景 PASS（合计39项）。包括真实PostgreSQL独立临时schema；SQLite仅用于纯reader兼容性测试。
- PostgreSQL：回复保事实；其他律师/助理/过期revision/伪造item/旧return拒绝；snapshot写入故障、回复后故障和模拟下游失败均回滚；清故障可重试；三轮双并发同一return重放每轮只增一个快照；原审批metadata及旧快照持久内容不改；原始SQL UPDATE/DELETE与重复round/return插入受数据库约束阻止。
- 迁移：up/down/up及重复up；旧重复历史快照保留；生成新提交后down拒绝、列和数据保留。
- `vet.log`：`go vet ./internal/services` PASS。
- `unit-first.log` 保留首次纯逻辑通过记录；`pg-first.log` 保留JSONB测试基线错误，改为入库后字节基线后`pg-second.log`通过。`business-role-first-red.log` 保留担保方/普通相关方不应因归档映射相同而放行的首红；最终已修复并通过。
- 测试环境为用户已启动的PG16本地socket/55437/`lawoa_intake_mvp_test`，每例独立schema，创建后清理，无真实外部调用。

复现命令：

```sh
GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache GOPROXY=off \
INTAKE_COLLAB_PG_DSN='host=/private/tmp/lawoa-intake-collaboration-runtime/socket port=55437 user=mac dbname=lawoa_intake_mvp_test sslmode=disable' \
go test ./internal/services -run '^(TestIntakeSubmissionPG|TestIntakeSupplement|TestApprovalSubmission)' -count=1 -v
```

## 交接与限制

- 完整签名、调用顺序及DTO见`contract.md`。所有InTx错误须外层rollback；caller负责approval授权/锁、tracking、指派、审批状态与inbox同事务。
- 主代理尚需确认两处集成点：普通metadata拒绝键包含`supplement_return_record_id`；初建在tracking INSERT前取得intake锁，遵循approval->intake->tracking。
- 两个新增`*_test.go`被仓库`.gitignore:282`忽略，但文件实际存在并运行。后续授权提交时须显式收录；本轮未改.gitignore或暂存。
- C12新审批争槽/两种有序交错、C13真实inbox故障、C14完整成案主体故障、浏览器闭环由主代理完成；本helper测试不代替这些端到端验收。
