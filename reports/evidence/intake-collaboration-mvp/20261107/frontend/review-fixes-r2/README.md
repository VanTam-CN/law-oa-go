# N2/N5 前端修复交接 R2

状态：作者修复与本地验证完成，源码及固定构建冻结；独立复审和真实浏览器验收待完成。C16 未通过，不改变验收矩阵。

工作目录：/private/tmp/lawoa-intake-collaboration-20261107。基线 a806831a7cef839c2b971cd4ab260cce77f42b8b，分支 codex/intake-collaboration-mvp；仍是未提交变更。冻结时间 2026-11-07T08:10:46.811Z。仅写 frontend 与前端证据；未编辑后端、安装依赖、改锁、启动服务、访问真实 API/外部服务、操作浏览器、派子代理或提交/推送。现有 PG/server 保持。

## 当前固定产物

- 静态目录：/private/tmp/lawoa-intake-collaboration-20261107/frontend/dist；实际入口 frontend/index.html → src/index.tsx。
- 主 JS：index-Cm1X_6UB.js。它取代 index-CxoHfApU.js；后者因错误使用 review.check_id 已由现场复验否定恢复链，保留归档只供复盘。
- [构建归档](frontend-dist.tar.gz)，内层 dist/。SHA256：78651531095b8b917c051a28df1bf904acc6fe593a71eeb400a220552fd71d1b。
- [源码归档](frontend-source.tar.gz)：基线上 25 个修改/新增前端文件的完整内容，不含依赖及其他人的后端改动。SHA256：ec81e39c57d56e5592c0d0a7ede4f8606eb22986895d97c8ebc603a33b6ca835。
- [源码清单](source-manifest.json) 绑定 25 个文件、12 个回归测试及构建输入；[构建清单](build-manifest.json) 绑定 7 个静态文件；[归档校验](integrity-check.json) 已逐文件比较 tar 内字节与当前源码/dist。
- node_modules 继续使用原 symlink，两树 package-lock 哈希一致。未 install。

## 确认项与修复对应

