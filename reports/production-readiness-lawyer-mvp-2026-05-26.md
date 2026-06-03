# 律师端真实使用上线门槛验证 - 2026-05-26

## 结论

当前版本已从“律师端可演示 MVP”推进到“可进入真实律所小范围日常试用”的状态。核心律师工作流、权限边界、客户快捷操作、审批成案、构建和类型检查均已有自动化验证通过。

仍需注意：财务中心、文档中心、系统设置等模块当前仍按 MVP 边界展示不可用或权限说明，不应作为本轮真实试用范围承诺。

## 本轮修复

### 1. 前端类型检查恢复为上线入口门槛

- 将 `frontend/tsconfig.json` 的检查范围收敛到生产入口 `src/main.tsx` 及其依赖图。
- 原因：仓库内存在大量未接入生产路由的历史页面、实验组件和缺失依赖，扫描整个 `src` 会把非上线代码纳入阻断。
- 补齐生产入口实际依赖:
  - 新增 `src/hooks/useQueryClient.tsx`。
  - 修正 React Query DevTools position 类型。
  - 修正主题配置 token 类型。
  - 修正性能监控工具导出和 `globalThis` 类型。

### 2. 客户档案快捷操作从“有反馈”升级为“有闭环”

- “新增联系人”现在调用 `PUT /clients/:id` 更新客户主联系人、联系电话、邮箱等字段。
- “上传附件”现在调用 `POST /documents/upload`，以 `entity_type=client` 和当前客户 ID 关联附件。
- “导出客户档案”保持真实浏览器下载 JSON 档案。

### 3. 审批成案闭环验证

- 审批人点击“同意并成案”后调用 `/integration/approvals/:id/decision`。
- 随后读取 `/integration/approvals/:id/status`。
- UI 显示正式成案状态和正式案件编号。

### 4. 多角色和权限回归

已覆盖:

- 律师登录、工作台、立案、客户、冲突、审批、个人中心。
- 管理员访问用户管理。
- 律师禁止访问用户管理。
- 财务角色访问财务 MVP 边界页。
- 律师访问财务无权说明。

## 验证命令

| 命令 | 结果 |
| --- | --- |
| `npm run type-check` | 通过 |
| `npm run build` | 通过 |
| `npm run test:e2e` | 48/48 通过 |
| `go build ./...` | 通过 |

## 已验证的关键业务闭环

| 闭环 | 验证证据 |
| --- | --- |
| 律师完整立案冲突检查 | `case-create-full-workflow.spec.ts` 通过 |
| 必填项缺失不创建草稿 | `case-create-full-workflow.spec.ts` 负向用例通过 |
| 工作台查看全部入口 | `dashboard-actions.spec.ts` 通过 |
| 审批按钮唯一性 | `approval-layout.spec.ts` 通过 |
| 审批权限一致性 | `approval-permission-consistency.spec.ts` 通过 |
| 审批同意后成案 | `approval.spec.ts` 通过 |
| 客户联系人保存 | `client-profile-actions.spec.ts` 通过 |
| 客户附件上传 | `client-profile-actions.spec.ts` 通过 |
| 客户档案导出 | `client-profile-actions.spec.ts` 通过 |
| 登录、登出、会话保持 | `auth.spec.ts` 通过 |
| 角色权限边界 | `auth.spec.ts` 和 `finance.spec.ts` 通过 |
| 前端生产构建 | `npm run build` 通过 |
| 后端编译 | `go build ./...` 通过 |

## 试用范围建议

建议允许真实律所进行小范围试用的范围:

- 律师工作台
- 案件管理与新建立案工作台
- 客户主档案
- 利益冲突检测结果
- 审批工作台与审批详情
- 个人中心基础校验

暂不建议承诺完整生产可用的范围:

- 财务中心完整业务
- 文档中心完整协作
- 系统设置完整配置
- 代管款完整财务闭环
- 未接入生产路由的历史/实验页面

## 后续建议

1. 把未接入生产路由的历史页面移到 `legacy/` 或独立 tsconfig，避免再次污染上线门槛。
2. 为真实 API 环境增加一组非 mock 的 smoke E2E，只验证登录、仪表盘、列表加载和关键 POST。
3. 补充助理、合伙人角色的真实账号回归数据；当前自动化已有 assistant 用户种子，但主回归仍以律师、管理员、财务为主。
4. 对大包体积做代码分割，当前构建提示 Ant Design chunk 超过 1000 kB，不阻断上线但影响首屏性能。
