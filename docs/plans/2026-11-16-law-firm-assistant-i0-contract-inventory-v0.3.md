# I0 附录：接案字段、写入口、持久化与证据清单

版本：I0-NP v0.3｜日期：2026-11-16。对应 [I0 v0.3 报告](2026-11-16-law-firm-assistant-i0-readonly-baseline-v0.3.md)及 [NP v0.3 设计](2026-11-16-law-firm-assistant-narrow-pilot-design-v0.3.md)。
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
| 来源/回执/上下文 | 未发现 NP v0.3 对应实现 | 按 v0.3 §11.3 新增有界接口；不新增助手业务写入口 |

草稿列表需服务端按当前接案/客户权限过滤后分页和计数，不能先返回全所结果再由浏览器隐藏。来源、回执、上下文、敏感接案详情明确绕过共用响应缓存；列表缓存也不能绕过当前授权。

### 1.2 五类接案写入

清点边界为实际业务源码中的 case_intakes、case_intake_parties、case_materials 写入及相关调用者，不含迁移建表/测试夹具；数据库触发器、外部脚本和真实部署任务仍未知。

| 写入家族 | 文件/入口 | 现有行为 | 必需控制 |
|---|---|---|---|
| WRT-1 新建 | demo_aggregate_handler.go:1753 CreateCaseIntake | 主记录及存在的子表事务写入，创建幂等键可选 | schema 必备；真实创建回执；客户/角色授权；同事务来源与幂等 |
| WRT-2 修改 | 同文件:1919 UpdateCaseIntake | 先读后写；正文出现子数组则删后重建；按部分字段存在性失效利冲 | expected_revision；真实变化计算；稳定子项 ID；当前客户授权；来源失效/继承 |
| WRT-3 事实确认 | 同文件:2349 ConfirmIntakeFacts | 改 created_by/status/metadata | 版本前置、事务内角色/对象/客户/状态复核；不改两字段时保留当前来源 |
| WRT-4 检测完成回写 | 同文件:2398 StartIntakeConflictCheck，尤其 2671/2686 | 启动不写；完成先 WRT-5，再旧 metadata 二次写 | 输入事实与 revision 一致取数；启动不额外改版；完成合并一次受控 CAS，旧版拒绝 |
| WRT-5 关联结果回写 | conflict_repository.go:148 LinkConflictCheckToCase，cases 约243，接案约249/276 | 同事务读/更新 cases 与 case_intakes；异步也经此函数 | 携带最初输入 revision；过期整次关联不生效，含 cases；不写 conflict_ready，不覆盖来源/metadata；无接案的正式案件路径必要回归 |

异步调用：async_conflict_check_service.go:147/198 的后台检测也调用 WRT-5，因此需要传递输入版本并在旧结果被拒时返回可理解的任务结果；无需重写检测算法或引入新消息队列。

注意：
- demo_aggregate_handler.go:2865 在审批创建中查询接案以补充审批关联信息，属于读取。
- integration_handler.go 的接案审批校验也是读取和状态消费，需回归，但不应误记为第六个接案写入口。
- 全局 Inbox/WorkItem 投影不是上述业务写入口；本批不修全局办理人推断。
- 版本规则应覆盖同一接案的全部这五类写入，而不是仅覆盖白名单律师的助手请求。

## 2. 顶层字段映射

现有 Create/Update 允许集来自 demo_aggregate_helpers.go:200/233。身份、所有权、状态及正式流程字段不允许通过新 DTO 越权。下列规则仅针对接案 DTO；不自动变更客户/正式案件 DTO。