| 项目 | 实际变更与回归证据 | 源码位置 |
| --- | --- | --- |
| F1：旧当前快照误判历史 | 默认 snapshot reader 的 viewed_snapshot_id 是经过服务端来源校验的持久行 ID。null active 不再直接等同历史；缺少 ID、SNAP_* 或 active/viewed 矛盾仍阻断。当前旧版本使用最新独立复核；显式选择历史仍只读。版本选择不再排除 conflict_approval，测试覆盖新旧混合历史。 | [ApprovalDecisionFlow](../../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L8247)；[currentSubmissionSnapshotID](../../../../../../frontend/src/services/intakeSupplement.ts#L109) |
| F2：旧补正发送空版本 | 补正恢复先读默认 snapshot，重提使用该持久 UUID，不取内嵌 SNAP_*。旧文字 item_id 原样保留；不确定响应重试仍使用相同版本和回复。服务端旧审批 409 封禁不绕过。 | [补正恢复](../../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L4522)；IntakeCollaborationPage.test.tsx |
| F3：迟到恢复覆盖/跨账号缓存 | URL、草稿列表、本地缓存关联恢复、409 重新加载统一校验账号、恢复代际、读取期间编辑签名及组件是否存活。逆序请求、账号切换、本地恢复及 409 重载期间编辑均有页面回归。 | [恢复边界](../../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L5092)；IntakeRecoveryRegression.test.tsx |
| F4：空白入口保留旧对象 | 账号和 URL 定义独立编辑会话。具体 A → 空白入口清空 runtime、intake ID、revision；回归继续选客户/律师并保存，确认 POST 新草稿而非 PUT A。 | [CaseIntakeWorkbench](../../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L4177) |
| 有效检测/独立复核恢复 | 仅 conflict_ready 且有服务端 conflict_check_id/facts_version 的草稿恢复当前绑定结果；验证 task 的 checkId/intakeId/clientId 及 review.checkId。复核 DTO 为直接 camelCase 对象。恢复形态形成内存比较基线，身份原文不写缓存；真实事实编辑立即过期。未改动草稿保存不再 PUT 事实，避免使既有复核失效。 | [conflictRecovery](../../../../../../frontend/src/pages/batch01/intake/conflictRecovery.ts)；[草稿状态](../../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L5015) |
| 长详情被 Sidebar 截获 | 冲突详情 portal zIndex=1050，高于 fixed 导航 1030；豁免子弹窗 1060。正文独立滚动，dvh 限高，窄屏复核指定换行。测试用真实 Modal 和编译后的真实 LESS 检查层级、滚动及下拉选择。尚无真实 390px 命中/键盘证明。 | [详情弹窗](../../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L7500)；[局部样式](../../../../../../frontend/src/pages/batch01/Batch01Prototype.less#L830) |
| N4 结果提醒 | completed 显示“已处理／结果提醒”，展示 workflow_status 的待成案/成案失败子状态。服务端 view_only 即使带 blocked_code 仍可进入审批详情；可写动作依旧禁用，详情自行重读权限。 | [WorkflowActions](../../../../../../frontend/src/components/workflow/WorkflowActions.tsx#L18)；InboxList.tsx；InboxWorkflowActions.test.tsx |
| 接案独立复核入口 | 按服务端 risk_queue.intake_id 识别接案，隐藏旧冲突审批入口并引导本页直接指定/复核。自动 CLEAR 也提供指定与复核，不把机器 CLEAR 当成独立复核完成。申请人回原草稿提交，前端新提交也要求已读到独立复核记录。 | [接案来源及复核入口](../../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L6499)；ConflictDetailModal.test.tsx |

## 验证及首红

执行目录 /private/tmp/lawoa-intake-collaboration-20261107/frontend。完整命令见 [verification-commands.txt](verification-commands.txt)。

| 检查 | 结果 | 日志 |
| --- | --- | --- |
| npm run type-check | exit 0 | typecheck-final.log |
| npm run lint | exit 0 | lint-final.log |
| npm run build | exit 0 | build-final.log |
| 12 文件定向 Jest | 75/75，exit 0 | jest-final.log |
| git diff --check -- frontend | exit 0 | diff-check.log |

现场确认 /review 的 checkId 后，先只把 mock 改为实际 camelCase DTO，旧 guard 确实复现“按钮仍禁用”：real-contract-first-red.log（1 失败）。修正 guard 后，75 项组合回归包含同一 checkId 可恢复、不同 checkId 拒绝、task 接案矛盾拒绝、复核读取失败拒绝；不通过删除编号校验来消除错误。task 的 intakeId 由后端 toTaskResponse 从 SearchParameters 提取，前端不回退到按客户猜最近历史。

之前 overlay 首红仍保存在 ../modal-overlay-fix/first-red.log。review-fixes/regressions-1.log 的两个失败分别是测试标签及图标参与 accessible name 的选择器错误，已校正，业务断言保留。JSDOM 伪元素/React act 警告与 Vite 大 chunk 提示仍保留，未改阈值或屏蔽日志。所有网络请求为 mock；这些结果是作者测试，不能冒充独立复审或真实全链。

## 主线程立即接手

1. 使用当前静态 dist 或此归档，reload A 草稿核验有效 check/review 恢复。A 未编辑时应显示当前复核结论并允许提交；不要为修复恢复而重跑检测。A 旧审批 d08a805d 的 409 属后端保护，不通过前端绕过。
2. C01/C02/C03/C04/C17 可按当前固定构建继续。检查 1200×1041 长详情复核选择不被侧栏截获，以及 390px 的滚动、Tab/Shift+Tab/Enter/Escape。作者没有接管浏览器。
3. 后端 B1/B2/B3 完成后，用两个全新样本分别跑同对象完整链；不要拼接 A 的旧协议成功片段。新接案冲突详情直接指定/独立复核，然后回草稿提交唯一接案审批。
4. 请独立 reviewer 复核 F1—F4 及本轮现场 DTO 修复。服务端 approval_id 和旧补正规范化 detail 两项早期投影缺口已有实现，不再列为当前未实现阻碍；真实 API 仍按主线程证据判断。

本次交接无已知未修复前端阻断；最新 R2 的真实 A 恢复、两条完整新链及 390px 键盘尚待主线程确认。C16 不标 PASS，C17 不标 PASS，production_ready=false。交接后停止前端编辑，等待具体复验/复审反馈。
