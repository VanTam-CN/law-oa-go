# 接案协作 MVP 迭代记录

- 用户范围：2026-11-07 明确继续接案协作 MVP，使用 superagent-iterate，按已有授权通过 PR 跟踪、审查和交付。
- 基线：`a806831a7cef839c2b971cd4ab260cce77f42b8b`，方案指定版本，未使用脏主树或旧 PR 45 作为基线。
- 实施目录：`/private/tmp/lawoa-intake-collaboration-20261107`，分支 `codex/intake-collaboration-mvp`。
- 持久交付目录：`/Users/mac/.codex/worktrees/intake-collaboration-mvp/law-oa-go`（目前仅基线；实施完成后本地同步）。
- 原目录 `/Users/mac/Desktop/FT/law-oa-go` 的未提交代码完整保留，未批量移植。
- 授权边界：本地实现、迁移文件、隔离虚构数据验证。未执行真实迁移、生产部署、GitHub 推送或合并；远端交付前核对已有具体授权。
- 最近 GitHub 核对：PR 45 是旧 pilot closure，head `4060edb40208c81790f31b9ef572036c71abb9e1`，不是本轮基线。

| 任务 | 责任范围 | 状态 | 证据或下一步 |
|---|---|---|---|
| N0 基线与契约 | 主代理、前端/待办只读核对 | 已核对，待固定完整 DTO | baseline/go.txt；OpenSpec strict 通过 |
| N1/N3 补正快照、受控重提 | 主代理，审批模型/业务/handler；迁移 000082 | 实施中 | 统一快照读者、版本校验、真实事务 |
| N2 客户后端 | Kuhn；client handler/service/repository 及新文件；迁移 000083 | 作者独立切片通过，待整体复核 | client/CLIENT_SLICE_REPORT.md；真实PG12路并发、65同名客户分页、回滚 |
| N4 待办 | Euclid；inbox 模型/仓储/服务/handler/辅助；迁移 000084 | 实施中 | 授权后计数/分页，事务交接函数 |
| N2/N5 前端 | Averroes；frontend 唯一写入者 | 实施中 | 共享客户表单、草稿/补正、动作投影 |
| N6 集成验证与独立审查 | 主代理、Kuhn事务测试、Linnaeus独立核心审查 | 进行中 | C01–C18逐项记录；baseline/integration-regression-2.txt保留旧fixture失败 |

## 已有证据

- 基线真实运行：`go test ./internal/services ./internal/handlers -run 'TestResubmitAtomic_|TestCaseTxGuard_|Test.*DraftRevision|TestListIntakeDrafts' -count=1 -v` PASS。仅证明日志中列出的既有测试，不代表整个 MVP。
- `openspec validate add-intake-collaboration-mvp --strict` PASS。
- 独立 PostgreSQL 16 临时集群已启动，Unix socket 位于 `/private/tmp/lawoa-intake-collaboration-runtime/socket`，端口 55437，未监听 TCP。只用于虚构数据；C12–C14 尚未运行。
- 前端复用相同 lockfile 的本机 node_modules，只读依赖，不安装或更改原目录依赖。

## 恢复要点

- 先核对上述目录实际差异和代理状态，不重新创建同名任务或覆盖用户修改。
- 新模型主代理负责 `ActiveSubmissionSnapshotID *string`；Inbox helper 由 N4 提供，主代理串行接入业务调用点。
- 所有提交/推送前重新核对最终差异、原目录保护和授权；本记录不构成新增授权。
- 工程验证与独立真人律师试用、生产验收分开：`production_ready=false`；真人试用 NOT_RUN。

## 15:16 检查点（Asia/Shanghai）

- Curie 快照helper交付并关闭，报告 snapshot/RESULT.md、contract.md；13个顶层/26个子场景作者自测（真实PG含），非整体验收。
- 首轮后端集成编译通过，针对性回归通过；扩大至相关审批/接案/待办回归后4项旧handler测试因极简fixture缺少真实申请人/完整事实复核被新门禁拒绝。Kuhn负责补齐fixture，保持原断言及并发顺序；没有绕过门禁或删除测试。
- 主代理已接审批事务、首提原子化、指派/复核/冲突提交待办hook、权威快照所有主要读者、回复专用受控保存、工作台同源待办、历史snapshot选择、生产bootstrap加法结构。
- initial lock order校正为approval -> intake -> tracking；PG时间精度通过tx重读后生成待办版本避免纳秒/微秒重复key。
- 原记录保护通过000082数据库触发器，bootstrap正同步相同保护；生产部署未执行。
- 临时PG主库lawoa_intake_mvp_test已由bootstrap初始化，自带qa-fixture虚构样本已写入并验证。
- 本地真实router测试服务127.0.0.1:18587，执行session20553；Vite13003 session16079（代理18587）。运行环境私有文件在runtime/env.json，不提交。
- 浏览器用已安装Playwright MCP（CLI无离线缓存，不安装依赖）：A虚构律师已登录；取消新增客户保留原立案输入、localStorage无测试身份原文实测。后续客户创建遭开发HMR重载，未作为业务失败/成功；完整C16等待固定前端构建再跑。
- Playwright默认少量自动日志写入原目录.playwright-mcp，原目录业务源码未改；最终截图显式保存本轮output/playwright。
- 独立核心审查代理Linnaeus：01a1b4da-5ab4-7d42-bc05-9554d539c65e。
- 仍在工作：Averroes前端、Euclid待办、Kuhn真实PG事务测试+旧handler fixture升级；不得重复派相同写入任务。

