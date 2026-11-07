# 接案协作 MVP 独立前端审查

## R4 专项独立复核 / 候选 7c5618f（2026-11-07 17:13 CST）

**当前结论：F5、F6 关闭（独立源码 + 固定构建 DOM/mock 证据层）；多阶段措辞与决定反馈通过专项复核。本次未发现新的前端阻断。R3 的 F1–F4 关闭结论保留，完整 MVP/真实浏览器两链仍待主线程验证。** 下方 R3/R1 的“未关闭”描述为原始失败记录，不再代表当前状态。

本次范围仅为 F5/F6、多阶段反馈及其直接契约边界；仅写本报告。未改产品代码、操作浏览器、执行真实 API/数据库写入、安装依赖、派子代理、提交或推送；没有重新开展全域审查或后端权限审计。

### 最终版本绑定

- 候选及当前 HEAD：`7c5618f26ed5600de76e4ed94a682252d10a8aef`，提交题目 `feat(intake): preserve drafts and coordinate revision approvals with workflow inbox`。
- 前端 [R4 交接](../frontend/approval-review-fix/README.md)，冻结时间 `2026-11-07T09:05:06.532Z`。固定 JS `index-fDMF1BW8.js`，SHA256 `1231ccd77e41527faa7491e8b8a04b212bff03647c0e734d5cef8ee21067c6df`。
- 前端源码归档 SHA256 `0caadc64bb0d059c45bb55106369d723d62029137968d557f1128216c69a8f2e`；构建归档 SHA256 `700670964c653995638de2586e69459dc9aa7029f3134fb18718281653879d5f`。独立重算一致，归档内 28 个源码文件、7 个 dist 文件逐项与清单匹配。
- 17:11 CST 对候选逐项核对：28 个前端源码、14 个关联测试、5 个构建输入同时匹配 R4 清单、工作目录及 `git show 7c5618f:<path>` 字节；7 个 dist 文件匹配 R4 清单。**本次提交没有改变已复核的 R4 前端。** dist/报告没有因此被当作已提交源码。
- 后端最终 [bootstrap/25-final-source-after.json](../bootstrap/25-final-source-after.json) SHA256 `b484b1091555275a70d1fae36b760e05cce6211e38c8fba07e4eacc5658a0708`。17:12 CST 独立核对其 601 个路径：工作目录内容 SHA256 全匹配，Git blob 全匹配候选树。该动作仅绑定文件版本，不是 601 文件代码审计或后端测试执行。
- 本次直接读取的后端契约均匹配候选：`approval_handler.go` SHA256 `15eab93c86d34ca71d1f4ec9d7844e49cc08ba99ecb4580cb8cf07ab18e6615e`；`approval_snapshot_mode.go` SHA256 `c762942156e3076549dd407c295f53e690a46c084f13c45b79459931a0820dc5`。
- 相对 R3 的四个前端变动文件为主 TSX、审批页面测试、approval DTO、approvalPresentation；独立归档 diff 未见草稿/客户流程扩展。主 TSX SHA256 `804bad90d90c6dc93c550438bd3dbb36acd70d9c2fd3182992267e9ff0c83569`，approvalPresentation SHA256 `59436a2167871ff965153fe60f75f08631fa40d0ba948ac1471d163a651f3e1b`。

### F5：普通无快照审批与受控旧业务边界 — 关闭

