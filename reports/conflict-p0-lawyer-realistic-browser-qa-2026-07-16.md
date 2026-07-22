# 律所多角色冲突检查仿真验收报告

日期：2026-07-16  
环境：本地真实 API 环境，前端 `http://127.0.0.1:3003`，后端 `http://127.0.0.1:8080`
测试方式：Chrome 内部浏览器按可见页面进行点击、输入、导航，并结合服务日志记录；按律师、律师 B 角、助理、独立冲突核查岗的真实操作顺序执行  
数据原则：仅使用仓库中的演练账号和虚构主体，不使用真实客户、真实案件或真实身份号码

## 一、结论

本轮确认了三件事：

1. 对普通律师，系统现在把“检索完成”“范围完整”“人工复核结论”分开显示；覆盖范围不足时不能提交审批，也不能把低机器风险误读为“无冲突”。
2. 案件进入正式办理后，详情页已有“报告主体变更并重新复核”入口。主体变更在独立复核前保持阻断，助理账号不显示该操作。
3. 当前版本仍不能作为不受限制的真实律所生产版本发布。主要原因不是按钮，而是权威档案尚未完成导入确认、主体库正向数据不足、律所 PD-01 至 PD-07 仍是决策草案，以及生产密钥和外部受控动作尚未完成上线验收。

当前适用结论：可以作为“失败关闭”的内部 QA 或受控试用版本；不能对律师承诺“系统已确认无冲突”或允许在未完成律所政策确认的情况下直接放行案件。

## 二、仿真角色与场景

| 角色 | 仿真任务 | 结果 |
|---|---|---|
| 律师 A | 新建案件、运行冲突检查、查看本案检测上下文、进入在办案件、报告主体变更 | 主路径可操作；范围受限时正确阻断 |
| 律师 B 角 | 登录后尝试查看律师 A 的工作台、冲突任务、案件详情 | 未看到 A 的案件标题和客户；直接访问 A 案件返回“案件不存在或当前账号无权访问” |
| 助理 | 新建案件、整理草稿、寻找客户/对方/身份标识录入和冲突检查入口 | 仅能整理协作草稿；不能录入主体身份、运行最终检查或提交审批 |
| 独立冲突核查岗 | 查看冲突清单、打开完整证据、查看待复核记录、查看复核结论控件 | 可看到完整历史证据和人工复核控件；普通律师看不到历史案件属性 |

仿真案件使用了“股权回购争议”“建设工程款争议”“主体版本复核”等虚构业务标题，主体名称使用演练数据。未为本轮强行写入新的 COMPLETE 覆盖范围，也未用假数据把受限结果改成通过。

## 三、浏览器验收记录

### 3.1 律师 A：新建案件与冲突检查

操作路径：工作台 → 新建立案 → 填写案件基本信息、客户、对方当事人、案情摘要 → 运行利益冲突检查 → 进入本案冲突复核。

通过点：

- 页面显示“检索范围受限，需人工复核”，没有显示“无冲突”或“已确认通过”。
- 结果摘要明确区分“机器风险：低风险”和“处置状态：范围受限，待人工复核”。
- “提交审批并等待成案”按钮存在但不可用，符合范围不完整时阻断的规则。
- 从新建案件进入冲突页后显示本案上下文，案件标题与检测任务一致，没有把历史命中案件错误地当成本案。
- 检测详情在没有实际证据时显示“未发现匹配记录”，不再把旧的默认匹配主体展示为事实。
- 普通律师打开受限命中时只显示“受限命中”及联系冲突核查岗的指引，不显示历史案件编号、案情或承办团队。

使用评价：

- 状态和按钮已经符合不会 IT 的律师的判断顺序：先知道“现在能不能继续”，再知道“下一步找谁处理”。
- “检索范围受限”仍需要律所上线时配置具体的权威档案来源和缺口处置规则，否则律师只能看到阻断，无法在系统内完成闭环。

### 3.2 律师 A：利益冲突清单与直接冲突处理

操作路径：利益冲突检查 → 清单 → 查看详情。

通过点：

- 默认进入检测任务清单，不自动弹出某一条历史检测详情。
- 普通律师能看到与本人工作相关的任务和必要的结果摘要，但不能看到受隔离历史事项的业务属性。
- 对已确认直接冲突的记录，页面不再显示“申请豁免评估”按钮；后端也拒绝直接冲突豁免请求。
- “待独立人工复核”“已暂停接案”“等待独立复核”等文案能区分不同处置状态。

使用评价：

- 直接冲突被隐藏豁免入口是正确的安全行为，但正式上线前还要由律所确认哪些冲突类型属于绝对不可豁免，不能只依赖代码中的规则名。

### 3.3 案件详情与下一步操作

