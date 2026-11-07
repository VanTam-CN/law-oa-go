# Bootstrap v16 与审批详情模式投影交接

源码已冻结。本轮不再修改 Go 源码或测试，不提交、不推送。当前冻结点为 `25-final-source-after.json`；300 个叶子测试通过、0 fail、0 skip，其中 111 个真实 PostgreSQL 叶子。race、vet、`go build ./...`、diff check 全部退出 0，601 个 Go/mod/迁移文件测试前后 hash 无变化。此为本地作者回归证据，不代替浏览器两链或完整 MVP 验收。

## B4：真实初始化与已有 v15 升级

- `internal/database/production_schema.go`：版本从 `postgres-mvp-2026-11-07-v15` 升为 v16；新表包含可空 `approval_claim_token VARCHAR(64)`；启动校验纳入 claim、tracking 的真实 `creation_status`，没有创建无用 `status` 列。
- `internal/database/intake_collaboration_schema.go`：无论已有版本标记是什么，正常 bootstrap 都执行幂等加法，补 claim、必要时补 tracking.intake_id，并安装 `idx_case_creation_tracking_intake_unique`。000080/000081 原迁移文件未改。
- 同一启动契约检查 claim 类型/长度/可空性、tracking.intake_id 可空性、完整且有效的单列唯一索引，拒绝同名普通/部分/错误列索引。
- 每个新测试从空的随机独立 schema 调用真实 `BootstrapProductionSchema`，创建 119 张表，再 seed 业务数据；没有 AutoMigrate 或手造运行时表来补缺。真实 HTTP 首提经历 claim、冻结快照、唯一 tracking 槽、决定待办；重复提交 409 且业务表不变。
- 已有 v15 测试从实际 bootstrap 结构恢复其缺 claim/缺 canonical index 的历史形状、写入 v15 标记，再以新连接执行 bootstrap，验证 v16、保留原数据和已有 claim，随后真实 HTTP 首提成功。新连接模拟启动命令，避免测试人为改形状后旧 prepared plan 缓存干扰。
- 多个 NULL intake 槽可共存，同 intake 第二槽触发 SQLSTATE 23505。遗留重复槽导致 bootstrap 失败，新增 claim 与版本升级整事务回滚，不清理原数据。
- 没有修改浏览器 runtime schema。主线程应以正常 `cmd/migrate -command bootstrap` 入口初始化/升级，再启动新候选；仅重启不执行 bootstrap 的 QA server 不会自动安装结构。

## F5：确认前后端契约

`GET /approvals/:id` 授权后响应 `data.submission_snapshot_required:boolean`：

- 仅确认普通业务且无持久 intake、历史/内嵌快照、active 指针时为 false。普通 leave 的 snapshot 404 合法，实际决定仍可成功。
- 任何持久 intake 关联、成案类型、历史/内嵌快照或 active 指针为 true。持久关联复用 `ResolveApprovalIntakeID`，覆盖 metadata、check、tracking、历史快照和 association；不依赖单一可变 metadata.intake_id。
- 读取错误返回错误，不输出 false。该字段只是只读业务模式，不是决定权限；原事务授权、版本与唯一槽门禁保持不变，旧 intake 缺指针/槽仍 409 无写入。
- 生产变更为 `internal/handlers/approval_handler.go` 与新 `internal/services/approval_snapshot_mode.go`；不改前端或 inbox 文件。

## 相对 298 PASS 冻结点的最后 delta

`13-final-source-after.json` 是第一版 DTO。随后收到完整前端契约，补齐“已有非 intake 历史快照、成案类型也为 true”及故障读不降级。原第一版历史快照被投影 false 的首红保存在 `18-existing-snapshot-mode-first-red.log`。

| 文件 | 最终 SHA-256 |
|---|---|
| internal/handlers/approval_handler.go | 15eab93c86d34ca71d1f4ec9d7844e49cc08ba99ecb4580cb8cf07ab18e6615e |
| internal/services/approval_snapshot_mode_postgres_http_test.go | 156750c4eea21e0c250447b1c4b876405992e003563a08a79b0ef0bc589fb957 |
| internal/services/approval_snapshot_mode.go（新增） | c762942156e3076549dd407c295f53e690a46c084f13c45b79459931a0820dc5 |

最后 delta 专项 6 叶子通过：普通 leave（含合法 snapshot 404、真实批准、注入读取失败）、非 intake 历史快照、仅检测/仅 tracking/仅历史快照三类旧 intake、成案类型与内嵌快照投影。最终联合复测已在收到“不再扩测”前启动并完成，没有再开一轮。

## 文件与证据

- 新测试：`approval_bootstrap_postgres_fixture_test.go`、`approval_bootstrap_postgres_contract_test.go`、`approval_bootstrap_postgres_http_test.go`、`approval_snapshot_mode_postgres_http_test.go`，均在 `internal/services/`。现有 `approval_collaboration_postgres_test.go` 仅提取纯业务 seed helper 供真实 bootstrap 测试复用，未改变旧断言。
- 仓库忽略规则可能隐藏新 `_test.go` 文件；提交整理应按 `28-task-file-manifest.json` 核对，不以普通 git status 作为新测试完整清单。
- `02-first-http-red.log`：真实 v15 初始化首提 500；`17-browser-r3-error-excerpt.log`：R3 SQL42703 原始日志摘录；其余首红均保留。
- `04-first-http-recheck.log`：修复过程中的 validator 字段拼写错误，已按实际 creation_status 修正。
- `05-bootstrap-contracts.log`：测试人为改表后重用 prepared plan 的测试设置失败；已用新连接模拟重启，并要求错误信息确实指向契约缺失，避免假阳性。
- `19-mode-contract-recheck.log`：最后 DTO 专项 6 PASS。
- `21-final-race.jsonl` / `.log`、`22-final-vet.log`、`23-final-build.log`、`24-final-diff-check.log`；结果聚合 `26-final-results.json`。
- `20-final-source-before.json` / `25-final-source-after.json`：601 文件前后同源；`30-delta-final.json`：相对 13 的两项变更、一项新增、最终现场 hash 再核对相同。
- `27-final-production-task.diff`：相对于接手 bootstrap 时的生产代码增量，保留其他作者改动；`29-final-preserved-red-sha256.json`：首红哈希。

所有 PG 测试仅访问已授权临时 socket/55437，使用独立 schema 清理自身数据；全部设置实际 PG 环境变量，skip 不计 PASS。未调用外部 API、安装依赖、修改生产数据、提交或推送。后续工作交由主线程进行候选提交、正常 bootstrap 与浏览器两条主链重跑。