| 字段 | 当前形态/差异 | v0.3 建议写入与清空规则 | 模型能力 |
|---|---|---|---|
| client_id | 可空客户 ID；律师创建仅在实际传入时授权；页面要求选现有客户 | 白名单律师新建必选已授权现有客户，不可清空。更新总查当前客户；无来源历史时改绑还需旧/新客户与动作授权，有来源历史普通改绑拒绝 | 无 |
| title | 字符串；PG 模型 255 字符、not null 不等于非空业务校验 | 可显式更新；建议草稿非空、长度与 PG/前端一致，不可隐式空串删除；是否允许纯手工空标题由业务确认 | 候选，必须人工应用 |
| description | 文本；目前无来源关联 | 可显式更新，可通过 clear_fields 清空为统一的空表示；清空同时使当前来源失效，不删除历史 | 候选，必须人工应用 |
| case_type | 字符串；正式事实确认会检查存在 | 人工控件更新，枚举沿用实际案件类型契约；是否允许草稿清空待确认，未确认前拒绝清空，不强迫补齐正式检测信息 | 无 |
| priority | 字符串，默认 medium | 人工控件更新；合法枚举与页面一致；不可清空，不以 null 恢复默认 | 无 |
| metadata | JSON；目前整体过滤后替换/特定状态合并 | 仅 metadata_changes 的路径 set/remove；未触达旧键保留；明确区分业务可编辑、服务端派生、流程专属 | 无 |
| parties | 当前更新可能整组替换 | 仅 party_changes 显式 create/update/delete，保留 ID/身份/关系；详见 §4.1 | 无 |
| materials | 页面固定 defaultMaterials；现后端可收数组 | POST 原生合法初始化；PUT 只要出现 materials/material_changes 或声明材料变化即整笔拒绝，保全既有记录；详见 §4.2 | 无 |
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

UpdateCaseIntake:1989 对 role==lawyer 且 assistant_draft 放开所有权，约2018 改 created_by；ConfirmIntakeFacts:2365 同类例外。全仓未见助理草稿指派给特定律师的字段/流程。代码路径成立，是否共享待接手池待业务决定；AI 本批不接手，原生仍须按定版政策回归，不擅自新增指派或放大接手权。

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

### 4.2 材料编辑后置与保全合同

- 当前页面为 defaultMaterials、文件管理禁用；本轮已查 qa-fixture 和 e2e 未发现用户修改材料数组的调用。该范围结论不覆盖未知外部客户端/脚本，G00 需确认。
- 新建 POST 沿原生角色/字段验证初始化合法材料；律师助手不生成材料数组。人员助理原生创建不强加律师试点条件。
- 更新 PUT 不接受 materials、material_changes，含空/null；changed_fields/clear_fields 含材料也拒绝 MATERIAL_EDIT_NOT_SUPPORTED。材料意图识别先于通用歧义校验；整笔请求不得部分保存、不得成功忽略，也不写成功/no_change 回执。
- 不实现材料 create/update/delete DTO，不新增上传、文件管理或解析界面。若已开放外部材料编辑调用存在，先重审后置与兼容范围，不能留下旧的无保护重建旁路。
- 仅改摘要时，既有材料 ID、name/type/status/required、storage_url、metadata/归属/时间保持；读取 >20 条应有完整分页/覆盖标识，更新不从截断列表推断删除。
- 主体的显式增量编辑/删除仍需正向验证，不能把材料后置扩大成停用已有主体功能。

## 5. 来源与回执的物理结构建议

目标是三类业务责任，不建立通用聊天/运行平台。名称均为拟名。

| 结构 | 最小字段/约束建议 | 不存什么 |
|---|---|---|
| case_intakes 扩展 | revision bigint；标题/摘要各自的 text_revision、coverage、snapshot 引用/字段键、binding_revision。可以用明确命名列实现，不放自由 metadata | 不存原始聊天或模型思考 |
| case_intake_source_snapshots | id；intake_id；saved_revision；schema/template 版本；操作者/时间；当时标题/摘要；必要原话、定位、更正及选择；覆盖模式；unique(intake_id,saved_revision) | 不存整段聊天、未采用候选、模型原始响应和凭据 |
| case_intake_write_requests | unique(actor_id,request_key)，不再按 action 分割；action/target/client 范围；摘要及版本；result_kind=committed/no_change/abandoned；结果 ID/revision/snapshot、终态时间。abandoned 无业务结果，可无原摘要 | 不存完整正文/案情回显/token；不伪造未到达请求的摘要 |

建议使用两个目标字段的明确列而非任意字段历史表，便于范围约束。快照字段的内容可以是版本化 JSONB，但它是有固定 schema 的受保护业务数据，不是任意 metadata。快照通过外键或等价事务校验保证归属；当前引用必须指向同一 intake 的合法快照字段。

旧字段再次纳入整理时，source_ref 保留 origin_intake_id/origin_field/origin_revision/source_kind/origin_coverage；由服务端计算，标明人工未关联等出处强度，不递归建链。新候选引用旧草稿不升级为原始证据。旧字段原录入人/时间没有可靠记录时标未知，不能以 created_by/updated_at 代填；记录本次纳入操作者与时间。