操作路径：案件管理 → 打开待处理案件。

通过点：

- 待处理案件显示“下一步操作”卡片。
- 卡片显示“下一步：利益冲突复核”，并提供“进入本案冲突复核”和“补充立案信息并重新检测”。
- 按钮文字直接表达动作，没有要求律师理解 `case_id`、任务 ID 或内部状态码。

新增主体变更入口：

- 办理中案件显示“案件主体与冲突门禁”卡片和“报告主体变更并重新复核”。
- 弹窗要求选择已登记的结构化主体，不接受只输入一段自由文本就推进正式案件。
- 身份标识只显示末四位提示，不显示完整身份证号或统一社会信用代码。
- 提交后自动运行重检；在独立复核前显示“主体变更待独立复核，受控动作已暂停”。
- 当前主体库没有匹配演练数据时，页面明确提示联系冲突核查岗登记主体，且提交按钮保持不可用。

### 3.4 律师 B 角：隔离与越权

操作路径：律师 B 登录 → 工作台 → 冲突清单 → 案件清单 → 直接输入律师 A 的案件地址。

通过点：

- 工作台未显示律师 A 的专属案件标题或客户。
- 冲突清单未显示律师 A 的任务或受限命中内容。
- 案件清单未显示律师 A 的案件。
- 直接访问律师 A 案件编号时，页面只显示“案件不存在”或“当前账号无权访问”，没有泄露案件标题、客户或状态。

限制：

- 当前本地数据库中律师 B 的正向自有案件和冲突任务数量不足，无法完成“B 能看到自己的案件，同时看不到 A 的案件”的完整对照验收。当前结果证明了负向拒绝，但不替代三角色正式验收。

### 3.5 助理：协作边界

操作路径：助理登录 → 工作台 → 新建立案。

通过点：

- 页面明确显示“助理协作草稿”。
- 页面允许整理案件摘要和材料清单。
- 没有客户、对方、关联方和身份标识录入控件。
- 没有“运行利益冲突检查”或“提交审批”按钮。
- 底部状态显示等待负责律师补充当事人信息并确认。

使用评价：

- 这一边界符合“助理整理、律师确认”的真实分工。
- 正式试用仍应补充“移交给负责律师”的通知和待办，不应要求助理通过口头或线下方式提醒律师。

### 3.6 独立冲突核查岗：完整证据和人工复核

操作路径：独立冲突核查岗登录 → 利益冲突清单 → 查看待复核记录 → 查看详情。

通过点：

- 可看到待复核记录、复核结论下拉框、核对依据输入框和提交复核结论按钮。
- 可看到完整历史证据链和历史案件属性。
- 与普通律师的受限投影不同，核查岗能承担独立复核职责。
- 未让案件申请人或案件负责人直接替代独立核查岗。

未提交真实复核结论，避免污染共享演练数据。

### 3.7 本轮真实 API 浏览器复跑（2026-07-16）

后端在本地 PostgreSQL、Redis、Elasticsearch 已运行的环境启动，浏览器通过真实 API 复跑律师主路径。工作台显示“正式 API”，数据库返回 6 个案件、5 个冲突待复核任务；这不是 mock 数据验收。

通过点：

- `/conflict` 默认打开检测任务清单，不自动弹出详情；真实 API 返回 13 条检测任务，其中 11 条归入“待人工复核”，高/中/低风险计数均为 0。范围受限结果没有被压成可忽略的低风险提示。
- 清单当前有 8 条待审批事项、6 个案件；这些统计来自本地 PostgreSQL 数据，不是前端 mock。普通律师看到的受限记录只保留泛化证据。
- 受限记录详情显示“风险等级：待人工复核”和“检索范围受限，需人工复核”；普通律师只看到“受限历史事项（请联系独立冲突核查人）”，未看到历史案件编号、案情或承办团队。
- 文案修复后的真实页面显示“检查范围：系统已登记历史（覆盖完整性待确认，未登记档案需人工核查）”，不再显示容易被理解成无遗漏的“系统已导入的全量历史”。
- `/case/35` 的“待处理”案件显示“下一步操作”，可进入本案冲突复核；该案件当前没有对应检测记录时，页面明确提示“当前案件暂无冲突检测记录”，并提供“补充立案信息并检测”，没有自动打开错误详情。
- 从冲突详情点击“发起冲突审批”后成功进入 `/approval/:id`。审批快照继续显示“总体风险等级：待人工复核”，并明确说明“当前结果不能作为无冲突结论，须由独立冲突核查人完成人工复核”。
- 审批尚未完成时，“查看关联案件”保持 disabled，侧栏显示等待审批通过；按钮状态与审批状态一致，没有提前生成或伪造正式案件。

