import { ApiRequestError } from './collaborationRequest'
import { objectFromJSON } from './intakeSupplement'

export function approvalCreatesCase(approval: Record<string, any>): boolean {
  const metadata = objectFromJSON(approval.metadata)
  return ['case_creation', 'conflict_approval'].includes(approval.type) ||
    Object.prototype.hasOwnProperty.call(metadata, 'case_creation_config') ||
    ['case_creation', 'both'].includes(String(metadata.integration_type || ''))
}

// Absence of frontend intake metadata is never proof of an ordinary approval.
// Only the authorized server projection can establish that no snapshot is required.
export function allowsSnapshotFreeDecision(approval: Record<string, any>, error: unknown): boolean {
  const metadata = objectFromJSON(approval.metadata)
  return approval.submission_snapshot_required === false &&
    !approval.active_submission_snapshot_id && !approval.intake_id && !metadata.intake_id &&
    !approvalCreatesCase(approval) &&
    error instanceof ApiRequestError && error.status === 404
}
