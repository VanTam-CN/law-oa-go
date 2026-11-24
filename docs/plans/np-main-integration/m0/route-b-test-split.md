# M0 路线 B 测试拆分清单（t_23c34a57；历史t_cbc0ab2c，v0.4审核修订）

- 日期：2026-11-20；状态：文档/证据收口，未实现、未运行；c8716b7纸面封板曾判HOLD，本轮主调度完成文档复审，运行/实施仍HOLD。
- 源仓库：/Users/mac/Desktop/FT/law-oa-go/output/np-v04-first-slice @ 2d659d9；主仓库基线 c8716b7（分支 codex/np-main-m0）；应用 main d9c92fb。
- 本轮（M0-R01..R08）仅修订规则与证据登记：不改业务代码/脚本/workflow，不运行测试/服务/DB；执行者不提交推送，由主调度统一复审和推送，未来验证一律 NOT_RUN。
- 源历史 63 项是来源项计数，不是集成通过；M0-R01..R08 与原 MI-R01..08 编号不混用。
- 依据：implementation-and-acceptance-plan.md §5/§6 已确认路线 B；复审追加裁决：continuation 正向用例 DEFER，鉴权/陈旧/缺对象场景并入真实 router 主干详情/恢复入口；logging 工厂/审计测试与前端 3 NP + 2 Login 套件列为必跑。
- 所有"计划新增/ADAPT"均为未来主干用例或夹具改造的命名冻结，非现有测试；原文件原断言在源仓库不变。

## 一、M0-R01..R08 问题编号对照（本轮登记，映射 P1-1..5 / P2-6..8）

| 编号 | 来源 | 范围 | 本轮处置要点 | 状态 |
|---|---|---|---|---|
| M0-R01 | P1-1 | draftrevision 详情可达性 | 源自注册路由不证明主干可达；M1 真实 router：本人 200 / 越权 403 / 业务缺失 404 / 陈旧 PUT 409 后同用户 GET 最新；详情 list alias 与脱敏全响应含客户+对方 sentinel；不能借未注册 404 过验；continuation 正向能力仍后置 | 规则修订，NOT_RUN |
| M0-R02 | P1-2 | 主体保全与授权挂接 | 由 [dependency-seal](dependency-seal.md) §三 承接：主体最小 helper、事务顺序、失败回滚与授权 db/tx 挂接；验收锚点 I04/I06/I09-MAIN | 已交叉引用，NOT_RUN |
| M0-R03 | P1-3 | I15 可达场景拆分 | M1 合法原生已确认事实修改→snapshot 失效/转 draft/revision 仅推进一次→旧 sync 与服务级 async 回写拒绝；intake status/metadata/source 指针与关联 cases 不变；旧公共 async intake 入口拒绝不算 late-write 证据；M3 再 NP 摘要来源 overlap；M5 真入口 | 规则修订，NOT_RUN |
| M0-R04 | P1-4 | useNarrowPilotSave 夹具 | 成功 abandon 缺 intake id、成功 committed 缺 request_id、旧单槽 scope 断言不适用 v2；ADAPT 成功夹具 echo 真实捕获 request_id/id（回执id与标识intake_id比较），committed/no_change需合法revision、abandoned不要求revision，action有返回才比；旧缺字段改负向保键保输入；按请求 v2 键断言 wrong ids/unknown/doublepages；保留原场景与源历史 33，不伪称逐字 KEEP 或新通过 | 规则修订，NOT_RUN |
| M0-R05 | P1-5 | B fixture 与 CI/M5 边界 | 四条 B fixture spec/config M3 准备、M4 验证；CI 正常 I11 在 fixture 完成后；M5 只真 root 入口 PG/Redis 外置 TCP proxy，不以 fixture 混为真 main | 规则修订，NOT_RUN |
| M0-R06 | P2-6 | B seed 供给 | 源 np-local-check seed 仅 target/control/client/party，seed-browser.py 仅一个材料两版，不满足 B；M3 主干 B seed（计划）由隔离 np-local-check seed 扩展/专用 B seed 入口提供：>=2 独立 material rows、非空 clients/parties、target+control 精确 ids、>2^53 未知 metadata，返回 ids 给 compare 不硬猜；M2 只消费；原 4 浏览器场景保持干净确定序列 source+3/receipt+4；negative 独立 seed/db 窗口不污染精确计数；M5 真实入口另独立合成数据，不开放正式故障头 | 规则修订，NOT_RUN |
| M0-R07 | P2-7 | 迁移与受控追加入口 | 由 [dependency-seal](dependency-seal.md) §四 承接：000086 冻结迁移、受控追加、已应用核验、回滚不删来源/回执/墓碑、root 唯一支持入口；验收锚点 I12（M1 方案+M6 演练） | 已交叉引用，NOT_RUN |
| M0-R08 | P2-8 | C10/SW 边界 | 主代理拆分：PROD 构建真实资产/注册观测（允许记录 sw.js 缺失/注册失败，不能凭 404 称无旧 worker）+同源受控旧 SW fixture 检验（须 200 正确 JS MIME/实际 controller）或旧终端 HOLD；不强制生产新增 sw.js/PWA；clean profile 证据与旧 worker 路径分开；M4 构建/fixture、M5 真 root 可复用同构建但证据分开；本轮两文件仅对齐引用，缓存策略归主代理 | 边界记录，NOT_RUN |

