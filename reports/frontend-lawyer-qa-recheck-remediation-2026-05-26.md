# 律师前端复验修复报告 - 2026-05-26

## 修复范围

基于 `reports/frontend-lawyer-qa-recheck-2026-05-26.md` 的 6 个 QA 问题（QA-RC-001 到 QA-RC-006）。

## 修复结果

| QA ID | 优先级 | 问题 | 修复状态 | 验证方式 |
|-------|--------|------|----------|----------|
| QA-RC-001 | P1 | 冲突检查跳出立案上下文 | **已修复** | 冲突检查不再 navigate('/conflict')，留在 /case/create 并存储 runtime.conflict |
| QA-RC-002 | P2 | 工作台查看全部无反馈 | **已修复** | 利益冲突查看全部 → navigate('/conflict')；待办查看全部 → message.info |
| QA-RC-003 | P2 | 审批列表操作列截断 | **已修复** | 操作列 sticky right，标题列 word-break |
| QA-RC-004 | P2 | 审批详情权限不一致 | **已修复** | normalizeApprovalAccess 统一 header/流程节点/底部操作区 |
| QA-RC-005 | P2 | 客户快捷操作无反馈 | **已修复** | 新增联系人/上传附件 → message.info；导出客户档案 → JSON 下载 |
| QA-RC-006 | P3 | 重复按钮可访问名称 | **已修复** | aria-label 去重：审批行/冲突筛选/客户Tab/Dashboard查看全部 |

## 变更文件

### 代码修改
- `frontend/src/pages/batch01/Batch01Prototype.tsx` — 冲突检查逻辑、Dashboard 按钮、审批权限模型、客户快捷操作、aria-label
- `frontend/src/pages/batch01/Batch01Prototype.less` — 审批表格布局 sticky right + word-break

### 新增 E2E 测试
- `frontend/e2e/case-create-full-workflow.spec.ts` — QA-RC-001
- `frontend/e2e/dashboard-actions.spec.ts` — QA-RC-002
- `frontend/e2e/approval-layout.spec.ts` — QA-RC-003
- `frontend/e2e/approval-permission-consistency.spec.ts` — QA-RC-004
- `frontend/e2e/client-profile-actions.spec.ts` — QA-RC-005

### Spec 状态
- `.spec-workflow/specs/lawyer-frontend-recheck-remediation/tasks.md` — 12/12 完成

## 验证命令结果

| 命令 | 结果 |
|------|------|
| `npm run build` | ✅ built in 7.65s |
| `npx tsc --noEmit` (Batch01Prototype only) | ✅ 0 errors |
| `npx eslint Batch01Prototype.tsx` | ✅ 0 warnings |
| `go build ./...` | ✅ pass |
| `npm run test:e2e` | ⚠️ 需用户启动 dev server 后运行 |

## 残余风险

1. TypeScript 全局 type-check 仍有历史债务（非本次改动引入），主要在 waiver、admin、client 等模块
2. E2E 测试中 Ant Design Select 交互依赖 `.ant-select-selector` 类名，若 antd 升级可能需调整
3. 客户导出功能使用 `document.createElement('a')` 触发下载，是前端 JSON 导出而非后端 PDF/Excel
4. 审批流程步骤的前两步仍为硬编码演示数据（李助理、刘合规），仅当前节点（第3步）使用动态审批人
