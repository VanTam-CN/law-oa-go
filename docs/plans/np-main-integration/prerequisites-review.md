# NP main 集成前置能力静态核验

- 日期：2026-11-19（Asia/Shanghai）。只读静态核验：未运行服务/数据库/Redis/测试，未执行 bootstrap，未改动本文件以外的任何文件。
- main 目标：`/Users/mac/Desktop/FT/law-oa-go/output/np-main-integration-prep` @ `d9c92fbd583a0cbbf7f9b4c071e1f2329981f4d7`（= origin/main，工作区干净）。
- 首切片源：`/Users/mac/Desktop/FT/law-oa-go/output/np-v04-first-slice` @ `2d659d9dd750f49ddb1f517746e22851392231a6`（工作区干净）。
- 谱系：首切片是独立冻结谱系，main 与首切片祖先集合交集为零，详见主控 `source-baseline.json`；不是仅凭跨库 bad object 得出结论。Go目录需从 `backend/` 映射到主干根目录，前端路径不剥前缀。
- PR46 = `origin/codex/intake-collaboration-mvp` @ `6ef9795`，与 main merge-base 为 `70d38b6`，未合并；其迁移 000082-000084 与首切片同号迁移存在冲突，必须登记具体名称/表结构再取舍。PR47 有独立分支 `codex/np-v04-first-slice`，证据与 PASS 绑定其树 `2d659d9`，不得沿用 main。
- 路径缩写：`M/` = main worktree；`F/` = first-slice worktree。

## 结论

与主控一致：main 已有持久认证会话（AuthTokenSession）与客户授权底座，可直接复用；原生草稿 revision 合同（`draft_revision`/`expected_revision`/事务行锁复核）在 main 完全缺失，是 NP 首切片落地的第一阻断。NP 处理链、回执/来源快照迁移、前端保存钩子在 main 为零。缓存边界按主控指令仅登记为必补项，方案由主控写。

## main 已有、可直接复用

| 能力 | 证据（M/ 内路径:行号） |
|---|---|
| 持久 access 会话模型+表 | `internal/models/auth_token_session.go:5-29`；`migrations/000078_auth_token_sessions.up.sql:1,17-18`；bootstrap 含该表 `internal/database/production_schema.go:939`；开发 AutoMigrate `main.go:252` |
| JWT 校验与用户上下文 | `internal/middleware/jwt.go:112`（ValidateToken）、`:182`（Set user_id）、`:258-259`（GetCurrentUserID）；首切片 session() 依赖的三个原语齐备 |
| 客户读授权 | `internal/services/authorization_service.go:198-221`（CanReadClient） |
| 接案表与当事人身份列 | `internal/database/production_schema.go:16`（bootstrap v14）、`:37,344-349,518-523` |
| 路由骨架 | `internal/router/router.go:274-275`（protected+AuthMiddleware）、`:378-384`（case-intakes 现有四路由） |
| Redis 可选接线 | `internal/database/optimized_components.go:243-248,280-289`；`main.go:76-84`（Redis 缺失即关闭响应缓存） |
| bootstrap 入口 | `cmd/migrate/main.go:40-47`（生产仅允许 -command bootstrap） |

## main 缺失：必须先补（阻断首切片）

1. **原生草稿 revision 合同（第一阻断）**：main `internal/` 对 `draft_revision|DraftRevision|np_description_source_revision|advisory` 零命中；`UpdateCaseIntake`（`internal/handlers/demo_aggregate_handler.go:1892`）无 expected_revision 校验，事务内无行锁/revision 复核（`:1973-1978`）。首切片实现与测试：`F/backend/internal/handlers/demo_aggregate_handler.go:46-50,85-97,2253`（函数）、`:2302,2367-2378,2390-2404,2438-2446`；`intake_draft_revision_test.go`（7 处）、`intake_draft_revision_contract_test.go`（12 处）。main 前端 `expected_revision` 零命中；PUT 保存无版本头（`frontend/src/pages/batch01/Batch01Prototype.tsx:4372-4379`，Idempotency-Key 仅创建用，`:4376`）。
2. **NarrowPilotHandler 全链路**：main 对 `NarrowPilot` 仅注释命中（`internal/services/authorization_service.go:51`）。需移植：handler 三件套（`F/backend/internal/handlers/narrow_pilot_types.go`；`narrow_pilot_context.go:25-30,44-53,67-93,95-121,153` 起；`narrow_pilot_write.go:46-59,60-66,70-104,126-171,189-217,247,278-288`）；路由（`F/backend/internal/router/router.go:164,383-388,403-409`）；`UpdateCaseIntake` 委托（`F/.../demo_aggregate_handler.go:2264` 附近 `narrowPilot.HandleUpdate`）。注意 main 同名函数有 `canViewAllMatterData` 放宽（`:1934`），NP 路径必须保持 owner-strict（created_by+CanReadClient+draft，`F/.../narrow_pilot_context.go:95-121`）。
3. **NP迁移与来源fixture基线不同**：main 迁移止于 000079+001。源fixture采用bootstrap → 82 → 83 → 84 → 85，再显式86；这是源验证环境的历史顺序，不是已证明的NP生产最小依赖。`000086_narrow_pilot.up.sql:1-24` 只建来源/回执两表与两个唯一约束，外键仅引用case_intakes。主控补读82–85：82涉及inbox/approval/client_creation_requests，83是审批补充，84含材料文件与case_tasks，85是任务审计。不能未审闭包就强制整组搬入。材料版本回归需要84相应结构及真实文件API，须单列为基础包范围；PR46同号迁移冲突须逐表/列合并，不重复执行两套同号SQL。
4. **缓存边界（主控指令，只登记）**：`main.go:188-191` CacheMiddleware 全局生效于 router/auth（`router.go:274-275`）之前；`internal/middleware/cache.go:63-96` 无 NP 路由排除（SkipRoutes 字段 `:28` 存在但 main 未传），`:104-118` 认证前生成键（user_id 为空），`:375-380` 响应含 no-store 仍缓存 1 分钟。要求：接案/助手/来源/回执路由显式跳过 HTTP 响应缓存读+写、加 `Cache-Control: private, no-store`（首切片已有先例 `F/.../narrow_pilot_context.go:29,68`），并以 main+Redis 双账号验收。静态观察：当前 main 配置下写缓存路径疑似失效（MaxBodySize 零值检查 `:288`、头捕获表从未填充 `:220-228,294-316`），未复现，不作为安全依据（NOT_RUN）。
5. **NP_ASSISTANT_ACTORS 白名单**：`F/.../narrow_pilot_context.go:44-53`，默认空名单；main 需同语义接线。
6. **前端保存钩子与恢复标识**：`F/frontend/src/components/intake/useNarrowPilotSave.ts:11-19,29-40,44-86`（scope 恢复、原键冻结、可清除的 409 码）、`NarrowPilotAssistant.tsx`、`Batch01Prototype.tsx:5598-5601,6574-6589`（eligible 门控→局部 PUT/全量 PUT 分流）+ 组件测试×2、hook 测试、`e2e/np-first-slice.live.spec.ts`。main 全部缺失。

