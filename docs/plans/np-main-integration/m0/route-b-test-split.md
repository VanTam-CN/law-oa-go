# M0 路线 B 测试拆分清单（t_cbc0ab2c，复审修订版）

- 日期：2026-11-20；状态：文档/证据收口，未实现、未运行、未推送。
- 源仓库：/Users/mac/Desktop/FT/law-oa-go/output/np-v04-first-slice @ 2d659d9；主仓库 HEAD 213da4c；应用 main d9c92fb。
- 依据：implementation-and-acceptance-plan.md §5/§6 已确认路线 B；复审追加裁决：continuation 正向用例 DEFER，3 个鉴权/陈旧/缺对象场景 SPLIT 到现有详情/恢复入口；logging 工厂/审计测试与前端 3 NP + 2 Login 套件列为必跑。
- 所有"计划新增"均为未来主干用例的命名冻结，非现有测试；原文件原断言在源仓库不变。

## 一、逐项处置表

### 后端 Go（直接 KEEP 24 + continuation 3 SPLIT + 1 DEFER + logging 2 KEEP）

| 源测试 | 实际函数（机器统计） | 处置 | 理由 | 所需新主干夹具 | 验收 ID |
|---|---|---|---|---|---|
| narrow_pilot_context_test.go（3） | TestNPMultiTurnInvalidationAndSessionExpiry；TestNPContextExpiryAndChangedDraft；TestNPContextEndReleasesExpiryTimer | KEEP | 来源绑定、多轮失效与到期是来源协议门禁 | 现有 sqlite/context 夹具等价移植 | I07 |
| narrow_pilot_review_pg_test.go（3） | TestNPPostgresNativeCannotForgeSourcePointer；TestNPPostgresLegacyNullMetadataIsInitialized；TestNPPostgresMalformedIntakeIDIsBadRequest | KEEP | 指针防伪造/旧库初始化为来源与 schema 门禁 | npPGFixture 等价物（NP_TEST_DSN 强制 127.0.0.1:18790 + np_first_slice_dev） | I06、I14 |
| narrow_pilot_pg_test.go（7） | TestNPPostgresSaveReopenConcurrencyAndReceipts；TestNPPostgresActualDisconnectAfterCommit；TestNPPostgresActualDisconnectBeforeCommit；TestNPPostgresParallelWritersAndBusyAbandon；TestNPPostgresSourcePermissionsAndStrictBoundary；TestNPPostgresLegacySaveAfterStrictManualStillChecksRevision；TestNPPostgresDescriptionWhitespaceIsExact | KEEP（scope 见下） | 并发、真实断线、回执、鉴权、revision 校验不后置；scope：TestNPPostgresSaveReopenConcurrencyAndReceipts 剔除 materialVersions 两次调用（:170/:190）及其文件流断言；preserved() 快照内 material_files join 改为主干 B 的 I09-MAIN 比较（见 §二） | npPGFixture 等价物；fixture 不再注册 case-materials 文件路由（源 :95-99 三条路由归 I09-FILES） | I05、I08、I13、I09-MAIN |
| intake_draft_revision_test.go（6） | newDraftRevisionDB（helper，:26）；TestUpdateIntakeRejectsMissingExpectedRevisionWithoutMutation；TestUpdateIntakeRejectsStaleRevisionWithoutOverwrite；TestUpdateIntakeSwitchClientUsesNewMasterIdentity；TestConfirmIntakeFactsRotatesRevisionAndRejectsStaleHeader；TestDraftRestoreReturnsRevisionAndProtectedIdentityStatus；TestDraftListCarriesRevisionTokens | KEEP | 原生 revision 门禁必跑；newDraftRevisionDB 实际为 AutoMigrate Client + 手工 DDL 五表（case_intakes/case_intake_parties/users/cases/case_ethical_wall_whitelist）+ seed 客户 33/999，与 material_files 无关 | 计划名：newDraftRevisionDB 等价移植（含 AutoMigrate Client，非只建五表） | I01、I05、I13 |
| intake_draft_revision_contract_test.go（5） | TestUpdateIntakeInvalidatesConfirmedSnapshotAndRotatesRevision；TestLegacyDraftWithoutStoredRevisionLoadsAndSavesOnce；TestConfirmAndConflictCheckAuthorizeBeforeRevisionHint；TestUpdateIntakeWithoutMetadataKeepsStoredBusinessFields；TestUpdateIntakeMergesProvidedMetadataWithoutErasingStoredFields | KEEP | 契约级 metadata/快照/鉴权顺序为原生保存合同 | 复用 newDraftRevisionDB | I05、I06、I13 |
| intake_continuation_handler_test.go（4） | TestIntakeContinuationHandlerReturnsVerifiedDecision | DEFER | 正向 verified-decision 依赖 can_submit/can_reuse_check、完整 facts_version/conflict 复核链与 /continuation 端点，超出首切片；原文件原断言保留，不为跑旧测试移植整套 IntakeContinuationService | 无 | 后置（DEFERRED/HOLD） |
| 同上（3） | TestIntakeContinuationHandlerRejectsStaleRevision；ForbidsOtherLawyer；MissingIntake | SPLIT语义 | 陈旧保护复用已KEEP的TestUpdateIntakeRejectsStaleRevisionWithoutOverwrite；原生详情/恢复GET可取最新，不新增409门槛 | 2个计划详情用例：TestMainDraftDetailForbidsOtherLawyer / TestMainDraftDetailMissingIntake；脱敏断言扩充现有恢复测试 | I04、I05 |
| middleware/narrow_pilot_logging_test.go（1） | TestNarrowPilotAuditSkipsSensitiveBodyBeforeRead | KEEP | 审计中间件不得读敏感正文，为工厂日志/审计门禁 | 现有 gin httptest 夹具等价移植 | I11、审计 |
| database/narrow_pilot_pool_logging_test.go（1） | TestOptimizedDatabaseSlowQueryLogDoesNotLeakParameters | KEEP | 慢查询日志不得泄漏绑定参数 | OptimizedDatabase 内存夹具等价移植 | I11、审计 |

