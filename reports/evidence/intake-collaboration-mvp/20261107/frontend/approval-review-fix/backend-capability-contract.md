# F5 详情能力投影对接

冻结前已只读观察到服务端实现，前端无后端写入：

- GET /approvals/:id 授权后调用 ResolveApprovalIntakeID，解析持久 check/tracking/snapshot 关联；异常返回错误，不按 metadata.intake_id 缺失降级。
- 当前详情输出 submission_snapshot_required:boolean；已有接案关联或 active 指针时为 true。服务端位置：[approval_handler.go](../../../../../../internal/handlers/approval_handler.go)，能力投影测试源码：[approval_snapshot_mode_postgres_http_test.go](../../../../../../internal/services/approval_snapshot_mode_postgres_http_test.go)。本实施者未执行后端测试或请求运行中的 API。
- 前端只在该字段严格等于 false、无可见接案/成案/active 关联矛盾且 snapshot 返回 404 时兼容。能力字段缺失，或 snapshot 403/409/500/传输失败，继续阻断。
- 普通无 snapshot 决定省略 expected_submission_snapshot_id，仍发送原始 expected_action_version，并遵守 available_actions。读到合法旧当前 snapshot 时仍携带服务端核验的 viewed_snapshot_id；明确历史只读。
- 不改决定端点、审批级数、后端权限/成案门禁。旧 intake 缺槽收到 409 后不降级、不重建、不重试其他接口。
- 主线程已确认将契约完整要点转交 Kuhn；真实运行服务需使用含上述字段的后端构建。