## 二、逐项处置表

### 后端 Go（直接 KEEP 24 + continuation 3 SPLIT + 1 DEFER + logging 2 KEEP）

| 源测试 | 实际函数（机器统计） | 处置 | 理由 | 所需新主干夹具 | 验收 ID |
|---|---|---|---|---|---|
| narrow_pilot_context_test.go（3） | TestNPMultiTurnInvalidationAndSessionExpiry；TestNPContextExpiryAndChangedDraft；TestNPContextEndReleasesExpiryTimer | KEEP | 来源绑定、多轮失效与到期是来源协议门禁 | 现有 sqlite/context 夹具等价移植 | I07 |
| narrow_pilot_review_pg_test.go（3） | TestNPPostgresNativeCannotForgeSourcePointer；TestNPPostgresLegacyNullMetadataIsInitialized；TestNPPostgresMalformedIntakeIDIsBadRequest | KEEP | 指针防伪造/旧库初始化为来源与 schema 门禁 | npPGFixture 等价物（NP_TEST_DSN 强制 127.0.0.1:18790 + np_first_slice_dev） | I06、I14 |
| narrow_pilot_pg_test.go（7） | TestNPPostgresSaveReopenConcurrencyAndReceipts；TestNPPostgresActualDisconnectAfterCommit；TestNPPostgresActualDisconnectBeforeCommit；TestNPPostgresParallelWritersAndBusyAbandon；TestNPPostgresSourcePermissionsAndStrictBoundary；TestNPPostgresLegacySaveAfterStrictManualStillChecksRevision；TestNPPostgresDescriptionWhitespaceIsExact | KEEP（scope 见下） | 并发、真实断线、回执、鉴权、revision 校验不后置；scope：TestNPPostgresSaveReopenConcurrencyAndReceipts 剔除 materialVersions 两次调用（:170/:190）及其文件流断言；preserved() 快照内 material_files join 改为主干 B 的 I09-MAIN 比较（见 §五） | npPGFixture 等价物；fixture 不再注册 case-materials 文件路由（源 :95-99 三条路由归 I09-FILES） | I05、I08、I13、I09-MAIN |
| intake_draft_revision_test.go（6） | newDraftRevisionDB（helper，:26）；TestUpdateIntakeRejectsMissingExpectedRevisionWithoutMutation；TestUpdateIntakeRejectsStaleRevisionWithoutOverwrite；TestUpdateIntakeSwitchClientUsesNewMasterIdentity；TestConfirmIntakeFactsRotatesRevisionAndRejectsStaleHeader；TestDraftRestoreReturnsRevisionAndProtectedIdentityStatus；TestDraftListCarriesRevisionTokens | KEEP | 原生 revision 门禁必跑；newDraftRevisionDB 实际为 AutoMigrate Client + 手工 DDL 五表（case_intakes/case_intake_parties/users/cases/case_ethical_wall_whitelist）+ seed 客户 33/999，与 material_files 无关 | 计划名：newDraftRevisionDB 等价移植（含 AutoMigrate Client，非只建五表） | I01、I05、I13 |
| intake_draft_revision_contract_test.go（5） | TestUpdateIntakeInvalidatesConfirmedSnapshotAndRotatesRevision；TestLegacyDraftWithoutStoredRevisionLoadsAndSavesOnce；TestConfirmAndConflictCheckAuthorizeBeforeRevisionHint；TestUpdateIntakeWithoutMetadataKeepsStoredBusinessFields；TestUpdateIntakeMergesProvidedMetadataWithoutErasingStoredFields | KEEP | 契约级 metadata/快照/鉴权顺序为原生保存合同 | 复用 newDraftRevisionDB | I05、I06、I13 |
| intake_continuation_handler_test.go（4） | TestIntakeContinuationHandlerReturnsVerifiedDecision | DEFER | 正向 verified-decision 依赖 can_submit/can_reuse_check、完整 facts_version/conflict 复核链与 /continuation 端点，超出首切片；原文件原断言保留，不为跑旧测试移植整套 IntakeContinuationService | 无 | 后置（DEFERRED/HOLD） |
| 同上（3） | TestIntakeContinuationHandlerRejectsStaleRevision；ForbidsOtherLawyer；MissingIntake | SPLIT 语义（M0-R01） | 陈旧保护复用已 KEEP 的 TestUpdateIntakeRejectsStaleRevisionWithoutOverwrite，并在 M1 真实 router 补"陈旧 PUT 409 后同用户 GET 最新"；本人 200/越权 403/业务缺失 404 经已注册主干详情入口验证，不能借未注册路由 404 过验 | 计划详情用例 3 个：TestMainDraftDetailSelfReturnsLatest / TestMainDraftDetailForbidsOtherLawyer / TestMainDraftDetailMissingIntake；脱敏与 list alias 断言扩充现有恢复/list 测试 | I04、I05 |
| middleware/narrow_pilot_logging_test.go（1） | TestNarrowPilotAuditSkipsSensitiveBodyBeforeRead | KEEP | 审计中间件不得读敏感正文，为工厂日志/审计门禁 | 现有 gin httptest 夹具等价移植 | I11、审计 |
| database/narrow_pilot_pool_logging_test.go（1） | TestOptimizedDatabaseSlowQueryLogDoesNotLeakParameters | KEEP | 慢查询日志不得泄漏绑定参数 | OptimizedDatabase 内存夹具等价移植 | I11、审计 |

