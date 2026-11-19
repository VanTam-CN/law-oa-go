# NP 首片后端独立复核 R1（58da474）

- 复核人角色：本轮未参与实现的独立复核/执行者
- 固定提交：`58da474bf917c60ff3107ce4aee56eb41c7441da`，父提交 `b3809f604dfa2b0b43e8108bc5644d0ba6ed8c29`
- 独立工作树：`/tmp/np-r1-independent-58da474`（detached，测试前后 clean）
- PR47 禁止合并部署；本复核不合并、不部署、不提交 Git/远端、不停服务、不创建新环境

## 源码复查发现

### 连接池参数化日志
`connection_pool.go`：GORM logger 配置新增 `ParameterizedQueries: true`，慢查询/错误日志参数化输出，不内联绑定参数值。新增测试 `narrow_pilot_pool_logging_test.go`（SQLite 内存库 + 金丝雀参数）断言日志含 SQL 且不含参数值。

### 原生创建/两条 metadata 合并保留来源路径
`demo_aggregate_handler.go`：
- 创建路径（~L2161）：`delete(createMetadata, npDescriptionSourceRevisionKey)` 在服务端盖章 `draft_revision` 前移除客户端伪造的来源指针。
- 更新路径（~L2387）：合并客户端 metadata 时跳过 `np_description_source_revision`；存储 metadata 起底、客户端键覆盖、服务端 `draft_revision` 最后盖章。
- 冲突重置路径（~L2407）：同样跳过伪造键，存储 metadata 起底，`draft_revision` 最后重盖。

新增测试 `TestNPPostgresNativeCannotForgeSourcePointer` 在 POST 创建 + PUT 更新（draft/lawyer_facts_confirmed/conflict_ready）下断言 `jsonb_exists(metadata,'np_description_source_revision')` 为假且普通 metadata 未丢。测试路由新增 `r.POST("/api/v1/case-intakes", h.CreateCaseIntake)`。

### UUID 防卫
`narrow_pilot_write.go` `HandleUpdate` 入口新增 `uuid.Parse(c.Param("id"))` 校验，失败返回 `INVALID_REQUEST` 400。位于 HasTable 检查之前，前置表探测与合法化前的数据库访问不会收到非法 ID。新增测试 `TestNPPostgresMalformedIntakeIDIsBadRequest` 覆盖两种请求体。

### COALESCE 及 PG 测试
`narrow_pilot_write.go` `strictUpdate`：两条 `jsonb_set` 写路径（手工保存/来源快照保存）均改用 `COALESCE(metadata,'{}'::jsonb)`，避免 NULL 传播导致 revision/来源静默丢失：旧表达式遇 metadata 为 NULL 时会 200 提交但 metadata 仍为 NULL，不是报错。新增测试 `TestNPPostgresLegacyNullMetadataIsInitialized` 以 `DROP NOT NULL` 构造 NULL 行，分别验证手工/来源保存：新 revision 持久化、来源指针仅来源路径写入、迟到旧版本 409。

## 独立测试执行
- 目录：`/tmp/np-r1-independent-58da474/backend`
- 环境：本机隔离 PG `127.0.0.1:18790/np_first_slice_dev`，`NP_TEST_DSN` 由 `scripts/np-first-slice/.env` 经 python3+shlex 注入，口令/DSN/token 未打印、未记录
- Go：go1.25.6 darwin/amd64，复用宿主 Go 缓存（同宿主共享缓存，非另一台机器）
- 开始 2026-11-19T09:35:10.977Z / 结束 09:35:41.137Z / 退出码 0
- 结果：27 RUN / 0 FAIL；database / handlers / middleware 三包全部 ok
- 完整 stdout：`independent-backend.log`；命令与元数据：`independent-backend.json`

## 结论（限定范围）
在本仓库 58da474 指定 diff 范围内，连接池参数化日志、原生创建/两条 metadata 合并的来源路径保留、UUID 防卫、COALESCE 修复与新增 PG 测试的实现与测试相互对应，独立 -race 测试通过。本结论仅覆盖上述 diff 与指定测试，不构成 PR47 合并或部署授权。

## 未测 / 限制
- 仅运行 -run 正则匹配的测试；包内其余测试未在本次独立执行范围，但其编译被三包全量构建覆盖。
- PG 测试依赖本机隔离实例与合成 .env，未验证生产 schema 迁移或真实数据行为。
- 同宿主共享 Go 缓存与隔离 PG：独立代理在同一台机器执行，不等同于另一台机器复现。
- 前端与文档集成由主控继续，不在本次范围。
- 复核期间仅新增独立工作树与本目录 3 个证据文件，未改动业务源码或他人文件。



## 前端独立复核（4c184d5，含 62ea83a）

- 完整提交：`4c184d5d46f1e101d3ff0ced04f62ebd0b16d6f3`（父 `62ea83a1be0e02a3e693389c13efdaccd499abdc`，祖父 `58da474bf917c60ff3107ce4aee56eb41c7441da`）
- backend 树一致性：`git rev-parse 4c184d5:backend` = `58da474:backend` = `b810d1390ff72c997b5e163d7e4e81328cc5ff02`，后端源码完全相同，后端结论继续适用
- 独立工作树：`/tmp/np-r1-independent-client-4c184d5`（detached，业务源码 clean；frontend/node_modules symlink 原工作区，锁文件 SHA-256 一致后建立，未重新拉依赖）