## 15:48 检查点（Asia/Shanghai）

- 独立核心review发现6项，本轮已修：真实HTTP claim状态/token、通用Submit唯一槽、legacy关联删除绕过、所有成功响应投影、重放同源、旧历史读取。Linnaeus随后负责handler修复并交接，已关闭；新独立后端审查Boyle进行中。
- 真PG事务专项原25叶子PASS；旧4个handler并发守卫抓到旧请求编号6位时钟后缀碰撞。主线程改用完整UUID编号，Kuhn正在最终复测；保留09/21首次红灯。最新增加成案/重放同事务待办交接。
- 冻结party成案+completed replay均取批准snapshot，批准后check被改不会改变案件主体；有真实PG专项证据。
- bootstrap触发器曾因一次prepared statement多命令失败，拆分后bootstrap-4/-5连续成功。另空白测试库lawoa_intake_migration_20261107完成84→83→82 down、82→83→84 up两轮及bootstrap重复成功，schema/results.json全exit0。
- 当前本地后端127.0.0.1:18587为qa-browser-next，session25718，PID60647；旧PID37286已仅停止本轮测试进程。日志runtime/qa-server-2.log。后续源码仍在修，此二进制尚非最后候选。
- 固定Vite preview127.0.0.1:13004，session81324；仍保留dev13003不用。Playwright只允许在原项目根写截图，故截图在原目录output/playwright/intake-collaboration-20261107，交付时复制到本轮证据。
- 浏览器A同对象：客户ID3 QA 协作客户 A 20261107；intake ec5b905a-c86b-477e-98ef-09e6de41f179、代码INT-20261107153302-EFEA10；check CCT_78878038-531f-4029-b0ab-aa0ea2f569cd。新客户原地建档成功、原字段保留；正式检测0匹配后officer独立指定自己并提交no_conflict成功，DB确认status conflict_ready、事实版本存在。
- A还走了旧“发起冲突审批”支路，旧approval d08a805d-6994-40cd-ac78-488485da80ab。当前旧snapshot被前端误判历史只读，独立review确认F1，本轮回归待修；它不是case_creation主审批。
- 浏览器阻断：1200×1041长冲突详情选择复核人被sidebar覆盖（截图02）；1440宽无遮挡可继续；退出重登A恢复同一草稿后页面显示未检测，而后台有效review存在。两项均交Averroes修复。此前部分ref失效是局部snapshot不含旧ref，不能仅据此断言HMR；固定preview已无HMR。
- Planck独立前端review3P1+1P2，report frontend-review.md，Averroes处理中：legacy当前/历史判定、旧轮次token、迟到恢复跨账号/覆盖、空白入口对象残留。前端不能标C16/C17通过。当前浏览器停在A已保存草稿，无未保存编辑。
- Euclid待办交接完成19作者SQLite tests PASS；已委托只写N0 contract-map，不再改源码。main修overdue统一授权service；frontend结果分类/blocked view-only入口也交Averroes。
- review-fixes-2.txt相关CoreReview/Inbox/ResubmitAtomic/CaseTxGuard通过；第一次仅constructor调用错误，已保留修复。
- GitHub实查main d9c92fbd583a0cbbf7f9b4c071e1f2329981f4d7，与a806831分叉，merge-base70d38b6；MVP同名PR搜索无结果。不得直接向main发包含历史回退的混合PR。将使用精确a806831基线分支承载范围清晰的Draft PR，明确基础修复尚未并入main；不自动合并或生产发布。尚未远端写入。

## 16:50 检查点（Asia/Shanghai）