Go 合计（机器统计）：直接 KEEP 24（context3 + review3 + pg7 + draft-rev6 + contract5）+ logging 2 = 26 个现有函数必跑；continuation 原 4 项中正向提交判断后置，其余 3 个场景并入真实 router 主干详情/恢复入口与已有陈旧 PUT 测试；原文件不变。

### 前端单元（5 套，历史展开 33 项；新 main 未运行）

| 套件 | 展开用例数 | 逐项证据 |
|---|---:|---|
| frontend/src/components/intake/__tests__/NarrowPilotAssistant.test.tsx | 3 | JSON frozen_frontend_expanded_cases 逐名列出 |
| frontend/src/components/intake/__tests__/NarrowPilotAssistantScope.test.tsx | 4 | JSON frozen_frontend_expanded_cases 逐名列出 |
| frontend/src/components/intake/__tests__/useNarrowPilotSave.test.tsx | 22 | JSON frozen_frontend_expanded_cases 逐名列出 |
| frontend/src/pages/auth/__tests__/Login.failure.test.tsx | 3 | JSON frozen_frontend_expanded_cases 逐名列出 |
| frontend/src/pages/auth/__tests__/Login.accessibility.test.tsx | 1 | JSON frozen_frontend_expanded_cases 逐名列出 |

