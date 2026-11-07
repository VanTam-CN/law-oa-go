# C06 最小缺口补证与 Boyle 复核交接

基线与当前产品候选：`3f970040f5e40b8016c18a070b5aa1345d2b2eec`。本轮只新增一个 Inbox 测试文件并保存证据，产品源码、原 acceptance-matrix.md、tasks 不变；无 commit/push、无真实环境操作。新增测试仅访问固定本机隔离PG库的随机schema。

## 结论

300项既有日志**没有完整覆盖本次两处请求**：助理已有服务层PG拒绝，但无补充/重提直接HTTP；核查岗已有详情/快照撤权HTTP，但无Inbox List/Stats/GET/home四入口联动。现新增2个实际PG/HTTP叶子，`-race -count=1`：**2 PASS、0 FAIL、0 SKIP，exit 0**，18:09:09结束。它们是作者补证，独立复核待Boyle；不把300/111改成302/113，也不混入独立40/20。

撤回前稿“全角色×全部字段×全部入口”的额外要求。主方案规定的是代表性六类角色、直接API负向及不泄漏受限内容；本轮以既有角色证据加两个实际缺口闭合该证据映射，不要求未规定的全笛卡尔积。C01全部字段、C05核查UI、C08记录配对仍归主线程。

已读取R5 A/B浏览器记录：候选3f97004，最终分别case3/case4，中间节点200/under_review/未成案，末节点200/approved/已成案。复杂主体SQL关系按主线程通知登记，由主线程汇入相关验收项；本补证没有重跑或接管浏览器/SQL验收，也不继续把新候选完整A/B链列为缺失。

## 既有300证据核对

下表每项均有bootstrap最终日志的pass事件；精确行号保存在 [05-results-and-audit.json](05-results-and-audit.json) 的existing_300_pass_refs。不是仅从测试名推断覆盖。

| 已有用例 | 实际覆盖 | 本次判断 |
|---|---|---|
| TestIntakeSubmissionPGControlledResponsePermissionsAndRollback/assistant | 直接调用SaveIntakeSupplementResponsesInTx，拒绝助理；PG | **不是HTTP**，本次补 |
| TestDenyIntakeWriteForPlainUser/assistant_allowed | 草稿写角色helper允许助理 | 不代表助理可保存律师补充/重提 |
| TestBackendFinalReviewPostgresSnapshotRevokedContext/conflict_officer | 核查岗当前/default与explicit快照403 | 不覆盖Inbox四入口，本次补 |
| TestBackendFinalReviewPostgresHistoricalSnapshotClientRevoked/conflict_officer | 核查岗旧客户历史快照撤权/恢复，受限历史ID不泄漏 | 已覆盖该历史边界，不重写/重跑 |
| TestInboxR1PostgresClientEditBeforeResubmitHTTP | 律师A/B客户分别撤权的List/Stats/GET/home | 角色是律师，不冒充核查岗 |
| TestInboxB2PostgresHTTPListStatsGetAndHome/lawyer、/partner | 律师、业务审批岗四入口动作/计数 | 保留既有PG证据 |
| TestIntakeWriteDeniesOtherLawyerViaObjectScope | 律师乙写律师甲接案HTTP403、无写入 | 保留既有负向 |
| TestIntakeWriteDeniesPlainUserAndConflictOfficerBeforeObjectScope | 普通用户、纯核查岗写律师接案HTTP403、无写入 | 保留既有负向 |
| TestConflictOfficerSeesOnlyAssignedConflictApproval | 纯核查岗按实际指派收敛审批可见范围 | 保留既有handler证据 |

## 新测试与断言

新增文件：[inbox_c06_postgres_http_test.go](../../../../../../internal/services/inbox_c06_postgres_http_test.go)。两项都使用实际Gin handler、实际授权/投影/业务事务；只通过测试middleware注入身份，没有模拟权限返回值。复用既有PG私有schema fixture（AutoMigrate + 协议迁移），**不是bootstrap、JWT登录或浏览器测试**。

### TestInboxC06PostgresHTTP/assistant_controlled_write