- 前端 R2 `index-Cm1X_6UB.js` 修复 F1–F4 与草稿恢复 camelCase；R3 `index-p6KdlFyE.js` 补移动 Drawer 层级和冲突列表回填，R2 75项mock通过，R3局部15项通过。Planck独立关闭F1–F4，但发现F5合法非接案无快照动作被禁、F6审批摘要复核/覆盖文案错误，已交Averroes最小修复并再冻结。
- B1/B3及R1作者Kuhn交接 `backend-final-fixes/HANDOFF.md`：262叶子/93实际PG、0fail/0skip，race/vet/build通过；原17结果13fail19skip与fixture错误全部保留，后37新结果为通过。新增R1解决换客户尚未重提时详情/快照/回复授权。Euclid正在汇合相同中间状态的Inbox投影，未以B2既有通过替代R1。
- Boyle独立复核B1/B2/B3原项，发现上述R1；修复后仍需最终重新绑定。没有向GitHub发送review评论或合并。
- 真实浏览器客户69：双击仅1请求，服务器创建后中断响应；同key改名409，改回原内容200同ID；原立案字段保留；SQL确认客户1/改请求0/取消0。`browser/client-response-loss.json`、client-db-counts.json。客户65条fixture查询单词“分页客户”，第二页15项，键盘选ID14，证据client-paging-mobile.json。
- 390×844客户Modal Tab/Shift+Tab/Escape可达、按钮在视口内；移动Drawer close最初被header遮挡，R3真实重测已修，证据mobile-drawer-retest.md。长补正仍未测。1200×1041新检测选择复核人实际点击已成功。
- 当前服务为 runtime/qa-browser-r3，session36373，PID记录qa-server-3.pid，日志qa-server-3.log；source/binary manifest `browser/candidate-r3.json`，构建16:30前后hash稳定，但随后R1改源，因此不是最终验收候选。Preview13004仍运行；dev13003未使用。
- 浏览器R3-A：客户70，intake ab040a71-c248-4529-8db2-f402d279c00b，编号INT-20261107163800-961E3C，check CCT_fa2e102a-218e-410e-80c0-7b770ead40e5，review2/no_conflict。UI录入5条角色（含共同委托、多对方、同名异身份、同主体多角色），保存/确认/检测及独立review均成功。重登精确恢复后首次提交500：bootstrap缺case_intakes.approval_claim_token，SQL42703。SQL确认intake conflict_ready、审批0、案件0、parties5；未手改样本修绿。Kuhn接手000080/81 bootstrap parity及真实新库首提测试。
- 隔离库原无流程，为虚构账号4 director→5 compliance建立CASE_APPROVAL双节点fixture，详见browser/fixture-workflow.json。没有修改正式审批规则或覆盖已有流程。
- 浏览器tabs：0律师甲，R3-A保存现场，URL含intake_id；1律师甲本地未保存“QA移动导航修复验证”（仅测试），390px；2隔离context intake-officer-20261107，officer已独立登录并复核R3-A。tab0令牌曾过期，16:46实际退出重新登录；auth_token在sessionStorage，getToken过期会不发header，不能把401误认为服务或数据库损坏。
- 操作细节：保存并退出返回/case（已改手册），继续草稿按钮名称包含“ · 继续编辑”。AntD选择项限定`.ant-select-dropdown:visible [title=...]`，避免隐藏旧菜单严格匹配冲突。相关方Select点div[aria-label]，内部readonly input被选中文字挡是自动化定位问题；不用强制点击。复核说明用getByPlaceholder。保存按钮loading图标动画离场后aria仍带loading但width0/opacity0、按钮实际未禁用，不是保存卡住。
- GitHub身份已用提升权限只读验证有效（默认限制下gh会误报tokeninvalid）；已fetch远端main。main与a806831：左3/右53个独有提交，merge-base70d38b6。计划精确基线anchor+feature Draft PR，暂未任何push/PR创建。Durableworktree仍pristine baseline。
- 最后需：关闭bootstrap/R1/F5F6；新候选hash/build；两个全新客户开始的完整主链C16、C04双窗口、C17长补正及C18手册；证据/矩阵/任务清单、显式纳入ignore的测试；提交、本地同步durable、授权push精确base+feature并Draft PR；不部署/merge，真人试用NOT_RUN。

## 17:42 检查点（Asia/Shanghai）

