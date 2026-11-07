import { apiRequest } from './collaborationRequest'

export const supplementFields: Record<string, { label: string; id: string; step: number }> = {
  parties: { label: '当事人', id: 'intake-client', step: 0 },
  party_name: { label: '指定当事人名称', id: '', step: 0 },
  party_role: { label: '指定当事人角色', id: '', step: 0 },
  party_identity: { label: '指定当事人身份', id: '', step: 0 },
  title: { label: '案件名称', id: 'intake-title', step: 0 },
  client_id: { label: '客户', id: 'intake-client', step: 0 },
  opponent_name: { label: '对方名称', id: 'intake-opponent', step: 0 },
  opponent_identity_number: { label: '对方身份', id: 'intake-opponent-identity-number', step: 0 },
  description: { label: '事项摘要', id: 'intake-description', step: 0 },
  case_type: { label: '案件类型', id: 'intake-case-type', step: 0 },
  business_area: { label: '业务领域', id: 'intake-business-area', step: 0 },
  sub_area: { label: '子领域', id: 'intake-sub-area', step: 0 },
  lawyer_id: { label: '负责律师', id: 'intake-lawyer', step: 2 },
}
export interface SupplementItem {
  item_id: string
  category: 'field_correction' | 'identity_verification' | 'matter_explanation'
  field_key?: string
  party_local_id?: string
  description: string
  required?: boolean
}
export interface SupplementResponse {
  item_id: string
  response: string
}
export interface SupplementReturn {
  recordId: string
  items: SupplementItem[]
  legacyText: string
}
export function objectFromJSON(value: unknown): Record<string, any> {
  if (typeof value === 'string') {
    try {
      return objectFromJSON(JSON.parse(value))
    } catch {
      return {}
    }
  }
  return value && typeof value === 'object' && !Array.isArray(value)
    ? (value as Record<string, any>)
    : {}
}
export function latestSupplementReturn(approval: Record<string, any>): SupplementReturn | null {
  if (approval.return_record_id && Array.isArray(approval.supplement_requests)) {
    return {
      recordId: String(approval.return_record_id),
      items: approval.supplement_requests,
      legacyText: String(approval.return_reason || ''),
    }
  }
  const records: Record<string, any>[] = Array.isArray(approval.records) ? approval.records : []
  const returned = records
    .filter((row) => row.decision === 'request_changes')
    .sort((a, b) =>
      String(b.approval_date || b.created_at || '').localeCompare(
        String(a.approval_date || a.created_at || ''),
      ),
    )[0]
  if (!returned?.id) return null
  const requirements = objectFromJSON(returned.imposed_requirements)
  return {
    recordId: String(returned.id),
    items:
      requirements.schema_version === 1 && Array.isArray(requirements.supplement_requests)
        ? requirements.supplement_requests
        : [],
    legacyText: String(returned.decision_reason || returned.decision_comments || ''),
  }
}
export interface ResubmitIntakeApproval {
  expected_intake_revision: string
  expected_submission_snapshot_id: string
  return_record_id: string
  supplement_responses: SupplementResponse[]
  revision_note: string
}
export const resubmitIntakeApproval = (id: string, payload: ResubmitIntakeApproval) =>
  apiRequest<any>(`/approvals/${encodeURIComponent(id)}/resubmit`, {
    method: 'POST',
    body: JSON.stringify({
      ...payload,
      supplement_responses: payload.supplement_responses.filter((reply) => reply.response.trim()),
    }),
  })

export const saveSupplementResponses = (
  intakeId: string,
  approvalId: string,
  expectedRevision: string,
  recordId: string,
  responses: SupplementResponse[],
) =>
  apiRequest<any>(`/case-intakes/${encodeURIComponent(intakeId)}`, {
    method: 'PUT',
    body: JSON.stringify({
      approval_id: approvalId,
      expected_revision: expectedRevision,
      return_record_id: recordId,
      supplement_responses: responses.filter((reply) => reply.response.trim()),
    }),
  })

// Only call on the default/current snapshot response. The reader has already
// verified legacy provenance and returns the persisted row UUID as viewed ID.
export function currentSubmissionSnapshotID(payload: Record<string, any>): string {
  const active = payload.active_submission_snapshot_id
  const viewed = payload.viewed_snapshot_id
  if (typeof viewed !== 'string' || !viewed || viewed.startsWith('SNAP_') || (active && active !== viewed)) {
    throw new Error('无法核实当前提交版本，请重新加载审批详情')
  }
  return viewed
}