Go 合计（机器统计）：直接 KEEP 24（context3 + review3 + pg7 + draft-rev6 + contract5）+ logging 2 = 26 个现有函数必跑；continuation原4项中正向提交判断后置，其余3个场景分别映射到已有陈旧PUT测试与2个计划详情用例；原文件不变。

### 前端单元（5套，历史展开33项；新main未运行）

| 套件 | 展开用例数 | 逐项证据 |
|---|---:|---|
| frontend/src/components/intake/__tests__/NarrowPilotAssistant.test.tsx | 3 | JSON frozen_frontend_expanded_cases逐名列出 |
| frontend/src/components/intake/__tests__/NarrowPilotAssistantScope.test.tsx | 4 | JSON frozen_frontend_expanded_cases逐名列出 |
| frontend/src/components/intake/__tests__/useNarrowPilotSave.test.tsx | 22 | JSON frozen_frontend_expanded_cases逐名列出 |
| frontend/src/pages/auth/__tests__/Login.failure.test.tsx | 3 | JSON frozen_frontend_expanded_cases逐名列出 |
| frontend/src/pages/auth/__tests__/Login.accessibility.test.tsx | 1 | JSON frozen_frontend_expanded_cases逐名列出 |

参数化test.each不能按字面test块计数。固定源历史Jest日志5套33项用于清单来源，不是新main复跑结果。

### 浏览器链（4 条场景，KEEP，逐项原字符串）

| # | 原 test 字符串（np-first-slice.live.spec.ts） | 依赖 | 验收 ID |
|---|---|---|---|
| 1 | 真实服务：多轮核对，仅应用摘要，原生保存后新会话从原菜单重开来源 | 18789 合成服务 + intake …9801 | I07 |
| 2 | 真实服务：A来源保存后，B旧版纯手工摘要不得覆盖并保留输入 | 同上 | I05/I07 |
| 3 | 真实上游断线：提交后断开TCP，重登按原键回查且重开来源 | 同上 + X-NP-Drop-Response 故障注入 | I08 |
| 4 | 生成期间保留未发送输入；人工摘要不被候选静默覆盖 | 同上 | I07 |

浏览器链运行前置不预置两版材料（seed-browser.py 归 I09-FILES）。

### 脚本

| 脚本 | 处置 | 理由 |
|---|---|---|
| init-db.sh | **SPLIT** | 源脚本明确按序执行 000082/000083/000084/000085（:56 循环），其中 000084 建 material_files 与 case_tasks，主干 B 不得带入；主干版 = v14 bootstrap（BootstrapProductionSchema，production_schema.go:16 版本 v14）+ 仅追加源 000086_narrow_pilot 原 SQL；源脚本原样保留 |
| seed-browser.py | SPLIT（材料两版预置归 I09-FILES） | 主干浏览器链不预置两版材料 |
| compare-browser-db.py | SPLIT（files==2 与 protected 内 material_files join 散列）+ KEEP（sources+3/receipts+4/revision 变化/source_revision 清空断言，主干版另立） | 来源/回执断言必跑；文件断言后置；**源完整脚本不变** |
| run-local-check.sh、probe-db.sh、compose.yml | KEEP | 隔离边界 127.0.0.1:18790/np_first_slice_dev、np_local_check 标签不后置 |