旧行初始化：revision 可从初始版本开始；两字段来源均 unlinked，历史为空；不能伪造录入人原话。仅改其他字段沿用现有关联；文字变化和明确来源动作遵循 v0.3 的版本转移。没有任何来源历史的纯手工草稿不必强造快照；已有来源历史即使当前指针为空，也不能允许普通客户改绑。

现有 case_intakes.idempotency_key 是全局唯一、可空的单列；它不能表达 update、多动作、同键异参和 no_change。新增回执表应独立于客户创建请求表；旧列保留/退役策略必须明确，避免新作用域和旧全局唯一索引冲突。原客户幂等的 HMAC 逻辑只能复用方法，须新建接案摘要域和动作/目标规范化。

### 5.1 共同原键锁、终态和业务事务

保存与明确 abandon 共用 actor_id+key 的 PG 事务 advisory 锁（稳定派生值，拟 try-lock 最长 3 秒）和唯一约束；未有回执行也能互斥。使用主库 READ COMMITTED，锁后读取终态，不能拿锁前旧快照宣告无结果。锁未取到返回忙/unknown，不放行新键。

原生保存：认证/DTO/摘要 → 共同原键锁 → 已有终态重放或拒绝 → 内存锁外完成授权查询 → 短内存锁比较可信绑定/期限/版本并接受/复制来源包 → 释放内存锁 → 当前客户/接案/子项业务锁及最终认证权限/状态/revision → 草稿/来源/回执同事务提交。无模型调用入事务，无内存锁跨 DB I/O；只提交终态，不新增持久化 pending/租约恢复。

普通 GET write-results 绝不写占位或墓碑。POST abandon 必须由用户明确动作，带原 action、intake_id 或新建 client_id 等最小范围并重新授权；不需要临时上下文/正文。锁内有 committed/no_change 则返回原真实结果，有 abandoned 返回原终态，无记录才插入 abandoned 并提交。提交确认前不许换键；错误/丢响应继续查原键。迟到原请求锁内遇墓碑直接拒绝，无业务写。

成功/无变化/已放弃回执不跟随临时 TTL 清理；试点全保留。未来清理需有永久拒绝旧键策略。no_change 无业务 revision/快照增加，但有持久化结果；来源重绑仍为真实变更。已回滚无成功回执；新意图前原键须闭合，不凭“没查到”推断回滚。

业务锁顺序仍需对照现有授权和关联调用定版：共同原键 → 必要客户保护 → 接案 → 子项 → 来源/回执；含 cases 的同次关联一并定顺序。先读客户 ID 后取得接案锁须复核其仍相同；变化即中止，不能逆序加锁。隔离墙关系不是锁客户行就自动锁住，需明确最终授权点及撤权线性化边界。

同步/异步检测均以一致取数时的最初 revision 完成 CAS；旧版不改接案，关联事务的 cases 更新也不得部分生效。启动不产生额外接案版本。源码静态不证明实际并发结果，A27/A29 的隔离 PG 证据仍全部 NOT_RUN。

## 6. 单实例、预算与日志

- 进程内上下文要求同一试点链的所有请求到同一进程。仅开启粘性会话不能证明故障/滚动升级时仍满足；不做跨实例恢复，实例切换应明确失效。
- 复用已登记 access UUID/会话行；expires_at 取 access/会话、创建+2h、最近接受动作+30m 的最小值。默认 3600 秒是签发时长，非本次可用一小时。只读/发布不续期，剩余不足45秒不接模型轮；来源接受后 TTL 不毁包，最终业务授权前 access 到期拒绝。新 token 不搬运临时正文，回执可按新登录当前权查询。
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

