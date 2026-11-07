# Inbox R1 修复交接：换客户但尚未重提

日期：2026-11-07；目录 `/private/tmp/lawoa-intake-collaboration-20261107`；HEAD `a806831a7cef839c2b971cd4ab260cce77f42b8b`，含未提交和被 ignore 的测试。

R1 作者修复及必要验证完成。它覆盖退回后编辑客户、尚未重提的合法中间状态，不能以 B2「成功重提后」的通过代替。本切片不宣布 C01–C18 最终通过，`production_ready=false`；最终共同冻结仍由主线程汇合 R1 与 Kuhn bootstrap parity。

## 变更范围与契约

生产源码仅修改 [inbox_workflow_action.go](../../../../../../internal/services/inbox_workflow_action.go)。`inbox_service.go`、shared DTO、业务调用点、bootstrap、router 及其他作者文件未改。

1. 继续使用 B2 的批量 active 快照校验及同轮内存投影，保留历史 metadata、快照字节与原来源一致性检查。
2. **仅有效 active pointer + needs_revision** 进入 `revisionVisible`：确认当前 intake/申请人绑定及可见性，再分别验证冻结客户、案件、检测引用。当前客户 B 不必等于冻结客户 A；任何一侧无权限、对象缺失、引用无效，均排除该审批的 Inbox 项及计数，单项 GET 返回原有 404。
3. 冻结对象范围与 Kuhn `AuthorizeApprovalSnapshotContext` 的字段来源对齐：root/metadata/client/conflict_record/search_parameters/case_creation_config 中的客户、案件与检测引用；补正对象授权与 `AuthorizeApprovalRevisionContext` 的双侧语义一致。实现只使用批量加载的对象与隔离墙结果，没有逐项调用数据库版授权函数，没有 N+1。
4. 普通 submitted/resubmitted/under_review/approved 与无 active 的 legacy 不获得该例外，仍保留 check/intake 客户一致性门。补充动作仍是 `resource_type=intake`、真实 intake `resource_id` 和显式字符串 `approval_id`，没有新 DTO 字段。

测试改动：

- [inbox_workflow_contract_test.go](../../../../../../internal/services/inbox_workflow_contract_test.go)：原 fixture 补实际 clients 表与客户行，以验证冻结客户存在性，原断言保留。
- [inbox_revision_authorization_test.go](../../../../../../internal/services/inbox_revision_authorization_test.go)：9 类状态/绑定边界、10 类冻结引用拒绝、76 个不同补正审批的批量证明。
- [inbox_revision_postgres_test.go](../../../../../../internal/services/inbox_revision_postgres_test.go)：隔离 PG fixture 与下一轮合法检测/复核输入。
- [inbox_revision_postgres_http_test.go](../../../../../../internal/services/inbox_revision_postgres_http_test.go)：真实 handler PUT 换客户，停在未重提状态验证读侧、撤权与恢复，再真实 HTTP 重提/决定/成案。

## 从失败到验证

| 记录 | 实际观察 |
|---|---|
| [01-first-red.log](01-first-red.log) | 真实 PG + HTTP PUT 客户 A→B 成功，intake=draft、approval=needs_revision、active 未变；补充 List 应有 1 项却为 0。R1 行为首红，未用 B2 失败替代。 |
| [02-inbox-regression.jsonl](02-inbox-regression.jsonl) | 初次修复后原 Inbox 与新 R1 PG 主链通过。 |
| [04-final-regression.jsonl](04-final-regression.jsonl) | 新负向 fixture 的第二个客户空邮箱重复，触发既有唯一约束；3 个顶层测试 setup 失败。PG R1 联测仍通过。保留此中间失败，不删约束或弱化断言。 |
| [07-fixture-recheck.jsonl](07-fixture-recheck.jsonl) | 给两个虚构客户设置不同 test.invalid 邮箱后，3 个顶层测试及 19 个边界子用例通过。 |
| [09-final-race.jsonl](09-final-race.jsonl) / [10-final-results.json](10-final-results.json) | 最终限定 race 联测：30 个顶层、54 个子用例，按叶子去重为 **77 pass / 0 fail / 0 skip**；services 与 handlers 均通过，16:58:10 +08:00 完成。其中 4 个顶层、6 个叶子实际使用隔离 PostgreSQL，包括 B2 与 Kuhn 的双角色补正授权用例。 |
| [06-vet.log](06-vet.log) | `go vet ./internal/services ./internal/handlers` exit 0。日志为空；后续只修复测试邮箱，没有继续修改生产源码。 |

