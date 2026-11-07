# I0 附录：接案字段、写入口、持久化与证据清单

日期：2026-11-07。对应 [I0 只读核验报告](2026-11-07-law-firm-assistant-i0-readonly-baseline.md)。  
“现有”表示本次源码观察；“建议”表示实施前合同建议，不表示代码已支持或业务方已批准。实际数据库、运行配置和测试均未核验。

## 1. 页面、接口和写入边界

### 1.1 页面与接口

| 对象 | 当前实现 | 本次建议范围 |
|---|---|---|
| 原菜单 | /case 案件中心，以 case_rows 为主；有创建入口和正式案件详情 | 在当前中心增加经授权的接案草稿入口，不建设全局工作台 |
| 接案页 | /case/create；读取 case_id/client_id/mode | 增加 intake_id 恢复分支；case_id 补录仍是不同语义，不能混用 |
| 真实接案读取 | GET /api/v1/cases/intake-workbench/:id | 复用其权限基础；补 revision/来源索引、子项完整性和恢复契约 |
| 原生创建/更新 | POST /api/v1/case-intakes；PUT /api/v1/case-intakes/:id | 保持入口，分别使用创建/更新 DTO，统一应用服务 |
| 事实确认/检测 | 原生接案动作和原页面 | 继续由人操作；接案回写参与版本规则；模型无调用工具 |
| 来源/回执/上下文 | 未发现 NP v0.2 对应实现 | 按 v0.2 §11.3 新增有界接口；不新增助手业务写入口 |

草稿列表需服务端按当前接案/客户权限过滤后分页和计数，不能先返回全所结果再由浏览器隐藏。来源、回执、上下文、敏感接案详情明确绕过共用响应缓存；列表缓存也不能绕过当前授权。

### 1.2 五类接案写入

清点边界为实际业务源码中的 case_intakes、case_intake_parties、case_materials 写入及相关调用者，不含迁移建表/测试夹具；数据库触发器、外部脚本和真实部署任务仍未知。

| 写入家族 | 文件/入口 | 现有行为 | 必需控制 |
|---|---|---|---|
| WRT-1 新建 | demo_aggregate_handler.go:1753 CreateCaseIntake | 主记录及存在的子表事务写入，创建幂等键可选 | schema 必备；真实创建回执；客户/角色授权；同事务来源与幂等 |
| WRT-2 修改 | 同文件:1919 UpdateCaseIntake | 先读后写；正文出现子数组则删后重建；按部分字段存在性失效利冲 | expected_revision；真实变化计算；稳定子项 ID；当前客户授权；来源失效/继承 |
| WRT-3 事实确认 | 同文件:2349 ConfirmIntakeFacts | 改 created_by/status/metadata | 版本前置、事务内角色/对象/客户/状态复核；不改两字段时保留当前来源 |
| WRT-4 检测完成回写 | 同文件:2398 StartIntakeConflictCheck，尤其 2671/2688 | 先调用 WRT-5，再用早先 metadata 二次写回 | 合并为一个受控接案结果写入；使用检测开始时的输入版本 |
| WRT-5 关联结果回写 | conflict_repository.go:148 LinkConflictCheckToCase，接案分支 249/276 | 校验客户和已有 check_id，更新 status/metadata | association 携带接案输入 revision；条件更新；事务冲突时不得覆盖新状态/metadata |

异步调用：async_conflict_check_service.go:147/198 的后台检测也调用 WRT-5，因此需要传递输入版本并在旧结果被拒时返回可理解的任务结果；无需重写检测算法或引入新消息队列。

注意：
- demo_aggregate_handler.go:2865 在审批创建中查询接案以补充审批关联信息，属于读取。
- integration_handler.go 的接案审批校验也是读取和状态消费，需回归，但不应误记为第六个接案写入口。
- 全局 Inbox/WorkItem 投影不是上述业务写入口；本批不修全局办理人推断。
- 版本规则应覆盖同一接案的全部这五类写入，而不是仅覆盖白名单律师的助手请求。