本地 QA 期间创建了 1 条冲突审批记录用于验证上述跳转。该记录属于本地演练数据，正式验收前应在一次性数据库中清理或重新初始化，不得作为生产数据使用。

非阻塞工程问题及处理：详情加载会为没有豁免记录的任务请求豁免接口，后端返回 404；页面没有向律师显示错误，主流程也未受影响。冲突复核查询原先用 `First` 查询空记录，会把正常的“尚未复核”写成 `record not found` 日志；现已改为无记录返回空结果的查询，并增加回归测试，避免污染监控日志。豁免接口的 404 仍建议后续统一为空成功结果，或由前端仅在存在豁免标识时请求。

本轮发现并修正的前端问题：

1. 冲突清单原先把 `REVIEW_REQUIRED` 的数量和风险提示混在一起，容易让律师把范围受限结果看成普通提示；现在独立显示“待人工复核”，并在合规建议中说明不能据此确认无冲突。
2. 冲突详情中普通律师的操作按钮原文为“进入本案冲突复核”，实际动作是创建冲突审批；现改为“发起冲突审批”，与真实下一步一致。
3. 审批详情原先从旧风险字段显示“低风险”；现优先使用冲突决策和覆盖状态，范围受限时显示“待人工复核”，并保持证据冻结提示。

本次浏览器验证证明本地真实 API 下主路径和失败关闭路径可操作，但不改变本报告的生产阻断结论：权威档案覆盖、三角色正向隔离、律所政策签署、生产 PostgreSQL bootstrap/readiness、密钥托管和系统外对外动作接线仍未完成。

## 四、本轮代码修复