最新 R1 PG 用例实际完成以下连续链，未拼接不同对象：

- 原审批正式提交并退回，真实 `PUT /case-intakes/:id` 把客户 A 改为 B。断言 intake=draft、approval=needs_revision、active 指针仍为旧轮次、原 metadata 不变；此时尚无重提。
- 申请人的 supplement List=1、Pagination.Total=1、Stats.needs_action=1；GET 返回正确 intake/approval 两个 ID，home todo=1，动作与计数同源。并联校验详情/default snapshot/explicit snapshot 200，与 Kuhn 双侧授权保持一致。
- 给冻结客户 A 新建隔离墙：List/Stats/home 全部归零，Inbox GET 404，审批详情/两类 snapshot 403；为申请人恢复白名单后，同一个补充项重新可见。
- 对当前可编辑客户 B 独立执行同样撤权/恢复，结果一致。各读阶段都比较 fixture 12 张表，无审批、快照或业务状态写回。
- 供应新客户的合法检测/独立复核 fixture 输入，仍在重提前验证补充项可见；随后真实 HTTP resubmit 得到新 active，审批人获得两侧授权后真实 HTTP approve 成功成案。旧快照字节、旧持久 metadata 仍不变。

批量证据：76 个不同 needs_revision 审批与单审批均为 **12 次查询**，其中 snapshot 1 次、clients 1 次；B2 的 76 个 submitted 审批仍为 11 次、snapshot 1 次。无逐审批数据库授权查询。

测试属于 **author**；上述 PG 路径同时属于 **realPG**。Gin httptest 不是 browser；新检测/独立复核为合法持久化 fixture 输入，没有声称实际检测器或复核页面已验收。本切片未启动服务器，没有操作真实业务环境，也未独立审查自己的修复。

复验命令（既有本机测试库，每项随机 schema，结束仅清理自己的 schema）：

```sh
env GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache GOPROXY=off \
  LAW_OA_COLLAB_PG_DSN='host=/private/tmp/lawoa-intake-collaboration-runtime/socket port=55437 user=mac dbname=lawoa_intake_mvp_test sslmode=disable' \
  go test -race -p 1 ./internal/services ./internal/handlers \
  -run '^(TestInbox|TestBackendFinalReviewPostgresPendingClientChangeHTTP)' -count=1 -json
```

## 共同冻结与文档同步

- [08 源码运行前哈希](08-final-source-before.json) 与 [10 运行后哈希](10-final-results.json) 覆盖 internal Go 文件及 go.mod/go.sum。运行期间唯一观察到变化的是 Kuhn 的 `approval_bootstrap_postgres_http_test.go`；本切片生产代码、测试及直接依赖均一致。该 bootstrap 测试不在本次选择集中，不将本结果描述成 bootstrap parity 通过。
- [contract-map](../../contract-map.md) 已替换旧「Inbox 依赖 metadata 镜像」说明，明确批量 active 投影、补正双上下文授权及真实状态来源；初版静态基线与历史哈希保留原时点。
- [acceptance-matrix](../../acceptance-matrix.md) 已链接 B2/R1 新报告；18 条主方案条件及最终判定不变，没有改 tasks。
- 没有已知阻断 R1 交接的剩余实现项；独立复核、Kuhn bootstrap parity、最终共同哈希及新浏览器候选仍由主线程汇合。已停止本切片源码写入，不扩展其他源码范围。
- 新增三份 `inbox_revision_*_test.go` 与原 contract test 被现有 `*_test.go` ignore 规则匹配；冻结时须显式纳入。没有修改 ignore、暂存、commit、push 或调用子代理。