## 2. 顶层字段映射

现有 Create/Update 允许集来自 demo_aggregate_helpers.go:200/233。身份、所有权、状态及正式流程字段不允许通过新 DTO 越权。下列规则仅针对接案 DTO；不自动变更客户/正式案件 DTO。

| 字段 | 当前形态/差异 | v0.2 建议写入与清空规则 | 模型能力 |
|---|---|---|---|
| client_id | 可空客户 ID；律师创建仅在实际传入时授权；页面要求选现有客户 | 白名单律师新建必选已授权现有客户，不可清空。更新总查当前客户；无来源历史时改绑还需旧/新客户与动作授权，有来源历史普通改绑拒绝 | 无 |
| title | 字符串；PG 模型 255 字符、not null 不等于非空业务校验 | 可显式更新；建议草稿非空、长度与 PG/前端一致，不可隐式空串删除；是否允许纯手工空标题由业务确认 | 候选，必须人工应用 |
| description | 文本；目前无来源关联 | 可显式更新，可通过 clear_fields 清空为统一的空表示；清空同时使当前来源失效，不删除历史 | 候选，必须人工应用 |
| case_type | 字符串；正式事实确认会检查存在 | 人工控件更新，枚举沿用实际案件类型契约；是否允许草稿清空待确认，未确认前拒绝清空，不强迫补齐正式检测信息 | 无 |
| priority | 字符串，默认 medium | 人工控件更新；合法枚举与页面一致；不可清空，不以 null 恢复默认 | 无 |
| metadata | JSON；目前整体过滤后替换/特定状态合并 | 仅 metadata_changes 的路径 set/remove；未触达旧键保留；明确区分业务可编辑、服务端派生、流程专属 | 无 |
| parties/materials | 当前更新可能整组替换 | 只接受局部 create/update/delete；详见 §4 | 无 |
| id/intake_code/created_by/status/timestamps | 服务端生成/原生流程控制 | 原生 DTO 不接受自由赋值；动作按真实角色和状态改变，接案编号需抗同秒碰撞 | 无 |
| revision/当前来源索引 | 当前模型缺失 | 只由服务端计算；expected_revision 是条件，不是可赋值字段 | 无 |
| source_selection/context_version/请求键 | 拟增协议字段 | 严格独立于业务字段；服务端解释，不得转存到自由 metadata | 无 |

当前律师页面只在 createIntake 入口统一检查客户 ID/名称及负责律师；当前服务端并没有完全对应的草稿必填检查。不能将“页面上有必填”当成服务端已经保证。人员助理的 assistant_draft 原本不允许录入 client_id/parties，这个角色语义应保留，不能被“AI 助手”一词混淆。

### 2.1 权限层次

- AI：仅真实 lawyer 角色、白名单、有效会话、当前对象/客户授权的交集；管理角色有业务阅读权不意味着自动拥有试点模型使用权。
- 普通原生保存：沿用当前人员助理、律师、业务管理、独立核查等已有角色规则；不能把所有原生用户都强制变成试点律师。
- 人员助理：保留现有允许字段和 assistant_draft 限制，不得新开放客户、主体或身份。
- 来源读取：按当前对象及字段可读性；若原生管理角色应读来源，需明确记录到角色矩阵，不能默认开放全量片段。
- 服务端从认证与当前授权事实取值，不采信前端提交的 role、actor、verified 标记或客户名称。

当前角色帮助函数列出的人员助理为 assistant/intake_assistant；业务管理为 director/partner/compliance/risk/risk_control/management；独立核查另含 conflict_officer。admin/super_admin 不因此自动获得专业业务权限。具体原生管理/独立核查角色能编辑哪些字段，需在冻结字段表时逐项与现有规则对齐；这不是本次授权调整。

## 3. metadata 路径建议表

本页实际产生的键见 Batch01Prototype.tsx:3910。表内“可移除”必须经 remove 操作，不接受 null/空串作为隐式删除；未触达键原样保留。

