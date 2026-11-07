# N2/N5 前端交接（2026-11-07）

状态：前端源码已冻结，固定构建可交给主代理。作者本地检查通过；完整 MVP 验收尚未完成。

- 工作目录：/private/tmp/lawoa-intake-collaboration-20261107
- 基线 HEAD：a806831a7cef839c2b971cd4ab260cce77f42b8b；分支 codex/intake-collaboration-mvp。交接的是未提交变更，不冒充新候选提交。
- 只写 frontend 和本证据目录；未编辑后端、未提交/推送、未安装依赖、未启动服务或调用外部服务、未委派其他代理。后端并行改动保留。
- node_modules 沿用既有主树 symlink；两树 frontend/package-lock.json SHA256 相同，详见 source-manifest.json。

## 固定构建

可直接提供 /private/tmp/lawoa-intake-collaboration-20261107/frontend/dist 的静态预览；不使用开发 HMR。另存 frontend-dist.tar.gz（内层为 dist/），可解压到独立目录作为固定快照。后续源码或 build 变化不会改变此归档。

归档 SHA256：90f94d92d467773cce9e335c7175ad8a07762fab2d7844aeafc9dc6b2c2cad16

source-manifest.json 列出本次 22 个前端修改/新增文件 SHA256；build-manifest.json 列出构建产物逐文件 SHA256 和实际检查结果。

## 已交付与文件所有权

| 能力 | 实现位置（均属本切片） |
| --- | --- |
| 客户共用创建弹窗、未确定响应保留幂等 key、防双击、成功使用响应 ID 回填、取消清身份 | frontend/src/components/client/ClientCreateDialog.tsx；frontend/src/services/client.ts；Batch01Prototype.tsx 的 ClientMasterProfile / CaseIntakeWorkbench |
| 授权远程搜索 + page_size=50 分页、同名按 ID 区分、旧搜索丢弃、新建/选中客户固定显示 | frontend/src/components/client/ClientSearchSelect.tsx |
| 精确 intake_id 恢复、四态保存、保存中后续编辑不标已保存、409 保留输入、账号隔离、旧身份 fingerprint 缓存移除 | frontend/src/pages/batch01/Batch01Prototype.tsx；intake/draftState.ts |
| 离开三选项 + beforeunload；真实入口改用 data router 支持 useBlocker | frontend/src/pages/batch01/intake/DraftNavigationGuard.tsx；frontend/src/main.tsx；frontend/src/index.tsx |
| needs_revision + 双 ID URL 校验；稳定 local_id 字段定位；逐项回复单独保存；原审批版本化重提 | frontend/src/pages/batch01/intake/SupplementChecklist.tsx；frontend/src/services/intakeSupplement.ts；Batch01Prototype.tsx |
| 结构化退回 imposed_requirements 对象；raw action_version；历史快照选择、当前标记、旧版及读取失败禁用决定 | frontend/src/pages/batch01/Batch01Prototype.tsx 的 ApprovalDecisionFlow；frontend/src/services/approval.ts |
| 工作台授权待办/等待分页、最近草稿先于统计；inbox 同源 action、详情重读、归档与完成区分 | frontend/src/components/workflow/PersonalWorkQueue.tsx；WorkflowActions.tsx；frontend/src/pages/inbox/InboxList.tsx |
| 请求保留状态码、业务 code/context，不记录身份请求正文 | frontend/src/services/collaborationRequest.ts |

主树源码只作只读参考，未将其未提交实现复制覆盖本候选。

## 与主代理待闭合的两项投影

1. 当前 internal/services/inbox_workflow_action.go 对 intake/supplement 输出 intake 资源 ID，没有审批关联 ID。前端 WorkflowAction 已支持显式 approval_id:string；有此字段时映射 /case/create?intake_id=...&approval_id=...&mode=revision。未提供时保持禁用，不从 source_id、标题或 workflow_group 猜测。另一可用契约是服务端输出 approval/supplement，以审批 ID 进入详情再跳补正。主代理需选择并完成服务端投影，再走真实待办链。
2. 旧文本退回的规范化条目由服务端产生。前端已优先读取审批详情顶层 return_record_id、supplement_requests、return_reason；保留其 item_id 原值。新协议原始 records JSON 同样可读。当前后端详情尚未见上述顶层规范化字段，旧文本只有文字意见，前端不会自己生成 legacy-* ID；不可据此宣称旧文本重提已打通。

