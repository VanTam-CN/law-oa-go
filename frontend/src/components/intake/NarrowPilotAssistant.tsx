import React from 'react'
import { Alert, Button, Checkbox, Input, Tag } from 'antd'

export interface NarrowPilotSource {
  id: string
  kind: string
  text: string
  turn?: number
  accepted_at?: string
}
export interface NarrowPilotCandidate {
  id: string
  description: string
  sources: NarrowPilotSource[]
}
export interface NarrowPilotContextView {
  id: string
  context_version: number
  expires_at: string
  candidate: NarrowPilotCandidate | null
}
export interface NarrowPilotApplySelection {
  context_id: string
  expected_context_version: number
  fields: {
    description: {
      candidate_id: string
      requested_mode: 'candidate_sources'
      reviewed: boolean
    }
  }
}
interface NarrowPilotSavedSource {
  description: string
  sources: NarrowPilotSource[]
  reviewed_by?: string
  reviewed_at?: string
  original_revision?: string
}
interface NarrowPilotSnapshot {
  revision?: string
  source_revision?: string | null
  source?: NarrowPilotSavedSource | null
}
export interface NarrowPilotAssistantProps {
  intakeId: string
  revision: string
  description: string
  apiRequest: <T>(path: string, options?: RequestInit) => Promise<T>
  onApply: (description: string, selection: NarrowPilotApplySelection) => void
  onInvalidate?: () => void
}

const INVALIDATING_CODES = new Set([
  'CONTEXT_STALE',
  'CONTEXT_UNAVAILABLE',
  'CONTEXT_EXPIRING',
  'CONTEXT_BUDGET_EXHAUSTED',
  'INTAKE_DRAFT_REVISION_CONFLICT',
  'ACCESS_SESSION_REQUIRED',
  'OBJECT_ACCESS_DENIED',
  'ASSISTANT_DISABLED',
])

function newRequestId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  // crypto.randomUUID 缺失时的 UUID v4 兜底（crypto.getRandomValues 可用则真随机）。
  const bytes =
    typeof crypto !== 'undefined' && crypto.getRandomValues
      ? crypto.getRandomValues(new Uint8Array(16))
      : Uint8Array.from({ length: 16 }, () => Math.floor(Math.random() * 256))
  bytes[6] = (bytes[6] & 0x0f) | 0x40
  bytes[8] = (bytes[8] & 0x3f) | 0x80
  const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, '0'))
  return `${hex.slice(0, 4).join('')}-${hex.slice(4, 6).join('')}-${hex.slice(6, 8).join('')}-${hex
    .slice(8, 10)
    .join('')}-${hex.slice(10, 16).join('')}`
}
function errorCodeOf(error: unknown): string {
  return error && typeof error === 'object' && 'code' in error
    ? String((error as { code?: unknown }).code ?? '')
    : ''
}
function errorTextOf(error: unknown): string {
  return error instanceof Error && error.message ? error.message : '操作未完成，输入已保留，请重试'
}

function SourceList({ items }: { items: NarrowPilotSource[] }) {
  return (
    <ul style={{ margin: '4px 0', paddingLeft: 20 }}>
      {items.map((item) => (
        <li key={item.id}>
          <Tag>{item.turn ? `第 ${item.turn} 轮录入原句` : '本次录入原句'}</Tag>
          <span>{item.text}</span>
        </li>
      ))}
    </ul>
  )
}