| 路径/键 | 建议归属、类型和动作 | 清空/保留规则 |
|---|---|---|
| business_area、sub_area | 人工业务分类字符串；复用分类联动校验 | 可移除，但需处理父子分类一致性；不自动补默认覆盖旧值 |
| dispute_amount | 人工金额输入，统一规范金额表示 | 可移除；未知与 0 区别；实际历史类型/精度需 schema/样本结构核验，不从模型推断 |
| source_channel、source_contact | 人工录入来源信息字符串 | 可移除；正文禁止记日志 |
| investment_agreement_date、dispute_date、breach_date、proposed_filing_date | 人工日期，校验允许格式/精度 | 可移除；不把缺日期补成当天，不因模板缺项阻断合法草稿 |
| jurisdiction | 人工管辖信息字符串 | 可移除；不由模型给出法律确认结论 |
| billing_method | 人工收费方式枚举 | 移除/切换与费用字段一致性需明确；不擅自删除旧费用 |
| fee_base、contingency_rate、minimum_fee | 人工金额/比例，需规范数值与范围 | 可移除；不能将 0 当缺项；类型/上限沿用正式字段约束，未形成约束时先定版 |
| lawyer_id | 人工选定的真实律师 ID | 律师草稿保存建议必选有效人员，不可清空；任意用户 ID/失效人员拒绝；具体指派权待角色表定版 |
| source | 页面/服务来源标记 | 服务端或受控客户端协议派生，不开放任意值来伪造来源 |
| subject_case_id、subject_case_number | 现有 case_id 补录语义 | 由已授权正式案件派生；窄版新建/已有普通 draft 不任意改挂正式案件 |
| party_relations | 子项关系派生 | 从通过校验的关系操作构造，与真实子项 ID 映射同事务维护 |
| lawyer_facts_confirmed_*、conflict_*、assistant_draft、created_by_role | 正式流程/角色派生 | 普通更新不可 set/remove；只能由相应服务端动作改变 |
| 其他历史未知键 | 不在本轮写入白名单 | 未触达保留；前端不得全量回传改写，也不能自动清理 |

不建立任意 JSON 路径编辑器。所有路径必须枚举；同路径、祖孙路径冲突、未知路径、声明无动作均拒绝。当前的“禁止明文身份证字段”检查仍保留，但不能替代路径白名单。

关于利冲失效：现代码用“正文是否带 client_id/title/case_type/metadata/parties”判断；完整 payload 因而容易使旧状态失效，description 则没有进入该集合。改造后按实际变化和原生业务依赖判定。描述、材料/收费类 metadata 哪些变化影响正式结果，需要明确规则，不允许把整个 metadata 容器一概当事实变化，也不能漏掉真正影响检测输入的字段。

## 4. 主体、材料、身份和关系

### 4.1 主体增量合同

| 元素 | 当前事实 | 建议局部规则 |
|---|---|---|
| id | PG 子项有真实主键；前端多为本地 localId | update/delete 必须携带属于当前接案的真实 ID；localId 仅用于本次新增和关系映射 |
| entity_name/entity_type/party_role | 现有校验与规范化 | 可按原生权限显式修改；必填不可清空；枚举沿用现有契约 |
| business_role、local_id、schema_version | 现有代码将关系/角色信息写入 metadata | 通过专用契约派生，拒绝自由 metadata 绕开 |
| relation_depth、relations | 现有深度、节点及关系类型校验 | 显式关系操作；删除节点不能遗留悬空引用，需明确删除相关关系或拒绝 |
| identity_type/身份输入 | 现有 helper 生成密文/摘要，部分主体要求真实身份 | 未改身份时保留原密文/摘要/引用；掩码不能当新原文；更改身份走原生验证，禁止模型写入 |
| entity_id/身份引用 | 后端保存/关联基础 | 不允许客户端以任意引用替换；所有引用须同对象/授权核验 |
| case_id/intake_id/created_at | 旧允许集包含这些键 | 新 DTO 不接受自由跨对象指定；由当前对象/动作派生，旧值保持合法归属 |

