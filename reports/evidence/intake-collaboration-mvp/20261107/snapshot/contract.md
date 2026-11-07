# N1/N3 helper 契约（2026-11-07）

独占：intake_collaboration.go、intake_submission_snapshot.go、intake_supplement.go、对应新增测试、000082迁移。不得回退其他写入者改动。

- `ResolveApprovalSubmission(db, approval) (map[string]interface{},error)` 返回快照根，其中 metadata 为成案同形数据；无 pointer 的 intake 历史数据必须与持久快照一致。不使用 created_at 最新覆盖。
- `ApplyApprovalSubmission(db, approval) error` 只投影内存。调用方保存前恢复原 Metadata；通用 UPDATE 不得持久化该投影。
- `FreezeIntakeSubmissionInTx(tx,approval,userID,req) error` caller 持审批锁；helper 再锁 intake；caller 随后才占 tracking。写新快照、设置内存 pointer/conflict 字段；不保存审批、不改状态、不创建 inbox。任何 error 外层必须 rollback。无 intake 且无 pointer legacy 审批 no-op。
- `ValidateResubmissionReplayInTx(tx,approval,req)(bool,error)`，与主代理最终签名一致。必须在 owner/subject 授权、审批锁之后，freeze/状态改变之前调用。true 不创建快照、不转换状态、不发待办。
- `NormalizeSupplementRequirements(input map[string]interface{},reason string)(string,error)`。input 仅接受 schema_version=1 和 supplement_requests。1—20条，item_id/source 为服务端生成，拒绝客户端赋值。非空 reason。identity_verification 说明统一为“需核验该主体身份”，不保存输入的身份详细说明。
- `ReadSupplementRequests(record)([]models.SupplementRequest,error)`。历史纯文字记录以 `legacy-<recordID>` 的只读虚拟项投影，source=legacy_text，不写旧 records。
- `SaveIntakeSupplementResponsesInTx(tx,intakeID,approvalID,userID,expectedRevision,returnRecordID,responses)(string,error)` 自行锁 approval -> intake。回复数组为全量替换，允许部分项保存；重提要求所有 required 项齐备。仅更新 supplement_return_record_id、supplement_responses、draft_revision、updated_at，保留事实版本和 status。

DTO `ResubmitApprovalRequest` Go字段 Responses，JSON字段 **supplement_responses**，另含 expected_intake_revision、expected_submission_snapshot_id、return_record_id、revision_note。SupplementResponse={item_id,response}。

field_key 白名单：client_id,title,description,case_type,business_area,sub_area,lawyer_id,opponent_name,opponent_identity_number,parties,party_name,party_role,party_identity。party_* 必须 party_local_id，且重提检验该主体仍存在。前端按此映射；无任意字段路径。

新增错误 `*services.IntakeCollaborationError` 含 Code/Message/Status；handler 需 errors.As 对应返回 400/403/409，而不是统一500。其他复用 ConflictReviewerError。依赖缺表原错误向外返回，不默许继续。

普通 intake handler 必须拒绝受控 metadata 键直接写入；补正接口只从顶层 supplement_responses/return_record_id/approval_id 分派给 helper。主代理负责 current object authorization（含 client ethical wall）与动作版本检查。

迁移 down 在存在 intake_submission 时明确拒绝，避免降级丢弃已产生轮次的解释能力。隔离库 up/down/up 在生成新数据前演练，生成数据后测试拒绝降级。

初次编译受到其他代理尚未完成的 inbox API 编译错误影响（CompletionMode/ArchiveInboxItem），未把空匹配或该失败报为测试通过。

## 协调观察（同轮共享文件只读）

- 主代理 intake handler 受控键名单务必包含 `supplement_return_record_id`（普通metadata写入不能覆盖）。
- 初建 `CreateIntegratedApproval` 在 CreateApproval 后若先 INSERT tracking 再 Submit/Freeze，会先持 tracking 锁后 intake，与重提顺序不一致。建议创建审批后先明确锁 intake，再插 tracking，之后 Submit/Freeze 在同事务复用该 intake 锁；或把 tracking 搬到 Freeze 之后。
- 000082 已添加 SQL UPDATE/DELETE 触发器：新 intake_submission 快照及绑定 submission_snapshot_id 的决定均不可改删；up/down/up + 有新轮次 down 拒绝已在首轮 PG 测试通过。
- PG 首次失败仅为测试使用了入库前 JSON 字符串与 jsonb 规范化读回字符串比较。已调整为入库后读取原字节作为基线，仍严格比对前后持久内容不变；未放松业务断言。