参数化 test.each 不能按字面 test 块计数。固定源历史 Jest 日志 5 套 33 项用于清单来源，不是新 main 复跑结果。

ADAPT 要求（M0-R04，计划非已存在）：useNarrowPilotSave 源成功 abandon 夹具缺 intake id、成功 committed 夹具缺 request_id，旧单槽 scope 断言不适用 v2；回执服务端身份字段是 id，恢复标识是 intake_id，compare 按 intake_id 匹配；成功夹具须 echo 真实捕获的 request_id + id；action 仅在响应实际返回时才比对；committed/no_change 夹具带合法 revision，abandoned 不要求 revision；旧缺字段场景改为负向保键保输入；按请求 v2 键断言 wrong ids/unknown/doublepages。保留原场景与源历史 33 项，不伪称逐字 KEEP，也不声称新 main 已通过。

### 浏览器链（4 条场景，KEEP，逐项原字符串）

| # | 原 test 字符串（np-first-slice.live.spec.ts） | 依赖 | 验收 ID |
|---|---|---|---|
| 1 | 真实服务：多轮核对，仅应用摘要，原生保存后新会话从原菜单重开来源 | 18789 合成服务 + intake …9801 | I07 |
| 2 | 真实服务：A 来源保存后，B 旧版纯手工摘要不得覆盖并保留输入 | 同上 | I05/I07 |
| 3 | 真实上游断线：提交后断开 TCP，重登按原键回查且重开来源 | 同上 + X-NP-Drop-Response 故障注入（仅 np_local_check 程序） | I08 |
| 4 | 生成期间保留未发送输入；人工摘要不被候选静默覆盖 | 同上 | I07 |

门禁与数据边界：四条场景先适配 B fixture，spec/config M3 准备、M4 验证（M0-R05）；CI 正常 I11 在 fixture 完成后执行；M5 只真 root 入口（真实登录/PG/Redis，外置合成 TCP 故障代理），fixture 结果不混称真 main；clean profile 证据与旧 worker 路径分开（M0-R08 仅引用）。浏览器链运行前置不预置两版材料；B seed 按 M0-R06 供给；negative cases 用独立 seed/db 窗口，不污染 source+3/receipt+4 精确计数；M5 真实入口另独立合成数据，不开放正式故障头。

### 脚本

| 脚本 | 处置 | 理由 |
|---|---|---|
| init-db.sh | **SPLIT** | 源脚本明确按序执行 000082/000083/000084/000085（:56 循环），其中 000084 建 material_files 与 case_tasks，主干 B 不得带入；主干版 = v14 bootstrap（BootstrapProductionSchema，production_schema.go:16 版本 v14）+ 仅追加源 000086_narrow_pilot 原 SQL；源脚本原样保留 |
| seed-browser.py | SPLIT（材料两版预置归 I09-FILES） | 主干浏览器链不预置两版材料；源 seed 仅一个材料两版，不满足 B，M3 由隔离 B seed 入口（计划）按 M0-R06 供给 |
| cmd/np-local-check seed（main.go） | SPLIT/扩展（计划） | 源 seed 仅建 target/control 两 intake + client + parties，无材料行；M3 主干 B seed 需 >=2 独立 material rows、非空 clients/parties、target+control 精确 ids、>2^53 未知 metadata，并返回 ids 供 compare 使用，不硬猜 |
| compare-browser-db.py | SPLIT（files==2 与 protected 内 material_files join 散列）+ KEEP（sources+3/receipts+4/revision 变化/source_revision 清空断言，主干版另立） | 来源/回执断言必跑；文件断言后置；**源完整脚本不变** |
| run-local-check.sh、probe-db.sh、compose.yml | KEEP | 隔离边界 127.0.0.1:18790/np_first_slice_dev、np_local_check 标签不后置 |

