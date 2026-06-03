# 律师端前端 QA 修复回归报告 - 2026-05-25

## 范围

基于 Spec Workflow `lawyer-frontend-qa-remediation` 修复 `reports/frontend-lawyer-qa-2026-05-25.md` 中 QA-001 到 QA-012。

## 已修复问题

- QA-001: 工作台“新建立案”直达 `/case/create`。
- QA-002: 工作台全局搜索增加结果/空态反馈。
- QA-003: 新建立案正式流程移除样例预填数据。
- QA-004: 文档材料归档入口改为当前流程内说明，不再中断立案。
- QA-005: 非当前审批人的律师账号不再看到审批决策按钮。
- QA-006: 冲突检测表格不再撑破主内容区。
- QA-007: 客户主档案不再横向溢出。
- QA-008: 新增客户入口打开新增客户表单。
- QA-009: 帮助中心入口展示帮助说明。
- QA-010: 修改密码空表单展示字段级校验。
- QA-011: 财务无权访问说明补充所需权限。
- QA-012: 通知为空时隐藏无效“全部已读”动作。

## 验证结果

| Command | Result | Notes |
| --- | --- | --- |
| `npm run test:e2e` | PASS | 37 passed |
| `npm run build` | PASS | Vite chunk size warning only |
| `npm run lint` | PASS | No ESLint errors |
| `go build ./...` | PASS | Backend build passed |
| `npm run type-check` | FAIL | Existing broad TypeScript debt outside this remediation |

## Type-check 说明

`npm run type-check` 仍失败，主要为既有问题：缺少 `@mui/material`、`date-fns`、旧组件接口类型不匹配、测试 mock 重复导出、旧页面服务类型不一致等。本轮改动已通过前端生产构建、Lint 和 E2E 验证。