## 二、I09-MAIN 主干 B 比较固定合同（冻结，计划非已存在）

1. 主干 B **不依赖 material_files**；I09-MAIN 保护对象固定为：case_materials、case_intake_parties、clients 及对照草稿（ID由夹具返回）现有记录。
2. 比较方式：整行 to_jsonb 按主键排序；Go测试用crypto/sha256、Python比较脚本用hashlib.sha256在内存计算相同PG规范文本摘要；不经过JS Number，不输出原始正文。
3. 仅排除目标intake的description、updated_at、metadata.draft_revision、metadata.np_description_source_revision；其他字段及保护表/对照草稿整行纳入。
4. 非空门禁：至少 2 条既有材料记录、metadata 含未知字段大整数（如 np_unrelated.big=9007199254740993 量级）；禁止以空表证明保护。
5. 冻结计划名：scripts/np-first-slice/compare-main-browser-db.py；源 scripts/np-first-slice/compare-browser-db.py 完整保留不变。

## 三、I09-FILES 后置清单（源断言不变，仅记录）

- narrow_pilot_pg_test.go materialVersions（:271-329：两版上传、files==2、download_url 逐字节比对）。
- npPGFixture 的 case-materials 文件路由注册（:95-99）与 StorageDir 临时目录。
- preserved() 快照内 material_files join 字段（源实现保留；主干版按 §二 替换）。
- seed-browser.py 两版材料预置链；compare-browser-db.py 的 files==2 与 protected files 散列断言。
- 后置验收：I09-FILES = DEFERRED/HOLD，在独立材料基础迭代中执行；不以 Skip、删除断言或假文件变为 PASS。

## 四、总数口径（机器统计，组数与用例数分开）

| 层 | 组数 | 现有用例数 | 计划新增用例数 |
|---|---|---|---|
| Go 直接 KEEP | 7个文件 | 26（context3+review3+pg7+draft-rev6+contract5+logging2） | 0（newDraftRevisionDB 为等价移植的 helper，非新用例） |
| Go continuation | 1个文件 | 0直接KEEP | 2个详情403/404适配；陈旧写入复用已有PUT拒绝用例，GET恢复可取最新 |
| Go DEFER | — | 1（continuation 正向） | 0 |
| 前端单元 | 5套件 | 33（固定源Jest证据展开参数化用例，不以11个字面test块计数） | 0 |
| 浏览器 | 1 spec | 4 场景 | 0 |
| 合计 | — | 63个既有必跑来源项 | 另有2个详情适配和新安全/缓存场景，不混入63 |

## 五、未决与 NOT_RUN

- 本轮未运行任何测试/构建/服务：全部执行项标 NOT_RUN。
- M1/M3 修改主干测试夹具前仍须按方案 §5 交独立审核；I09-MAIN排除清单在§二及JSON已冻结，后续审核核对实际符合，不再放宽。
- 主干版 init 脚本与 compare-main-browser-db.py 的实际编写属后续阶段，本轮不实现代码。


## 六、主代理最终复核

26个Go函数逐名核对固定源；5套前端实际33个展开用例取自固定源历史Jest记录（不是本轮复跑），4条浏览器场景另列，共63个既有必跑来源项。JSON另列I09/I14/I15/I16/C01–C14新增计划场景，不能只跑历史63项就宣称完整验收。

续办原测试在源完整保留：正向提交/复核能力后置，陈旧保护复用现有PUT拒绝用例，不能给正常详情/恢复GET新增409门槛导致无法取最新。403/404映射到两个计划详情用例；恢复脱敏断言覆盖客户及对方主体。

I09-MAIN整行保护、字段排除、B比较脚本名已在M0冻结；只有代码/运行属后续。main B初始化v14+86，不带82–85；源材料文件断言不删、不Skip。规范文本的Go/Python摘要输入一致，避免大整数重新解析丢精度。

主干脚本路径也须适配：run-local-check.sh:13/:18的backend目录改为root，隔离规则原样保留。源4条浏览器场景先适配B fixture；源X-NP-Drop-Response故障头仅用于np_local_check标签程序。M5真实main必须真登录/PG/Redis并用外置合成TCP故障代理，正式handler不接收测试故障头；两种入口的结果分别记录。
