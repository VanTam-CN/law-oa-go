# 后端最终复核：候选 3f97004（含 B5、C06 补证）

## 双 tracking 配置形态：独立测试审查冻结（2026-11-07 18:24 +08:00）

**本次窄审查通过，可冻结新增测试及其证据；未发现阻断。** 仅审查 `approval_multistage_postgres_http_test.go` 相对 `3f97004` 的测试 diff（98 行增加、3 行删除）、[multistage-config/HANDOFF.md](../multistage-config/HANDOFF.md) 和实际六叶子运行日志。产品实现未重新审查，也未执行任何测试或数据库请求。**新增两个 config 用例属于“独立测试审查通过 + 作者实际运行通过”，不属于本审查者独立执行。** 原 B5 四叶子的独立执行结论仍以其原测试 hash、时间和下文记录为准，不扩计为六个。

### 新断言是否支撑交接结论

- [测试:41](../../../../../internal/services/approval_multistage_postgres_http_test.go#L41) 从真实 fixture intake 读取配置，首提请求同时携带顶层及 metadata `case_creation_config`。仍通过真实 HTTP 创建审批，未手工创建、更新或重排 tracking；使用既有真实 PG production bootstrap fixture。新增普通/integrated 两个 `config_direct_two_stages` 均为直接两级批准，不退回、不重提，原四个场景保留。
- [测试:86](../../../../../internal/services/approval_multistage_postgres_http_test.go#L86) 显式断言两行及时间先后：较早 `INTAKE_BIND_<approval>` 绑定正确 intake，较晚 `CASE_TRACK_<approval>_…` 的 intake 为 NULL，均 pending；真实 repository 的 latest getter 返回较晚 generic ID。该用例验证了实际选择到 generic 的形态，支持交接中的时间排序结论，未证明或宣称 canonical 优先。
- [测试:166](../../../../../internal/services/approval_multistage_postgres_http_test.go#L166) 中间决定 HTTP 200、under_review/compliance/44；tracking 总数为 2，整表完整 JSON 与决定前完全相等，case INSERT 计数和 cases 均为 0，下一审批人仅一个开放决定待办。由初始两行 pending 与整表相等共同证明没有隐藏的 failed 行，不依赖未指定排序的单行 `tracking()` getter 得出双行结论。
- [测试:228](../../../../../internal/services/approval_multistage_postgres_http_test.go#L228) 最终与再次批准后都按固定 ID 分别重读两行：canonical 全结构等于最初记录（因此仍 pending），generic 为 completed、NULL intake 并指向同一 `CreatedCaseID`；总数仍为 2。案件总数为 1、决定数为 2；案件 client/lawyer/check 与 fixture 对应，同案有效 CLIENT/OPPOSING 角色及非空身份 digest 精确匹配，snapshot 数为 1。重放 HTTP 200 且审批仍关联原案，未用单纯行数冒充关联正确。
- [状态查询断言:68](../../../../../internal/services/approval_multistage_postgres_http_test.go#L68) 实际 GET integration status 在初始/中间返回 pending、空 case ID，最终/重放后返回 completed、相同 case ID；每次读取前后 12 张业务表相等。重放调用 `body()` 读取当前审批版本后再次批准，证明该调用方式下的同案重放；不延伸为旧版本原请求重发、并发重放或双 tracking 故障重试。

| 阶段 | canonical（有 intake） | generic（NULL intake） | GET integration status | 案件 |
|---|---|---|---|---|
| 首提及中间节点 | pending，保留原行 | pending，保留原行 | pending，空 case ID | 0 |
| 最终及同案重放后 | 原行完全不变，仍 pending | completed，绑定同一 case ID | completed，同一 case ID | 1 |

### 作者运行日志的独立核对

逐条解析 [02-pg-race.jsonl](../multistage-config/02-pg-race.jsonl)：六个叶子 PASS，0 FAIL、0 SKIP，与 `04-results.json` 六个结果逐项一致；包终态 `2026-11-07T18:18:10.15628+08:00`，15.061s、PASS。作者记录使用 `-race -count=1`、exit 0；日志无 race 报告，有既有 LC_DYSYMTAB 链接警告。

- 普通 config 叶子：日志第 55 行为独立空 schema 的 v16 bootstrap/119 表；第 59、60、66、68 行分别记录初始 pending、较早 canonical/较晚 generic、中间 200/两行/零案及中间 pending；第 76、81 行最终与重放均 completed/case=1，第 84 行 PASS（1.96s）。
- integrated config 叶子：对应第 129、133、134、139、141、148、153 行，第 156 行 PASS（1.88s）。两个 config 叶子各在自己的 schema 内得到 case=1，不把跨 schema 的相同数值当成同案证明；同案结论来自各叶子的变量与数据库关联断言。
- 原四叶子 PASS 事件位于第 28、52、105、126 行；本次仅核对日志与保留的断言，没有重跑、替换或合并此前的独立执行计数。

### 冻结绑定与修改边界

独立比对作者前后清单、当前文件与 `3f970040f5e40b8016c18a070b5aa1345d2b2eec` Git blob：**321 个受跟踪非 test Go 产品文件全部一致，清单覆盖完整、漂移 0**。保存的 `05-test-only.diff` 与当前实际 diff 字节一致，diff-check 通过。

| 本次冻结对象 | SHA-256 |
|---|---|
| approval_multistage_postgres_http_test.go（新增两叶子后的完整文件） | `bd432dd45b244a236d563a8f96f7017f31dab4c3fe936c5e7cf83aabc7a8d563` |
| multistage-config/02-pg-race.jsonl | `2bf20119ff3b1853cfc1c6c3d7ae499e7908e886ca0649ddddefb6f591f27235` |
| multistage-config/04-results.json | `aa96931547212cd0fc1907bf84f110ed288eb07b0b34c723ddb96a3a96c83aa2` |
| multistage-config/05-test-only.diff | `0df8f45d5f3742938c8a375d7046df7207fb827ffba5009186135c938ec5a220` |

本结论绑定产品提交与上述测试/证据 hash；新增测试 diff 尚在工作树，未包含在 `3f97004` 原提交中，最终纳入由主线程处理。本轮只追加本报告，未改产品、测试或作者日志，未浏览器操作、派发代理、暂存、提交或扩审。此前 C06、B5 和原始红灯记录原样保留；本专项审查至此冻结。

## C06 补证独立复核（2026-11-07 18:18 +08:00）

**本轮指定的 C06 两处证据缺口已关闭：助理补充/重提的直接 HTTP 拒绝，以及核查岗对象权限撤销后的 Inbox 四入口收敛与恢复。** 审查限于 Euclid 新增的一个测试文件、直接 fixture/HTTP router 和原始日志；未发现需要修复的断言缺陷或本范围阻断。没有扩展为全角色、全字段、全接口的排列组合，也没有重开 B1–B5 产品审查。

### 实际断言核对

- [助理用例：inbox_c06_postgres_http_test.go:47](../../../../../internal/services/inbox_c06_postgres_http_test.go#L47)：数据库中实际存在 active assistant 44；原审批由真实服务创建并退回，补充 DTO 带有效 return record、revision 与 responses。助理 List/Stats/home 返回 200 且不暴露目标 approval/intake ID 和标题，List 空且 total=0，原律师待办 GET=404。实际 PUT 补充、POST 重提均 403，每次比较 12 张业务表的完整有序 JSON，不仅比较行数；拒绝响应不含测试中的标题、退回记录 ID 和身份值。随后申请律师以同一补充 payload PUT=200，读取保存后的新 revision 后 POST=200，持久状态为 resubmitted 且 active snapshot 指针改变，排除了无效请求或端点失效造成的假阳性。助理 Stats/home 这里只断言成功状态及目标内容不泄漏，没有将其宣称为所有计数字段均显式断言为零。
- [核查岗用例：inbox_c06_postgres_http_test.go:107](../../../../../internal/services/inbox_c06_postgres_http_test.go#L107)：新建有真实 case 上下文的 check 和 active 独立指派，由 `ReconcileConflictInboxInTx` 生成待办；没有删除已完成复核或改写冻结快照来伪造待办。授权有效时 List/home 各一项、needs_action=1、GET=200，三处动作一致为 `review_conflict` 并指向实际 check ID。启用案件隔离墙后 List/home 为空、List total 与 Stats total/needs_action 为 0、GET=404；List/Stats/home 三者 stats 一致，四入口完整响应均不含标题、客户名、身份/意见哨兵或 check ID。为同一核查岗恢复白名单后，原同一任务 ID 再次可见、动作恢复；三阶段读取前后 12 表均不变，任务始终未完成。授权态 GET 显式包含标题、身份及意见哨兵，避免“从未暴露该字段”的空断言；客户名仅为撤权态附加哨兵，不推称全面字段泄漏验证。
- [真实 HTTP 接线：approval_backend_final_review_http_test.go:20](../../../../../internal/services/approval_backend_final_review_http_test.go#L20) 使用实际 Gin handler、授权服务及业务事务，只有身份由测试 middleware 注入。新路由的 Inbox 与首页也接入真实 service，没有模拟权限返回值。fixture 的 [state():219](../../../../../internal/services/approval_collaboration_postgres_test.go#L219) 比较审批、快照、tracking、Inbox、接案、案件及主体等 12 张表；这个边界不等于全数据库无写入。

### 独立执行结果

```sh
env GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache GOPROXY=off GOTOOLCHAIN=local \
  LAW_OA_COLLAB_PG_DSN='host=/private/tmp/lawoa-intake-collaboration-runtime/socket port=55437 user=mac dbname=lawoa_intake_mvp_test sslmode=disable' \
  go test -mod=readonly -race -p 1 ./internal/services \
  -run '^TestInboxC06PostgresHTTP$' -count=1 -timeout=120s -json
```

本审查者实际执行，**2 个 realPG/HTTP 叶子 PASS，0 FAIL、0 SKIP，exit 0**；包终态 `2026-11-07T18:17:07.436011+08:00`、耗时 4.168s。数据库实际报告 PostgreSQL 16.10；两个叶子各用随机私有 schema，fixture 校验固定本机 socket/端口/库名/用户，清理回调正常完成。无 race 报告。链接器出现 LC_DYSYMTAB warning，未妨碍构建及测试执行，不把运行描述为无警告。

| 独立用例 | 结果 | 直接观察 |
|---|---|---|
| assistant_controlled_write | PASS，0.38s | PUT=403、POST=403、12 表不变；申请律师 save=200、resubmit=200、new_active=true |
| conflict_officer_revoked_context | PASS，0.30s | 同一 task=4；List total/needs_action/home 数量 1→0→1，GET 200→404→200；三阶段 read_only=true |

该结果单独登记，**不重写作者既有 300/111，也不混入此前独立 40/20 或 B5 的 4 个叶子统计**。测试使用 AutoMigrate、最小 intake schema 与协议迁移，属于真实 PG/handler 授权和业务验证；不代表 production bootstrap、JWT 登录、浏览器验收或所有岗位/指派撤销排列组合。

### 内容绑定、原始失败与交付边界

- 产品绑定 `3f970040f5e40b8016c18a070b5aa1345d2b2eec`。独立核对 `01-source-before.json` 的受跟踪产品 Go 文件及 go.mod/go.sum（合计 323 项），清单覆盖与 Git 跟踪范围一致；文件内容、该提交 Git blob 与清单 SHA-256 全部匹配，复跑后产品漂移 0。
- 新测试 [inbox_c06_postgres_http_test.go](../../../../../internal/services/inbox_c06_postgres_http_test.go) 复跑前后 SHA-256 均为 `fdeda3a62b9beaa237180892faced5f357ba89c97d190d1da972900e2c4316bb`，与作者最终格式化版本一致。本次独立执行覆盖了作者最后那一处 tab 格式调整后的实际文件，未只依赖格式调整前的作者 PASS。
- **新增测试仍被 `.gitignore:282` 的 `*_test.go` 忽略，尚未包含在 3f97004 中。** 本结论绑定“产品提交 + 上述新增测试 hash”，测试需由主线程显式纳入交付；本审查者未暂存或提交。复跑后观察到并行修改 `approval_multistage_postgres_http_test.go`，它属于范围外的 B5 测试，未被此次 C06 选择器执行，也未被本结论认证；没有据此扩审或覆盖他人工作。
- 已读取并核对 [HANDOFF.md](../inbox/c06/HANDOFF.md)、`05-results-and-audit.json` 及其证据 hash，漂移 0。作者 `02-http-race.jsonl` 是 socket 权限导致 setup 失败；`03-http-race.jsonl` 助理已通过，核查岗因缺少可验证上下文失败，作者记录为 fixture 的 subjectCaseId 数字类型错误，已核对当前源码使用字符串值；`04-http-race.jsonl` 才是两叶子全部通过。两轮失败和格式差异原件均保留，没有改记为产品首红或全绿。

本轮只更新本复核报告，无产品/测试源码修改，无浏览器、子代理、真实环境变更或 Git 提交。C06 本次要求的缺口已关闭；C01、C05、C08 及最终验收矩阵仍由主线程按各自证据处理。

## B5 最终专项结论（2026-11-07 17:46 +08:00）

**B5 已独立复核通过并关闭；当前限定后端审查范围可冻结到 `3f970040f5e40b8016c18a070b5aa1345d2b2eec`。** 本轮只读复核一个生产文件的 8 行 diff、新增两级 PG/HTTP 测试及直接调用链，未扩全库。新候选真实浏览器两链仍由主线程执行，本结论不代替该验收。

原问题保留：主线程在 `7c5618f` 浏览器观察中间节点已经推进、接口却因尝试成案返回 409。作者 `multistage/01-first-red.log` 的 integrated 分支确有 HTTP 409、under_review/compliance/44、tracking=failed、cases=0；普通分支该轮先遇到错误 envelope 的测试预期问题，没有把它算作同一业务首红。首红文件 SHA-256 仍为 `a5d5a2b5098f21df7d9436503a437d364cba80c71d0e35c6204204d4175d6b2f`，未修改。浏览器旧失败样本未由本审查者读取或修写。

### 实际修复与调用链

- [approval_conflict_integration_service.go:1166](../../../../../internal/services/approval_conflict_integration_service.go#L1166) 在真实 `ProcessApproval` 成功返回后，要求 `decision=approve && requiresCaseCreation && updatedApproval.Status==approved` 才进入成案；并向 `ensureApprovedApprovalHasCase` 传决定后的对象。中间节点直接返回已更新的 under_review 与下级审批人，不再执行成案及 failed-tracking 分支。
- 核对 [approval_service.go:524](../../../../../internal/services/approval_service.go#L524)、`approval_assigner.go:206`：非最后一级设置下一节点/审批人及 under_review，最后一级才设 approved。决定事务提交成功才返回 integration；本 delta 没有改变审批阶段定义、独立审批人校验、版本与唯一槽门禁。
- 决定返回对象通过 `GetApproval` 重新加载并执行 active snapshot 内存投影（`approval_service.go:162`、`:575` 附近），因此换用 updatedApproval 不会把重提后的 frozen 内容降级为旧 metadata。实际成案事务仍重新解析 active 快照。
- 每次 approve 前的独立复核、风险及成案材料/承办律师预检保留；最终成案失败的 tracking/retry 和已成案同案重放分支未放宽。普通错误 envelope 与 integrated 错误 envelope 的既有差异未在本次修改。

### 独立实际执行：4 PASS / 0 FAIL / 0 SKIP，全部 realPG

```sh
env GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache GOPROXY=off GOTOOLCHAIN=local \
  LAW_OA_COLLAB_PG_DSN='host=/private/tmp/lawoa-intake-collaboration-runtime/socket port=55437 user=mac dbname=lawoa_intake_mvp_test sslmode=disable' \
  go test -mod=readonly -race -p 1 ./internal/services \
  -run '^TestApprovalMultistagePostgresHTTP$' -count=1 -timeout=120s -json
```

独立 run 终态为 `2026-11-07T17:45:31.549292+08:00`，包耗时 `10.998s`，exit 0。实际调用 production bootstrap 创建各自空私有 schema，再使用真实 Gin handler 初次提交及两级决定；身份通过测试 middleware 注入，未经过登录 UI，检测/复核资料为合法 fixture 输入。没有替换授权、决定、tracking 或案件事务。

| 独立 PG 叶子 | 结果 | 核心断言 |
|---|---|---|
| ordinary/complete | PASS，1.89s | 普通入口中间 200，最终 200、同一案件重放 |
| ordinary/final_case_failure_retry | PASS，1.93s | 中间 200，最终 INSERT 故障明确失败，恢复后真实 retry 唯一成案 |
| integrated/complete | PASS，1.82s | 集成入口中间 200，最终 200、同一案件重放 |
| integrated/final_case_failure_retry | PASS，1.95s | 中间 200，最终 INSERT 故障明确失败，恢复后真实 retry 唯一成案 |

四个场景均逐项验证：

1. 停用承办律师时预检返回 409，12 表状态不变、case INSERT 未触达；恢复后继续原节点。
2. 首节点 HTTP 200、under_review/compliance/44，tracking 表只有原来一行 pending，**整行及整表内容都不变**；case INSERT 计数=0、cases=0、approval_records=1，下一审批人仅一个开放决定待办。它不仅检查“最新 tracking 不是 failed”，也排除了另增 NULL-intake failed 行。
3. 旧节点重放 HTTP 409，12 表不变；最终节点才触达实际 case INSERT。
4. 正常最终批准只有一个案件、两个 parties、两条决定；故障注入发生在真实案件写事务，cases/parties 均无残留，审批为 approved、两条决定、tracking=failed。移除故障后实际 HTTP retry 得到唯一案件；再次决定 replay 不新增案件/决定且复用原 case ID。

成功输出中的直接观察为：`first node HTTP=200 approval=under_review next=compliance approver=44 tracking=pending tracking_rows=1 cases=0`；四场景最终均为 `records=2 cases=1 parties=2 same-case replay=true`。这些是本审查者独立运行结果，不是转述作者 `03-http-recheck.log`。

### 冻结绑定和剩余边界

17:46:21 核对：`3f97004` 的父提交正是已复核的 `7c5618f`，提交差异只有以下两文件；二者测试前、测试后、Kuhn `multistage/06-results-and-freeze.json` 与 Git blob 的 SHA-256 完全相同。`internal/` 工作树无改动；原 601 文件清单只有预期 service 一项改变，新增测试已实际提交，未遗漏在 ignore 中。

| 文件 | 最终 SHA-256 |
|---|---|
| internal/services/approval_conflict_integration_service.go | `7690825d8accd27d47132f78cd003332293f544727acd10ed00185f6f2e5373e` |
| internal/services/approval_multistage_postgres_http_test.go | `40d8de57765cb8030288061130f639010eef72de98251fe0bc322e6f1f837f84` |

`06-results-and-freeze.json` 的 SHA-256 为 `cb945d359dac9d031d5798ae0654283b86cf0745e38c978e1a1454728a4d374e`。作者另有 build/vet/diff-check 通过及四个既有 completion/inbox 回滚叶子通过，本轮未冒充独立重跑这些检查；`02-targeted-recheck.log` 整体失败仍按交接记录保留，没有因部分叶子通过而改判整组 PASS。

本轮没有新增发现需阻断 B5 交付；只更新本报告，没有源码编辑、浏览器操作、代理派发或 Git 提交。独立测试只写已授权本机测试库的随机私有 schema，并清理自身数据。下文为 `7c5618f` 已完成的 R1/B4/F5 与此前 B1/B2/B3 复核记录；其范围和证据边界保留，当前候选按本节绑定到 `3f97004`。

## 7c5618f 阶段判定与冻结范围（B5 之前）

**2026-11-07 专项独立复核通过，当前限定后端范围未发现剩余阻断。R1 已关闭；B4 bootstrap 与 F5 服务端模式 DTO 增量通过，可随候选 `7c5618f26ed5600de76e4ed94a682252d10a8aef` 冻结。** 此结论继承下文已完成的 B1/B2/B3 源码复核，只追加 R1、B4、F5 及直接调用链，不是全库或前端复审。

- 独立执行最终候选专项：**12 个顶层 / 40 个叶子 PASS，0 FAIL，0 SKIP，其中 20 个叶子实际使用隔离 PostgreSQL**；`-race -count=1`，services 包耗时 `20.360s`，终态时间 `2026-11-07T17:11:44.953653+08:00`。
- 17:12:03 核对：`bootstrap/25-final-source-after.json` 所列 **601 文件**全部存在，当前工作树内容、提交 `7c5618f` 的对应 Git blob、清单 SHA-256 三方一致；缺失/漂移均为 0。原被 ignore 的专项测试已实际包含在提交中，不再仅存在于作者工作目录。
- 该清单文件 SHA-256：`b484b1091555275a70d1fae36b760e05cce6211e38c8fba07e4eacc5658a0708`。新增 `internal/services/approval_snapshot_mode.go` 已纳入。
- 冻结范围：原 B1/B2/B3 已复核后端修复、R1 补充阶段双侧对象授权及 Inbox 四入口、B4 PostgreSQL bootstrap v16 与启动结构验证、F5 `submission_snapshot_required` 只读投影及相关测试。**没有授权或执行真实环境升级、部署；不把本结论升级为浏览器两链、真实律师试用或生产验收。** 主线程仍须对新构建执行既定浏览器验收；仅重启旧 QA 服务不代表已运行正常 bootstrap 升级。

## R1 关闭证据：新旧客户权限与四入口

审读实际实现后确认，修复没有简单取消客户一致性检查：

1. [approval_subject_authorization.go:29](../../../../../internal/handlers/approval_subject_authorization.go#L29) 仅对 `needs_revision` 且 active 指针存在的审批调用 [AuthorizeApprovalRevisionContext](../../../../../internal/services/approval_snapshot_authorization.go#L36)。后者核验审批参与资格、持久 intake 关联及申请人/接案所有者绑定，再分别核验当前 editable intake 和 active frozen 客户/案件。详情与 snapshot 共用该入口；其他状态、无 active 的 legacy 继续原严格路径。
2. [intake_supplement.go:232](../../../../../internal/services/intake_supplement.go#L232) 附近在 approval→intake 事务锁下调用同一补充授权，随后才校验 revision/return record、写回复。申请人不能因只拥有当前客户而保存旧轮次回复或重提；读侧修复也没有赋予审批人写补充的能力。
3. [inbox_workflow_action.go:241](../../../../../internal/services/inbox_workflow_action.go#L241) 只为有效 active 的 `needs_revision` 收集 frozen 对象，批量加载 client/check/case/隔离墙；[revisionVisible](../../../../../internal/services/inbox_workflow_action.go#L439) 分别核验当前 intake、旧冻结客户/案件/检测引用，仍拒绝错误所有者、串 intake、缺对象和无权对象。普通状态及 legacy 仍经过 `checkVisible` 的客户一致性门。
4. List 和 Stats 都使用授权后的 `visibleInbox`；GET 经 `projectOneInbox`，无权 workflow 项返回 404；[首页](../../../../../internal/handlers/demo_aggregate_handler.go#L580) 使用同一 `ListInboxItems(needs_action)` 的动作和 stats。没有四条独立的授权捷径，也没有读请求写回历史内容。

独立真实 PG/HTTP 用例先完成真实 PUT 客户 A→B，确认 intake=draft、approval=needs_revision、active 仍是旧轮次，然后在**尚未重提**时检查：

| 状态 | List / Stats / 首页 | Inbox GET | 详情 / default snapshot / explicit snapshot |
|---|---|---|---|
| A、B 都可访问 | 补充项/计数均为 1，动作和真实 intake/approval ID 一致 | 200，supplement | 均 200 |
| 只撤销冻结客户 A 权限 | 项目与计数归零 | 404 | 均 403 |
| 为 A 恢复白名单 | 同一项和计数恢复 | 200 | 均 200 |
| 只撤销当前客户 B 权限 | 项目与计数归零 | 404 | 均 403 |
| 为 B 恢复白名单 | 同一项和计数恢复 | 200 | 均 200 |

各读取阶段比较 fixture 12 张业务表，均无变化；律师、合伙人的详情/快照双侧授权均独立跑过。律师在任一客户撤权时 PUT 回复与 POST 重提均 403 且无写入；恢复后可实际保存回复并重提。Inbox 连续链最终实际 HTTP 重提、批准并关联正式案件，旧快照字节和持久 metadata 不变。检测/独立复核输入为持久 fixture，本次没有调用检测器或复核 UI。

额外独立边界共 20 叶子：9 种状态/绑定、10 种 frozen 引用拒绝、76 个不同补充审批的批量测试。76 项与单项同为 12 次查询，snapshot/client 各一次；不是逐条数据库授权。

## B4 专项：bootstrap 升级及启动验证

仅审查相对作者接手快照的 delta、其直接调用方及现有专项测试：

- [production_schema.go:16](../../../../../internal/database/production_schema.go#L16) 升至 v16，新表模型含可空 `VARCHAR(64)` claim；[intake_collaboration_schema.go:12](../../../../../internal/database/intake_collaboration_schema.go#L12) 在既有表上显式补 claim、tracking.intake_id 与全量单列唯一索引，未修改 000080/000081 原迁移。bootstrap 的加列、索引、验证及版本标记同一事务，重复槽错误不会清理原数据。
- [validateIntakeApprovalClaimContract](../../../../../internal/database/intake_collaboration_schema.go#L82) 验证 claim 类型/长度/可空性、intake_id 可空性、索引唯一/有效/就绪/立即生效/单列/无 predicate/无表达式及 NULL 兼容性。`IF NOT EXISTS` 没有将同名错误索引伪装为成功；[production_schema.go:470](../../../../../internal/database/production_schema.go#L470) 将该检查接入 bootstrap 与运行时共有 validator，`main.go` 启动门同时要求 v16。
- 独立执行 **11 个 PG 叶子**：空私有 schema 调用真实 `BootstrapProductionSchema`，建成 119 表，再只 seed 业务数据；实际首提经历 claim→冻结→唯一槽→审批待办，重复请求 409、业务表不变。没有用 AutoMigrate/手工运行时 DDL 补齐被测初始化缺陷。
- 已有 v15 测试保留原数据与进行中的 claim，正常升级到 v16、再次 bootstrap 幂等，并实际首提成功。真实唯一索引拒绝第二个相同非 NULL intake（SQLSTATE 23505），允许多个 NULL 槽。
- 缺 claim、短 claim、非空 claim、缺索引、普通索引、部分索引、错误列索引这七种损坏均被启动契约拒绝；遗留重复槽使整个升级回滚，claim 新列与 v16 标记不残留，重复数据未删除。

未发现本 delta 的高风险阻断。上述已有 v15 测试是空库正常 bootstrap 后恢复到相关历史缺陷形状，**不是任意历史数据库或真实环境迁移验收**；不扩大其证据边界。

## F5 专项：服务端快照模式 DTO

[approval_handler.go:246](../../../../../internal/handlers/approval_handler.go#L246) 在对象授权之后调用 [ApprovalRequiresSubmissionSnapshot](../../../../../internal/services/approval_snapshot_mode.go#L14)，`:271` 显式返回 boolean。只有确认无持久 intake、active、成案类型、内嵌及持久历史快照的普通审批才返回 false；任何查询/解析错误返回错误，不向前端给出可降级的 false。

独立 **6 个 PG 叶子**验证：普通 leave 为 false、snapshot 404 合法、原决定仍能成功；普通业务已有 legacy snapshot 为 true，原决定仍可成功；仅 check / 仅 tracking / 仅历史 snapshot 留有 intake 关联均为 true，即使顶部 type/metadata 看似普通，决定仍 409 且无写入。成案类型、内嵌快照亦为 true；注入 snapshot 读取故障时详情 500，不出现 false。DTO 读取不写表、不占槽；字段没有被接入权限放行或替代事务版本/唯一槽校验。

本次最终采用新增 helper 的版本，而非 `14-final-results.json` 对应的第一版 DTO。未发现新 DTO 导致旧单绕过成案守卫或普通审批被一律禁止。

## 独立运行记录及作者证据区分

最终成功命令（实际独立执行；复用编译缓存，`-count=1` 不复用测试结果）：

```sh
env GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache GOPROXY=off GOTOOLCHAIN=local \
  LAW_OA_COLLAB_PG_DSN='host=/private/tmp/lawoa-intake-collaboration-runtime/socket port=55437 user=mac dbname=lawoa_intake_mvp_test sslmode=disable' \
  go test -mod=readonly -race -p 1 ./internal/services \
  -run '^(TestInboxR1.*|TestBackendFinalReviewPostgresPendingClientChangeHTTP|TestApprovalBootstrapPostgres.*|TestApprovalSnapshotModePostgres.*)$' \
  -count=1 -timeout=180s -json
```

| 独立终态测试组 | 叶子结果 | 证据类型 |
|---|---:|---|
| TestInboxR1OnlyActiveRevisionMayChangeClient / FrozenReferencesRemainAuthorized / BatchAuthorizationAcrossRevisions | 20 PASS | 本地 SQLite/race |
| TestBackendFinalReviewPostgresPendingClientChangeHTTP + TestInboxR1PostgresClientEditBeforeResubmitHTTP | 3 PASS | realPG + Gin HTTP |
| TestApprovalBootstrapPostgresClaimAndSlotParity / V15UpgradeFirstHTTPSubmission / StartupRejectsBrokenClaimContract / DuplicateLegacySlotsFailAtomically / FirstHTTPSubmission | 11 PASS | realPG + 真实 bootstrap；两条真实 HTTP 首提 |
| TestApprovalSnapshotModePostgresHTTP / CaseCreationAndEmbeddedMode | 6 PASS | realPG；含实际 HTTP 决定及注入读取错误 |

读取及测试均由本审查者执行，没有浏览器、子代理或源码修改。测试只访问经源码核对的固定本机测试库，随机 `approval_collab_*` / `bootstrap_intake_*` schema，结束清理自己的数据；bootstrap fixture 先确认既有 pgcrypto，不迁移数据库级扩展。不是无写入测试，但写入仅限获准的临时测试数据；实际业务读取路径的无写入由状态断言验证。

保留过程边界：17:06 开始的首次本地 race 构建失败且同期 F5 文件发生变化，未进入用例，不作为候选证据；17:10 默认执行权限下 20 本地叶子通过，PG setup 的 14 个失败终态均为 `dial unix ... connect: operation not permitted`，尚未执行 PG 业务断言。按既有隔离验证授权获得本机 socket 执行权限后，相同最终专项在 17:11 完整通过，没有跳过、改测试或弱化断言。

作者结果另列：`inbox/r1/09-final-race.jsonl` 的 77 叶子/0 fail/0 skip 已按原终态核对；`bootstrap/21-final-race.jsonl` 与 `26-final-results.json` 的 300 叶子/0 fail/0 skip、race/vet/build/diff-check 通过已核对，作者将实际 PG 归类为 111 叶子。本审查不把该 300/111 记为独立执行，也未重复全库回归。

原红灯均保留：R1 `inbox/r1/01-first-red.log` 的真实 PUT 后 List 为空；B4 `bootstrap/02-first-http-red.log` 的真实首提失败；F5 `bootstrap/18-existing-snapshot-mode-first-red.log` 的已有历史快照错误 false。`bootstrap/29-final-preserved-red-sha256.json` 五份相对路径证据逐个核对无漂移，之前 `16-preserved-red-sha256.json` 的证据也未改写。setup/fixture 错误仍按各原日志分类，不算业务首红。

## 最终关键文件指纹

所有下列文件均包含在 `7c5618f` 与 601 文件冻结清单内，测试前后内容一致；全清单负责覆盖其余直接依赖和测试，不将提交 SHA 代替内容核对。

| 文件 | SHA-256 |
|---|---|
| internal/services/approval_snapshot_authorization.go | `105a66d0b10e82ac332ba663d5749a8afdb1bff453dccba01a2c8795b0614a66` |
| internal/handlers/approval_subject_authorization.go | `590705d027eb3c5ab5efab9211562cd4ceee186b835adbd9ec64fb8badff659b` |
| internal/services/intake_supplement.go | `267a053a36fcc92f463bd71514b022e757ef111e357a107f059a341048ba9349` |
| internal/services/inbox_workflow_action.go | `1f7880faa2b02e67725bf15bc6a461ea21332c82ec81f62df798ddc937771df9` |
| internal/database/production_schema.go | `bcfa04989c1a865f0a068e26cb8de92a4d98e97bf61cd5028a13fa55a8fcf51b` |
| internal/database/intake_collaboration_schema.go | `bbcd2b9522454fb7cbd93d82a581e24d00ef2afad3f7e0cb8c6c57fe28677d14` |
| internal/handlers/approval_handler.go | `15eab93c86d34ca71d1f4ec9d7844e49cc08ba99ecb4580cb8cf07ab18e6615e` |
| internal/services/approval_snapshot_mode.go | `c762942156e3076549dd407c295f53e690a46c084f13c45b79459931a0820dc5` |

## 以下为 16:36 交接前审查记录（保留历史，当前结论以上文为准）

2026-11-07 16:36 +08:00。只读审查工作树 `/private/tmp/lawoa-intake-collaboration-20261107`，基线/当前 HEAD 为 `a806831a7cef839c2b971cd4ab260cce77f42b8b`，实际对象含未提交、新增和 ignored 源码/测试。复核原 `backend-final-review.md` 的三个 finding，加入 Euclid 已交接的 B2；不扩展为全库或前端审计。

**16:36 当时结论：原 B1/B2/B3 的主要修复在实际源码中成立；仍有 1 项 P2 正常补充流程缺口（R1），因此当时未出最终 PASS。** 下列旧观察指纹不是最终冻结签字；R1 的当前关闭证据和最终绑定见报告顶部。

## R1 原始 P2：退回后改客户、尚未重提，合法补充上下文被拒，待办消失（已专项复核关闭）

**触发条件：** 接案审批已冻结客户 A / 检测 C1，随后退回 `needs_revision`；申请律师对 A、B 均有权限，按现有草稿接口将 intake 客户改为 B，尚未完成新一轮重提；旧审批仍指向 C1。此状态合法：`demo_aggregate_handler.go:2198` 附近授权新客户，`:2239` 附近使变更后的检测失效；`intake_supplement.go:42` 也明确允许要求补充 `client_id`。

**实际代码路径和结果：**

- [approval_subject_authorization.go:59](../../../../../internal/handlers/approval_subject_authorization.go#L59) 将旧 C1 的 `record.ClientID` 交给当前接案授权；[conflict_context.go:133](../../../../../internal/services/conflict_context.go#L133) 要求其等于当前 intake 客户 B，条件必定失败。详情 [approval_handler.go:244](../../../../../internal/handlers/approval_handler.go#L244) 和 snapshot GET 的 `canAccessApprovalSnapshot` 都先走这条路径，因此返回 403，尚未进入各自的 frozen 投影/快照对象授权。
- [inbox_workflow_action.go:383](../../../../../internal/services/inbox_workflow_action.go#L383) 同样比较旧检测客户与当前 intake 客户；`:411` 的 `approvalVisible` 因此为 false，`:441` 附近在生成 supplement action 前已排除该项。申请人的补充项从 List/Stats/单项 GET 及同源首页计数中被过滤。
- 新 [approval_snapshot_authorization.go:13](../../../../../internal/services/approval_snapshot_authorization.go#L13) 已明确承认这种合法换客户场景，`AuthorizeApprovalResubmission` 将当前草稿和旧 frozen 对象分别授权；但目前只用于重提 POST，读取路径与待办未同步。因此不能以重提服务测试成功证明刷新/恢复补充页面可用。

**业务影响：** 修改客户后刷新或从待办重新进入，不能取得退回要求与旧轮次上下文；直接持有全部参数的重提服务可能仍成功，但正常恢复入口中断。这里不是原 B2 的“重提成功后 top-level C2 与 metadata C1 不一致”，两者应分别验收。

**最小修复边界：** 在受控审批的 `needs_revision` 补充读取场景，将审批参与资格、当前 editable intake 对象权限、所选 frozen 轮次客户/案件权限分别核验；Inbox 保持批量对象加载并执行等价权限。不要全局删除检测/客户一致性约束，不要仅因申请人身份放开历史客户隔离墙，也不要写回旧 snapshot/metadata。现有重提 helper 可作为授权语义依据，不能直接用于所有读者，因为它限定申请人。

**需要补的证据：** 真实 handler 下先成功保存客户 B，停在未重提状态，验证详情、default/explicit snapshot、supplement List/Stats/GET/home；对旧 frozen 客户 A 和当前客户 B 分别撤权应拒绝，再授权应恢复；随后实际重提及新轮次审批仍成功。现有 `ApprovalBackendFinalReviewHistoricalClientFixture` 在约 `approval_backend_final_review_postgres_test.go:198` 已调用成功重提，之后才测历史客户撤权，不能覆盖此中间状态。

**证据边界：** 这是当前源码确定的条件分支结论，已核对草稿写入口；审查者没有新建 fixture、执行 PG 写入或使用浏览器复现。已在本会话告知用户，并通过可用的会话消息工具成功发送给 Kuhn `01a1b4c2-7276-73d1-8414-942429355917`，提醒协调 Euclid 的 Inbox 所有权及主线程冻结影响。Kuhn 随后明确确认问题成立、影响生产源码冻结，R3 暂不能作为两链最终验收候选；他将补首红并修读取授权，Inbox 留给 Euclid。此确认不等于修复或独立动态复现完成。未替作者修改源码。

## 原三项修复核对

| 原 finding | 当前实际源码判断 | 证据与限制 |
|---|---|---|
| B1 旧 conflict_approval 两案旁路 | 核心阻断已接入入口与消费边界，不再依赖旧 metadata 存在 intake_id；暂未发现原旁路残留 | `approval_intake_identity.go` 从持久 snapshot、check、association、tracking 归并 intake；旧单仍存续但无 active 时拒绝成案。PG/HTTP 为作者执行，最终交接未到。 |
| B2 重提后 Inbox 消失 | 原 C1→C2 问题已修；批量 active 投影先于上下文加载及授权，无效 active 不回退 | 实际代码及新增测试已审读；独立本地回归通过。B2 自身源码和四测试与 Euclid 的 `13-pg-recheck-result.json` 相符。R1 是另一个尚未重提的状态缺口。 |
| B3 历史快照对象授权 | 原 applicant/current approver 直接放行已移除；当前资格及每个历史版本的对象权限均验证 | 历史版本名单按权限过滤，隐藏不可访问的 previous snapshot ID；作者用例覆盖律师/合伙人/冲突复核人撤权与恢复。R1 仍影响合法读取，不把其误判为权限旁路已恢复。 |

### B1：旧单拒绝、既有旧案占用、正常新链

- [demo_aggregate_handler.go:2883](../../../../../internal/handlers/demo_aggregate_handler.go#L2883)：intake 来源的备用 CreateConflictApproval 在建审批/快照前返回 `INTAKE_CONFLICT_REVIEW_REQUIRED`，指向独立冲突复核和原接案受控提交；未将尚未完成独立复核的旧入口强行 Freeze。
- [approval_intake_case_guard.go:57](../../../../../internal/services/approval_intake_case_guard.go#L57)：从持久来源解析 intake 后锁同一 intake 行；未成案时要求有效 active 冻结及本审批唯一槽。`:87` 封住已存在的 nil-pointer 旧单，不能借旧 legacy 兼容分支自动成案。
- [approval_workflow_transactions.go:230](../../../../../internal/services/approval_workflow_transactions.go#L230) 在旧 nil-pointer 返回前调用新守卫；[approval_conflict_integration_service.go:500](../../../../../internal/services/approval_conflict_integration_service.go#L500) 在成案 claim 前调用守卫，普通决定、集成决定、自动成案和 retry 不只依赖 handler 的新入口限制。
- [approval_intake_case_guard.go:18](../../../../../internal/services/approval_intake_case_guard.go#L18) 查询全部同 intake 历史检测所关联的未删除案件及 tracking 关联，不只看当前 C2；[intake_submission_snapshot.go:377](../../../../../internal/services/intake_submission_snapshot.go#L377) 在冻结前拒绝已成案 intake。因此旧 C1 已成案，即使后来准备了 C2，也不能新提交另一成案审批；已经存在的新审批在消费时也被拒。
- 同一旧案的恢复仅在 `approved`、唯一案件、审批/已完成 tracking 指向同一 case 时允许；`:510` 的早期恢复分支先验证原案，不重新 claim 新案。真正无 intake 关联的旧业务保留原流程。
- 审读作者 PG 测试的实际断言：两种先后顺序及三次并发；普通决定/auto/retry/仅剩历史 snapshot 引用；旧案完整标志和过期标志重放；旧案后换检测提交、已经提交新审批再消费；真实非 intake legacy 兼容；失败前后表状态相同、旧单无决定记录、唯一槽和案件均只有一条。HTTP 测试验证原旧单各入口返回具体 409 且无写入。未把“测试名称存在”作为通过依据。

### B2：active 批量读取及同源投影

- [inbox_workflow_action.go:89](../../../../../internal/services/inbox_workflow_action.go#L89) 验证 active row 的归属、类型、内外 snapshot/approval/intake/check ID、轮次与同轮 task ID；仅替换内存副本 metadata 与 check 字段，旧持久 metadata 不变。
- `:175` 批量取 active rows，`:191` 先投影再收集 check/intake/case；不逐审批调用数据库版 resolver。缺失、串属或损坏 active 排除，schema/查询错误返回错误；无 active 的 legacy 保留来源一致性校验。
- 14 类无效/越权 active、缺表、legacy 不一致，以及 76 个不同审批的批量查询测试均已读断言；批量断言为快照查询 1 次、总查询 11 次，与单审批相同。独立本地测试覆盖这些断言。
- 作者 PG 用例在实际退回和重提后验证 C2、旧 metadata/旧 snapshot 不变，读操作比较 12 张表未改变；再退回的 supplement ID、再次重提、实际集成成案为一案一结果任务。HTTP 用例用真实 Gin Inbox/CommandCenter handler 核对两角色 List/Stats/GET/home；不是浏览器证据。
- 四个 ignored 测试已显式读取并确认 ignore 命中：`inbox_workflow_contract_test.go`、`inbox_active_snapshot_test.go`、`inbox_workflow_postgres_test.go`、`inbox_workflow_postgres_http_test.go`。按用户要求由主线程处理纳入候选，本审查不暂存、不改 ignore。

### B3：历史客户与隔离墙撤权

- [demo_aggregate_handler.go:3310](../../../../../internal/handlers/demo_aggregate_handler.go#L3310) 附近统一查询审批参与资格并检查当前业务对象，已无申请人/当前审批人直接 return true；历史审批人也不因此免除对象授权。
- [approval_snapshot_authorization.go:52](../../../../../internal/services/approval_snapshot_authorization.go#L52) 从所选 frozen 内容及实际持久 check 收集客户/案件，校验客户存在、客户隔离墙与案件权限；缺受控上下文拒绝。当前客户可访问不能授权旧客户。快照默认读取、explicit 历史读取、history 清单均走该检查。
- 历史撤权测试实际先完成换客户并重提，再对旧客户案件启墙：current snapshot 200、old snapshot 403、history 无旧版本 ID；恢复白名单后 old snapshot 200，读取前后状态不变。另测当前案件撤权时 default/explicit 两入口都 403。以上逐角色断言已审读，原 B3 权限绕过修复有依据；不覆盖 R1。

## 实际验证及证据边界

**审查者独立执行，仅本地临时数据，无 PG/外部连接：**

```sh
env GOCACHE=/private/tmp/lawoa-intake-independent-review-go-cache GOPROXY=off GOTOOLCHAIN=local \
  go test -mod=readonly ./internal/services ./internal/handlers \
  -run '^(TestInbox(Active|Legacy|Workflow|Approval|Supplement|Authorization|MoreThan|Rejects|Reassignment|Reconcile|Trusted|Conflict|Projection|Public)|TestCoreReview|TestProjectApproval|TestApprovalSubmissionResolver|TestApprovalSubmissionPreservesBusinessRole|TestIntakeSupplement)' \
  -count=1
```

结果：exit 0；services `2.565s`，handlers `2.077s`。`git diff --check -- internal` exit 0，仅涵盖已跟踪差异，不冒充 untracked 检查。没有新写测试或运行审查者 PG 用例；共享树仍有作者并发写测试，因此这次执行不是最终候选哈希绑定。

**只读核对的作者日志：**

- `inbox/b2/09-final-inbox-verification.jsonl`：解析实际 terminal 事件，25 顶层、33 子用例 pass，无 fail/skip；两个 package pass。
- `inbox/b2/12-pg-recheck.jsonl`：两个 PG 顶层及两个 HTTP 角色子用例 pass，无 fail/skip；与 `13-pg-recheck-result.json` 对照。B2 六个源码/测试文件当前指纹一致；部分共享 B1/B3 依赖后续变化，因此不将 Euclid 16:15 的整组依赖当作当前最终候选。
- 最新可见 `backend-final-fixes/23-final-race.jsonl`：独立解析终态，260 叶子用例 pass、0 fail、0 skip，四个 package pass；B1/B3 共 31 个顶层/子用例终态均 pass，含 B2 的 PG/HTTP 回归。`28-final-results.json` 记录 race/vet/build/diff-check 均 exit 0；与实际 race 日志相符。16:35 前后只读比较 `27-final-source-after.json` 所列文件，未发现漂移。
- 上一轮 `12-final-race.jsonl` 中 B1/B3 专项通过，不代表那次扩大回归整体通过；旧失败和跳过不覆盖为成功。最终使用作者新一轮完整结果，仍明确为 **author / realPG**，不是本审查者独立执行，也不是 browser 或生产验收。

本轮只新增本报告，没有改源码、前端、原审查报告、数据库或外部环境，没有使用浏览器、派子代理、暂存或提交。作者正式交接、R1 修复复核和最终 hash 绑定仍待完成。

## 观察指纹（非最终冻结）

读取时间 `2026-11-07T16:32:47+08:00`。这里只保存辨认本次结论所需的文件；不把 HEAD 等同于未提交候选内容。

| 文件 | SHA-256 |
|---|---|
| internal/services/approval_intake_case_guard.go | `aca8d2ada77afcb9a3d6174d3c0a3080e231f612854a230cc78bdcb682e6f39f` |
| internal/services/approval_intake_identity.go | `34d6af9f6fbfd423b19dbeac8e46e9c26440124c9927bc3017315ba382d7328e` |
| internal/services/approval_snapshot_authorization.go | `7ba1e53130ac219d6def6987fc22a1ae6c28480fe8262bf7d3acbbb844a7ff1a` |
| internal/handlers/approval_subject_authorization.go | `301df2d649e94b1d3dc1792cb8c29855ae49cfb2d7454dfbe1502b71e14be490` |
| internal/handlers/approval_handler.go | `8121329c35bbe8b6d82e228084cc98c4ff55a48d96b97e20b07dd693c1435937` |
| internal/handlers/demo_aggregate_handler.go | `c1a3b91190d4a80d0cb8bd1798fb9a7cb6b84e1cecd9aff7ec4f66e75b0e987f` |
| internal/services/inbox_workflow_action.go | `b9d800edaec9b38fbc8919f0ee79b7e32e51c1a52a8dc44bd1ac570d8d6ad590` |