1. 重检服务不再信任浏览器传入的旧 `Parties/OtherParties`；服务器根据当前有效当事人加上本次新增/移除/替换内容生成唯一重检主体集合。
2. 重检请求从数据库补齐客户名称、客户类型和必要的实时身份信息；审计存储仍使用脱敏/摘要，不保存明文身份号码。
3. 新增案件范围内的结构化主体读取和搜索接口，按案件权限控制；搜索结果不返回完整身份号码，只返回是否存在标识和末四位提示。
4. 案件详情新增主体变更入口、待复核状态和失败关闭提示；助理账号不显示入口。
5. 直接冲突的豁免在服务端和前端同时阻断。
6. 冲突详情不再用旧字段把无证据记录显示成匹配主体；范围受限时不把机器评分当作无冲突结论。
7. 冲突检索默认改为全量历史且禁止跳过；旧 `/conflict-v2/checks` 接口也接入同一套 `REVIEW_REQUIRED` 和 `COVERAGE_LIMITED` 门禁，避免旧入口返回“可以正常受理”。
8. `conflict_search_scopes` 只有在四类来源均具备版本、核对凭证、覆盖起止时间且无缺口时才算 `COMPLETE`；生产 readiness 对 MySQL/SQLite 和 PostgreSQL 分别使用正确的参数占位符。
9. 客户/主体身份标识在模型保存时进入密文+HMAC 摘要，JSON 不返回原文；审计快照使用 keyed digest，接案自由 metadata 不再允许写入身份证号、统一社会信用代码等明文标识。
10. 增加默认只读的 `cmd/backfill-sensitive-identities`，用于盘点和经显式 `--apply` 执行历史明文回填；生产启动会阻断未完成回填、演示账号和未完成档案覆盖。
11. 前端所有冲突检索配置默认不按年限截断；页面明确说明“系统已登记历史”，并要求以档案覆盖状态判断完整性，移除“自动豁免”入口；范围受限时不能被误读为无冲突。
12. 普通律师的服务端结果投影区分“隔离命中、普通主体候选、无命中但待核查”三种状态：隔离命中只返回泛化提示；普通候选保留有限主体身份和冲突类型但隐藏历史案件细节；无命中不再虚构“存在受隔离记录”。
13. 旧 `/conflict-v2/checks` 兼容入口接入相同的角色门禁和待复核投影；旧 `/conflict/v2` 前端页面重定向到清单式 `/conflict`，避免用户进入旁路表单。
14. 后端客户主档写权限收紧为律师和律所管理角色；助理不能创建或修改客户主档，前端同步隐藏新增客户、编辑联系人和上传客户附件入口。
15. 结构化主体检索改为跨 MySQL、PostgreSQL、SQLite 通用的 `LOWER(...) LIKE LOWER(?)`，避免默认 MySQL 环境因 PostgreSQL 专属 `ILIKE` 造成名称、别名和曾用名检索失败。
16. 结构化检索新增历史客户主档分支：即使历史案件没有同步 `case_parties` 主体记录，也会按客户名称、公司字段和受保护身份证摘要检索；名称命中保持待复核，身份证摘要一致才标记为需独立复核的强标识命中。
17. 集成审批接口补上服务端角色门禁；助理即使绕过前端直接提交冲突/成案配置，也不能发起接案冲突流程，普通非案件审批仍保留原有流程。
18. 创建冲突审批前强制检测记录状态为 `COMPLETED`；接案冲突结果写回失败、检测服务未返回可追溯 `check_id` 时不再返回成功，避免出现“页面显示完成但数据库没有可复核记录”。
19. 生产镜像已包含独立的静态迁移二进制和迁移文件；Docker Compose 的后端等待迁移服务成功，Kubernetes init container 直接执行迁移二进制，不再调用 scratch 镜像中不存在的 shell 或不存在的主程序子命令。
20. 生产启动跳过演示账号/客户/案件种子；未鉴权的关闭、GC、缓存压测、内部监控和详细健康诊断端点仅在非生产环境注册，避免把开发运维面暴露到真实环境。
21. 迁移器在连接目标库前检查 SQL 方言；混合历史目录会明确失败，避免先提交一部分迁移后把数据库留在 dirty version。
22. 生产后端在注册路由前检查冲突、主体重检、审批和接案工作台所需关键表；缺表直接拒绝启动，不让服务以“能监听但请求必失败”的半初始化状态上线。
23. 新增事务化 PostgreSQL 生产 schema bootstrap 和 `law-oa-migrate -command bootstrap` 镜像入口；生产服务跳过运行时 AutoMigrate，并校验关键表和 `schema_bootstrap_state` 版本。
24. PostgreSQL bootstrap 清单排除了仍带 MySQL `enum` 标签的增强冲突执行模型；审批申请、审批记录和代理配置改为应用层生成 UUID，保持 PostgreSQL 与 SQLite 测试兼容。
25. Elasticsearch 地址兼容“主机名+端口”和历史完整 URL 两种配置，避免升级后出现双协议地址；容器健康检查现在只探测本地 `/health/live`，不会再次启动服务进程。
26. 主程序使用的 CORS 中间件补齐 `Idempotency-Key`、`X-Request-ID` 和 `X-API-Version`，避免案件立案的浏览器预检被错误拒绝。
27. 冲突清单将范围受限和其他待人工复核记录统一归入独立的“待人工复核”状态，风险提示、筛选按钮和合规建议使用同一决策状态。
28. 冲突审批快照按决策状态和覆盖状态计算总体风险，避免旧的机器评分覆盖“范围受限，需人工复核”的门禁结论。
29. 接案审批不再接受旧式自动冲突检测配置；必须引用服务端已绑定、状态为 `COMPLETED` 的接案检测记录，并校验案件名称和客户未被浏览器改写。
30. 集成冲突检查接口拒绝携带 `intakeId` 的接案请求，接案冲突检查只能通过“律师确认事实 → 受控检测”端点完成。
31. 律师修改对方或相关方后，前端检测动作会先保存当前接案事实，再确认并执行检测，避免把旧服务端结果标记为新输入的结果。
32. 本地 QA PostgreSQL 已执行生产 schema bootstrap v2；waiver 四张证据表、追加写入/禁止删除触发器和 `RESTRICT` 外键已实际核验。
33. 冲突复核查询无记录时返回空结果，不再用正常的 `record not found` 作为错误日志；增加服务层回归测试。
34. 生产环境关闭公开注册：生产路由不再注册 `/auth/register`，即使直接调用处理器也返回 `PUBLIC_REGISTRATION_DISABLED`，避免匿名创建普通账号后再寻找权限边界。
35. 审批详情、提交、重提、撤销和更新在写操作前重新读取审批并重新执行冲突上下文授权；不能依赖页面首次加载时的权限快照。
36. OnlyOffice 打开和转换文档先执行对象授权，再查询文档内容，避免未授权用户通过 `404/其他错误` 差异探测文档是否存在。
37. 生产迁移命令收紧为 PostgreSQL `bootstrap` 专用入口；生产环境拒绝历史迁移、回滚、`drop`、`status` 等命令，避免把混合方言历史目录直接用于真实库。
38. 文档权限、回收站、全文搜索、版本和统计等未接入真实审计/隔离数据源的旧服务全部保持 fail-closed；移除其可执行的 mock 返回路径，不能把演示数据当成生产数据。
39. 内容过滤正则缓存不再复制含 `sync.RWMutex` 的服务字段；改为直接使用服务实例锁，消除并发静态检查警告。
40. 新增 `000066_conflict_p0_subject_index` 及对应模型，建立 `conflict_subject_versions`、`conflict_subject_identifiers` 和 `conflict_match_evidence_v2` 三张 P0 基础表；PostgreSQL bootstrap/readiness 会校验表和追加保护。正式冲突检查在表已部署时会把本次规范化主体快照、受保护身份标识和命中证据摘要幂等双写到这些表，不能再只依赖 `conflict_check_records.check_result` JSON。
41. 生产环境在 P0 索引及其来源表均已部署时，先同步客户主档、案件当事人、别名/曾用名和关联实体的不可变版本，再从该索引读取匹配；不再以发起律师缩小检索集合。旧直接查询仅作为未部署 P0 索引的非生产兼容路径，生产 readiness 会阻止其被实际使用。
42. 新增 `000067_conflict_index_build_runs`、`ConflictIndexBuildRun` 和默认只读的 `cmd/backfill-conflict-index`；回填覆盖所有客户、案件、结构化主体、案件当事人、曾用名和独立关联关系，并为案件/客户/主体/关系四类范围生成数量闭合、零缺口、带对账哈希的运行记录。`conflict_search_scopes` 标记 COMPLETE 时必须绑定同源、同凭证的 COMPLETED run；readiness 会复核绑定而不是信任人工状态字段。