## 三、I15 场景拆分（M0-R03，计划非已存在）

- M1 可达场景：合法原生已确认事实修改 -> snapshot 失效、转 draft、revision 仅推进一次；旧 sync 写回与服务级 async 回写拒绝；intake status/metadata/source 指针与关联 cases 保持不变。
- 证据边界：旧公共 async intake 入口的拒绝只记录为入口拒绝，不算 late-write 证据。
- M3：再 NP 摘要来源 overlap 场景；M5：真入口验证（真 root main）。
- 计划组名维持 TestNPMainConflictCompletionUsesInputRevision；子场景清单以 JSON planned_main_regression_groups.I15 为准。

## 四、总数口径（机器统计，组数与用例数分开）

| 层 | 组数 | 现有用例数 | 计划新增用例数 |
|---|---|---|---|
| Go 直接 KEEP | 7 个文件 | 26（context3+review3+pg7+draft-rev6+contract5+logging2） | 0（newDraftRevisionDB 为等价移植的 helper，非新用例） |
| Go continuation | 1 个文件 | 0 直接 KEEP | 3 个详情用例（本人 200/403/404，均真实 router）；陈旧写入复用已有 PUT 拒绝用例并补同用户 GET 最新；GET 恢复可取最新 |
| Go DEFER | — | 1（continuation 正向） | 0 |
| 前端单元 | 5 套件 | 33（固定源 Jest 证据展开参数化用例，不以 11 个字面 test 块计数） | 0（useNarrowPilotSave 为夹具 ADAPT，不新增历史计数） |
| 浏览器 | 1 spec | 4 场景 | 0（另有 negative 独立 seed/db 窗口计划，不混入 63） |
| 合计 | — | 63 个既有必跑来源项（来源项计数，不是集成通过） | 另有详情 router 计划用例、I15 拆分场景、save ADAPT、B seed 等新增/改造项，不混入 63 |

## 五、I09-MAIN 主干 B 比较固定合同（冻结，计划非已存在）

1. 主干 B **不依赖 material_files**；I09-MAIN 保护对象固定为：case_materials、case_intake_parties、clients 及对照草稿（ID 由夹具返回）现有记录。
2. 比较方式：整行 to_jsonb 按主键排序；Go 测试用 crypto/sha256、Python 比较脚本用 hashlib.sha256 在内存计算相同 PG 规范文本摘要；不经过 JS Number，不输出原始正文。
3. 仅排除目标 intake 的 description、updated_at、metadata.draft_revision、metadata.np_description_source_revision；其他字段及保护表/对照草稿整行纳入。
4. 非空门禁：至少 2 条既有材料记录、metadata 含未知字段大整数（如 np_unrelated.big=9007199254740993 量级）；禁止以空表证明保护；ids 由 seed 返回，不硬猜。
5. 冻结计划名：scripts/np-first-slice/compare-main-browser-db.py；源 scripts/np-first-slice/compare-browser-db.py 完整保留不变。

## 六、I09-FILES 后置清单（源断言不变，仅记录）

- narrow_pilot_pg_test.go materialVersions（:271-329：两版上传、files==2、download_url 逐字节比对）。
- npPGFixture 的 case-materials 文件路由注册（:95-99）与 StorageDir 临时目录。
- preserved() 快照内 material_files join 字段（源实现保留；主干版按 §五 替换）。
- seed-browser.py 两版材料预置链；compare-browser-db.py 的 files==2 与 protected files 散列断言。
- 后置验收：I09-FILES = DEFERRED/HOLD，在独立材料基础迭代中执行；不以 Skip、删除断言或假文件变为 PASS。