- 实际数据库用户44/assistant，真实律师退回与有效当前补充请求。
- 助理List/Stats/home返回200但不含目标审批ID、intake ID及标题；List为空/total0，目标Inbox GET404。
- PUT补充与POST原审批重提均403；错误响应不包含被测审批标题、退回记录ID和虚构身份；每次与fixture的12张业务表全文快照比较均不变。
- 阳性对照：同一有效补充payload由真实申请律师执行200；取得新revision再重提200，状态resubmitted且active指针变为新轮次。排除“请求无效/端点不可用导致拒绝”的假阳性。
- 实际终态：[04-http-race.jsonl:16](04-http-race.jsonl#L16)。

### TestInboxC06PostgresHTTP/conflict_officer_revoked_context

- 新建可验证case上下文的pending check及有效独立指派；不删除/改写旧已完成复核和冻结快照。真实ReconcileConflictInboxInTx产生核查待办。
- 核查岗33权限有效：List/Stats/home为1，GET200；同一任务、真实check字符串ID、review_conflict动作四入口一致。
- 启用对象隔离墙而不关闭任务：List/Stats/home为0，GET404。完整响应不含被测私有标题、客户名、身份/意见哨兵与check ID；阳性GET明确包含任务中已种入的标题/身份/意见哨兵。
- 为同一核查岗恢复白名单后，原同一任务ID再次List/Stats/home为1、GET200、review_conflict。三个阶段均比较12张业务表；读取不改变历史/任务，存储任务IsCompleted始终false。
- 这是**对象权限撤销**，不是声称覆盖所有岗位变更、指派撤销或身份字段的排列组合；转交等边界继续引用既有用例。
- 实际终态：[04-http-race.jsonl:26](04-http-race.jsonl#L26)。

## 原始运行和保留失败

| 文件 | 分类 |
|---|---|
| 02-http-race.jsonl / 02-run-result.json | 本机Unix socket权限拒绝；两个setup失败，尚未进入业务，不算产品首红 |
| 03-http-race.jsonl / 03-run-result.json | 助理实际通过；核查fixture把subjectCaseId写成数字，违反既有字符串契约，reconcile拒绝。仅修测试输入为fmt.Sprint(caseID)，未放宽断言/改产品 |
| 04-http-race.jsonl / 04-run-result.json | 最终两项实际PG/race通过，0 fail/skip；包耗时4.086s |
| 05-format-only.diff | 最终测试后gofmt只删去一处多余缩进tab，不改测试token或断言；被测和最终文件hash分别保留 |
| 01-source-before.json / 05-results-and-audit.json | 323个受跟踪产品Go文件及go.mod/go.sum前后无漂移；新增测试最终hash与既有/新增pass行索引 |

执行命令：
```sh
env GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache GOPROXY=off GOTOOLCHAIN=local \
LAW_OA_COLLAB_PG_DSN='host=/private/tmp/lawoa-intake-collaboration-runtime/socket port=55437 user=mac dbname=lawoa_intake_mvp_test sslmode=disable' \
go test -mod=readonly -race -p 1 ./internal/services -run '^TestInboxC06PostgresHTTP$' -count=1 -timeout=120s -json
```

最终新增测试SHA-256：`fdeda3a62b9beaa237180892faced5f357ba89c97d190d1da972900e2c4316bb`。本次format检查及git diff --check通过；没有扩大回归范围。该新文件被.gitignore:282的`*_test.go`忽略，**尚未纳入候选**，由主线程整理提交时显式纳入；本次不暂存/提交。

## 给 Boyle 的限定交接

仅复核上述**一个新测试文件的两项断言**及其直接fixture/日志，不重开B1–B5或产品实现审查。重点检查：

1. 助理请求是真实handler、有效DTO，403后无写入；申请律师阳性对照确实成功。
2. 核查岗动作在撤权前真实可办；撤权后四入口与计数收敛、受限哨兵不输出；恢复同一任务，排除删任务/手工完成制造“不可见”。
3. PG隔离、两项pass计数、fixture修正及纯格式差异的证据边界准确；没有用mock/SQLite代替PG。

建议独立报告只写本目录`BOYLE-REVIEW.md`；不改测试/产品，不代主线程填写最终C01–C18。作者结论不自动变为独立通过。当前工具未提供Boyle内部代理入口；本文件及最终回复作为主线程转交的明确材料，不声称已向未定位的会话发送消息。