- 已提交源码/测试候选 `7c5618f26ed5600de76e4ed94a682252d10a8aef`，98文件，21个被忽略的新Go测试显式纳入。持久worktree已快进同SHA。原脏主目录业务未改。
- 已按用户现有授权推送精确base a806831与feature，Draft PR https://github.com/VanTam-CN/law-oa-go/pull/46 ，目标base是codex/intake-collaboration-mvp-base，不是main。GitHub当前无status checks，不能称CI通过；无merge/deploy/@review。
- Kuhn bootstrap/HANDOFF.md：v16补000080claim/000081index、真实初始化/已有v15升级/首提HTTP，300叶子PASS，111realPG，0fail/skip。最终source601文件见25-final-source-after.json。F5服务端能力DTO确认普通业务可false，持久intake/历史snapshot/成案true，读取错误不降级。
- 前端R4 index-fDMF1BW8.js冻结；32作者与29独立DOM专项通过。Planck关闭F1–F6无剩余前端代码阻断。Boyle独立40PASS/20PG关闭R1/B4/F5；这些不替代浏览器。
- 新运行服务runtime/qa-browser-final，session13455，PID qa-server-final.pid，日志qa-server-final.log；正常migrate-final bootstrap升级成功，未手改schema或业务样本。binarySHA66f3f6113bd4d05d17554ea0b639d0041d4312d6cec2558deceb70c702a0eb73；browser/candidate-final.json。
- 浏览器最终A（失败样本，不算C16）：client71，intake eba99230-1fa1-48ee-b9a0-b0b532879462，code INT-20261107171510-67D9A3，approval18e5b8f062cb0200，初snapshot e903f608-1687-4ad6-8a11-1bcf25f16790，第二snapshot c3696e36-b505-4246-bbcc-78975add9939。初check CCT_ecb20634-76c9-497f-8328-30c3b85e9be6，重检CCT_d39c57dd-6262-4f34-a98a-5bc2e24c460e。
- A已真实完成原地客户/5角色、保存退出重登、律师乙不串草稿、两窗口旧revision409且输入保留、正式检查独立复核、首提200、director退回2项、home补充入口、390x844两项回复保存刷新稳定ID、keyboard定位正确焦点、改身份后重提409 INTAKE_FACTS_STALE、新检测独立复核、原审批第二轮重提200。详情见browser/final-a-events.json，截图08–13已复制。
- B5首红17:35:23：director同意第一阶段，integration decision409但已推进compliance。ProcessApprovalWithConflict约1168未检查updatedApproval最终approved，错误尝试成案并新增failed tracking。SQL case0，approval under_review next5，tracking pending+failed两行。browser/final-a-stage-first-red.md/db.json/server-excerpt.log保留，不继续该样本最终审批、不手改DB。
- Kuhn已接B5最小后端修复+真实PG两级HTTP测试（新ignored approval_multistage_postgres_http_test.go），Boyle待专项复查。Euclid只整理C01–C15 final-matrix-draft，C16–C18主线程。Averroes/Planck源码已冻结，勿重复派写入。
- PR46正文已追加B5阻断。修复后需新commit/build/只停止本轮qa-server-final进程启动新binary，两套全新客户完整主链；现A部分证据仅支持未变功能，不拼成完整链。前端bundle可沿用相同哈希。
- Playwright tabs：0律师甲/A审批详情；1旧未保存移动探针(不用)；2officer隔离context intake-officer-20261107，17:21重新登录，/conflict第二review已完成；3director隔离context intake-director-final-20261107，A审批确认框保留409；4compliance隔离context intake-compliance-final-20261107，A审批详情；5同律师甲独立context intake-lawyer-duplicate-final，A原草稿旧版。勿在旧样本上假修绿。
- UI教训：登录后立即快速导航可能遇延迟dashboard导航，登录/下一步分两次工具调用等待稳定。AntD选择改点div[aria-label]避免readonly input遮挡。批准按钮实际可访问名有icon，用button.filter({hasText:'同意并通过本阶段'})；弹窗确认button exact为确认同意并通过本阶段。回复保存是PUT /case-intakes/:id，不是supplement URL。定位使用setTimeout0，必须waitForFunction activeElement.id后核验（intake-description/intake-opponent-identity-number）。刷新有修改说明未保存时会beforeunload，处理真实对话后再继续，不把工具等待当产品失败。

## Final delivery checkpoint (2026-11-07, Asia/Shanghai)

- Runtime source frozen at 3f970040f5e40b8016c18a070b5aa1345d2b2eec; final additions are tests, documentation and evidence. All 601 baseline manifest entries match commit 7c5618f; all 52 R4 frontend entries match final files.
- Fresh R5 A/B each completed actual browser new-client-to-case chains (case3/case4); frozen parties match SQL multisets, unique cases/slots, zero open workflow tasks. Final C01 field retention, C05 officer dashboard and C08 record/item pairing gaps closed.
- B5: final author multistage 6 PG PASS (includes original4); independent original4 PG PASS and additional two-test code/log review. C06: author2 and independent2 PG PASS. Product source unchanged. No counts combined with earlier300/111 or independent40/20.
- C01-C18 engineering matrix,115 subconditions,14-chapter manual,6 screenshots and original failure artifacts delivered. OpenSpec strict exit0. Runtime credential value and JWT pattern scan found zero matches, including archives.
- Draft PR46 targets exact a806831 baseline branch; main prerequisites remain unresolved, CI no checks, no merge/deploy. Real human trial, production and disaster recovery NOT_RUN; production_ready=false.
- Final report, matrix and SHA256SUMS are the delivery entry points; older checkpoints intentionally retain historical states. GitHub reviewer messages were not sent.
