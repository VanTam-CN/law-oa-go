## 1. Intake Conflict-Check Guardrail

- [ ] 1.1 Read `frontend/src/pages/batch01/Batch01Prototype.tsx` and identify the exact local state fields used for matter title, client, counterparty, responsible lawyer, case type, business domain, and subdomain.
- [ ] 1.2 Add a single helper that returns missing required fields for the conflict-check action. The helper must be pure and easy to unit-test or assert through E2E.
- [ ] 1.3 Wire the "运行利益冲突检查" and equivalent footer "保存并进行利益冲突检查" actions to run the helper before any draft-create, draft-update, or conflict-check API call.
- [ ] 1.4 When required fields are missing, keep the user on the relevant intake context, show field-level errors where possible, show one top-level message listing missing fields, and leave conflict status as "未检测".
- [ ] 1.5 Ensure no intake draft ID, intake number, conflict task ID, or success toast is created when preflight validation fails.
- [ ] 1.6 When all required fields are present, preserve the fixed behavior from the previous remediation: remain inside `/case/create`, mark conflict status completed or in progress according to API result, and allow entry to team/fees.

## 2. Dashboard Action Feedback and Names

- [ ] 2.1 Read `frontend/src/pages/dashboard/Dashboard.tsx` and locate all dashboard action buttons tested by the lawyer workflow.
- [ ] 2.2 Change "查看全部待办" to route to the existing inbox/task-center route, or visibly expand the full todo list if no route is available. The action must be observable in URL, content, or toast state.
- [ ] 2.3 Change the conflict review "查看全部" action to have a unique visible label or accessible name such as "查看全部冲突任务"; keep its route to `/conflict`.
- [ ] 2.4 Ensure all dashboard buttons that look clickable either navigate, refresh, filter, expand, or show an explicit unavailable message.
- [ ] 2.5 Update `frontend/e2e/dashboard-actions.spec.ts` to assert exact accessible names and expected outcomes for every dashboard action.

## 3. Approval Detail Action Clarity

- [ ] 3.1 Read `frontend/src/pages/approval/ApprovalDetail.tsx` and locate the header action group, material/action tabs, and decision action group.
- [ ] 3.2 Rename or label the header "更多操作" as "更多审批操作".
- [ ] 3.3 Rename or label the decision-area "更多操作" as "更多处理方式".
- [ ] 3.4 Make "查看关联案件" enabled only when a valid related case ID and target route are available.
- [ ] 3.5 If no related case is available, keep the button disabled and expose a visible tooltip or helper text that explains the reason.
- [ ] 3.6 Ensure top header, process node, and action-area approval identity still use the normalized approval access model from prior remediation.
- [ ] 3.7 Update approval E2E tests to assert unique button names, disabled reason visibility, and successful navigation when related case data exists.

## 4. Client Profile Quick Actions

- [ ] 4.1 Read `frontend/src/pages/client/ClientManagement.tsx` and map the existing handlers for "新增联系人", "上传附件", "导出客户档案", "发起新案件", and "发起冲突检查".
- [ ] 4.2 Implement "新增联系人" so it opens the established modal/drawer/form pattern. If a full save API is already available, wire save and success feedback; otherwise show a clearly labeled unavailable state inside the modal.
- [ ] 4.3 Implement "上传附件" so it opens an upload control or shows a clear "附件上传暂不可用" message.
- [ ] 4.4 Implement "导出客户档案" so it either starts the existing export/download flow or shows a clear success/error/unavailable message.
- [ ] 4.5 Ensure "发起新案件" and "发起冲突检查" preserve current working navigation behavior.
- [ ] 4.6 Update `frontend/e2e/client-profile-actions.spec.ts` to verify modal/drawer/toast/download feedback for all quick actions.

## 5. Regression Test Matrix

- [ ] 5.1 Add a negative E2E case to `case-create-full-workflow.spec.ts`: with missing matter title and counterparty, selecting only client and lawyer then clicking conflict check must not create a draft and must show validation feedback.
- [ ] 5.2 Keep the positive full intake E2E case passing: filled matter title, client, counterparty, lawyer, and required select fields can run conflict check and continue to team/fees.
- [ ] 5.3 Add dashboard route/feedback assertions for "查看全部待办" and "查看全部冲突任务".
- [ ] 5.4 Add approval detail assertions for unique "更多审批操作" and "更多处理方式" names.
- [ ] 5.5 Add client profile assertions for "新增联系人", "上传附件", and "导出客户档案" feedback.
- [ ] 5.6 Run targeted E2E:
  - `npm run test:e2e -- case-create-full-workflow.spec.ts dashboard-actions.spec.ts approval-layout.spec.ts approval-permission-consistency.spec.ts client-profile-actions.spec.ts`

## 6. Verification and Documentation

- [ ] 6.1 Run frontend verification from `frontend/`: `npm run type-check`, `npm run lint`, and the targeted E2E command above.
- [ ] 6.2 If backend code changes are needed, run repository backend verification: `go build ./...` and targeted Go tests for touched packages.
- [ ] 6.3 Re-run internal browser smoke test for the exact retest paths in `reports/frontend-lawyer-qa-internal-browser-retest-2026-05-26.md`.
- [ ] 6.4 Create a short remediation report under `reports/` listing fixed issue IDs `IB-RT-001` through `IB-RT-006`, verification commands, and any remaining limitations.
- [ ] 6.5 Do not mark this change complete until all required tests pass or failures are documented as unrelated pre-existing defects with evidence.
