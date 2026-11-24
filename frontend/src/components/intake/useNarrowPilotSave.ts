import React from 'react'
import type { NarrowPilotApplySelection } from './NarrowPilotAssistant'

type API = <T>(path: string, options?: RequestInit) => Promise<T>
type Receipt = {
  state: string
  id?: string
  revision?: string
  request_id?: string
  action?: string
  source_revision?: string
}
type Pending = { request_id: string; action: 'update'; intake_id: string }
type Selection = { description: string; value: NarrowPilotApplySelection }

const uuid = /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i
export function isIntakeRevision(value: unknown, intakeID: string): boolean {
  return typeof value === 'string' && (uuid.test(value) || value === 'legacy-' + intakeID)
}

// MI-R08: a terminal receipt must prove it belongs to this pending request.
// request_id is required; the receipt object field is `id` and must match the
// pending intake_id; `action` is optional and compared only when present.
function receiptMatchesPending(receipt: Receipt, pending: Pending): boolean {
  if (receipt.request_id !== pending.request_id) return false
  if (receipt.id !== pending.intake_id) return false
  if (receipt.action !== undefined && receipt.action !== pending.action) return false
  if (receipt.state === 'abandoned') return true
  return ['committed', 'no_change'].includes(receipt.state) && isIntakeRevision(receipt.revision, pending.intake_id)
}

export function sameExceptDescription(a: string, b: string): boolean {
  if (!a || !b) return false
  try {
    const left = JSON.parse(a); const right = JSON.parse(b)
    delete left.form.description; delete right.form.description
    return JSON.stringify(left) === JSON.stringify(right)
  } catch { return false }
}