新增继续复用 intakePartyContract、intakePartyRowsWithIdentityGuard 等规则；局部修改需补“既有身份不变可保留”的分支，不能要求浏览器再次提交原身份证号才能保存摘要。对子项删除和关系联动的具体业务规则列为 G00 未闭环项，不默认“删节点就全删关系”。

### 4.2 材料增量合同

- 当前页面仅显示 defaultMaterials，文件管理按钮禁用（Batch01Prototype.tsx:5342）；没有已开放的材料增量编辑界面。本批必须做好后端协议与未改材料保全，不能把新增材料编辑/文件管理界面混入既有功能回归或 W07。
- name、material_type、status、required：后端原生增量 DTO 的显式动作字段，使用原生业务枚举/布尔类型；名称不可空，required=false 不是删除。其正向协议可在服务测试验证；将来开放控件需单独裁定。
- storage_url、metadata：保持既有值；本批不新增文件上传/解析。已有引用的人工修改权限和允许地址来源需要在字段表中明确，不能把任意外链变成可执行链接。
- case_id/intake_id、ID、时间：服务端控制；update/delete 必须归属于当前接案。
- 材料未编辑时不发送 material_changes，不能每次保存都重新生成 defaultMaterials。
- 子项列表分片读取要返回覆盖情况/分页，不能把 GET 的最多 20 条当完整集合。
- 集合删除、空集合、未提供集合三种状态必须区分；只有逐个明确 delete 才能删除记录。

## 5. 来源与回执的物理结构建议

目标是三类业务责任，不建立通用聊天/运行平台。名称均为拟名。

| 结构 | 最小字段/约束建议 | 不存什么 |
|---|---|---|
| case_intakes 扩展 | revision bigint；标题/摘要各自的 text_revision、coverage、snapshot 引用/字段键、binding_revision。可以用明确命名列实现，不放自由 metadata | 不存原始聊天或模型思考 |
| case_intake_source_snapshots | id；intake_id；saved_revision；schema/template 版本；操作者/时间；当时标题/摘要；必要原话、定位、更正及选择；覆盖模式；unique(intake_id,saved_revision) | 不存整段聊天、未采用候选、模型原始响应和凭据 |
| case_intake_write_requests | actor_id+请求键唯一；action/target；带密钥规范摘要及摘要版本；结果 intake_id/revision/snapshot_id、result_kind、提交时间 | 不存完整原请求、完整回显案情、裸 JWT 或模型正文 |

建议使用两个目标字段的明确列而非任意字段历史表，便于范围约束。快照字段的内容可以是版本化 JSONB，但它是有固定 schema 的受保护业务数据，不是任意 metadata。快照通过外键或等价事务校验保证归属；当前引用必须指向同一 intake 的合法快照字段。

旧行初始化：revision 可从初始版本开始；两字段来源均 unlinked，历史为空；不能伪造录入人原话。仅改其他字段沿用现有关联；文字变化和明确来源动作遵循 v0.2 的版本转移。没有任何来源历史的纯手工草稿不必强造快照；已有来源历史即使当前指针为空，也不能允许普通客户改绑。

现有 case_intakes.idempotency_key 是全局唯一、可空的单列；它不能表达 update、多动作、同键异参和 no_change。新增回执表应独立于客户创建请求表；旧列保留/退役策略必须明确，避免新作用域和旧全局唯一索引冲突。原客户幂等的 HMAC 逻辑只能复用方法，须新建接案摘要域和动作/目标规范化。

### 5.1 事务与锁顺序待定稿

拟顺序：请求键占位 → 授权保护所需客户记录 → 接案 → 按 ID 排序的受影响子记录 → 当前来源关联/快照与结果。更新先读客户 ID 再取锁时，取得接案锁后再次核对客户及 revision；不一致则中止，不能临时逆序追加旧客户锁。

必须与现有授权/隔离墙更新路径对照：仅锁客户行并不能自动证明隔离墙策略不会并发变化；若授权依赖其他关系表，需要明确事务内重验和撤权生效点。该点本次没有完成全量锁路径证明，因此不宣称锁顺序已定版。

