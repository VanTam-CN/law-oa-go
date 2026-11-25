import { isIntakeRevision, sameExceptDescription } from '@/components/intake/useNarrowPilotSave'

// I13 (implementation-and-acceptance-plan 118-130): the strict
// description-only channel is a frontend convenience. The server keeps
// enforcing session, ownership, draft state and revision on every write.
// History facts come from the server detail (has_source_history); a missing
// field means unknown and must never be treated as "no history".
export type NarrowPilotSaveChannel = 'strict' | 'native' | 'blocked'

export interface NarrowPilotEligibilityInput {
  isAssistant: boolean
  isSupplementMode: boolean
  intakeID: string
  revision: string
  currentUserID: string
  role?: string
  userIsActive?: boolean
  ownerID: unknown
  intakeStatus?: string
  projection: string
  baseline: string
  hasSourceHistory: boolean
  sourceHistoryKnown: boolean
}

export interface NarrowPilotChannelDecision {
  mode: NarrowPilotSaveChannel
  reason?: string
}

function strictlyQualified(input: NarrowPilotEligibilityInput): boolean {
  if (input.isAssistant || input.isSupplementMode) return false
  if (!input.intakeID || !isIntakeRevision(input.revision, input.intakeID)) return false
  if (input.role !== 'lawyer') return false
  if (input.userIsActive !== true) return false
  if (!input.currentUserID || String(input.ownerID ?? '') !== input.currentUserID) return false
  if (input.intakeStatus !== 'draft') return false
  if (!input.baseline || input.baseline === input.projection) return false
  return sameExceptDescription(input.baseline, input.projection)
}

export function narrowPilotIntakeEligibility(input: NarrowPilotEligibilityInput): NarrowPilotChannelDecision {
  if (!input.intakeID) return { mode: 'native' }
  if (!input.sourceHistoryKnown) {
    return { mode: 'blocked', reason: '来源保护状态未知：请重新加载服务端草稿后再保存' }
  }
  if (strictlyQualified(input)) return { mode: 'strict' }
  if (!input.hasSourceHistory) return { mode: 'native' }
  return { mode: 'blocked', reason: '此草稿已有来源保护，仅支持严格摘要保存；存在其他改动或条件不符时请重新加载草稿核对后由负责律师处理' }
}

// After a strict description-only save the new anchor equals the previous
// anchor with only the description replaced. Non-editable identity fields
// never leave memory: they live inside the projection object only.
export function projectionWithDescription(baselineProjection: string, description: string): string {
  try {
    const parsed = JSON.parse(baselineProjection) as { form?: Record<string, unknown> }
    parsed.form = { ...(parsed.form || {}), description }
    return JSON.stringify(parsed)
  } catch {
    return ''
  }
}
