# M0最小依赖封板 v0.4（审核修订）

- 输入`codex/np-main-m0 @ c8716b7a964e4035b938761b8d4201b28b501cba`；应用main `d9c92fb`；源独立对象库`np-v04-first-slice @2d659d9`。
- 当前任务`t_5a4a0afa`，历史`t_7f6ad58e`保留。GLM完成后端增补初稿，主调度纠正引用并接手本清单/JSON整合，不称GLM已独立完成封板或复审。
- 本轮只是设计/依赖记录；任何新文件、路由、helper、脚本与检查均为后续计划，未实施、未运行；不进入M1。
- [结构化依赖账](../../../np-evidence/m0/dependency-seal.json)保留**33条原来源映射**，另列**9个已有原生目标、4个计划基础文件**及现有router/handler内读取hunk；这些集合可能引用同一文件，不机械相加为文件总数。旧8目标/2计划的blob复核文件保持历史不改。

## 一、工作包与交付顺序

| 包 | 具体依赖 | 验证时点 |
|---|---|---|
| M1 | 原生revision/锁内历史、真实list/detail/alias/dashboard恢复读取、主体保全最小helper、检测输入revision和关联事务、86受控追加及共享router schema保护 | 原生真实router/权限/主体回归、I14原生和I15-NATIVE先于M3；不要求NP来源联动提前可用 |
| M2 | CI局部隔离、PG/关键Skip失败、fuzz输入/汇总、产物白名单及消费链 | 早期配置/独立负向可先行；完整正常/负向I11待M3/M4 fixture后、M5前完成，只计一次 |
| M3 | JSON n2–5/7–16/18–26/28/33的NP、日志、B seed/spec/config/比较器与回执ADAPT；db/tx授权最小构造 | 源适用场景和新增负向同阶段必跑，B fixture不能延到M5才提供；I15-NP此后追加 |
| M4 | 仅内存、旧正文/每请求恢复键、编辑B/C防覆盖、响应缓存及PROD/受控旧worker | 消费M3稳定入口/夹具；组件和合成fixture先验，C10-PROD/OLD分开，不把dev当前端生产证据 |
| M5 | root main+PG/Redis、同候选PROD产物、外置TCP故障代理与浏览器/PG保全 | 真入口新增作者联调，消费前包产物，不再计fixture实现或完整I11 |
| M6 | 独立复审及必要复跑、I12合成追加升级/已应用核验/安全回滚/单实例 | 不回退未保护候选，不执行真实环境发布 |

## 二、33条原映射与新增依赖的计数

33条来源路径、slice_blob、main_blob保持固定对象身份；变更的是n18/19的交付从旧M5前移M3，以及夹具适配处置。它们不是新增两个来源项，不把原映射改称34条。

9个原生目标在旧8项基础上增加`internal/services/authorization_service.go`：源`NewAuthorizationServiceWithDB`在main不存在，M3复用main四repository按当前db/tx构造或最小工厂，撤权仍由服务端判定。不复制完整授权系统。

4个计划基础文件如下（均未创建）：

| 路径 | 固定来源/目的 | 归属 |
|---|---|---|
| internal/handlers/intake_draft_revision.go | 源demo_aggregate_handler中最小revision helper/恢复合同，不带完整continuation平台 | M1 |
| internal/database/narrow_pilot_schema.go | PG/两表/列/唯一索引/外键及异常失败关闭；调用放共享router装配点 | M1 |
| internal/handlers/intake_party_identity_guard.go | 源同名helper文件blob `5aab40a0ebee94cfe9faf7dbaeed1256f2bc1691`，仅必要主体保全符号 | M1 |
| scripts/np-first-slice/apply-main-np86.sh | 受控合成候选离线psql事务入口，固定86/SQL校验和/目标身份及schema，已应用只核验；不是通用迁移平台 | M1 |

## 三、原生读写、身份与I15最小符号闭包

**M0-R01 / P1-1 读取恢复**：main接案组只有POST/PUT等（router:378–384），工作台别名在:399。源router:385/:394的ListIntakeDrafts/IntakeWorkbench读取应按最小DTO/投影移植至真实原生list/detail/alias及dashboard读取，不能只在测试中注册。真实router须本人200、他人403、路由命中后的业务对象404；陈旧PUT409后同用户GET返回最新摘要/revision。客户/对方身份明文、密文和digest哨兵不得出现在全响应含嵌套metadata中，仅保留获准保护状态。GET不增加旧revision拒绝门槛。不新增continuation端点或can_submit/can_reuse_check平台。

**M0-R02 / P1-2 主体保全**：源原生handler:2450–2481在事务内先准备全部party rows，再delete/insert；main:1996–2013先删除再generic prepare。M1须区分切换客户取新client master与空对方身份继承旧受保护值；准备/插入失败时回滚全部写入。