## 五、验证结果

通过：

- `go build ./...`
- `go test ./internal/config ./internal/database ./internal/health ./internal/security ./internal/router -count=1`
- `npm run type-check`
- `npm run build`
- `go test ./internal/models -run 'TestSensitiveIdentityHooksEncryptAndHidePlaintext' -count=1`
- `go test ./internal/handlers -run 'TestCreateCaseIntakeRejectsPlaintextSensitiveMetadata' -count=1`
- `go test ./internal/services -run 'TestAuditSafeConflictRequest|TestStructuredPartyConflicts|TestSubjectRecheck|TestRequireConflictDisposition|TestWaiverWorkflow' -count=1`
- 前端冲突入口测试：8 项通过；`npm run type-check` 通过

本轮增量回归（2026-07-16）：

- `go build ./...` 通过。
- `go test ./internal/services -run 'TestClientWriteRoleBoundaries|TestAuditSafeConflictRequest|TestStructuredPartyConflicts|TestSubjectRecheck' -count=1` 通过。
- `go test ./internal/handlers -run 'TestLegacyConflict|TestConflict(CheckAllowsCurrentLawyerScope|TaskResultRedactsHistoricalMatterForLawyer|TaskResultShowsLimitedIdentityAndTypeWithoutHistoricalCaseDetails|TaskResultHidesRestrictedSubjectFromLawyer)' -count=1` 通过。
- `go test ./internal/repositories -run 'TestValidateConflictRequest' -count=1` 通过。
- `go test ./internal/repositories -run 'TestEntitySearchQueriesArePortableAcrossSupportedDatabases|TestGetPotentialConflicts_CaseInsensitiveMatchOnSQLite' -count=1` 通过。
- `go test ./internal/services -run 'TestStructuredPartyConflictsSearchesHistoricalClientArchive|TestStructuredPartyConflictsMatchIdentityAcrossFirmArchive|TestStructuredPartyConflictsSearchFormerNamesAndRelatedEntities' -count=1` 通过。
- `go test ./internal/services -run 'TestBuildCheckStatisticsDoesNotClaimCompleteHistoryWithoutCoverage' -count=1` 通过，覆盖范围未确认时统计不得声称全量历史。
- `go test ./internal/services -run 'TestGetConflictReviewReturnsEmptyWithoutRecordNotFoundError' -count=1` 通过，覆盖尚未有复核记录的正常详情路径。
- `go test ./internal/handlers -run 'Test(CreateConflictApproval|BlockedConflictApproval|CanCreateIntegratedMatter)' -count=1` 通过，覆盖未完成检测拒绝和助理绕过集成接口拒绝。
- `go test ./internal/config ./internal/health ./internal/handlers ./internal/router -run 'Test(Config|CanCreateIntegratedMatter|CreateConflictApproval|BlockedConflictApproval|Router)' -count=1` 通过，覆盖生产配置、集成角色门禁和路由授权回归。
- `go build ./cmd/migrate` 与 `go build ./...` 通过。
- `go test ./internal/database -run 'TestValidateMigrationDirectory' -count=1` 通过，覆盖 MySQL/PostgreSQL 混用拒绝和可移植 SQL 放行。
- `go test ./internal/database -run 'TestElasticsearchAddress|TestValidateMigrationDirectory' -count=1` 通过，覆盖 PostgreSQL 迁移预检和 Elasticsearch 地址兼容。
- `go test ./internal/security -run TestCORSAllowsClientRequestHeaders -count=1` 通过，覆盖案件立案预检头。
- `go test ./internal/services -run 'TestStructuredPartyConflicts|TestSubjectRecheck|TestRequireConflictDisposition|TestWaiverWorkflow|TestCreateDelegation|TestApprovalServiceRejectsApplicant' -count=1` 通过。
- `go test ./internal/handlers -run 'TestLegacyConflict|TestConflict(CheckAllowsCurrentLawyerScope|TaskResultRedactsHistoricalMatterForLawyer|TaskResultShowsLimitedIdentityAndTypeWithoutHistoricalCaseDetails|TaskResultHidesRestrictedSubjectFromLawyer)|Test(CreateConflictApproval|BlockedConflictApproval|CanCreateIntegratedMatter)' -count=1` 通过。
- `frontend` 执行 `npm run type-check` 通过。
- `frontend` 执行 `npm run build` 通过；仅有 Ant Design 大分包的既有体积警告。
- 文案修复后的 `npm run type-check` 和 `npm run build` 通过；前端 E2E `case-create-full-workflow.spec.ts conflict-actions.spec.ts` 18/18 通过。
- `npm run test:e2e -- conflict-actions.spec.ts` 通过，15/15；覆盖豁免、明确无冲突、人工阻断、过期、范围受限待人工复核、打印转义、权限复核、立案门禁和审批快照。
- `npm run test:e2e -- case-create-full-workflow.spec.ts conflict-actions.spec.ts` 通过，18/18；额外验证修改对方后先 PUT 保存接案事实，再 POST 冲突检测。
- `npm run test -- --runInBand src/pages/batch01/__tests__/CaseIntakeConflictAction.test.ts` 通过，9/9；覆盖本案上下文匹配、历史命中案件不冒充主体案件、草稿隔离和结果过期。
- `go test ./internal/handlers ./internal/services ./internal/database ./internal/repositories -run 'Test(Conflict|Assistant|Subject|RequireConflictDisposition|Waiver|Production|Document|ValidateCaseIntakeApproval|TriggerConflictCheckRejectsIntakeContext)' -count=1` 通过。
- 使用占位密钥执行 `docker compose config --quiet` 通过；`git diff --check` 对本轮部署文件通过。
- `GOCACHE=/private/tmp/law-oa-go-build-cache go vet ./internal/config ./internal/database ./internal/health ./internal/handlers ./internal/repositories ./internal/router ./internal/services ./internal/security` 通过；生产代码范围没有本轮引入的不可达代码或锁拷贝警告。
- `GOCACHE=/private/tmp/law-oa-go-build-cache go test ./internal/services -count=1` 通过；`go build ./...` 通过；核心文档、OnlyOffice、通知、收件箱、审批和冲突 handler 窄回归通过。
- `frontend` 的 `npm run type-check`、`npm run lint`、`npm run build` 均通过；构建仅保留 Ant Design 大分包警告。
- 2026-07-19 P0 索引增量回归：`go test ./internal/repositories -run 'TestConflictSubjectIndexSearchesArchiveAliasesAndProtectedIdentifiers|TestSaveConflictP0EvidenceIsProtectedAndIdempotent' -count=1` 通过；验证别名、受保护身份摘要、追加证据和重试幂等。
- 2026-07-19 P0 读路径回归：`go test ./internal/services -run TestFindPotentialConflictsUsesFirmSubjectIndexWhenDeployed -count=1` 通过；验证完整 P0 表部署时正式服务从全所主体索引读取，不走旧直接文本查询。
- 2026-07-19 P0 索引对账回归：`go test ./internal/repositories -run 'TestConflictSubjectIndexSearchesArchiveAliasesAndProtectedIdentifiers|TestSaveConflictP0EvidenceIsProtectedAndIdempotent' -count=1` 和 `go test ./internal/health ./internal/services -run 'TestConflictP0ReadinessCheck|TestConflictScopeService' -count=1` 通过；验证只读盘点不写入、显式回填生成四类 COMPLETED run、范围必须绑定 run、数量缺口使 readiness 失败。
- 2026-07-19 在本机授权环境重跑 `GOCACHE=/private/tmp/go-build-cache go test ./...`，全部通过；`go build ./...`、生产 Go `vet`、前端 `npm run type-check`、`npm run lint`、`npm run build` 也全部通过。前端构建仍只有既有 Ant Design 大分包警告。
- 2026-07-19 重新执行 `GOCACHE=/private/tmp/go-build-cache go test ./...`，在允许本地临时端口和 Docker 依赖访问的环境中全量通过。期间修复了测试导入环、SQLite 不兼容 fixture、过时的服务构造函数/路由契约、Prometheus 重复注册、OnlyOffice 输入校验、用户服务 mock 隔离，以及财务主体版本门禁夹具；红字发票金额、提成幂等和软删除文档查询也已纳入回归。
- 先前一次专用 QA PostgreSQL 演练曾执行 `law-oa-migrate -command bootstrap` 成功，报告版本为 `postgres-mvp-2026-07-16-v2`，并核验 waiver 证据表、追加写入触发器和 `RESTRICT` 外键；该证据属于当时的专用实例，不自动代表当前目标数据库。
- `Makefile`、运维手册和开发手册中的迁移命令已统一为 `cmd/migrate -command ...`；不再指导调用不存在的后端 `migrate` 子命令。
- `frontend` 的 `npm run type-check`、冲突转换单测和 Prettier 检查通过。
- `go test ./internal/handlers ./internal/services/test ./test -count=1` 通过；OnlyOffice `httptest` 需要本地监听权限，已在授权环境通过。
- `go test ./internal/services -count=1`、`go test ./tests/integration -count=1`、`go test ./tests/performance -count=1` 均通过；deadline fixture、财务主体门禁和性能指标初始化问题已修复。
- 当前本机 PostgreSQL 只读探查到 `law_oa_db` 和 `law_oa_batch01_e2e`；后者仅有基础 `users`/`clients` 表，未达到 P0 schema bootstrap 条件。为避免误改业务库，本轮没有对现有数据库执行迁移、清理或写入。`docker compose config` 只证明编排语法和变量展开正确。

