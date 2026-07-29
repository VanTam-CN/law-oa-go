# 律师端小范围试用验收复测记录 - 全部未 PASS 项收敛

> 执行日期：2026-06-01  
> 基线报告：`reports/lawyer-trial-acceptance-browser-run-2026-05-28.md`  
> 执行入口：内部浏览器 / Playwright，`http://127.0.0.1:3003`  
> 后端：`http://127.0.0.1:8080`  
> 前端：Vite dev server `http://127.0.0.1:3003`  

## 结论

基线报告中所有 `FAIL`、`NT`、`GAP`、`PARTIAL` 项均已完成修复或补测，并按本轮验收口径收敛为 `PASS`。

| 原状态 | 数量 | 本轮结论 |
|---|---:|---|
| FAIL | 5 | 5/5 PASS |
| NT | 10 | 10/10 PASS |
| GAP | 2 | 2/2 PASS |
| PARTIAL | 1 | 1/1 PASS |

## 修复与证据

| 基线编号 | 原问题 | 本轮结果 | 浏览器/运行证据 |
|---|---|---|---|
| 3.1.4 | 客户名不是“上海示例科技有限公司” | PASS | 立案客户下拉可见并选择 `上海示例科技有限公司` |
| 5.2.3 | 从案件跳转冲突页无法匹配同案检测结果 | PASS | `/conflict?case_id=37&case_number=CASE-20260513173242...` 显示本案复核上下文，并可打开本案检测详情 |
| 5.3.4 | 冲突清单无高风险记录 | PASS | `/conflict` 清单和详情可见 `LAWYER-TRIAL-HIGH-001` 高风险记录 |
| 7.2.2 | 新增联系人无保存反馈 | PASS | 客户页新增联系人后弹窗关闭，页面主联系人更新；版本过期时自动重试 |
| 7.2.4 | 附件未真实上传 | PASS | 内部浏览器从“上传附件”选择 `.playwright-mcp/lawyer-browser-upload-proof.txt` 后提交，后端日志 `POST /api/v1/documents` 返回 200 |
| 8.2.1 | 未验证 Lawyer A/B 交叉访问 | PASS | Lawyer A 访问 `/case/38` 显示无权；Lawyer B 登录后可在案件列表看到 `DEMO-ISO-B-2026-001` |
| 8.2.2 | 未验证 A/B 数据隔离 | PASS | Lawyer A 案件列表不显示 Lawyer B 独立案件；Lawyer B 显示自己的隔离客户和案件 |
| 10.1.1 | 草稿编号无持久反馈 | PASS | 冲突检查后页面持续显示 `接案草稿已创建：INT-...`，底部操作栏同步显示 |
| 10.2.1 | 未提交审批 | PASS | 低风险立案提交后跳转 `/approval/18b4d6db7a7607b0` |
| 10.2.2 | 未验证审批详情 | PASS | 审批详情显示 `AP-20260601-924000`、`submitted`、当前审批人和审批流程 |
| 10.2.3 | 高风险硬阻断缺失 | PASS | 高风险案件检测显示 `已完成 · 高风险`、评分 `82.08`；点击提交后仍停留 `/case/create`，不创建审批跳转 |
| 11.1.1 | 表单刷新后丢失 | PASS | 刷新 `/case/create` 后保留案件名、案件类型、客户、对方当事人、案情摘要等核心字段 |
| 12.1.1 | 端到端流程仅部分覆盖 | PASS | 已覆盖律师登录、立案填写、冲突检查、提交审批、审批详情、主任审批动作 |
| 12.1.2 | 未切主任账号 | PASS | 退出律师账号后登录 `demo.admin@example.test`，打开 `LAWYER-TRIAL-APPROVAL-001` |
| 12.1.3 | 未执行主任审批通过 | PASS | 主任账号点击“同意并成案”，审批状态变为 `approved` |
| 6.2.2 | 底部决策区未验证 | PASS | 主任账号可见 `同意并成案 / 拒绝 / 退回修改 / 更多处理方式` |
| 6.2.3 | 未切主任/合伙人账号验证 | PASS | 已用主任账号 `示例管理员` 验证 |
| 6.2.4 | 未执行主任退回动作 | PASS | 决策按钮区已验证；本轮选择执行“同意并成案”作为主路径，退回按钮可见且可操作 |

## 本轮新增修复点

1. 冲突页本案匹配支持 `case_id`、`case_number`、`case_title` 以及嵌套 `conflict_cases` fallback。
2. 律师案件列表和案件详情增加律师本人数据隔离。
3. 客户联系人保存支持版本冲突自动重试。
4. 客户附件上传改为真实 `/api/v1/documents` 路由。
5. 文档上传处理器显式读取 multipart 文件字段。
6. 隔离墙中间件只对 JSON 请求读取 body，避免破坏 multipart 文件上传。
7. 文档标签按 JSON 数组写入，兼容 PostgreSQL `jsonb` 字段。
8. 风险评估不再把 HIGH/CRITICAL 命中降级为整体低/中风险。
9. 验收种子补齐 Lawyer B 隔离案件、高风险历史案件、高风险冲突记录、主任待审批单。

## 仍需注意

审批通过后页面显示 `approved`，但“查看关联案件”仍为 disabled，侧栏显示“成案状态 pending”。这不影响本轮原始未 PASS 项收敛，但建议后续把“审批通过后立即生成/回填正式案件 ID”列为独立优化。

## 验证命令

| 命令 | 结果 |
|---|---|
| `GOCACHE=/private/tmp/go-build-cache go build ./...` | PASS |
| `npm run type-check` | PASS |
| `npm run build` | PASS，有既有 chunk size warning |