同键 PG 冲突处理可借鉴 client_service 的事务内占位+savepoint，但来源冻结和数据库占位之间必须避免长时间持有上下文锁。数据库事务内有最终权限、状态和 revision 检查；成功/无变化回执同事务提交。已回滚不得留成功回执；查询“无记录”不得伪称确定未提交。

## 6. 单实例、预算与日志

- 进程内上下文要求同一试点链的所有请求到同一进程。仅开启粘性会话不能证明故障/滚动升级时仍满足；不做跨实例恢复，实例切换应明确失效。
- 临时上下文 TTL 不得超过绑定 access token 有效期；登出/撤权后的访问仍重新验证。新 token 不搬运旧会话正文。
- 接案页当前使用本地 apiRequest/fetch，不能把其他模块 axios 的 10 秒默认超时误认为本页现有超时。
- 当前前端 Dockerfile 中 Nginx 代理读超时 30 秒。若采用设计的 45 秒总模型预算，需统一为本链路可容纳的预算/代理配置，或降低总预算；不是全站无条件加长超时。
- 实际访问日志会记录路径、状态等，URL/query 不能放案情、引用原文或 token。错误码和结果键不能夹带正文。
- GORM 参数禁记、panic/异常安全包装、模型提供方错误清洗、缓存排除都应从真实挂载路径落实。
- 已安装 GORM v1.30.0 的 callbacks.go:133 和 logger/logger.go:193 显示 SQL 展开及参数过滤链；不能只屏蔽 SQL 参数而继续原样记录可能含数据的驱动错误详情。Gin v1.10.1 的 logger.go:274 会拼入 query，recovery.go:72 会处理请求头和 panic 信息，均需纳入隔离日志探针。这些依赖读取未替代运行证据。
- 业务来源表允许存必要原话；这是受控业务留存，不等于允许诊断日志记录正文。不要用“日志不能有正文”误删业务来源。

## 7. 现有测试资产及不足

以下只读过测试/夹具，均未运行。

| 资产 | 可复用内容 | 不能代替 |
|---|---|---|
| demo_aggregate_handler_security_test.go | 创建同键、部分权限/客户检查 | 同键异参、稀疏 PUT 当前客户撤权、持久化回执、来源事务 |
| intake_write_role_guard_test.go 及 identity/co_client/party_contract 系列 | 角色、身份、共同客户和关系契约 | 稳定真实 ID 的局部 CRUD 和未改字段保全 |
| conflict_repository_test.go、异步检测测试 | 显式接案/客户关联与现有检测边界 | 检测输入版本过期后禁止写回 |
| client_idempotency_test.go | 请求规范化、摘要、同/异参及重复键处理模式 | 接案多动作/来源冻结/结果回查 |
| CaseIntakeConflictAction/PartyContract/ClientRemoteSearch/CaseTypeContract | 页面必填、身份不存储、本地 ID/角色、类型/远程客户 | 页面编辑序号、服务端真实恢复、来源及结果待定 |
| frontend/e2e/cases.spec.ts 等 | 保存/检测/审批操作可回归 | 当前部分使用固定 Mock 响应和 localStorage，不能证明隔离 PG 全链 |

新增测试应围绕实际失效风险：旧事实检测回写、事务故障、跨会话/撤权、超过 20 个子项、身份掩码保留、PG 键冲突、保存顺序竞态及 SQL 日志哨兵。不得通过禁用原有子项编辑、审批操作或改弱断言换取“窄版通过”。

## 8. 证据索引

行号对应本次观察工作树；后续若变化以源文件指纹和符号名重新定位。

