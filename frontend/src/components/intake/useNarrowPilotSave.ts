import React from 'react'
import type { NarrowPilotApplySelection } from './NarrowPilotAssistant'

type API = <T>(path: string, options?: RequestInit) => Promise<T>
type Receipt = { state: string; id?: string; revision?: string; request_id: string; source_revision?: string }
type Pending = { request_id: string; action: 'update'; intake_id: string }
type Selection = { description: string; value: NarrowPilotApplySelection }

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
  const scope = `np-save:${actorID}:${intakeID}`
  const scopeRef = React.useRef(scope); scopeRef.current = scope
  const [pending, setPending] = React.useState<Pending | null>(null)
  const [selection, setSelection] = React.useState<Selection | null>(null)
  const [notice, setNotice] = React.useState('')
  const [confirmManual, setConfirmManual] = React.useState(false)
  const [busy, setBusy] = React.useState(false)
  const frozen = React.useRef<{ scope: string; key: string; body: string } | null>(null)
  const inFlight = React.useRef(false)
  React.useEffect(() => {
    frozen.current = null; setSelection(null); setNotice(''); setConfirmManual(false)
    try {
      const item = JSON.parse(localStorage.getItem(scope) || 'null')
      setPending(item?.action === 'update' && item.intake_id === intakeID && typeof item.request_id === 'string' ? item : null)
    } catch { setPending(null) }
  }, [scope, intakeID, revision])
  React.useEffect(() => {
    if (selection && selection.description !== description) {
      setSelection(null); setConfirmManual(true); setNotice('摘要已人工修改。请重新核对候选，或明确改为人工摘要保存（不关联来源）。')
    }
  }, [description, selection])
  const clear = (startedScope: string) => {
    localStorage.removeItem(startedScope)
    if (scopeRef.current === startedScope) { setPending(null); frozen.current = null }
  }
  const save = async () => {
    if (inFlight.current) throw new Error('原保存仍在处理，请等待或回查原结果')
    if (confirmManual || (selection && selection.description !== description)) throw new Error('来源已失效，请重新应用候选，或明确改为人工摘要保存（不关联来源）')
    const body: Record<string, unknown> = { expected_revision: revision, changed_fields: ['description'] }
    if (description === '') body.clear_fields = ['description']
    else body.description = description
    if (selection && selection.description === description) body.source_selection = selection.value
    const encoded = JSON.stringify(body)
    const stored = localStorage.getItem(scope)
    if (stored && (!frozen.current || frozen.current.scope !== scope || frozen.current.body !== encoded)) {
      throw new Error('原保存结果待核实；当前文字已保留，请先回查结果或明确结束原保存')
    }
    const request = frozen.current?.scope === scope ? frozen.current : { scope, key: crypto.randomUUID(), body: encoded }
    const ref: Pending = { request_id: request.key, action: 'update', intake_id: intakeID }
    // Persist only the recovery identifier before sending. Failure here prevents an unrecoverable write.
    localStorage.setItem(scope, JSON.stringify(ref)); setPending(ref); frozen.current = request
    inFlight.current = true; setBusy(true)
    try {
      const result = await apiRequest<Receipt>(`/case-intakes/${intakeID}`, { method: 'PUT', headers: { 'Idempotency-Key': request.key }, body: request.body })
      if (!['committed', 'no_change'].includes(result.state) || result.id !== intakeID || !result.revision) throw new Error('保存回执无法确认，请按原请求回查')
      clear(scope); if (scopeRef.current === scope) { setSelection(null); setNotice('摘要已由原生接口保存，可重新打开核对来源。') }
      return result
    } catch (error) {
      const api = error as { status?: number; code?: string }
      // Only these server refusals prove this key cannot still commit; all other errors retain recovery.
      if (api.status === 409 && ['INTAKE_DRAFT_REVISION_CONFLICT', 'SOURCE_SELECTION_STALE', 'SAVE_REQUEST_ABANDONED'].includes(api.code || '')) clear(scope)
      else if (scopeRef.current === scope) setNotice('原保存结果待核实；不能换键重发，请回查结果或明确结束原保存。')
      throw error
    } finally { inFlight.current = false; if (scopeRef.current === scope) setBusy(false) }
  }
  const resolve = async (abandon: boolean) => {
    if (!pending || inFlight.current) return
    const stored = localStorage.getItem(scope)
    try { if (!stored || JSON.parse(stored).request_id !== pending.request_id) return } catch { setNotice('恢复标识无法核对，请保留原请求标识联系管理员。'); return }
    inFlight.current = true
    setBusy(true)
    try {
      const result = await apiRequest<Receipt>(`/case-intakes/write-results/${pending.request_id}${abandon ? '/abandon' : ''}`, abandon ? { method: 'POST', body: JSON.stringify({ action: pending.action, intake_id: pending.intake_id }) } : {})
      if (scopeRef.current !== scope) return
      if (['committed', 'no_change', 'abandoned'].includes(result.state)) {
        clear(scope)
        setNotice(result.state === 'abandoned' ? '原保存已明确结束；未提交内容可核对后另存。' : '已查到原保存结果；当前输入保留，请点击“重新加载服务端草稿”核对。')
      } else setNotice('尚无终态回执，结果仍未知；不能换键。可继续回查或明确结束原保存。')
    } catch { if (scopeRef.current === scope) setNotice('结果尚未核实，请保留原请求标识并稍后回查。') }
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
    acceptManual: () => {setSelection(null); setConfirmManual(false); setNotice('本次将按人工摘要保存，不关联候选来源。')},
    query: () => resolve(false), abandon: () => resolve(true),
  }
}
