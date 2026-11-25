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
  const scope = 'np-save-v2:' + actorID + ':' + intakeID
  const legacyScope = 'np-save:' + actorID + ':' + intakeID
  const [pending, setPending] = React.useState<Pending | null>(null)
  // True when a same-scope record cannot be strictly verified. The page must
  // block the strict intake write on this signal; it must not reroute the
  // write to another path.
  const [storageBlocked, setStorageBlocked] = React.useState(false)
  const [selection, setSelection] = React.useState<Selection | null>(null)
  const [notice, setNotice] = React.useState('')
  const [confirmManual, setConfirmManual] = React.useState(false)
  const [busy, setBusy] = React.useState(false)
  const frozen = React.useRef<{ scope: string; key: string; body: string } | null>(null)
  const inFlight = React.useRef(false)
  // Each operation binds to the generation that started it; late receipts
  // from older generations settle their own storage and touch nothing else.
  const generationRef = React.useRef(0)
  // Live scope for stale closures: a late operation from a previous scope
  // must never enumerate its old keys into the switched page's view.
  const scopeRef = React.useRef(scope); scopeRef.current = scope
  const recordKey = (requestId: string) => scope + ':' + requestId
  // Single strict validator: exactly the three identifier fields. Mixed or
  // missing fields make the record unverifiable: it blocks, never reads as
  // absent and never gets auto-deleted.
  function readRecord(key: string): Pending | null {
    const raw = localStorage.getItem(key)
    if (raw === null) return null
    const item = JSON.parse(raw)
    const fields = Object.keys(item ?? {}).sort().join(',')
    if (fields !== 'action,intake_id,request_id' || item.action !== 'update' ||
        item.intake_id !== intakeID || typeof item.request_id !== 'string' || !uuid.test(item.request_id) ||
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
    const generation = ++generationRef.current
    frozen.current = null; setSelection(null); setNotice(""); setConfirmManual(false)
    const rediscover = () => {
      try {
        setStorageBlocked(false)
        setPending(readStoredPending())
      } catch {
        setPending(null); setStorageBlocked(true)
        setNotice('恢复标识无法核对，请保留原请求标识联系管理员。')
      }
    }
    rediscover()
    // Same-origin tabs share this localStorage: follow adds/removes so pending
    // records created elsewhere stay discoverable here. The frozen in-flight
    // request keeps priority and page input/selection stay untouched.
    const onStorage = (event: StorageEvent) => {
      if (event.key !== null && !event.key.startsWith('np-save')) return
      if (generationRef.current !== generation) return
      rediscover()
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [scope, legacyScope, intakeID, revision])
  React.useEffect(() => {
    if (selection && selection.description !== description) {
      setSelection(null); setConfirmManual(true); setNotice('摘要已人工修改。请重新核对候选，或明确改为人工摘要保存（不关联来源）。')
    }
  }, [description, selection])
  // Removes only this verified request's storage. View state (frozen/pending)
  // is released by the initiating generation alone; a late receipt from an
  // older generation settles its own storage and leaves the current view
  // untouched. Re-enumeration failure after a confirmed terminal receipt
  // surfaces as storageBlocked, never as an unknown result; malformed legacy
  // data stays for manual handling.
  function clearResolved(requestId: string, generation: number) {
    localStorage.removeItem(recordKey(requestId))
    try {
      if (readRecord(legacyScope)?.request_id === requestId) localStorage.removeItem(legacyScope)
    } catch { /* malformed legacy stays for manual handling */ }
    // Re-enumerate only for the live scope: a stale generation must not leak
    // its other pending records into a page that has switched scope. Within
    // the same scope, a newer generation still gets the refreshed view.
    if (scopeRef.current !== scope) return
    if (generationRef.current === generation && frozen.current?.key === requestId) frozen.current = null
    try { setPending(readStoredPending()) }
    catch { setPending(null); setStorageBlocked(true) }
  }
  const save = async () => {
    if (inFlight.current) throw new Error('原保存仍在处理，请等待或回查原结果')
    if (storageBlocked) throw new Error('恢复标识无法核对；严格保存已阻止，请先处置异常恢复标识。')
    const generation = generationRef.current
    if (confirmManual || (selection && selection.description !== description)) throw new Error('来源已失效，请重新应用候选，或明确改为人工摘要保存（不关联来源）')
    const body: Record<string, unknown> = { expected_revision: revision, changed_fields: ['description'] }
    if (description === '') body.clear_fields = ['description']
    else body.description = description
    if (selection && selection.description === description) body.source_selection = selection.value
    const encoded = JSON.stringify(body)
    // A pending record blocks a new key. Resending is allowed only when it is
    // the same frozen request with the exact same body (unknown-result retry).
    let stored: Pending | null
    try { stored = readStoredPending() }
    catch {
      setStorageBlocked(true)
      throw new Error('恢复标识无法核对，请保留原请求标识联系管理员。')
    }
    const frozenRequest = frozen.current?.scope === scope ? frozen.current : null
    const ownFrozen = Boolean(stored && frozenRequest && stored.request_id === frozenRequest.key)
    if (stored && (!ownFrozen || frozenRequest!.body !== encoded)) {
      throw new Error('原保存结果待核实；当前文字已保留，请先回查结果或明确结束原保存')
    }
    const request = frozenRequest ?? { scope, key: crypto.randomUUID(), body: encoded }
    const ref: Pending = { request_id: request.key, action: 'update', intake_id: intakeID }
    // Persist only the recovery identifier before sending. Failure here prevents an unrecoverable write.
    try { localStorage.setItem(recordKey(request.key), JSON.stringify(ref)) }
    catch { throw new Error('恢复标识无法写入，严格保存请求已阻止；内容保留在页面。') }
    setPending(ref); frozen.current = request
    inFlight.current = true; setBusy(true)
    try {
      const result = await apiRequest<Receipt>('/case-intakes/' + intakeID, { method: "PUT", headers: { "Idempotency-Key": request.key }, body: request.body })
      if (!['committed', 'no_change'].includes(result.state) || !result.revision || !receiptMatchesPending(result, ref)) throw new Error('保存回执无法确认，请按原请求回查')
      clearResolved(request.key, generation)
      // A late receipt from a previous scope generation must not relabel the
      // current page's input as saved.
      if (generationRef.current === generation) { setSelection(null); setNotice("摘要已由原生接口保存，可重新打开核对来源。") }
      return result
    } catch (error) {
      const api = error as { status?: number; code?: string }
      // Only these server refusals prove this key cannot still commit; all other errors retain recovery.
      if (api.status === 409 && ['INTAKE_DRAFT_REVISION_CONFLICT', 'SOURCE_SELECTION_STALE', 'SAVE_REQUEST_ABANDONED'].includes(api.code || '')) {
        clearResolved(request.key, generation)
      }
      else if (generationRef.current === generation) setNotice('原保存结果待核实；不能换键重发，请回查结果或明确结束原保存。')
      throw error
    } finally { inFlight.current = false; setBusy(false) }
  }
  const resolve = async (abandon: boolean) => {
    if (!pending || inFlight.current) return
    const generation = generationRef.current
    const requestId = pending.request_id
    inFlight.current = true
    setBusy(true)
    try {
      const storedKey = recordKey(requestId)
      const storageKey = localStorage.getItem(storedKey) === null && localStorage.getItem(legacyScope) !== null
        ? legacyScope : storedKey
      if (readRecord(storageKey)?.request_id !== requestId) return
      const result = await apiRequest<Receipt>('/case-intakes/write-results/' + requestId + (abandon ? '/abandon' : ''), abandon ? { method: "POST", body: JSON.stringify({ action: pending.action, intake_id: pending.intake_id }) } : {})
      if (result.state === 'unknown' && result.request_id === requestId) {
        if (generationRef.current === generation) setNotice('尚无终态回执，结果仍未知；不能换键。可继续回查或明确结束原保存。')
        return
      }
      if (!receiptMatchesPending(result, pending)) {
        if (generationRef.current === generation) setNotice('原保存结果无法与请求标识核对，请保留原请求标识联系管理员。')
        return
      }
      if (['committed', 'no_change', 'abandoned'].includes(result.state)) {
        clearResolved(requestId, generation)
        if (generationRef.current === generation) setNotice(result.state === "abandoned" ? "原保存已明确结束；未提交内容可核对后另存。" : "已查到原保存结果。当前输入已保留，后续编辑尚未保存。请重新加载服务端草稿核对。")
      } else if (generationRef.current === generation) setNotice('尚无终态回执，结果仍未知；不能换键。可继续回查或明确结束原保存。')
    } catch { if (generationRef.current === generation) setNotice("结果尚未核实，请保留原请求标识并稍后回查。") }
    finally { inFlight.current = false; setBusy(false) }
  }
  return {
    eligible: Boolean(intakeID && revision && sameExceptDescription(projection, baseline)),
    selection, pending, storageBlocked, notice, busy, save, confirmManual,
    apply: (text: string, value: NarrowPilotApplySelection) => {setConfirmManual(false); setSelection({ description: text, value })},
    invalidate: () => {
      if (selection) {setConfirmManual(true); setNotice('旧候选来源已失效，已应用文字保留。请重新整理核对，或明确改为人工摘要保存（不关联来源）。')}
      setSelection(null)
    },
    acceptManual: () => {setSelection(null); setConfirmManual(false); setNotice("本次将按人工摘要保存，不关联候选来源。")},
    query: () => resolve(false), abandon: () => resolve(true),
  }
}