新增测试围绕：真实断线/迟到原请求与墓碑两种先后顺序、只读查询、TTL 后旧键、原始期限、任一新输入旧候选失效、同步/异步关联 CAS、主体正向编辑与材料拒绝/保全、事务/身份/日志。材料编辑后置是显式范围裁定，不得据此停用已开放主体编辑/审批或削弱原断言。

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
| E13 | internal/config/config.go:136；internal/auth/token_manager.go:418～431；frontend/src/hooks/useAuthService.ts:197 及调用搜索 | access 默认3600秒、新 UUID/会话、前端未见刷新调用；不证明实际运行配置 |
| E14 | demo_aggregate_handler.go:1774/2150；main.go:501；nginx/nginx.conf:31 | 创建/更新 DB 未绑定请求 context，HTTP 只有 ReadHeaderTimeout，单上游无 proxy_next_upstream；只证明断线后可能提交，不是已复现 |
| E15 | cmd/qa-fixture/main.go；Batch01Prototype.tsx:695/3873/3880/3994/4183/5342；frontend/e2e/cases.spec.ts:144～145、utils/test-helpers.ts:427/432、ux-remaining-probes.spec.ts:139～148、conflict-actions.spec.ts:534起 | 夹具/前端/e2e 未发现用户编辑材料数组；外部调用未知，G00 确认 |
| E16 | conflict_repository.go:148～285；demo_aggregate_handler.go:2671/2686；async_conflict_check_service.go:198 | 关联事务先写 cases 再写接案；同步二次写、异步间接写均须条件版本；不能只查调用者文件 |

“未发现实现”的搜索范围为 internal、frontend/src、migrations 的 NP 专用对象/协议符号及实际路由，不等于证明整个仓库没有任何名为 revision 的无关机制。源码默认配置、部署模板和历史记录不作为当前运行环境的事实。

## 9. G00 核心与真实资料条件分开

核心未闭环项（虚构技术阶段也需要）：

1. 实际 PG/schema/约束、主库一致读、单实例及新旧协议切换窗口。
2. 独立源码候选含必要已跟踪未提交修改和未跟踪文件正文，附全部清单/指纹；HEAD 和散列不够。
3. 草稿必填/可清空、原生各角色动作、助理共享接手池政策、技术来源可读角色。
4. 主体删除/关系联动、稳定身份引用；外部材料编辑调用方核实，确认可后置材料 CRUD。
5. metadata 白名单/类型和实际事实变化对检测失效的规则。
6. 客户/接案/授权关系与含 cases 关联的锁顺序、撤权最终判断点；共同原键锁派生/有界等待。
7. 回执 actor+key 唯一作用域、旧全局索引兼容、无摘要墓碑、明确放弃/未知/TTL 保留。
8. 实际 access/会话期限与取消/超时；日志/代理/缓存可验证；虚构数据清理及测试访问。

真实资料条件另列：来源保留/删除/备份/退出政策、业务可读人员与外发授权、提供方训练/保留配置。当前 PENDING/HOLD，不阻塞已满足核心条件的纯虚构验证，也不替代技术访问角色定版。

上述核心未闭环前 G00 不填 PASS。本轮补读和规则定稿不授权实施或部署，全部运行门禁仍为 NOT_RUN。

## 10. 2026-11-16 补读文件指纹

以下为本轮观察的补充索引，不扩称为可复现源码候选；原 858 项清单不改写。

| 路径 | SHA-256 |
|---|---|
| cmd/qa-fixture/main.go | 28341612ab603088485c2522f7fd5c13892b2a488c4c4626285fea5eb2f4dca7 |
| nginx/nginx.conf | 408ecc4a270957626522d994ad7ae63a87090cd51c35959b77c4af6e0336eb45 |
| main.go | da42d9d127963a608b975d7fb5adeb3352707285bab185bde5e90bd7e54868af |
| frontend/src/hooks/useAuthService.ts | 2bf05e147541c6327bc4f826e5eab6ec8440b849ecfb70b8733d6e230148c820 |
| frontend/e2e/cases.spec.ts | 14d3908624e4f6350f4809d39a5e0cf3cd4b59b0d9a9b5ae8691f5c3e7d66e51 |
| frontend/e2e/utils/test-helpers.ts | 02d94c9136780c5714645ede9161e6474ea1cefb49265e955479484923a7fb0e |
| frontend/e2e/ux-remaining-probes.spec.ts | a0710ff843c37cf309410759c9f630a502611e7df2e1a16355f9c41c146777d2 |
| frontend/e2e/conflict-actions.spec.ts | 949e96ad2c660a8abcf9aa2329108cca6a2ad7ce84f7358edf93aa1b8c2c9fc1 |
| internal/repositories/conflict_repository.go | 8d43cbce6892eb60e6af68851a996179ebfda0987e08b9fa8ff6f4ce1acedfe3 |
| internal/services/async_conflict_check_service.go | 098752f70be90bebc77f515e983a749803d9ac164c151d4e4ca427d89ff7668d |