**实现位置**：[approvalPresentation.ts:13](../../../../../frontend/src/services/approvalPresentation.ts#L13)–18；[Batch01Prototype.tsx:8391](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L8391)、8476–8477、8527。

服务端字段必须严格为 `submission_snapshot_required === false`；前端同时排除 active 指针、顶层/metadata intake、成案类型或配置等可见矛盾，且只接受 `ApiRequestError.status === 404`。仅此兼容模式省略快照令牌，仍要求并原样发送 `expected_action_version`。加载新详情先清除兼容模式；读取到有效快照仍采用正常持久版本路径。

**真实代码契约核对**：[approval_handler.go:249](../../../../../internal/handlers/approval_handler.go#L249) 在既有对象授权后调用模式解析，271 以非 omitempty 的 bool 输出字段。[approval_snapshot_mode.go:14](../../../../../internal/services/approval_snapshot_mode.go#L14)–36 先解析持久 intake 关联，再处理 active、case_creation、内嵌/持久快照；解析/查询失败返回错误。响应投影保留该 bool。前端并非把“没看到 intake_id”当普通审批，也没有把业务模式字段当作决定权限。现有真实处理端点及后端事务门禁未改。

**独立 R4 固定 bundle 探针结果**：

| 触发 | 实际行为 |
| --- | --- |
| 普通 leave，明确 boolean false，snapshot 404，合法当前处理人与版本 | 3 个处理动作可见；只发 1 次原集成决定 POST；没有 expected_submission_snapshot_id；精确保留 `2026-11-07T02:03:04.123456789Z`；刷新得到 approved 后提示“审批已通过”；确认框和 decision_comments 不承诺成案。 |
| 能力字段缺失 / true / null / 字符串 false | 全部显示快照读取错误，无处理动作、无写请求。 |
| false 与 active / 顶层 intake / metadata intake / case_creation / conflict_approval / case_creation_config / integration_type=both 矛盾 | 全部阻断，无无快照降级、无写请求。 |
| false + snapshot 403 / 409 / 500 / 传输失败 / 200 但缺版本字段 | 全部阻断，无写请求；成功状态码本身不替代可核验快照。 |
| false + 合法 404，但 action_version 为空 | 提交确认提示“审批版本尚未加载，请刷新详情后处理”，零写请求。 |
| 可读取当前旧快照、旧 intake 决定返回 409 INTAKE_SUBMISSION_REQUIRED | 只发一次原端点请求，仍携带持久 viewed UUID 和原始 action_version；错误显示、确认框保留；未改走另一接口、清除令牌、重建或重试审批。 |

以上最后一行保持原受控接案边界；不把“前端读到历史事实”解释为“该旧接案单可以批准”。没有访问或决定现场旧单 d08a805d。

### F6：冻结检索、独立复核与覆盖状态 — 关闭

**实现位置**：[Batch01Prototype.tsx:8702](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L8702)–8725、8935–8941、8954–8956。

R4 将冻结自动检索结果、独立复核来源及覆盖状态分开展示；零匹配仅显示“未发现匹配证据”。当前页面优先使用当前可读复核，读取被拒绝时明确展示本次提交冻结复核；显式历史只读，读取历史自己的冻结复核。

独立固定 bundle 四种场景均通过，无写请求或执行异常：

1. `REVIEW_REQUIRED + coverage=COMPLETE + live review=no_conflict` 同时显示“冻结自动检索结果：待人工复核”“当前可读独立复核：无冲突”“检索覆盖：完整”；同意本阶段可用；原“总体风险等级：待人工复核”和零命中推断“检索覆盖受限”均不再出现。
2. live review=no_conflict 但 coverage=PARTIAL，显示“检索覆盖：未完整（PARTIAL）”，同意仍禁用。
3. live review 403、当前提交冻结 review=no_conflict，显示“本次提交冻结独立复核：无冲突”，没有伪称已读取最新复核。
4. 从当前切历史，显示“该版本冻结独立复核：信息不足”、历史 PARTIAL，无当前 live review 文案、无决定按钮；切回后恢复当前无冲突复核与允许的处理动作。

### 多阶段措辞与状态反馈 — 通过

**实现位置**：[Batch01Prototype.tsx:8485](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L8485)–8497、8537–8558、9187。按钮/弹窗/提交 comments 均限定当前阶段；提示最终成案的前提包含全部审批通过及成案条件满足。刷新详情优先决定反馈，刷新失败才用原决定响应。只读核对 [integration_handler.go:864](../../../../../internal/handlers/integration_handler.go#L864)：成功 data 是 updatedApproval，探针的刷新失败样本也按此真实 envelope 返回对象。

| 服务端返回状态（全量 mock） | 固定构建观察到的反馈 |
| --- | --- |
| 决定成功，刷新为 under_review / compliance_review / 合规员 | “本阶段已通过，下一节点：合规复核（合规员）”；没有“审批已通过”或“已成案”；确认框关闭。 |
| 最终 approved，集成状态有 case_id=991 | “最终审批已通过，已成案：CASE-REVIEW-991”。 |
| 最终 approved，仍无 case_id、成案 pending | “最终审批已通过，成案状态请以关联案件为准”；没有“已成案”。 |
| 最终 approved，成案 failed、无 case_id | “最终审批已通过，成案失败，请查看成案状态”；没有“已成案”。 |
| 刷新详情 500，决定响应明确为下一阶段 under_review | 保留读取错误，同时只提示本阶段通过/下一节点；没有最终通过/已成案承诺。 |
| HTTP 200 决定响应没有可确认状态，刷新详情 500 | “决定已记录，审批状态尚未确认，请刷新核对”，不假定阶段或整体通过。 |

所有决定探针均只发一次请求，持久快照 UUID 和纳秒 action_version 不变；成功后关闭确认框，旧接案 409 保留错误/确认框。没有改变审批级数、决定端点或用客户端状态替代后端最终状态。

### 证据层级与最终剩余阻断

- **本次独立执行**：从 R4 tar 在内存加载固定 bundle，完整 App 路由、JSDOM、虚构身份、全量 mock fetch/XHR，共 29 个专项场景（F5 18、F6 4、多阶段/旧单拒绝 7），三个脚本均正常结束，场景均无 JSDOM 执行异常。未新建测试文件或运行真实浏览器。这些是独立行为探针，不能记作真实 HTTP 或浏览器 PASS。
- **作者证据**：已读取 R4 regression.log 的 2 文件 32/32 结果，type-check/lint/build/diff 日志与清单；本审查未重跑该 32 项，也未把旧全套 75 项计作 R4 重跑。主线程通知后端最终 300 pass；本审查只绑定最终源码清单，不冒充后端测试执行者或独立后端审查者。
- **前端专项剩余代码阻断：无。** F5/F6 原问题可关闭，多阶段直接相关回归未见新问题。R3 已关闭的 F1–F4 没有重新扩大为本次全量验收。
- **完整交付剩余验收项**：运行服务需实证输出当前 candidate 的 submission_snapshot_required；bootstrap 初始 schema 与后端候选的实际运行状态、两条新 case_creation 完整链、真实双节点中间/最终决定、真实旧受控接案拒绝、最终 C16/整体 C17，仍由主线程按最新证据完成。当前消息表明浏览器主链即将启动，故不写已通过。先前用户确认通过的 A 恢复与移动场景继续保留对应版本证据，不被本报告回退为未测。

本结论支持主线程继续候选 `7c5618f` 的真实链验证；不构成部署、生产或真实律师试用验收。

## R3 独立复核结论（2026-11-07 16:38 CST）

**F1–F4 原始阻断行为已通过源码与独立 DOM 复核，可在该证据层关闭；仍有两项 P2 未关闭：F5 非接案无快照审批被误阻断、F6 审批详情摘要错误表达覆盖/复核状态。不能据此宣告完整 MVP 或真实浏览器主链通过。** F1 关闭的是 nil active 被误当历史、决定/补正入口误禁用、旧历史不可选的 reader 问题；用户原先看到的“待复核”文字还有独立的既有展示原因，见 F6。

本次仅更新本报告。未修改源码、接管浏览器、调用真实 API/数据库、派子代理、安装依赖、提交或推送。后端只读用于核对字段与拒绝条件，不作为权限审计或运行证明。下方 R1 原文保留为首次失败证据；其状态、行号、旧入口与缺失证据描述均为历史记录，以本节及主线程最新证据为准。

### 冻结版本与执行边界

- 最终绑定 [R3 交接](../frontend/mobile-nav-fix/README.md)，冻结时间 `2026-11-07T08:25:55.807Z`；基线 `a806831a7cef839c2b971cd4ab260cce77f42b8b` 上未提交实现。
- 固定入口 `index-p6KdlFyE.js`，SHA256 `9221fc5e276b9375a744b51dcbc652b8e12ca6422f07eea76e3df79df9d823d5`。
- 构建归档 SHA256 `ef0dcd9b5cee75717585aeddd487edadee337378609a83722c4e80c802c0254b`；源码归档 SHA256 `31f28869cd5706c3d733038abf5bcb5655b1a7fb0e7a6ba54334f585da6c6860`。独立重算一致，归档内 27 个源码文件和 7 个 dist 文件逐一与清单匹配。
- 16:35 CST 末次工作目录校验：27 个源码、7 个构建文件、5 个构建输入、14 个所列回归测试文件均与 R3 清单匹配。主 TSX SHA256 `eabb5c25c92516a04862a835f0149f787af0d4fa1a0393082d9703ab8b9d8b44`。
- 独立对比 R2/R3 归档：主 TSX 差异只在 `ConflictCheckResults` 的已读复核回填及文字；其他实现差异为 Drawer / 离开提示局部层级。F1–F4 逻辑未变。R2 `index-Cm1X_6UB.js` 及旧归档未被用来冒充 R3。
- 下列固定构建探针从 **R3 tar 在内存读取并打包**，使用 JSDOM、虚构账号、全量 mock fetch/XHR；未写探针文件。账号切换探针使用 R3 归档源码与现有项目依赖，在内存暴露 store 驱动同组件身份切换。没有执行浏览器，也没有真实接口成功证据。
- 已核对作者日志：R2 12 文件 75 tests；R3 增量 3 文件 9 tests + 1 文件 6 tests。**75 项没有在 R3 全量重跑，也不是本审查独立执行的测试。** 作者 typecheck/lint/build 不替代下列行为探针或浏览器验收。

### F1–F4 修复逐项判定

| 编号 | 本次结论 | 独立证据与具体实现 |
| --- | --- | --- |
| F1 · 原 P1 | 原 reader 阻断关闭；另保留 F5/F6 | [Batch01Prototype.tsx:8270](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L8270) 比较已解析的当前版本与正在查看的版本；8379–8385 将默认 reader 的 `viewed_snapshot_id` 作为经服务端核验的当前版本，不再丢弃旧类型历史。合法非接案旧当前版本（active=null）与新当前版本均显示 3 个处理动作；切旧历史后为 0、显示只读，切回恢复 3 个。当前 live review=no_conflict 时同意按钮可用，未把 review 更新当 ID 漂移。旧样本使用事实相同的持久旧记录与可解析当前 UUID，未以相互冲突的旧来源假造合法兼容。 |
| F2 · 原 P1 | 关闭 | [Batch01Prototype.tsx:4535](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L4535) 读取默认快照，[intakeSupplement.ts:109](../../../../../frontend/src/services/intakeSupplement.ts#L109) 拒绝空 ID、`SNAP_*` 和 active/viewed 不一致；4753 的重提使用已解析 `submissionSnapshotID`。旧文字补正先 PUT 受控回复、获得 `rev-reply`，随后两次 POST 均携带持久 UUID `00000000-0000-4000-8000-000000000011`、同一 return/item ID，响应中断后的请求 body 完全相同，没有额外事实 PUT。三种不可核验快照样本均保持重提禁用、无写请求。 |
| F3 · 原 P1 | 关闭 | [Batch01Prototype.tsx:5092](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L5092) 用请求代次、owner、挂载状态、编辑签名统一保护；URL、本地关联草稿、列表恢复及 409 重载使用该边界。独立结果：B 先回→A 迟回后仍为 B；读取期间新增输入保留；409 重载期间新输入保留；账号 7 请求未完成时切 8，旧响应后标题仍为空，账号 8 草稿缓存不存在旧内容。 |
| F4 · 原 P2 | 关闭 | [Batch01Prototype.tsx:4177](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L4177) 按 owner + URL 参数重建编辑会话。固定 bundle 从 `/case/create?intake_id=A` 经 SPA history/popstate 切 `/case/create` 后标题为空；重新选客户/律师并保存只产生 `POST /case-intakes`，没有 expected_revision、没有 PUT A。 |

草稿恢复另验证了真实 DTO 字段形状：`task.checkId/intakeId/clientId`、`review.checkId`，以及当前 `conflict_ready` 绑定。完整虚构样本恢复后“提交审批并等待成案”可用，未改动即点保存不产生多余 PUT。用户已报告 **A 真实草稿恢复由主线程 PASS**；这是主线程证据，本审查没有重复执行或扩大成其他浏览器场景 PASS。

### 旧业务、旧接案与只读边界

1. **合法非接案旧审批，有可核验持久快照**：默认 reader 返回的当前版本可正常处理；显式历史只读。恢复的是既有合法能力，不是把 nil active 当作授权凭据。
2. **intake 关联的旧 conflict_approval，缺新快照/唯一槽**：后端 [approval_intake_case_guard.go:87](../../../../../internal/services/approval_intake_case_guard.go#L87) 与 100–105 拒绝同意/成案。固定 bundle 的隔离探针让当前快照可读、决定返回 `409 INTAKE_SUBMISSION_REQUIRED`：前端只发一次原集成决定请求，携带原始 action_version 和持久 viewed ID，显示服务端错误、停留当前页；没有降级调用旧接口、清空令牌、重建新审批或自动重试。**读取到旧快照不证明该旧接案单可批准；当前 UI 仍可能展示按钮，真实后端拒绝必须保留。** 本审查没有批准或访问现场 `d08a805d-6994-40cd-ac78-488485da80ab`。
3. **来源无法核验/所选历史**：快照出错或历史选择均不提供处理通路。只读历史 reader 与成案授权不是同一判断；不得以修复 F5 为由放行所有 404、nil active 或缺少前端 intake 字段的记录。

### F5 · P2 · 未关闭：合法非接案无快照审批被通用提交版本门禁误阻断

**位置**：[Batch01Prototype.tsx:8386](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L8386)–8388；同文件 8514–8515、8796。前者把所有快照读取失败都变成处理错误；后两者还无条件要求提交快照令牌。

**触发**：有效的普通非接案审批（例如 `type=leave`），无 intake/冲突/成案关联、无 active 指针、从未生成提交快照；详情返回当前审批人和允许的 approve/reject/request_changes，`GET /approvals/:id/snapshot` 按真实契约返回 404。

**依据**：这类对象不属于新接案门禁拒绝对象。[intake_submission_snapshot.go:111](../../../../../internal/services/intake_submission_snapshot.go#L111)–118 允许非接案无快照解析元数据，311–315 对首次普通提交直接返回；[approval_submission_history.go:65](../../../../../internal/services/approval_submission_history.go#L65)–66 则明确没有历史快照时返回 not found；[approval_workflow_transactions.go:154](../../../../../internal/services/approval_workflow_transactions.go#L154) 只在存在 active 指针时强制匹配快照令牌。处理仍须通过本来的身份/状态/版本校验，本审查未重复审计这些权限。

**独立复现**：R3 完整 App `/approval/approval-test`，干净普通审批数据（无 conflict_result、无 intake）+ mock 404，显示“提交版本读取失败，请重新加载后再处理审批”，`decisions=[]`、无执行异常。基线 a806831 中快照失败只清空展示（8383–8385），决定门没有强制 snapshot 条件（8780）；这是本轮通用详情 reader 的兼容性回归，非原有普通审批权限问题。刷新不能补出本来就不要求存在的快照，用户无法同意、拒绝或退回。

**需要修复/补证**：按服务端可核验的业务模式/能力区分普通旧业务与被门禁封闭的旧接案。不能只删除 `snapshotError`，因为提交函数仍强制 active ID；也不能仅凭前端 `metadata.intake_id` 为空放行，后端可能从持久检测/跟踪关联识别 intake。补充普通审批无快照可正常决定、旧接案缺快照/槽仍被拒绝两组契约与 UI 证据。该问题不直接阻断已有合法 active 快照的新 case_creation 链，但阻止整体兼容验收。

### F6 · P2 · 未关闭：审批详情仍把覆盖完整/已复核样本显示成覆盖受限/待复核

**位置**：[Batch01Prototype.tsx:8713](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L8713)–8718、8925、8945–8947。最新 review 只参与处理门计算；风险摘要仍沿用机器 `overallRisk=REVIEW_REQUIRED`。`approvalNoMatch` 为真就硬编码“未发现匹配证据（检索覆盖受限）”，没有检查 coverageStatus。

**触发/独立复现**：当前 `case_creation` 持有有效 active/viewed 相同 UUID；冻结检测 `decision.status=REVIEW_REQUIRED`、`coverageStatus=COMPLETE`、`riskAssessment.overallRisk=REVIEW_REQUIRED`、零匹配；同 checkId 的最新 review 和冻结 review 均为 no_conflict。R3 实际 DOM 同时出现：

```text
同意并成案：enabled；拒绝/退回修改：存在
总体风险等级：待人工复核
未发现匹配证据（检索覆盖受限）
```

**影响与归因**：审批人已可正常处理，但页面仍给出相反的覆盖信息，且没有独立呈现已读取的复核结论。这是业务证据含义错误，不是纯样式意见。它解释了原现场“仍待复核”文字可能残留的部分；并不表示 reader 仍把 review 更新当证据漂移。基线 a806831 的 8909、8930 已有相同摘要逻辑，故**不是 R2/R3 新引入的回归**。R3 两处文字修复在 `ConflictCheckResults`，没有覆盖 `ApprovalDecisionFlow`。

**需要修复/补证**：审批详情分别表达冻结的自动检索结果与该版本/当前可读的独立复核，不从零匹配推断覆盖受限；显式历史必须继续展示该历史事实而非混入最新 review。至少补 coverage=COMPLETE/非 COMPLETE、current review=no_conflict、历史只读三类展示检查。该问题不阻断本次已修复的处理动作，但不能把原“待复核”展示症状也报告为全部关闭。

### 固定入口与尚未关闭的证据

- **两个真实业务入口正常挂载（DOM/mock 层）**：R3 完整路由 `/client` 显示客户主档案，`/case/create` 显示立案工作台，两处均存在共享“新增客户”入口；未见路由/执行异常。实际构建仍由 `frontend/index.html → src/index.tsx` 进入；没有把未参与 HTML 构建的 main.tsx 计成另一次通过。
- **行为问题未关闭：F5、F6**。旧接案审批收到后端拒绝属于应保留的限制，不能作为待“恢复按钮”的缺陷处理。前端仍可能显示处理按钮这一体验不等于门禁失效，本次未证明任何绕过。
- **真实浏览器/后端闭环仍由主线程补齐**：R3 drawer 关闭点击与焦点、离开保护真实导航、C16 两条新完整 case_creation 链、整体 C17；本审查不接管、不关闭已知遮挡项。主线程已验证的 A 恢复及其他客户/移动场景，按其实际证据更新，不以本报告的旧 R1 清单覆盖主线程新结果。
- **本次探针不证明**：数据库幂等重放、跨窗口竞争、真实登录退出/存储隔离、真实旧审批迁移重提及版本过期拒绝、后端最终重启后的真实 HTTP 行为。75 项作者 mock、内存 DOM、代码契约和主线程 browser 证据必须分别计数。真实律师试用、部署和生产资格均未验证。

## R1 首次审查原始记录（历史）

R1 历史结论：**未通过，确认 3 项 P1、1 项 P2。** 两个业务入口的固定构建可挂载；这不等于真实浏览器主链通过。

审查时间：2026-11-07 15:43 CST。范围限共享客户表单/幂等重试、草稿隔离与导航、补正保存/重提、历史快照及 WorkflowAction。只读相关后端契约，没有重复后端权限审计。审查者仅写本报告；没有修改源码、操作浏览器、访问真实 API/数据库、安装依赖、派子代理、提交或推送。

## 版本及证据边界

- 目录：`/private/tmp/lawoa-intake-collaboration-20261107`。
- 基线 HEAD：`a806831a7cef839c2b971cd4ab260cce77f42b8b`；审查对象为其上的未提交实现。
- 作者交接：`../frontend/README.md`，冻结时间 `2026-11-07T07:23:28.424Z`。
- 首次独立校验：source-manifest 的 22 个文件、build-manifest 的 7 个文件全部匹配。
- 固定构建归档 SHA256：`90f94d92d467773cce9e335c7175ad8a07762fab2d7844aeafc9dc6b2c2cad16`；独立计算匹配。
- 固定 JS `frontend/dist/assets/index-CRjDLMzv.js` SHA256：`6b3c3a05f3ae21910b9be1037493372627cb5531644e0f802bec5a59591fcbd3`。
- 审查期间用户通知 Averroes 修复弹窗被侧栏遮挡。末次校验发现 `Batch01Prototype.less`、`Batch01Prototype.tsx` 已变化，其他清单文件及全部固定 dist 文件仍匹配。以下行为复现主要绑定上述固定 dist；账号切换探针绑定当时源码。报告行号按末次源码读取，作者后续修改可能使行号移动。
- 末次读取 TSX SHA256：`78dc4704818f2ae11413968c0c1b397b5ed7b8695a24410e492df34bedc0bdfc`；LESS SHA256：`c60383045c3ea1282306ea4e1016cbd9e37ed6510ab1d6d03a13baa4971b7dad`。

作者 typecheck/lint/build 及 9 文件 47/47 mock tests 日志已核对。本审查没有重复将其计为独立执行。新增复现使用既有 esbuild 的 `write:false` 在内存装载固定产物，运行于 JSDOM，使用虚构身份及全量 mock fetch；没有向网络发送请求、没有新建测试源码文件。JSDOM 不提供真实布局、遮挡、焦点可达性或浏览器存储隔离验收。

## F1 · P1：无 active 指针的当前旧快照被误判为历史，合法处理动作全部消失

**位置**：[Batch01Prototype.tsx:8316](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L8316)，相关位置 8327、8430、8733–8737、8841、9195。

**触发**：合法旧审批的 `/snapshot` 返回 `active_submission_snapshot_id=null`、非空 `viewed_snapshot_id`，例如仍由现有接口新建的 `type=conflict_approval`。这不是损坏记录：旧 writer 在 [demo_aggregate_handler.go:2990](../../../../../internal/handlers/demo_aggregate_handler.go#L2990) 建立审批，在 3021–3024 写 `snapshot_type=conflict_approval`，不会写新协议 active 指针。现 reader 在 3264–3268 分别返回实际 active 与 viewed ID。

**行为与影响**：`Boolean(viewedSnapshotID && viewedSnapshotID !== activeSnapshotID)` 为 true。`canDecisionActions` 因此为 false，即使当前处理人为本人、后端最新复核已是 `no_conflict`，也没有同意/拒绝/退回按钮。`currentConflictReview` 同时跳过已取回的最新复核，改读旧快照根级 `conflict_review`，造成“仍待复核”的展示。旧审批进入 `needs_revision` 时，“补正并重新提交”也被同一个条件禁用。版本列表又过滤掉全部 `conflict_approval` 行，页面没有切回有效旧版本的入口；混合新旧历史中的旧行同样不可选择。

**独立复现**：固定产物模拟 GET 的检测 ID 全部为同一值，review 为 `no_conflict`，申请人 2、当前审批人/登录人 7：

```text
legacy-conflict_approval: decisions=[]; mismatchBanner=false; jsdomErrors=[]
active-case_creation:    decisions=[同意并成案,拒绝,退回修改]; mismatchBanner=false; jsdomErrors=[]
legacy-needs_revision:  supplementButtonPresent=true; disabled=true; versionSelectorPresent=false
```

对照基线的前端代码，旧 `canDecisionActions` 没有 history 条件，`currentConflictReview` 优先使用已读取的独立复核；因此这是**本轮前端 reader/版本状态引入的回归**。它不是必然由新后端权限门造成的拒绝。

**关于用户报告的 d08a805d-6994-40cd-ac78-488485da80ab**：用户报告的“冻结时待复核 + 按钮全部缺失”与上述分支吻合。`hasConflictEvidenceMismatch` 在 8767–8778 只比较检测 ID，既不比较 review 内容，也不比较 review 更新时间；“复核变为 no_conflict 被误当证据漂移”没有代码支持。本审查未读取这个现场对象的 HTTP 响应，不能断言现场不存在另一个 ID 不一致；但在 ID 全部一致的最小样本中已复现同样按钮问题。

**对必需链的判断**：新的 `case_creation + intake_submission + active指针` 不被这一条件直接阻断，对照样本也支持该结论；可继续独立验证真正的新链。旧协议办理/旧文字退回/多版本可查阅仍是 C15 必需兼容范围，不能据此标记整个 MVP 通过。另需按实际契约区分类型与副作用：现有 `RequiresCaseCreationApproval`（[approval_conflict_integration_service.go:1379](../../../../../internal/services/approval_conflict_integration_service.go#L1379)）连基线也将 `conflict_approval` 列为可触发成案的类型，不能仅凭旧审批名称断言批准它没有成案行为。本审查没有触发决定。

**修复后应验证**：合法旧当前版本、新当前版本、明确选中的旧历史版本三种状态分别处理；旧当前审批能消费最新复核且保留原有处理权限，历史仍只读。缺少/冲突的历史来源仍按真实错误阻断，不能一概把所有 nil active 记录放行。

## F2 · P1：旧接案审批重提只取新 active 字段，发送空版本令牌

**位置**：[Batch01Prototype.tsx:4737](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L4737)，恢复流程 4507–4524。

**触发**：可核验的旧接案审批已经 `needs_revision`，后端能返回规范化旧文字补充项，但 `active_submission_snapshot_id` 为空。即使修复 F1，或从合法 WorkflowAction 双 ID 链接直接进入补正页，此处仍只发送 `revisionApproval.active_submission_snapshot_id || ''`；补正页没有读取兼容快照版本。

**后端契约**：[intake_submission_snapshot.go:143](../../../../../internal/services/intake_submission_snapshot.go#L143) 用经过核验的持久旧快照 UUID 返回兼容 `snapshot_id`，没有伪造 active 指针。重提在 242–244 调用 fingerprint，271–272 要求 expected snapshot 非空；随后 334 才比较具体旧快照 ID。不能用内嵌旧 `SNAP_*` 标识替代持久行 UUID。

**独立复现**：固定产物加载合法形状的旧文字退回、填写回复和修改说明后点击重提，捕获的 mock 请求依次为：

```json
{"method":"PUT","path":"/case-intakes/intake-A","expected_revision":"rev-intake-A","return_record_id":"return-old"}
{"method":"POST","path":"/approvals/approval-old/resubmit","expected_intake_revision":"rev-reply","expected_submission_snapshot_id":"","return_record_id":"return-old"}
```

**影响**：逐项回复可以保存，但原审批无法进入首个新协议轮次；真实后端静态契约会以 `SUPPLEMENT_ITEM_INVALID` / 400 拒绝空令牌。本探针证明前端实际生成了该空字段，没有调用真实 API 验证 400。F1 仅修正显示逻辑不足以关闭本项。

**修复后应验证**：补正页读取服务端确认的当前/兼容版本，旧文字回复原 ID 保持不变；首次迁入新协议与同请求重放都使用正确令牌；来源无法确认时保留具体错误。

## F3 · P1：列表恢复草稿的迟到响应未绑定账号和选择序号，会覆盖当前内容并串入新账号缓存

**位置**：[Batch01Prototype.tsx:5223](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L5223)–5242；账号清理 4363–4376；缓存写入 4383–4385。相似未绑定回调见 5158–5169 和 5282–5289。

**触发一（同账号，固定产物独立复现）**：空白立案页有 A、B 两个服务端草稿；连续点两次“恢复此草稿”，让 A 的 GET 延迟、B 先返回。B 显示后 A 才返回。

```text
先返回当前选择 B：title=草稿B
旧请求 A 随后返回：title=草稿A
```

恢复按钮未在请求中锁定，响应也没有 sequence/intake ID 校验，最后到达的旧请求可覆盖后选草稿或等待期间的编辑。对应 C04 精确恢复及保留当前输入。

**触发二（同一组件存活时账号切换，源码 DOM 探针独立复现）**：账号 7 点恢复 A，暂停 GET；按现有账号隔离测试相同方式把 store 切到账号 8，等待表单清空，再释放账号 7 的恢复响应。结果：

```text
切换账号后：title=""
旧账号响应后：title="账号七独有事项"
localStorage["law-oa-case-intake-draft-v2:8:new"]:
  intakeId="A", title="账号七独有事项", description="账号七摘要", revision="rev-A"
```

**影响**：新增加的账号清理 effect 和 `formOwner` 写缓存判断，不能拦住这些旧回调：它们重新 setForm/setDraftActive，当前 effect 随即按账号 8 的 key 写入账号 7 的事项。已复现的是组件未卸载的身份切换状态，**没有把它等同于正常退出登录后新页面必然泄漏**；正常退出通常会卸载组件，该浏览器路径未测。作者的账号测试只切换同步本地输入，未覆盖在途恢复响应。

这是既有恢复回调没有纳入本轮账号隔离/异步恢复边界的问题，不是后端跨账号授权漏洞的审计结论。建议所有恢复回调使用一致的 owner + 请求代际校验；仅检查缓存 key 或在切换时清一次表单不足。

## F4 · P2：同组件从具体草稿 URL 切回空白入口时保留旧 intake ID，保存仍覆盖旧草稿

**位置**：[Batch01Prototype.tsx:4497](../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L4497)–4503，保存目标选择 4630–4638。

**触发**：已在 `/case/create?intake_id=A` 恢复 A；同一 SPA 路由实例转到 `/case/create`（例如该路由的 history/popstate 变化）。effect 因 `!requestedIntakeID` 直接返回，没有清理 form/runtime 中的 A。

**固定产物独立复现**：使用内存 DOM 的 history.pushState + popstate 模拟同路由历史导航；观察 URL 已无 intake_id，但标题仍为“已有草稿A”。把标题改为“原以为是新案件B”后点“保存草稿”，捕获的是：

```text
PUT /api/v1/case-intakes/A
body.title="原以为是新案件B"
body.expected_revision="rev-A"
```

**影响**：空白入口与页面所持对象不一致，继续保存可能改写原草稿。新加的 deep-link 恢复必须定义离开该对象后的状态重置。此探针确认路由状态变化后的行为，没有操作真实浏览器后退；尚缺具体导航控件/历史链的浏览器复现。硬刷新 `/case/create` 会重挂载，此问题不涵盖该情形。

## 固定构建入口及其他核对结果

实际 Vite HTML 入口是 `frontend/index.html` → `/src/index.tsx`，使用单个 `createBrowserRouter`、`RouterProvider`，`App.tsx` 保留 `/client` 和 `/case/create` 业务路由。`main.tsx` 也是 data router 形式，但不在这次 HTML 构建入口中；不能把它另计为已运行的固定产物。

在固定 dist、完整 App 路由、虚构已登录律师、mock API 的独立 JSDOM 环境中：

| 路由 | 实际挂载结果 | 共用客户入口 | 异常 |
| --- | --- | --- | --- |
| `/client` | 客户主档案；请求 clients 列表和指定 master-profile | “新增客户”存在 | 未见路由/执行异常 |
| `/case/create` | 新建案件立案工作台；请求 clients、lawyers、drafts | “新增客户”存在 | 未见路由/执行异常 |

以上仅说明固定构建在 DOM 模拟环境中路由可达，没有确认真实 HTTP 深链接回退、浏览器渲染或弹窗遮挡已通过。

本范围其他代码检查未发现新的确定行为缺陷：

- 两入口确实复用 ClientCreateDialog；并发 busy 门在校验后再次检查；不确定响应保留 key，同 key 改请求仍由后端拒绝；成功以响应 ID 回填。取消清空表单/key，失败后明确新请求按钮另换 key。客户端不按名称合并客户；远程搜索/后页使用 ID，旧搜索响应受 sequence/abort 保护。
- 保存签名排除 revision/对象 ID，保存期间新增编辑不会被当成已保存。导航 guard 有三选项和 beforeunload；保存失败留页。其真实键盘焦点、遮挡与页面退出体验仍待浏览器验证。
- 补正回复专用 PUT 仅发 approval_id、return_record_id、expected_revision、supplement_responses；与后端受控 payload 一致。回复的轮次及 item ID 来自服务端；事实 PUT 不混入回复字段。
- 新协议当前/历史选择采用独立 snapshot 请求，历史或读取失败禁用决定；新决定携带原始 action_version，没有经 Date 丢精度。
- WorkflowAction 按字符串 resource_id 映射，明确 intake + approval 双 ID 才生成补正链接；缺失/blocked 动作禁用。待办详情会重新读取；业务 workflow 项没有手工完成按钮。
- 作者交接的两处后端投影缺口在本轮只读时已见代码闭合：`inbox_workflow_action.go:382` 输出 ApprovalID；`approval_handler.go:256` 合入服务端规范化 supplement context。不把早期交接备注继续当作当前未实现结论，也不把代码存在当作 HTTP 验证通过。

## 尚缺证据与交接

1. 用户已报告真实浏览器中新客户弹窗被 sidebar 遮挡，Averroes 正在修复；本审查不重复排查、不关闭该项。**C16 仍未通过**，后续 CSS 构建及浏览器复测由主代理绑定新产物。
2. C01–C03 仍需真实 API 下响应丢失后的同 key 重放、改请求 409、超过 50 条授权结果及新客户回填的浏览器证据。本报告的 DOM/mock 结果不是客户端幂等数据库证明。
3. F1/F2 修复后的旧文字退回、新旧混合历史，以及用户报告旧 conflict_approval 的真实响应和复测尚缺；本报告不读取/更改该现场对象。
4. C04 的在途恢复、跨账号、双窗口和真实导航应补复测；F3 的同账号逆序和组件存活账号切换已独立模拟复现。
5. C08–C11 的真实补正保存→刷新→事实重检/独立复核→原审批重提→版本过期拒绝，及 C16 两条独立同对象链尚不能由 47 项 mock tests 判为通过。
6. C17 390×844、Tab/Shift+Tab/Enter/Escape、长补充清单定位及离开弹窗焦点均未在本审查执行。真实律师试用、部署、生产资格也未验证。

应先关闭本报告确认项并绑定修复后的源码/构建，再补相应验证。本报告不修改验收矩阵，不替主代理宣告其他后端检查完成。