### 源码复查发现

**hook 错误白名单（62ea83a）**：`useNarrowPilotSave.ts` 清键条件从「4xx 且非 SAVE_GUARD_BUSY/SAVE_RESULT_UNKNOWN」收窄为「仅 409 且 code 属于 INTAKE_DRAFT_REVISION_CONFLICT / SOURCE_SELECTION_STALE / SAVE_REQUEST_ABANDONED」。其余错误（含 408/429/401/403/404/400、409 IDEMPOTENCY_PAYLOAD_MISMATCH、409 未知 code、503 SAVE_RESULT_UNKNOWN、无 status）保留恢复键，notice 提示不能换键重发。新增测试 62ea83a-前缀 15 条参数化「不明结果保留原键、冻结体、正文不落 localStorage」+ 3 条终态清键 + 1 条重挂载回查/放弃 + 2 条成功回执清键 + 1 条来源选择不入缓存，共 22 条，覆盖原错误白名单与新语义的边界。

**组件 epoch（4c184d5）**：`NarrowPilotAssistant.tsx` 新增 `scope = intakeId|revision` + `requestEpochRef`，scope 变化时递增 epoch；上下文重建 effect 在 scope 变化时重置 busy/snapshot/enabled；生成与应用两处请求均以 `stillCurrent() = aliveRef && epoch 匹配` 把关，迟到响应（无论成功或失败）不再 setState、不再调用 onApply、不再解除新 scope 的 busy。新增 Scope 测试覆盖：同挂载切换 intakeId 后旧 GET 成功/失败均不影响新草稿、迟到 turn 响应不显示旧候选且不锁 busy、切换后返回同一草稿旧 GET 仍失效。

### 独立执行结果

- jest：`npm test -- --runInBand --runTestsByPath <5 文件> --coverage=false`，5 suites / 33 tests 全部 PASS，退出码 0（09:42:19Z–09:42:27Z）
- type-check：`npm run type-check`（tsc --noEmit）无错误，退出码 0（09:43:14Z–09:43:19Z）
- Playwright：`npx playwright test e2e/np-first-slice.live.spec.ts`（未传 --config，使用默认 playwright.config.ts：baseURL 127.0.0.1:3003、reuseExistingServer=!CI 默认允许复用，但本轮 [WebServer] 日志显示新启动了 Vite 且 BABEL 编译路径为独立 worktree，证明实际为新服务未复用旧前端；该证据属运行时观察，不能声称配置层强制 reuseExistingServer=false），4 条真实浏览器用例全 PASS，退出码 0（09:44:51Z–09:45:11Z），VITE_API_PROXY_TARGET=http://127.0.0.1:18789（vite 代理指向 18789 后端），NP_FIXTURE_PASSWORD 私下注入未记录

### 浏览器-PG 前后取证

- seed-browser.py 先行确认两版虚构材料（SYNTHETIC_BROWSER_MATERIALS_OK versions=2）
- compare-browser-db.py 复制到 /tmp/compare-browser-db-r1.py，仅改第 9 行 ROOT（固定主仓根路径）与第 10 行 REPORT（指向 docs/np-evidence/review-r1/browser-pg-comparison.json），查询与断言一字未改（diff 全文 2 行，已记录）
- before：sources=15, receipts=18, files=2, protected SHA 前缀 5373dedc89c7
- after：sources=18, receipts=22, files=2, protected SHA 前缀 5373dedc89c7
- 断言通过：protected/当事人/材料/文件/D2 控制组不变、来源快照 Δ=3、回执 Δ=4、revision 变化、最后手工保存未关联来源
- 历史证据 docs/np-evidence/browser-pg-comparison.json 已先备份并在 finally 恢复，恢复后 SHA-256 与原文件一致（69e719d4…）
- 4 条浏览器用例执行期间无其他写入（主控不同时跑 PG 或浏览器）

## 前端/浏览器结论（限定范围）

在 4c184d5 的前端 diff 范围内，hook 错误白名单与 22 条回归、组件 epoch 与 4 条回归的实现与测试相互对应，独立 jest/type-check/Playwright 全部通过，浏览器-PG 前后取证满足既定断言且目标外数据未变化。本结论不构成 PR47 合并或部署授权。

## 前端/浏览器未测 / 限制

- jest 仅独立执行指定的 5 个测试文件（33 用例）；全量前端套件未在本次独立范围
- Playwright 仅运行 np-first-slice.live.spec.ts 4 条用例；其余 e2e 未跑
- 同宿主共享 npm/jest/Playwright 缓存与浏览器二进制，不等同另一台机器复现
- backend 服务由主控独占启动（18789），本复核只做 HTTP 调用与浏览器交互，未检查服务进程
- 4 条用例产生的 trace/screenshot/video 未保存提交（retain-on-failure 且全部通过）
- 历史证据文件恢复以 SHA-256 一致为准；review-r1/browser-pg-comparison.json 为本次独立新文件，未覆盖历史