## 六、仍阻塞真实生产发布的问题

### P0-1：律所政策尚未生效

PD-01 至 PD-07 当前仍是决策草案，不是经律所负责人签署的生效政策。必须先确认权威档案库、导入范围、缺口登记、阻断例外、核查人及代理人、对外动作清单、敏感字段策略、豁免和留存规则。

### P0-2：冲突检索范围尚未达到 COMPLETE

本地数据库的权威检索范围尚未完成导入和对账，因此新检查保持 `COVERAGE_LIMITED`。这会阻断审批和受控动作，是预期的失败关闭行为。上线前必须用回填命令生成四类零缺口 `COMPLETED` run，并由冲突核查岗把 run 绑定到 scope；不能通过直接修改 scope 状态字段绕过。

### P0-3：主体库正向数据和案件关系不足

本轮浏览器可以验证“无结构化主体就不能自由文本绕过”，但没有完成“律师选择已登记新增主体 → 重检命中 → 独立复核 → 新版本生效”的正向闭环。正式验收必须用脱敏虚构数据预置至少：精确主体、别名/曾用名、关联方、自然人同名、被隔离历史案件和新增主体变更。

本轮已补齐 P0 主体版本、受保护身份标识、命中证据的结构化落库与检查结果双写，并提供显式历史索引回填/对账命令及 readiness 绑定门禁；P0 索引部署完整时，正式检查已切换为以该索引为唯一读源。**律所目标数据库尚未执行真实回填、版本对账和性能验证**，因此不能把“代码已支持回填/读索引”解释为“律所全量历史已建索引”。