## 七、CI 清单引用口径

- CI 清单统一引用：26 Go + 33 展开前端 + 4 浏览器历史来源，及 ADAPT/新增项；源历史 63 项是来源项，不是集成通过。
- 主计划已改为引用逐名26 Go/33展开前端/4浏览器来源及ADAPT/新增组，不再使用“原生7项”泛称。

## 八、未决与 NOT_RUN

- 本轮未运行任何测试/构建/服务：全部执行项标 NOT_RUN；本轮只修订规则，不声称运行漏洞关闭或 M1 已放行。
- M1/M3 修改主干测试夹具前仍须按方案 §5 交独立审核；I09-MAIN 排除清单在 §五 及 JSON 已冻结，后续审核核对实际符合，不再放宽。
- 主干版 init 脚本、B seed 入口与 compare-main-browser-db.py 的实际编写属后续阶段，本轮不实现代码。
- M0-R02/M0-R07：规则由 dependency-seal 主体保全与迁移条目承接，见 dependency-seal.md §三/§四。

## 九、前轮复核记录（保留）

26 个 Go 函数逐名核对固定源；5 套前端实际 33 个展开用例取自固定源历史 Jest 记录（不是本轮复跑），4 条浏览器场景另列，共 63 个既有必跑来源项。JSON 另列 I09/I14/I15/I16/C01–C14 新增计划场景，不能只跑历史 63 项就宣称完整验收。

续办原测试在源完整保留：正向提交/复核能力后置，陈旧保护复用现有 PUT 拒绝用例，不能给正常详情/恢复 GET 新增 409 门槛导致无法取最新。403/404 并入真实 router 详情用例；恢复脱敏断言覆盖客户及对方主体。

I09-MAIN 整行保护、字段排除、B 比较脚本名已在 M0 冻结；只有代码/运行属后续。main B 初始化 v14+86，不带 82–85；源材料文件断言不删、不 Skip。规范文本的 Go/Python 摘要输入一致，避免大整数重新解析丢精度。

主干脚本路径也须适配：run-local-check.sh:13/:18 的 backend 目录改为 root，隔离规则原样保留。源 4 条浏览器场景先适配 B fixture；源 X-NP-Drop-Response 故障头仅用于 np_local_check 标签程序。M5 真实 main 必须真登录/PG/Redis 并用外置合成 TCP 故障代理，正式 handler 不接收测试故障头；两种入口的结果分别记录。

## 十、本轮修订记录（M0-R01..R08 登记版）

- 新增 §一 问题编号对照表，统一 M0-R01..R08 映射 P1-1..5/P2-6..8，与原 MI-R01..08 区分；R02/R07 交叉引用 dependency-seal，见 §一。
- R01：continuation 3 场景并入真实 router 主干详情入口，计划详情用例由 2 个扩为 3 个（新增本人 200），明确未注册 404 不得作为证据，陈旧 PUT 409 后补同用户 GET 最新。
- R03：新增 §三 I15 场景拆分，区分服务级 async 回写拒绝与旧公共 async intake 入口拒绝（后者不算 late-write 证据）。
- R04：新增前端 ADAPT 要求（成功夹具 echo 真实 request_id/id（回执id与标识intake_id比较）、v2 键断言、负向保键保输入）。
- R05：新增浏览器 fixture/CI/M5 门禁与数据边界段。
- R06：脚本表新增 np-local-check seed 行与 B seed 供给要求。
- R02/R07：按用户裁定改为交叉引用 dependency-seal 主体保全与迁移规则，不写范围外；R04 回执字段合同落实为 id/intake_id/request_id/action/revision 具体规则。
- R08：按主代理 C10 拆分记录 PROD 资产/注册观测与受控旧 SW fixture 边界，本轮仅对齐引用。
- 主仓库基线指针由 213da4c 更新为 c8716b7；JSON revision 更新为 route-b-r1-issue-register。
