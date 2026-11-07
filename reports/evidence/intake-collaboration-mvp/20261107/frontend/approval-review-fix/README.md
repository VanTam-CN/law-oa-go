# R4：F5/F6 与多阶段决定文案冻结交接

生产前端源码已冻结于 **2026-11-07T09:05:06.532Z（17:05 CST）**，工作目录 /private/tmp/lawoa-intake-collaboration-20261107，基线 a806831a7cef839c2b971cd4ab260cce77f42b8b。入口 **index-fDMF1BW8.js**。本轮相对 R3 仅四个前端文件变化，无依赖、锁文件、后端、审批级数或接口流程变更。

- [source-manifest.json](source-manifest.json)：28 个累计前端文件、14 个关联测试文件、5 个构建输入的 SHA256；delta_from_r3 精确列出本轮四文件。
- [build-manifest.json](build-manifest.json)：7 个 dist 文件、命令、日志和证据边界。
- [frontend-source.tar.gz](frontend-source.tar.gz)，SHA256 0caadc64bb0d059c45bb55106369d723d62029137968d557f1128216c69a8f2e。
- [frontend-dist.tar.gz](frontend-dist.tar.gz)，SHA256 700670964c653995638de2586e69459dc9aa7029f3134fb18718281653879d5f。
- [archive-verification.json](archive-verification.json)：已逐文件解包比对源码与产物字节哈希。

## Planck 专项复查入口

1. **F5**：[approvalPresentation.ts](../../../../../../frontend/src/services/approvalPresentation.ts) 中 allowsSnapshotFreeDecision 要求服务端明确 false + 无可见矛盾关联 + snapshot 404。Batch01Prototype.tsx:8391 处理普通合法无快照；8476 仅此模式省略 snapshot token，8527 仍强制原始 action_version。能力缺失/true、active 矛盾、旧 intake、成案类型、403/409/500/网络错误继续阻断。已有快照及显式历史门禁保留。服务端字段已在共享源码出现，详见 [backend-capability-contract.md](backend-capability-contract.md)。
2. **多阶段决定**：Batch01Prototype.tsx:8487 用重新读取的 approval.status/current_stage/current_approver_name 决定反馈；若刷新详情失败，才使用决定响应的审批对象，不从 HTTP 200 推断最终 approved。中间节点只提示本阶段通过和后续节点；最终 approved 且真实 case_id 才提示已成案。普通 leave 不显示成案承诺；按钮、弹窗、decision_comments 均改为同意本阶段。成功后关闭确认弹窗。仍调用原决定端点，没有改审批级数。
3. **F6**：Batch01Prototype.tsx:8935 分别显示冻结自动检索结果、当前可读/本次提交冻结/所选历史冻结独立复核、明确覆盖状态。零命中仅写“未发现匹配证据”。覆盖非 COMPLETE 仍阻止同意；历史展示自身冻结 review，不混入最新 review。
4. [ApprovalCollaborationPage.test.tsx](../../../../../../frontend/src/pages/batch01/__tests__/ApprovalCollaborationPage.test.tsx) 共 22 项，保留旧当前、历史只读、结构化退回与精确纳秒令牌；新增兼容拒绝边界、两阶段反馈、旧 intake 409 无降级、F6 三类显示。

## 本轮作者验证

| 检查 | 结果 | 证据 |
| --- | --- | --- |
| 审批页面 + 接案契约回归 | 2 文件，32/32 通过 | [regression.log](regression.log) |
| TypeScript | exit 0 | [type-check.log](type-check.log) |
| 四个改动文件 lint | exit 0 | [lint.log](lint.log) |
| Vite 构建 | exit 0 | [build.log](build.log) |
| frontend diff whitespace | exit 0 | [diff-check.log](diff-check.log) |

测试命令（仓库根目录）：

`npm --prefix frontend test -- --runInBand --coverage=false --runTestsByPath src/pages/batch01/__tests__/ApprovalCollaborationPage.test.tsx src/pages/batch01/__tests__/IntakeCollaborationContracts.test.ts`

另运行 npm --prefix frontend run type-check / build；lint 在 frontend 使用已有 node_modules/.bin/eslint 对上述四文件执行。未 install。首红 [first-red.log](first-red.log) 保留：首版 20 项中 8 失败；包括新文案与 F5/F6 未实现，后续修正了按钮带 icon 的 accessible name 查询，最终 22 项审批测试 + 10 项契约测试通过。未重跑旧全套 75 项。

## 未宣告通过的层级

本轮只运行 mock API/JSDOM、静态检查和构建；未触碰用户浏览器、服务或真实数据。服务端能力仅按当前代码只读核对，运行服务须重建并验证对应字段。F5/F6 关闭须由 Planck 专项复查；两条新样本真实全链、双节点实际审批、真实旧接案 409、C16 等由主线程继续，**本报告不标 C16 PASS**。构建仍有原 AntD chunk 大小警告，构建成功。

主线程可固定此 dist 验证：普通 leave 无快照可处理；CASE_APPROVAL 第一节点不最终 approved、不承诺已成案；最终节点按真实状态反馈；接案/历史/读取错误门禁保持。旧 R3 及用户报告已通过的移动导航与恢复证据保留，本轮未扩展这些实现。