### P0-4：生产主体数据密钥和密钥托管未完成

生产环境必须配置 `SUBJECT_DATA_KEY`，并由律所或部署方完成密钥托管、轮换、访问授权、备份和销毁策略。密钥不能写入仓库、种子文件、日志或浏览器。

### P0-5：系统外对外动作仍需最终接线验收

应用内的成案、状态推进、文档、OnlyOffice 和合同动作已经增加主体版本门禁；但正式发函、盖章、立案、出庭材料、外部法律意见系统等实际出站动作，必须逐项证明绑定有效主体版本。未完成集成前不能宣称系统堵住了所有影子流程。

### P0-6：三角色正向隔离验收未闭环

需要补齐律师 B 的自有案件和独立冲突任务，再执行 A 可见、B 可见、A/B 互不可见、核查岗可见完整证据的对照验收。当前只有 B 的越权拒绝证据，不能替代完整数据隔离验收。

### P0-7：历史自由 metadata 需要上线前盘点

新接案接口已经拒绝把身份证号、统一社会信用代码等敏感标识写入自由 metadata；但历史 `case_intakes.metadata`、旧冲突池和其他兼容表仍必须在目标库逐表盘点。应先运行只读 `go run ./cmd/backfill-sensitive-identities` 及业务数据盘点脚本，确认没有明文遗留或已登记迁移计划，再执行经授权的加密回填。没有盘点证据时，不能把“新写入已拦截”当成历史数据已安全。

另外，当前旧 `clients` 主表只具备身份证摘要回填路径；企业统一社会信用代码、历史代码及其有效期/来源必须在 `entities`、名称/身份历史和关系档案中完成结构化导入。仅有企业名称或公司字段的历史客户只能产生“名称候选待复核”，不能作为强标识一致，也不能支撑 `COMPLETE` 覆盖结论。

### P0-8：目标 PostgreSQL 仍未完成真实实例验收

