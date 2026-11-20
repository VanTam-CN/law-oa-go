# M0 后端闭包核验（backend-closure）

- 工作树 `output/np-main-m0`，分支 `codex/np-main-m0`，基线 HEAD `2481a47fb8bc628a73449ba57e721311a7030eb8`（与main `d9c92fb`应用源码一致已由主代理以排除docs的git diff独立核对；本执行者未自行验证）。
- 参照源首切片 `output/np-v04-first-slice` 固定 `2d659d9dd750f49ddb1f517746e22851392231a6`。本报告只读核验，未改业务代码/迁移；未运行测试或构建（主控已跑 go build 与 handlers/auth 回归，PG 为显式 bootstrap 独立合成库）。

## 结论

**MI-R03**：主干无独立 ConflictCheckTask 模型或入队列；异步任务即 `conflict_check_records` 行，队列状态内嵌 `check_status`（`internal/services/async_conflict_check_service.go:17-22`），持久上下文只有 `search_parameters` JSON 列（`internal/models/conflict.go:219-229`）。`ConflictCheckRequest` 无 revision/facts 版本字段（`internal/models/conflict.go:12-22`），接案冲突入口组装请求也不带 revision（`internal/handlers/demo_aggregate_handler.go:2325-2343`）。落库 payload 是审计安全快照（`async_conflict_check_service.go:91` → `conflict_detection_service.go:1629-1642`），因此 `LinkConflictCheckToCase` 读 `search_parameters` 时拿不到任何 revision（`internal/repositories/conflict_repository.go:143-202`）。同步回写在 Link 事务成功后另起一次 intake 更新（`demo_aggregate_handler.go:2378-2387`），两段无共享事务，存在中间状态或晚到覆盖风险；本轮未运行复现。

**MI-R02/M0-G00**：主干 `UpdateCaseIntake` 无 expected_revision、无事务行锁复核，metadata 覆盖写（`internal/handlers/demo_aggregate_handler.go:1892-2053`，尤其 :1926 与 :1997-1999）；`rg draft_revision|expected_revision|source_revision|DescriptionSource` 在 `internal/` 零命中。来源快照/回执两表只存在于源切片 000086（`backend/migrations/000086_narrow_pilot.up.sql:2-24`，外键仅引用 case_intakes），主干迁移止于 000079+001。源切片 revision 合同实现在 `backend/internal/handlers/demo_aggregate_handler.go:2295-2460`（必填校验 :2367、行锁复核 :2437-2446、来源键保护 :2390-2404）。

**I09**：用户已确认路线 B（2026-11-20）。主干首轮只验现有 `case_materials` 及实际关联记录保全；`case_materials` 为单表无版本列（`internal/database/production_schema.go:61-72`），更新是删除重建（`demo_aggregate_handler.go:2017-2035`）。两版文件流 I09-FILES 记 DEFERRED/HOLD；不声称与源首切片完整材料能力等价，不授权 M1。

**auth/session**：主干已有持久会话 `AuthTokenSession`（`internal/models/auth_token_session.go:5-29`）+ `TokenSessionStore`（`internal/auth/token_session_store.go`）+ 000078 迁移，M1 可直接复用，无需移植源会话代码。

## 最小 M1 工作范围（按依赖排序，不估人日）

1. 接案写入者：沿用源合同的expected_revision请求体、事务行锁复核及metadata.draft_revision；主干对可达新旧写入统一保护，不新增接案版本列；来源快照/回执两表与回执语义按 000086 移植（PostgreSQL-only；与 000080-000085 的编号冲突未逐一审计，属待核项，不能整组同号重复执行）。
2. 冲突链路：`ConflictCheckRequest`缺少检测输入revision载体；M1建议在内部请求/association增加服务端取数revision，并落入已有search_parameters审计快照。不得相信客户端自报值；审计复制、同步/异步传递和完成时对照都须验证，不能仅凭浅拷贝推定全链正确。沿用现有模型，不为本切片新增任务平台或一套facts-version体系。
3. 同步回写合并：Link 事务已写 status=conflict_ready 与 metadata（`conflict_repository.go:246-280`），同步路径删除二次 Update 或并入单事务。
4. 材料路线 B：不新增版本栈；验收 I09-MAIN = 现有材料表/关联记录保全测试（具体断言夹具待核），I09-FILES 保持 DEFERRED/HOLD。
5. 前端保存钩子/expected_revision请求体字段为前端分组工作，本报告不覆盖。

## 未查/未运行

- 未运行：全部 Go 测试/构建（主控基线运行）、PG/浏览器夹具、真实环境。
- 待核：000080-000085 与 000086 逐表/列迁移依赖与同号合并；facts-version（lawyer_facts_version/conflict_check_facts_version）在主干 metadata 是否已有约定（源切片 intake_continuation_service.go:236-254 有读取，主干未确认）；前端 expected_revision 现状。
- 假设：基线与 main 应用源码一致（主控给定）。

## 主代理复核补充

本报告已按主干实际行号修正同步二次回写位置（2378–2387）、repository事务回写（246–280）及expected_revision是请求体字段的表述。当前主干无draft_revision/expected_revision字段合同，异步使用conflict_check_records和search_parameters，不需要新增独立任务表；来源两表实际名称为case_intake_source_snapshots、case_intake_write_requests。M1改造建议不等于实施授权，静态风险不算已运行复现。

主代理新增结构证据：在全新主干v14合成库直接执行源2d659d9的000086原SQL成功，case_intake_source_snapshots/case_intake_write_requests及两个唯一索引/接案外键均存在，未执行82–85。故“86核心两表依赖必须整组移植82–85”不成立；完整源fixture、应用读写与迁移编排仍待封板，不能据结构成功报告M1完成。