## 已对齐的请求

- 回复 PUT /case-intakes/:id 只发 approval_id、return_record_id、expected_revision、supplement_responses；不混事实或 metadata。返回 revision 立即用于后续请求。
- 事实保存沿用业务 metadata 白名单。保存回复不靠前端伪造事实版本；版本效应由后端受控服务承担。
- 决定 imposed_requirements 是对象；expected_action_version 直接使用 detail.action_version，未经 Date 转换。
- resubmit 使用原 approval_id，发送 expected_intake_revision / expected_submission_snapshot_id / return_record_id / supplement_responses / revision_note。
- snapshot 查询使用 snapshot_id；选择器只开放 intake_submission 协议快照。历史字段不混入当前申请事实，历史模式及快照读取失败禁用决定。
- 工作台与待办均消费服务端 workflow_action。列表计数/分页来自 /inbox 的授权投影；不从第一页估算全量。

## 实际验证

执行目录：/private/tmp/lawoa-intake-collaboration-20261107/frontend

| 命令 / 范围 | 实际结果 | 证据 |
| --- | --- | --- |
| npm run type-check | exit 0，最终源码 | typecheck-handoff.log |
| npm run lint | exit 0，最终源码 | lint-handoff.log |
| npm run build | exit 0，最终源码 | build-handoff.log |
| npm test -- --runInBand --coverage=false --runTestsByPath（下列 9 文件） | 47/47，exit 0 | jest-final.log |
| 最后补充账号异步隔离后重跑 IntakeCollaborationPage.test.tsx | 5/5，exit 0 | intake-handoff.log |
| git diff --check -- frontend | exit 0 | 本地终端确认 |

9 文件：

- src/components/client/__tests__/ClientCollaboration.test.tsx
- src/pages/batch01/__tests__/IntakeCollaborationContracts.test.ts
- src/pages/batch01/__tests__/ApprovalCollaborationPage.test.tsx
- src/pages/batch01/__tests__/IntakeCollaborationPage.test.tsx
- src/pages/inbox/__tests__/InboxWorkflowActions.test.tsx
- src/pages/batch01/__tests__/IntakeDraftRestore.test.ts
- src/pages/batch01/__tests__/IntakePartyRoleRoundTrip.test.ts
- src/pages/batch01/__tests__/CaseIntakeConflictAction.test.ts
- src/pages/batch01/__tests__/DashboardPermission.test.tsx

新页面测试 mock fetch 与身份，未访问真实 API；data-router 测试补 Node 内置 Request/Response/Headers，fetch 仍由测试 mock。测试日志保留 jsdom 的伪元素布局/React act 警告，不能替代真实布局与键盘验收。构建仍有体积较大 chunk 的提示，无新增依赖或通过调高阈值消除提示。

首次失败记录保留：page-jest-1.log / pages-jest-2.log（测试路由目标及 jsdom 缺 Request，已补真实路由测试环境）；lint-1.log / lint-2.log（本次格式问题，已修）；combined-jest-1.log 包含三个误写的旧测试路径、实际 21 项执行通过，修正命令后最终 47/47。未删除或弱化失败断言。

## 交给主代理的下一步

1. 闭合上面两项服务端投影，使用固定 dist 预览与 127.0.0.1:18587 虚构数据后端。
2. 两条真实、同对象全链各自独立执行：客户→接案→复核→原审批退回两项→待办直达→事实/回复分开保存→必要复核→原审批重提→唯一成案。绑定每条链的对象 ID，不能拼接成功片段。
3. 390px 视口验证客户搜索/分页、弹窗取消/失败、补正定位、固定底栏和 Tab/Shift+Tab/Enter/Escape；确认离开阻拦不吞焦点。
4. 真实 API 核查补正回复刷新恢复、事实变更失效、历史只读和 action_version 并发拒绝；作者 mock 不能证明后端事务/权限。
5. 独立 review 随后进行。当前不宣称 C01—C18 全量通过、真实律师试用通过或生产可用；production_ready=false。