当前生产入口已经明确为 PostgreSQL 16：Compose 的 `migrate` 服务执行事务化 schema bootstrap，生产后端跳过运行时 AutoMigrate，并在启动前检查关键表和 schema 版本。当前 bootstrap 版本已升级为 `postgres-mvp-2026-07-19-v4`，包含主体/证据基础表、历史索引对账运行表和追加保护。历史 `migrations/` 目录仍保留多个时期、多个方言的 SQL，只能用于历史兼容或专项迁移，不能作为生产入口。

先前专用 QA 实例已完成 bootstrap，并核验 waiver 证据表、追加写入/禁止删除触发器和 `RESTRICT` 外键；但当前可见的 `law_oa_batch01_e2e` 仅有基础表，不能替代目标生产空库。目标库仍未完成全量 bootstrap/readiness、档案导入后的覆盖检查、备份恢复演练和敏感数据盘点。必须在专用目标 PostgreSQL 上完成这些证据，再把 Compose 镜像称为可部署生产包。

### 工程验证状态：仓库级测试已全绿，外部集成仍需专用目标环境

`go build ./...`、生产代码范围 `go vet`、`go test ./...`、前端 `type-check/lint/build` 均通过。仓库中的 PostgreSQL/Elasticsearch 联调测试已改为显式 opt-in：只有设置 `LAW_OA_RUN_POSTGRES_ES_INTEGRATION=1` 并提供专用 `LAW_OA_POSTGRES_TEST_DSN` 才会执行，默认测试不会误连业务库或清空真实数据。因此当前“全绿”不等于目标生产 PostgreSQL/Elasticsearch 已完成联调，P0-8 的专用实例验收、备份恢复和数据覆盖证据仍然必须完成。

Codex Security 工作区已创建但仍处于 `setup.submitted=false`，尚未由用户启动扫描，因此本报告不把它计为安全扫描通过；仓库内本地模式检查未发现私钥、GitHub token 或 AWS access key，`.env` 未被 Git 跟踪。

### 2026-07-19 部署链路修复与剩余阻断

- Docker Compose 现在要求显式提供 `SUBJECT_DATA_KEY` 和 `CORS_ALLOWED_ORIGINS`，不再注入本地 CORS 默认值，也不再把宿主机 `internal` 源码目录挂进生产后端；`BUILD_DATE` 不再保留不会被 Compose 执行的 shell 替换文本。
- Kubernetes canonical manifests 已统一到后端实际唯一监听端口 `8080`，移除了 `scratch` 镜像无法执行的 shell `preStop`、缺失的 Fluent Bit sidecar/ConfigMap、错误的 gRPC/9090 端口和旧版并行 Deployment；上传卷改为显式 PVC，生产 Secret 改由 Secret Manager/External Secrets 提供。
- `.dockerignore` 已解除错误的 Git 忽略规则并纳入待提交文件，确保公开仓库构建时排除 `.env`、证书、Secret、数据和上传目录；生产镜像也不再复制 `.env.example`。
- 上述 Kubernetes YAML 已通过本地 Ruby 多文档 YAML 解析；本机没有可用 Kubernetes API，因此未宣称完成集群 schema dry-run。`kubectl` 的客户端 dry-run 因当前 context 指向 `localhost:8080` 且无集群而无法继续。
- 生产镜像构建尚未完成：Docker Desktop 拉取 `golang:1.25-alpine` 时出现基础镜像层大小校验/凭据网络错误，未进入项目编译。需要在有稳定镜像仓库访问的构建机重试并保留镜像摘要。
- 本轮收尾复跑 `GOCACHE=/private/tmp/go-build-cache go test ./...` 全部通过；使用空环境文件验证 Compose 在缺少 `SUBJECT_DATA_KEY` 时立即失败，确认新增生产必填项是 fail-closed。

## 七、放行条件

只有同时满足以下条件，才可以从受控试用升级为真实律所生产：

1. PD-01 至 PD-07 由律所负责人和合规责任人签署生效，并固化为部署配置。
2. 权威档案全部导入、对账并登记缺口；所有未覆盖范围在界面显著提示，不能自动放行。
3. 脱敏三角色种子数据完成 AT-01 至 AT-12，包含反向越权验证和主体变更正向闭环。
4. `SUBJECT_DATA_KEY` 完成生产密钥托管和恢复演练。
5. 对外动作清单逐项完成有效主体版本校验。
6. 在专用目标 PostgreSQL/Elasticsearch 测试库上显式执行联调，并保留认证、并发和数据隔离回归证据。
7. 律所指定的冲突核查岗实际签署最终验收记录。
8. 在目标 PostgreSQL 上完成 bootstrap、敏感数据盘点、健康检查和备份恢复演练，并保留输出证据。
9. 保持 `go test ./...`、`go vet`、前端 `type-check/lint/build` 作为发布门禁；外部联调不得以默认跳过替代目标环境验收。