| 编号 | 源文件与定位 | 支持的观察 |
|---|---|---|
| E01 | internal/handlers/demo_aggregate_handler.go:1753、1771、1833、1856、1910、1919、1984、2010、2043 | 创建/重放/缺表路径、更新权限与无 revision 条件 |
| E02 | 同文件:290、316；internal/router/router.go:383、404 | 接案读取权限、20 条子项限制、真实原生入口 |
| E03 | demo_aggregate_helpers.go:200、209、220、233；demo_aggregate_handler.go:2047、2061、2100、2173、2204 | 字段允许集、整组重建、身份与关系契约 |
| E04 | demo_aggregate_handler.go:2349、2398、2671、2688、2865；internal/repositories/conflict_repository.go:148、249、276；internal/services/async_conflict_check_service.go:147、198 | 五类写入与异步调用；审批片段是读取 |
| E05 | frontend/src/pages/batch01/Batch01Prototype.tsx:832、2377、3574、3670、3781、3873、4005、4090、4342、4395、5494；frontend/src/App.tsx:232 | 页面路由、fetch 错误、全量 payload、保存及本地恢复 |
| E06 | frontend/src/pages/batch01/intake/conflictCheck.ts:171、175、231、254；Batch01Prototype.tsx:3622、3665 | 草稿 key、冲突指纹、不完整本地持久化 |
| E07 | internal/database/production_schema.go:22、39、58、326；migrations/000042_case_intake_idempotency.up.sql；000070_case_intake_party_identity.up.sql；000078_auth_token_sessions.up.sql；000082_intake_collaboration_mvp.up.sql | 当前表模型、幂等/身份/会话已有结构；未见 NP revision/来源结构 |
| E08 | internal/router/router.go:106、112、279；internal/middleware/jwt.go:75、138、181；internal/auth/token_revocation.go:119、225、320；internal/auth/token_manager.go:197、355、383；internal/models/auth_token_session.go:7 | 实际认证挂载、类型/会话兼容分支、刷新 UUID 和撤销范围 |
| E09 | main.go:173、187；internal/database/connection_pool.go:59、141；internal/database/optimized_components.go:213；internal/middleware/audit.go:24、67；cache.go:199、280；common.go:257 | 实际日志/ORM/缓存入口，AuditLogger 不能代表实际挂载 |
| E10 | internal/services/mcp_client.go:61、94、124、262、318；frontend/Dockerfile:161；internal/config/config.go:119；docker-compose.yml:40、218；k8s/deployments/backend.yaml:25；helm/law-oa-go/values.yaml:34 及 templates/deployment.yaml:17 | 现有客户端并非生成模型适配器、模板超时/数据库/后端实例差异 |
| E11 | internal/services/client_service.go:171、199、206、216、248；internal/services/client_idempotency.go:103、122 | 可借鉴的 HMAC/幂等与 PG savepoint 模式 |
| E12 | internal/database/migrator.go:29、110；internal/services/authorization_service.go:50、82、94；internal/handlers/intake_write_role_guard.go:14 | 迁移工具会连接数据库；角色基础需保留 |

“未发现实现”的搜索范围为 internal、frontend/src、migrations 的 NP 专用对象/协议符号及实际路由，不等于证明整个仓库没有任何名为 revision 的无关机制。源码默认配置、部署模板和历史记录不作为当前运行环境的事实。

## 9. G00 尚需决策的最小清单

1. 确认实际 PG/schema 和后端单实例，包括升级期间的实例行为。
2. 在独立候选中纳入必要未跟踪文件并生成可复现构建清单，不能只给 HEAD。
3. 草稿标题、类型、负责律师的必填/可清空规则，以及原生各角色动作权限。
4. 主体删除与关系联动、材料引用的可编辑/可删除规则。
5. metadata 类型/范围及哪些变化触发正式检测失效。
6. 接案/客户/授权关系的事务复核和锁顺序，及撤权生效点。
7. 回执键作用域与旧 idempotency_key 索引的兼容方案；明确摘要版本和失败/未知结果处理。
8. 实际日志/代理链及敏感读取缓存排除可验证。
9. 保存来源的读取角色、保留/删除和备份政策；真实资料与虚构资料阶段分别裁定。

这些条目未闭环前 G00 不填 PASS。本次报告已经完成源码差距调查，不能将“G00 未通过”误读为本轮什么都没做；也不能将本轮完成误读为可开始部署。

