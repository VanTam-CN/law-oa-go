# 接案协作迁移验证

本地 PostgreSQL 16.10，Unix socket `/private/tmp/lawoa-intake-collaboration-runtime/socket:55437`。测试库 `lawoa_intake_migration_20261107` 为本轮新建空白库，没有真实资料。

1. 使用当前 schema bootstrap 初始化空白库。
2. 执行 000084 → 000083 → 000082 down，再执行 000082 → 000083 → 000084 up。
3. 重复上述完整 down/up 一轮。
4. 再次执行 schema bootstrap。
5. 查验两项不可变触发器与四项唯一索引。

所有16个子命令退出码均为0，见 `results.json` 和对应原始日志。回退前无新协议业务数据；有数据时的000082拒绝回退及不变性由 snapshot/RESULT.md 与本轮合跑 integration-affected-3.jsonl 对应真实PG用例证明，不将空库回退当真实资料回退。

浏览器测试库首次重复bootstrap遇到prepared statement不接受多语句（runtime/bootstrap-3.log）；将每个函数/触发器拆为独立执行后，bootstrap-4/-5连续成功。本目录保留的是修复后的完整三迁移验证。

生产迁移、备份恢复及异机灾备均未执行。新轮次数据存在时不能直接down82；应先停止新写入、备份并按维护窗口另行制定迁移方案。本轮只提供迁移文件与隔离验证，不构成真实迁移授权。