export function useNarrowPilotSave(props: {
  actorID: string; intakeID: string; revision: string; description: string
  projection: string; baseline: string; apiRequest: API
}) {
  const { actorID, intakeID, revision, description, projection, baseline, apiRequest } = props
  // v2 keys isolate every request (np-save-v2:<actor>:<intake>:<request>):
  // one pending save must never overwrite or clear another page record.
  // The legacy single slot stays readable-only for upgrade compatibility.
  const scope = `np-save-v2:${actorID}:${intakeID}`
  const legacyScope = `np-save:${actorID}:${intakeID}`
  const scopeRef = React.useRef(scope); scopeRef.current = scope
  const [pending, setPending] = React.useState<Pending | null>(null)
  const [selection, setSelection] = React.useState<Selection | null>(null)
  const [notice, setNotice] = React.useState('')
  const [confirmManual, setConfirmManual] = React.useState(false)
  const [busy, setBusy] = React.useState(false)
  const frozen = React.useRef<{ scope: string; key: string; body: string } | null>(null)
  const inFlight = React.useRef(false)
  const recordKey = (requestId: string) => `${scope}:${requestId}`
  function readRecord(key: string): Pending | null {
    const raw = localStorage.getItem(key)
    if (raw === null) return null
    const item = JSON.parse(raw)
    if (item?.action !== 'update' || item.intake_id !== intakeID || typeof item.request_id !== 'string' ||
        !uuid.test(item.request_id) ||
        (key !== legacyScope && key !== recordKey(item.request_id))) {
      throw new Error('恢复标识无法核对，请保留原请求标识联系管理员。')
    }
    return { action: 'update', intake_id: intakeID, request_id: item.request_id }
  }
  function readStoredPending(): Pending | null {
    if (frozen.current?.scope === scope) {
      const own = readRecord(recordKey(frozen.current.key))
      if (own) return own
    }
    for (let index = 0; index < localStorage.length; index++) {
      const key = localStorage.key(index)
      if (key?.startsWith(scope + ':')) return readRecord(key)
    }
    return readRecord(legacyScope)
  }
  React.useEffect(() => {
    frozen.current = null; setSelection(null); setNotice(""); setConfirmManual(false)
    try { setPending(readStoredPending()) } catch {
      setPending(null)
      setNotice('恢复标识无法核对，请保留原请求标识联系管理员。')
    }
  }, [scope, legacyScope, intakeID, revision])
  React.useEffect(() => {
    if (selection && selection.description !== description) {
      setSelection(null); setConfirmManual(true); setNotice('摘要已人工修改。请重新核对候选，或明确改为人工摘要保存（不关联来源）。')
    }
  }, [description, selection])
  // v2 clears only the resolved request record; the legacy slot is cleared in
  // the same terminal step only when it stores this same request id.
  function clearResolved(requestId: string) {
    for (const key of [recordKey(requestId), legacyScope]) {
      if (readRecord(key)?.request_id === requestId) localStorage.removeItem(key)
    }
    if (scopeRef.current !== scope) return
    frozen.current = null
    setPending(readStoredPending())
  }
  const save = async () => {
    if (inFlight.current) throw new Error('原保存仍在处理，请等待或回查原结果')
    if (confirmManual || (selection && selection.description !== description)) throw new Error('来源已失效，请重新应用候选，或明确改为人工摘要保存（不关联来源）')
    const body: Record<string, unknown> = { expected_revision: revision, changed_fields: ['description'] }
    if (description === '') body.clear_fields = ['description']
    else body.description = description
    if (selection && selection.description === description) body.source_selection = selection.value
    const encoded = JSON.stringify(body)
    // A pending record blocks a new key. Resending is allowed only when it is
    // the same frozen request with the exact same body (unknown-result retry).
    const stored = readStoredPending()
    const frozenRequest = frozen.current?.scope === scope ? frozen.current : null
    const ownFrozen = Boolean(stored && frozenRequest && stored.request_id === frozenRequest.key)
    if (stored && (!ownFrozen || frozenRequest!.body !== encoded)) {
      throw new Error('原保存结果待核实；当前文字已保留，请先回查结果或明确结束原保存')
    }
    const request = frozenRequest ?? { scope, key: crypto.randomUUID(), body: encoded }
    const ref: Pending = { request_id: request.key, action: 'update', intake_id: intakeID }
    // Persist only the recovery identifier before sending. Failure here prevents an unrecoverable write.
    localStorage.setItem(recordKey(request.key), JSON.stringify(ref)); setPending(ref); frozen.current = request
    inFlight.current = true; setBusy(true)
    try {
      const result = await apiRequest<Receipt>(`/case-intakes/${intakeID}`, { method: "PUT", headers: { "Idempotency-Key": request.key }, body: request.body })
      if (!['committed', 'no_change'].includes(result.state) || !result.revision || !receiptMatchesPending(result, ref)) throw new Error('保存回执无法确认，请按原请求回查')
      clearResolved(request.key)
      if (scopeRef.current === scope) { setSelection(null); setNotice("摘要已由原生接口保存，可重新打开核对来源。") }
      return result
    } catch (error) {
      const api = error as { status?: number; code?: string }
      // Only these server refusals prove this key cannot still commit; all other errors retain recovery.
      if (api.status === 409 && ['INTAKE_DRAFT_REVISION_CONFLICT', 'SOURCE_SELECTION_STALE', 'SAVE_REQUEST_ABANDONED'].includes(api.code || '')) {
        clearResolved(request.key)
      }
      else if (scopeRef.current === scope) setNotice('原保存结果待核实；不能换键重发，请回查结果或明确结束原保存。')
      throw error
    } finally { inFlight.current = false; if (scopeRef.current === scope) setBusy(false) }
  }
  const resolve = async (abandon: boolean) => {
    if (!pending || inFlight.current) return
    const storedKey = recordKey(pending.request_id)
    const isLegacy = localStorage.getItem(storedKey) === null && localStorage.getItem(legacyScope) !== null
    const storageKey = isLegacy ? legacyScope : storedKey
    try { if (readRecord(storageKey)?.request_id !== pending.request_id) return } catch { setNotice("恢复标识无法核对，请保留原请求标识联系管理员。"); return }
    inFlight.current = true
    setBusy(true)
    try {
      const result = await apiRequest<Receipt>(`/case-intakes/write-results/${pending.request_id}${abandon ? "/abandon" : ""}`, abandon ? { method: "POST", body: JSON.stringify({ action: pending.action, intake_id: pending.intake_id }) } : {})
      if (scopeRef.current !== scope) return
      if (result.state === 'unknown' && result.request_id === pending.request_id) {
        setNotice('尚无终态回执，结果仍未知；不能换键。可继续回查或明确结束原保存。')
        return
      }
      if (!receiptMatchesPending(result, pending)) { setNotice('原保存结果无法与请求标识核对，请保留原请求标识联系管理员。'); return }
      if (['committed', 'no_change', 'abandoned'].includes(result.state)) {
        clearResolved(pending.request_id)
        setNotice(result.state === "abandoned" ? "原保存已明确结束；未提交内容可核对后另存。" : "已查到原保存结果；当前输入保留，请点击“重新加载服务端草稿”核对。")
      } else setNotice('尚无终态回执，结果仍未知；不能换键。可继续回查或明确结束原保存。')
    } catch { if (scopeRef.current === scope) setNotice("结果尚未核实，请保留原请求标识并稍后回查。") }
    finally { inFlight.current = false; if (scopeRef.current === scope) setBusy(false) }
  }
  return {
    eligible: Boolean(intakeID && revision && sameExceptDescription(projection, baseline)),
    selection, pending, notice, busy, save, confirmManual,
    apply: (text: string, value: NarrowPilotApplySelection) => {setConfirmManual(false); setSelection({ description: text, value })},
    invalidate: () => {
      if (selection) {setConfirmManual(true); setNotice('旧候选来源已失效，已应用文字保留。请重新整理核对，或明确改为人工摘要保存（不关联来源）。')}
      setSelection(null)
    },
    acceptManual: () => {setSelection(null); setConfirmManual(false); setNotice("本次将按人工摘要保存，不关联候选来源。")},
    query: () => resolve(false), abandon: () => resolve(true),
  }
}