必要源符号：lookupCaseIntakePartyRows、caseIntakePartyIdentityInputPresent、inheritCaseIntakePartyIdentity、prepareCaseIntakeClientPartyRow、clientEntityTypeMatchesMaster、caseIntakePartyRoleValue、caseIntakePartyIsClient、prepareCaseIntakePartyUpdateRow；不因为同一文件存在create包装就整组引入新平台。其依赖main已存在的Client.EffectiveIdentityType/IdentityDigestValue/IdentityCiphertextValue、prepareCaseIntakePartyRow/normalizeIntakeEntityType、intakeClientID/firstNonEmptyValue/stringValue与NewSubjectWorkflowError，JSON逐路径登记复用，不能再算模型/服务重建。

M3授权构造将NP调用绑定当前db/tx的四个现有repo，客户撤权后须拒绝；源factory只是参考，不能把存在现有AuthorizationService当作同名factory已经存在。

**M0-R03 / P1-3 I15分层**：main正式检测要求lawyer_facts_confirmed（handler:2188），源严格NP要求draft（narrow_pilot_context.go:113）。M1合法原生事实更新使确认快照失效、回draft并推进revision；旧同步和服务级异步完成必须因输入revision不符拒绝，修改后基准的intake状态/metadata/指针及关联cases原子不变。公共async intake入口被拒绝不证明在途完成保护。匹配版本正常完成只推进一次，检测启动不额外推进。输入revision走现有ConflictCheckRequest/search_parameters/association及conflict_check_records，不建队列/任务表；同步repository与handler二次整体写合并为同一受保护事务。NP来源联动在M3，真入口在M5。

## 四、86追加、装配入口与回滚（M0-R07 / P2-7）

- 固定main止于79+001，86号未占；旧M0全新v14合成库执行源86成功，只证明两表/两个唯一索引/接案外键结构兼容，不是正式升级/回滚验收。
- 旧`cmd/migrate`production只允许`-command bootstrap`，不能假设执行它会追加86，不放宽历史混合方言up/down限制。
- M1拟用上表受控离线入口：确认合成目标PG/数据库身份、v14结构和固定源SQL SHA-256；排他事务内仅首次追加86，部分存在/结构不符/权限错误立即失败；已应用只核对实际列、外键和唯一索引定义，不无条件重复DDL，不依赖AutoMigrate补表。M1定点检查与M6实际演练分开计量。
- root main是本轮唯一支持/构建发布启动目标。`cmd/server/main.go:82`仍调用router.Init，NP schema检查必须放共享装配点，在路由可用前覆盖误启动；仅文档写“不支持”或只检查root不足以防旁路。备用entry不列为产品支持目标，也不增加独立服务实现。
- 新候选必须86就绪，即使助手关闭/白名单空也不解除历史保护。原从未启用NP的未升级旧版本可保留旧合同，但不报NP通过；已升级缺表/查询失败按未知拒绝写入。
- 回滚只至保留revision/来源保护的候选，保留来源、回执、墓碑；不执行down删表，不退回未保护main。升级/重复状态/失败后恢复/重启/单实例及旧客户端切换均在M6合成I12演练，尚NOT_RUN；真实环境仍未授权。

## 五、回执、B seed和SW交叉规则

M0-R04 / P1-4：成功回执fixture用捕获的request_id及服务端id（与恢复intake_id比较）；committed/no_change有合法revision，abandoned不要求保存revision，action返回才比；缺身份改负向保键保输入，按v2逐请求键断言。M0-R05 / P1-5：B spec/config前移M3/M4，完整I11此后由M2执行。

M0-R06 / P2-6：M3 B seed必须>=2独立case_materials、非空客户/主体/目标及对照精确IDs、大整数未知metadata；四场景来源+3/回执+4在干净确定窗口，新负向独立seed不污染。M0-R08 / P2-8：M4同候选PROD构建及受控旧worker，M5同产物真root链，不凭缺sw.js推定无旧worker；不扩建PWA。

上述用例与JSON见[路线B清单](route-b-test-split.md)，缓存规则见[缓存策略](../cache-policy-proposal.md)。源历史26+33+4来源计数不改，ADAPT/新增组另列，全部集成执行NOT_RUN；I09-FILES仍DEFERRED/HOLD。

## 六、复审、工作量与边界

[M0净工时账](net-effort-estimate.md)把读取/I15/迁移基础、授权挂接及缓存的原已计部分与主体净新增分开，不按9目标/4计划文件条数加价。源码编号/基线前进必须重核；保留[旧blob复核](../../../np-evidence/m0/seal-blob-verification.json)历史，新静态校验在[本轮复审证据](../../../np-evidence/m0/m0-review-revision.json)。

本轮仅完成文档规则，不关闭运行缺陷、不替代独立集成验收或给M1实施授权。外部旧全量调用方、实际升级回滚单实例、真实main+Redis/TLS及律师试用仍待后续验证；不承诺外部旧客户端长期透明兼容。