export function NarrowPilotAssistant({
  intakeId,
  revision,
  description,
  apiRequest,
  onApply,
  onInvalidate,
}: NarrowPilotAssistantProps) {
  // A monotonic epoch also rejects late responses after switching away and back.
  const scope = `${intakeId}|${revision}`
  const scopeRef = React.useRef(scope)
  const requestEpochRef = React.useRef(0)
  if (scopeRef.current !== scope) { scopeRef.current = scope; requestEpochRef.current++ }
  const [enabled, setEnabled] = React.useState(false)
  const [snapshot, setSnapshot] = React.useState<NarrowPilotSnapshot | null>(null)
  const [ctx, setCtx] = React.useState<NarrowPilotContextView | null>(null)
  const [input, setInput] = React.useState('')
  const [reviewed, setReviewed] = React.useState(false)
  const [busy, setBusy] = React.useState(false)
  const [error, setError] = React.useState('')
  const [appliedHint, setAppliedHint] = React.useState('')
  const [stalePair, setStalePair] = React.useState<{
    candidate: NarrowPilotCandidate
    pageText: string
  } | null>(null)
  // 基线：当前候选生成那一刻的页面摘要；人工修改（含清空）后应用前必须重新对比。
  const baselineRef = React.useRef<{ version: number; candidateId: string; description: string } | null>(null)
  const descriptionRef = React.useRef(description)
  descriptionRef.current = description
  const aliveRef = React.useRef(true)
  const invalidateRef = React.useRef(onInvalidate)
  invalidateRef.current = onInvalidate
  React.useEffect(() => {
    aliveRef.current = true
    return () => {
      aliveRef.current = false
    }
  }, [])
  React.useEffect(() => {
    if (!ctx) return
    const timer = window.setTimeout(() => {
      setCtx(null); setReviewed(false); setStalePair(null); baselineRef.current = null
      setError('本次整理已到期，已应用文字保留，请重新开始整理并核对来源。')
      invalidateRef.current?.()
    }, Math.max(0, Date.parse(ctx.expires_at) - Date.now()))
    return () => window.clearTimeout(timer)
  }, [ctx?.id, ctx?.expires_at])

  React.useEffect(() => {
    let cancelled = false
    setBusy(false)
    setSnapshot(null)
    setEnabled(false)
    setCtx(null)
    setInput('')
    setReviewed(false)
    setError('')
    setAppliedHint('')
    setStalePair(null)
    baselineRef.current = null
    if (!intakeId || !revision) {
      setEnabled(false)
      setSnapshot(null)
      return
    }
    apiRequest<{ enabled: boolean }>('/case-intake-assist/capabilities')
      .then((data) => {
        if (!cancelled) setEnabled(data?.enabled === true)
      })
      .catch(() => {
        if (!cancelled) setEnabled(false)
      })
    apiRequest<NarrowPilotSnapshot>(`/case-intakes/${intakeId}/source-snapshots/${revision}`)
      .then((data) => {
        if (!cancelled) setSnapshot(data ?? null)
      })
      .catch(() => {
        if (!cancelled) setSnapshot(null)
      })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [scope])

  const invalidateCandidate = (text: string, notifyParent: boolean) => {
    setCtx(null)
    setReviewed(false)
    setStalePair(null)
    setAppliedHint('')
    baselineRef.current = null
    setError(text)
    if (notifyParent) onInvalidate?.()
  }
  const clearDeniedContent = (error: unknown) => {
    if (['ACCESS_SESSION_REQUIRED', 'OBJECT_ACCESS_DENIED'].includes(errorCodeOf(error))) {
      setSnapshot(null); setEnabled(false); setInput('')
    }
  }

  const sendTurn = async () => {
    // 仅判空用 trim；原句原样提交，服务端 Stub 逐字保存为可核对来源。
    const hasText = input.trim().length > 0
    if (!hasText || !intakeId || !revision || busy) return
    const epoch = ++requestEpochRef.current
    const stillCurrent = () => aliveRef.current && requestEpochRef.current === epoch
    const sentText = input
    // 新输入一旦发出即保守失效旧选择：即便请求失败、状态不明，
    // 旧 selection 也不能再随保存提交；表单文字保持不动。
    onInvalidate?.()
    setCtx(current => current ? { ...current, candidate: null } : null)
    setReviewed(false)
    setBusy(true)
    setError('')
    setAppliedHint('')
    const pageDescription = description
    try {
      const current = ctx
      const view = current
        ? current
        : await apiRequest<NarrowPilotContextView>('/case-intake-assist/contexts', {
            method: 'POST',
            body: JSON.stringify({ intake_id: intakeId, expected_revision: revision }),
          })
      if (!stillCurrent()) return
      const updated = await apiRequest<NarrowPilotContextView>(
        `/case-intake-assist/contexts/${view.id}/turns`,
        {
          method: 'POST',
          body: JSON.stringify({
            request_id: newRequestId(),
            expected_context_version: view.context_version,
            text: sentText,
          }),
        },
      )
      if (!stillCurrent()) return
      setCtx(updated)
      setInput(current => current === sentText ? '' : current)
      if (updated.candidate) {
        baselineRef.current = {
          version: updated.context_version,
          candidateId: updated.candidate.id,
          description: pageDescription,
        }
        setReviewed(false)
        setStalePair(null)
      }
    } catch (turnError) {
      if (!stillCurrent()) return
      clearDeniedContent(turnError)
      if (INVALIDATING_CODES.has(errorCodeOf(turnError))) {
        invalidateCandidate('整理上下文或草稿版本已失效，请核对当前草稿后重新开始', true)
      } else {
        setError(errorTextOf(turnError))
      }
    } finally {
      if (stillCurrent()) setBusy(false)
    }
  }

  const confirmApply = (candidate: NarrowPilotCandidate) => {
    const baseline = baselineRef.current
    const context = ctx
    if (!baseline || !context) return
    setStalePair(null)
    baselineRef.current = { ...baseline, description: candidate.description }
    setAppliedHint('候选已写入表单摘要；请用原“保存草稿”按钮保存，来源才会随保存写入服务器')
    onApply(candidate.description, {
      context_id: context.id,
      expected_context_version: baseline.version,
      fields: {
        description: {
          candidate_id: candidate.id,
          requested_mode: 'candidate_sources',
          reviewed: true,
        },
      },
    })
  }

  const requestApply = async (force = false) => {
    const context = ctx
    const candidate = context?.candidate
    const baseline = baselineRef.current
    if (!context || !candidate || !baseline || !reviewed || busy) return
    const epoch = ++requestEpochRef.current
    const stillCurrent = () => aliveRef.current && requestEpochRef.current === epoch
    setBusy(true)
    setError('')
    try {
      const view = await apiRequest<NarrowPilotContextView>(`/case-intake-assist/contexts/${context.id}`)
      if (!stillCurrent()) return
      setCtx(view)
      const current = view.candidate
      if (!current || view.context_version !== baseline.version || current.id !== baseline.candidateId) {
        invalidateCandidate('候选已失效（已有新输入或上下文变化），请重新生成并核对来源', true)
        return
      }
      // await 期间人工可能又改了摘要：一律读 ref 里的最新页面值，
      // 不用闭包里的旧值，避免覆盖用户在等待期间输入的文字。
      const latestPage = descriptionRef.current
      if (force) {
        // 覆盖确认打开后页面又变了：回到对比界面重新确认，不能按旧差异直接覆盖。
        if (stalePair && latestPage !== stalePair.pageText) {
          setStalePair({ candidate: current, pageText: latestPage })
          return
        }
      } else if (latestPage !== baseline.description) {
        setStalePair({ candidate: current, pageText: latestPage })
        return
      }
      confirmApply(current)
    } catch (applyError) {
      if (!stillCurrent()) return
      clearDeniedContent(applyError)
      if (INVALIDATING_CODES.has(errorCodeOf(applyError)) || errorCodeOf(applyError) === 'SOURCE_SELECTION_STALE') {
        invalidateCandidate('整理上下文已失效，请重新生成候选', true)
      } else {
        setError(errorTextOf(applyError))
      }
    } finally {
      if (stillCurrent()) setBusy(false)
    }
  }

  const candidate = ctx?.candidate ?? null
  const baseline = baselineRef.current
  const pageChanged = Boolean(candidate && baseline && description !== baseline.description)
  // 来源读取不依赖助手开关：白名单未启用时仍展示已保存来源供重开核对。
  const saved = snapshot?.source ?? null
  const showAssistant = enabled && Boolean(intakeId && revision)

  return (
    <section className='np-assistant' aria-label='窄版摘要助手'>
      <h3>窄版摘要助手（首切片：仅案情摘要）</h3>
      {saved ? (
        <div style={{ margin: '8px 0', padding: 8, background: '#fafafa' }}>
          <strong>已保存摘要来源</strong>
          {saved.description !== description && <Alert type='info' message='当前表单摘要已变化；以下仅为已保存版本的出处，不代表当前输入仍有自动来源关联。' />}
          <p style={{ margin: '4px 0' }}>{saved.description}</p>
          <SourceList items={saved.sources} />
          {saved.reviewed_by && (
            <p style={{ margin: '4px 0', color: '#666' }}>
              核对人：{saved.reviewed_by}
              {saved.reviewed_at ? `（${saved.reviewed_at}）` : ''}
            </p>
          )}
        </div>
      ) : (
        snapshot && <p style={{ margin: '4px 0', color: '#999' }}>当前版本尚无已保存来源</p>
      )}
      {showAssistant ? (
        <>
          {error && <Alert type='warning' message={error} showIcon />}
          {ctx && <p>本次整理到期时间：{new Date(ctx.expires_at).toLocaleString()}</p>}
          <div className='batch-field'>
            <label htmlFor='np-assist-input'>摘要对话输入</label>
            <Input.TextArea
              id='np-assist-input'
              rows={3}
              value={input}
              onChange={(event) => setInput(event.target.value)}
              placeholder='输入案情事实、补充或更正；原句会作为可核对来源保留'
            />
          </div>
          <Button
            aria-label={candidate ? '生成摘要候选' : '开始摘要整理'}
            disabled={!input.trim() || busy}
            loading={busy}
            onClick={sendTurn}
          >
            {candidate ? '生成摘要候选' : '开始摘要整理'}
          </Button>
          {candidate && baseline && (
            <div style={{ marginTop: 12 }}>
              <strong>摘要候选</strong>
              <pre style={{ whiteSpace: 'pre-wrap', background: '#f6f6f6', padding: 8 }}>{candidate.description}</pre>
              <strong>可核对来源（{candidate.sources.length} 条）</strong>
              <SourceList items={candidate.sources} />
              {pageChanged && (
                <Alert
                  style={{ margin: '8px 0' }}
                  type='warning'
                  showIcon
                  message='生成候选后页面摘要又被人工修改（含清空）。应用会覆盖人工文字，请先对比下方差异再确认。'
                />
              )}
              <Checkbox checked={reviewed} onChange={(event) => setReviewed(event.target.checked)}>
                已核对来源
              </Checkbox>
              <Button
                type='primary'
                aria-label='应用摘要'
                style={{ marginLeft: 12 }}
                disabled={!reviewed || busy}
                loading={busy}
                onClick={() => requestApply(false)}
              >
                应用摘要
              </Button>
              {appliedHint && <p style={{ margin: '8px 0', color: '#1677ff' }}>{appliedHint}</p>}
            </div>
          )}
          {stalePair && (
            <div role='alertdialog' aria-label='覆盖确认' style={{ marginTop: 12, border: '1px solid #faad14', padding: 12 }}>
              <strong>页面当前摘要与生成时不一致，请对比后明确确认</strong>
              <div style={{ display: 'flex', gap: 12, marginTop: 8 }}>
                <div style={{ flex: 1 }}>
                  <p>页面当前（将被覆盖）</p>
                  <pre style={{ whiteSpace: 'pre-wrap', background: '#fff7e6', padding: 8 }}>{stalePair.pageText || '（已清空）'}</pre>
                </div>
                <div style={{ flex: 1 }}>
                  <p>候选摘要（应用后）</p>
                  <pre style={{ whiteSpace: 'pre-wrap', background: '#f6ffed', padding: 8 }}>{stalePair.candidate.description}</pre>
                </div>
              </div>
              <Button danger onClick={() => requestApply(true)} disabled={busy}>
                确认覆盖人工修改并应用
              </Button>
              <Button style={{ marginLeft: 8 }} onClick={() => setStalePair(null)} disabled={busy}>
                取消
              </Button>
            </div>
          )}
        </>
      ) : (
        !enabled && intakeId && revision && (
          <p style={{ margin: '8px 0 0', color: '#999' }}>当前账号未启用摘要助手，仅展示已保存来源。</p>
        )
      )}
    </section>
  )
}

export default NarrowPilotAssistant