## 最小依赖闭包（首切片在 main 落地）

实施前须锁定的闭包：main v14与所需表/列/索引核对 → 选择必要的追加结构（不能把82–85整组视为已证前提）→ 原生详情/列表/恢复/创建/更新的revision合同与相关写入者 → NarrowPilot三件套、路由、委托、白名单 → 接案缓存绕过 → 前端局部保存与恢复。

测试闭包还需包括源原生7项、上下文3项、PG10项、日志工厂/审计、前端5套33用例、fixture浏览器4条。这里是已有固定候选的用例清单，不使用rg命中处数当用例数。main未找到对应MatterCollaborationHandler/material_files和材料版本保护用例，源PG/浏览器夹具依赖真实材料上传/版本/下载栈；该缺项须在M0明确复用子集或另立基础范围，不能删去文件保护断言来让移植通过。

## 阶段可后置

- PR46 整体功能不是默认前提；但本轮尚未证明源测试和所有基础能力均可独立于它移植。先确认必需子集，不能称“整体肯定不依赖”或直接整包合并。
- G02+ 跨角色来源阅读、五类状态写入、异步迟到检测、真实模型/预算（`F/docs/np-first-slice-contract.md` §3 明确不在首切片范围）。
- `internal/middleware/v2/v2_security_middleware.go:110` 的 no-store 路径：main.go 未接线，本切片不依赖。

## 未核项 / NOT_RUN

- 未运行任何构建/测试/服务/数据库/Redis；未执行 bootstrap（只读授权边界）。
- main bootstrap v14 与首切片 bootstrap 未逐行比对（双方均含 case_intake_parties 身份列，等价性未证）。
- 迁移 82-85 能否精简未逐文件核；000086 与 82-85 的依赖只核到"86 仅外键引用 case_intakes"。
- 缓存写路径"疑似失效"仅为静态读取，未发请求复现；不能据此宣称越权已复现或已排除。
- main 中 PUT /case-intakes 的其他调用方未全量枚举（只核了 Batch01 页面）。
- PR46 与首切片在 handler 层的语义冲突（超出迁移编号）未分析。
- Redis 缺失时 CacheService 降级行为未实测。

## 阻断条件（任一成立即不得进入实施验收）

- main 无 draft_revision/expected_revision 事务锁合同而直接接通 NP 保存路径。
- 原材料/文件版本基础缺失却删减验收、以模拟文件替代真实版本API，或把整组82–85当作无需审查的必需迁移。
- 000086 两表/两唯一键未建即放开 HandleUpdate（strictReady 将 503 SAVE_SCHEMA_REQUIRED，`F/.../narrow_pilot_write.go:60-66`）。
- NP 路由未显式跳过 HTTP 响应缓存读+写，或缺 no-store，或未完成 main+Redis 双账号验收。
- 把 PR47 PASS 或首切片本地证据直接当作 main 集成验收。

## 主控复核更正

本报告由GLM Flash提供初稿，主代理核对并修正：PR47独立分支、谱系证明方法、迁移82–85的条件性质与材料版本测试闭包。以上更正为最终计划依据。Redis缺失仅有现有降级接线与日志，实际缓存行为仍NOT_RUN；不以日志“保持关闭”代替运行证明。

2026-11-20补充：材料闭包在M0出口按实施计划v0.2裁决A（独立材料基础包后满足完整I09）或B（经明确范围确认，主干现有材料保护必测，两版文件流DEFERRED/HOLD）。当前未选择B，原完整要求保留；B不是通过删测试或Skip规避失败，也不继承源首切片的完整材料PASS。前端分流还必须检查本人/在职律师/draft及来源历史保护状态；服务端逐次鉴权与版本保护不因前端eligible成立而省略。
