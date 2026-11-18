import React from 'react'
import { Alert, Button, Checkbox, DatePicker, Form, Input, Modal, Select, Space, Tag } from 'antd'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  HistoryOutlined,
  PlusOutlined,
  ReloadOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import dayjs, { type Dayjs } from 'dayjs'
import { message } from '@/utils/messageHelper'
import { getToken } from '@/utils/storage'

const { Option } = Select
const { TextArea } = Input

export interface TaskCapabilities {
  can_adjust: boolean
  can_cancel: boolean
  can_complete: boolean
  can_view_changes: boolean
}


const CaseTaskPanel: React.FC<Props> = ({ caseID, assigneeOptions, currentUserId, focusTaskID }) => {
  const [rows, setRows] = React.useState<CaseTaskRow[]>([])
  const [loading, setLoading] = React.useState(false)
  const [loadError, setLoadError] = React.useState('')
  const [busyID, setBusyID] = React.useState<number | null>(null)
  const [creating, setCreating] = React.useState(false)
  const [title, setTitle] = React.useState('')
  const [detail, setDetail] = React.useState('')
  const [assigneeID, setAssigneeID] = React.useState<number | null>(null)
  const [dueDate, setDueDate] = React.useState<Dayjs | null>(null)
  const [createError, setCreateError] = React.useState('')
  const [completingTask, setCompletingTask] = React.useState<CaseTaskRow | null>(null)
  const [completionNote, setCompletionNote] = React.useState('')
  const [pendingCreate, setPendingCreate] = React.useState<PendingCreate | null>(null)
  const [pendingOutcome, setPendingOutcome] = React.useState<PendingOutcome>('')
  const [verifying, setVerifying] = React.useState(false)
  const [editingTask, setEditingTask] = React.useState<CaseTaskRow | null>(null)
  const [editingMode, setEditingMode] = React.useState<'adjust' | 'cancel'>('adjust')
  const [editAssigneeID, setEditAssigneeID] = React.useState<number | null>(null)
  const [editDueDate, setEditDueDate] = React.useState<Dayjs | null>(null)
  const [editClearDue, setEditClearDue] = React.useState(false)
  const [editReason, setEditReason] = React.useState('')
  const [editBusy, setEditBusy] = React.useState(false)
  const [editError, setEditError] = React.useState('')
  const [expandedTaskID, setExpandedTaskID] = React.useState<number | null>(null)
  const [changeRows, setChangeRows] = React.useState<CaseTaskChangeRow[]>([])
  const [changeLoading, setChangeLoading] = React.useState(false)
  const [changeError, setChangeError] = React.useState('')
  const [editServerRow, setEditServerRow] = React.useState<CaseTaskRow | null>(null)
  const [editRefreshing, setEditRefreshing] = React.useState(false)
  // Task id of the verified deep-link focus target: set only after the row
  // has been proven to belong to this case, cleared on any case switch.
  const [focusedRowID, setFocusedRowID] = React.useState<number | null>(null)
  const idempotencyKeyRef = React.useRef(generateIdempotencyKey())
  // Scope generation: bumped on every caseID change. Each async callback
  // captures its generation and refuses to touch state after the user
  // moved to another case, so no late response bleeds into the new scope.
  const scopeGenRef = React.useRef(0)
  // History requests serialize: only the newest toggle may render.
  const changesReqRef = React.useRef(0)

  // Every in-flight request is registered here; switching cases or
  // unmounting aborts them so stale responses can never call setState
  // into the new case's scope.
  const abortRef = React.useRef<AbortController[]>([])
  const registerAbort = React.useCallback((controller: AbortController) => {
    abortRef.current.push(controller)
    return controller
  }, [])
  const abortPending = React.useCallback(() => {
    abortRef.current.forEach((controller) => controller.abort())
    abortRef.current = []
  }, [])

  const loadRows = React.useCallback(async () => {
    if (!caseID) return
    const gen = scopeGenRef.current
    const controller = registerAbort(new AbortController())
    const stale = () => gen !== scopeGenRef.current || controller.signal.aborted
    setLoading(true)
    setLoadError('')
    try {
      const payload = await taskRequest(`/cases/${caseID}/tasks`, { signal: controller.signal })
      if (stale()) return
      setRows(Array.isArray(payload?.tasks) ? payload.tasks : [])
    } catch (error) {
      if (stale()) return
      setLoadError(error instanceof Error ? error.message : '任务清单加载失败')
    } finally {
      if (!stale()) setLoading(false)
    }
  }, [caseID, registerAbort])

  // Deep link (inbox ?task_id=...): once this case's rows exist, expand and
  // highlight the requested task. The row must still belong to the case the
  // server returned for THIS panel (row.case_id === caseID), so an unknown,
  // stale, or foreign-case id can never expand another task or trigger a
  // cross-case history read. History loads through the normal toggleChanges
  // path, so no extra data route bypasses its permission handling. A task
  // whose capabilities deny history must not be auto-read by a deep link;
  // and when the requested id changes or disappears the previous highlight,
  // expansion, and in-flight history read are cleared instead of lingering.
  const focusedFocusIDRef = React.useRef<number | null>(null)
  const clearTaskFocus = () => {
    focusedFocusIDRef.current = null
    setFocusedRowID(null)
    setExpandedTaskID(null)
    setChangeRows([])
    setChangeLoading(false)
    setChangeError('')
    changesReqRef.current += 1
  }
  React.useEffect(() => {
    if (!focusTaskID) {
      if (focusedFocusIDRef.current !== null) clearTaskFocus()
      return
    }
    if (rows.length === 0) return
    if (focusedFocusIDRef.current === focusTaskID) return
    if (focusedFocusIDRef.current !== null) clearTaskFocus()
    const target = rows.find((row) => row.id === focusTaskID && row.case_id === caseID)
    if (!target) return
    focusedFocusIDRef.current = focusTaskID
    setFocusedRowID(target.id)
    if (capabilitiesOf(target).can_view_changes) {
      setExpandedTaskID(target.id)
      void toggleChanges(target)
    }
    const rowEl = document.querySelector<HTMLElement>("[data-testid='task-row-" + String(target.id) + "']")
    if (rowEl && typeof rowEl.scrollIntoView === 'function') {
      rowEl.scrollIntoView({ block: 'center', behavior: 'smooth' })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [focusTaskID, rows, caseID])
  React.useEffect(() => {
    // New case scope: kill every in-flight request and reset all state so
    // the previous case's rows, modals, pending-check banner, and form
    // inputs can never leak into this scope.
    scopeGenRef.current += 1
    abortPending()
    setRows([])
    setLoadError('')
    setBusyID(null)
    setCreating(false)
    setVerifying(false)
    setTitle('')
    setDetail('')
    setAssigneeID(null)
    setDueDate(null)
    setCreateError('')
    setPendingCreate(null)
    setPendingOutcome('')
    setCompletingTask(null)
    setCompletionNote('')
    setEditingTask(null)
    setEditingMode('adjust')
    setEditAssigneeID(null)
    setEditDueDate(null)
    setEditClearDue(false)
    setEditReason('')
    setEditBusy(false)
    setEditError('')
    setEditServerRow(null)
    setEditRefreshing(false)
    setExpandedTaskID(null)
    setFocusedRowID(null)
    focusedFocusIDRef.current = null
    setChangeRows([])
    setChangeLoading(false)
    setChangeError('')
    loadRows()
    return () => {
      abortPending()
    }
  }, [loadRows, abortPending])

  const createTask = async (override?: PendingCreate) => {
    let bodyObj: Record<string, unknown>
    let key: string
    if (override) {
      bodyObj = override.body
      key = override.requestId
    } else {
      const trimmedTitle = title.trim()
      if (!trimmedTitle) {
        setCreateError('任务标题必填')
        return
      }
      if (!assigneeID) {
        setCreateError('请选择任务负责人')
        return
      }
      key = idempotencyKeyRef.current
      bodyObj = {
        title: trimmedTitle,
        detail: detail.trim() || undefined,
        assignee_id: assigneeID,
        due_date: dueDate ? dueDate.toISOString() : undefined,
        idempotency_key: key,
      }
    }
    setCreating(true)
    setCreateError('')
    const controller = registerAbort(new AbortController())
    const gen = scopeGenRef.current
    const stale = () => gen !== scopeGenRef.current || controller.signal.aborted
    try {
      const payload = await taskCreateRequest(`/cases/${caseID}/tasks`, {
        method: 'POST',
        body: JSON.stringify(bodyObj),
        signal: controller.signal,
      })
      if (stale()) return
      // A 200 without a usable task row is NOT a success: the write may
      // or may not have landed, so the outcome is unknown and the pending
      // banner stays up instead of silently clearing the form.
      // The envelope is {task: {...}} (or the task row itself on older
      // deployments); validate the actual task row, not the wrapper.
      const createdTask = payload?.task ?? payload
      if (!isValidCreatedTask(createdTask, caseID)) {
        setPendingCreate({ requestId: key, body: bodyObj })
        setPendingOutcome('unknown')
        setCreateError(
          '服务器返回内容异常，无法确认任务是否已创建。原内容已保留，请先核对创建结果；不会重新创建，也不会自动重试。',
        )
        return
      }
      idempotencyKeyRef.current = generateIdempotencyKey()
      setPendingCreate(null)
      setPendingOutcome('')
      setTitle('')
      setDetail('')
      setDueDate(null)
      message.success('任务已创建，待办同步生成')
      await loadRows()
    } catch (error) {
      if (stale()) return
      const kind = classifyCreateFailure(error)
      if (kind === 'conflict') {
        // 409 idempotency conflict: the key may already have created a
        // task. Never rotate and never offer an instant reset here; the
        // SAME verification flow as an unknown outcome decides next.
        setPendingCreate({ requestId: key, body: bodyObj })
        setPendingOutcome('unknown')
        const raw = error instanceof Error ? error.message : ''
        setCreateError(
          (raw ? raw + ' ' : '') +
            '原请求可能已创建过任务。请先核对创建结果；暂未找到时可用原内容重试，请勿重新建同一任务。',
        )
      } else if (kind === 'rejected') {
        setPendingCreate(null)
        setPendingOutcome('')
        const text = error instanceof Error ? error.message : '任务创建失败'
        setCreateError(text)
        message.error(text)
      } else {
        setPendingCreate({ requestId: key, body: bodyObj })
        setPendingOutcome('')
        setCreateError(
          '创建请求未收到服务器确认（网络超时或中断）。原任务内容已保留，请先核对创建结果，避免重复创建。',
        )
      }
    } finally {
      // A stale old-case create must never clear the new case's in-flight
      // loading state; only this scope's own completion may reset it.
      if (!stale()) setCreating(false)
    }
  }

  const checkCreationResult = async () => {
    if (!pendingCreate) return
    setVerifying(true)
    setCreateError('')
    const gen = scopeGenRef.current
    const controller = registerAbort(new AbortController())
    const stale = () => gen !== scopeGenRef.current || controller.signal.aborted
    try {
      const payload = await taskRequest(
        `/cases/${caseID}/tasks/creation-result?request_id=${encodeURIComponent(pendingCreate.requestId)}`,
        { signal: controller.signal },
      )
      if (stale()) return
      const state = payload?.state
      const task = payload?.task
      // The SAME validator gates the create response and the found check:
      // a found without a usable same-case task cannot prove the write, so
      // the key only rotates on real evidence from either path.
      if (state === 'found' && isValidCreatedTask(task, caseID)) {
        // The original create actually landed: show the original task and
        // start fresh with a new key. This is the only rotation point, and
        // it happens because the key is now provably consumed.
        setPendingOutcome('found')
        setPendingCreate(null)
        idempotencyKeyRef.current = generateIdempotencyKey()
        setTitle('')
        setDetail('')
        setDueDate(null)
        setAssigneeID(null)
        message.success('已找到先前创建的任务')
        await loadRows()
      } else if (state === 'not_found') {
        // A lookup cannot rule out an original write still in flight. Keep
        // the frozen body and key; only replay that same request manually.
        setPendingOutcome('not_found')
        setCreateError('本次查询暂未找到原任务。可用原内容重试；原请求可能仍在处理，请勿重新建同一任务。')
      } else if (state === 'found') {
        // state === 'found' but the payload is not a usable same-case task.
        setPendingOutcome('unknown')
        setCreateError(
          '核对结果缺少有效任务信息，暂无法确认创建结果。原内容已保留，请稍后再次核对；不会重新创建，也不会自动重试。',
        )
      } else {
        // unknown: keep the pending state; the user can check again later.
        setPendingOutcome('unknown')
        setCreateError(
          '暂时无法确认创建结果。原内容已保留，请稍后再次核对；不会重新创建，也不会自动重试。',
        )
      }
    } catch (error) {
      if (stale()) return
      setCreateError(error instanceof Error ? error.message : '核对创建结果失败')
      setPendingOutcome((prev) => prev || 'unknown')
    } finally {
      if (!stale()) setVerifying(false)
    }
  }

  const retryPendingCreate = () => {
    if (!pendingCreate) return
    void createTask(pendingCreate)
  }

  const completeTask = async () => {
    if (!completingTask) return
    setBusyID(completingTask.id)
    const gen = scopeGenRef.current
    const controller = registerAbort(new AbortController())
    const stale = () => gen !== scopeGenRef.current || controller.signal.aborted
    try {
      await taskRequest(
        `/case-tasks/${completingTask.id}/complete`,
        {
          method: 'POST',
          body: JSON.stringify(
            completionNote.trim() ? { completion_note: completionNote.trim() } : {},
          ),
          signal: controller.signal,
        },
      )
      if (stale()) return
      message.success('任务已完成，关联待办同步关闭')
      setCompletingTask(null)
      setCompletionNote('')
      await loadRows()
    } catch (error) {
      if (stale()) return
      message.error(error instanceof Error ? error.message : '完成任务失败')
    } finally {
      if (!stale()) setBusyID(null)
    }
  }

  const capabilitiesOf = (row: CaseTaskRow): TaskCapabilities => {
    if (row.capabilities) return row.capabilities
    // Back-compat fallback for servers that do not send capabilities:
    // keep the original complete-button inference and hide the newer
    // adjust/cancel/history actions. The server still re-validates.
    return {
      can_complete:
        row.status === 'open' &&
        currentUserId !== undefined &&
        (currentUserId === row.assignee_id || currentUserId === row.created_by),
      can_adjust: false,
      can_cancel: false,
      can_view_changes: false,
    }
  }

  const openEdit = (row: CaseTaskRow, mode: 'adjust' | 'cancel') => {
    setEditingTask(row)
    setEditingMode(mode)
    setEditAssigneeID(row.assignee_id)
    setEditDueDate(row.due_date ? dayjs(row.due_date) : null)
    setEditClearDue(false)
    setEditReason('')
    setEditError('')
  }

  const submitEdit = async () => {
    const row = editingTask
    if (!row) return
    if (editingMode === 'cancel' && !editReason.trim()) {
      setEditError('取消任务请填写原因')
      return
    }
    const body: Record<string, unknown> = {}
    if (typeof row.version === 'number') {
      body.version = row.version
    } else if (row.updated_at) {
      body.expected_updated_at = row.updated_at
    } else {
      setEditError('无法核对最新任务内容，请刷新任务清单后重试')
      return
    }
    if (editingMode === 'cancel') {
      body.reason = editReason.trim()
    } else {
      if (editClearDue) {
        body.clear_due_date = true
      } else if (editDueDate) {
        body.due_date = editDueDate.toISOString()
      }
      if (editAssigneeID && editAssigneeID !== row.assignee_id) {
        body.assignee_id = editAssigneeID
      }
      body.reason = editReason.trim()
      if (!('assignee_id' in body) && !('due_date' in body) && !('clear_due_date' in body)) {
        setEditError('请至少修改负责人或截止时间')
        return
      }
    }
    setEditBusy(true)
    setEditError('')
    const controller = registerAbort(new AbortController())
    const gen = scopeGenRef.current
    const reqID = ++changesReqRef.current
    const stale = () => gen !== scopeGenRef.current || controller.signal.aborted
    try {
      const isCancel = editingMode === 'cancel'
      await taskRequest(isCancel ? `/case-tasks/${row.id}/cancel` : `/case-tasks/${row.id}`, {
        method: isCancel ? 'POST' : 'PUT',
        body: JSON.stringify(body),
        signal: controller.signal,
      })
      if (stale()) return
      message.success(
        isCancel ? '任务已取消，关联待办同步关闭' : '任务已调整，变更已记录',
      )
      setEditingTask(null)
      await loadRows()
    } catch (error) {
      if (stale()) return
      const isTaskChangeError = error instanceof TaskCreateError
      const text = error instanceof Error ? error.message : '任务变更失败'
      setEditError(text)
      message.error(text)
      // The submit request belongs to this latest interaction; close out the
      // trailing loading state explicitly so it cannot mask a newer one.
      if (reqID === changesReqRef.current) setEditBusy(false)
      const isStaleChange =
        (isTaskChangeError && error.code === 'CASE_TASK_CHANGE_STALE') ||
        text.includes('CASE_TASK_CHANGE_STALE')
      if (isStaleChange) {
        // Read the latest server row for the next manual submission, keep
        // the user's edits, and never resubmit automatically.
        setEditRefreshing(true)
        try {
          const list = await taskRequest(`/cases/${caseID}/tasks`, {
            signal: controller.signal,
          })
          if (stale()) return
          const fresh = (Array.isArray(list?.tasks) ? (list.tasks as CaseTaskRow[]) : []).find(
            (item) => item?.id === row.id,
          )
          if (fresh) {
            setEditServerRow(fresh)
            setEditingTask(fresh)
          } else {
            setEditServerRow(null)
          }
        } catch {
          if (stale()) return
          setEditServerRow(null)
        } finally {
          if (!stale()) setEditRefreshing(false)
        }
        setEditError(
          '任务已被他人修改。已读取最新任务内容，请对照后再修改；不会自动重提，只有下一次手动提交才按最新任务内容处理。',
        )
        return
      }
      void loadRows()
    } finally {
      if (reqID === changesReqRef.current && !stale()) setEditBusy(false)
    }
  }

  const toggleChanges = async (row: CaseTaskRow) => {
    if (expandedTaskID === row.id) {
      setExpandedTaskID(null)
      return
    }
    setExpandedTaskID(row.id)
    setChangeRows([])
    setChangeError('')
    setChangeLoading(true)
    const controller = registerAbort(new AbortController())
    const reqID = ++changesReqRef.current
    const gen = scopeGenRef.current
    const stale = () => gen !== scopeGenRef.current || controller.signal.aborted
    try {
      const payload = await taskRequest(`/case-tasks/${row.id}/changes`, {
        signal: controller.signal,
      })
      if (reqID !== changesReqRef.current || stale()) return
      setChangeRows(Array.isArray(payload?.changes) ? payload.changes : [])
    } catch (error) {
      if (reqID !== changesReqRef.current || stale()) return
      setChangeError(error instanceof Error ? error.message : '变更记录加载失败')
    } finally {
      if (reqID === changesReqRef.current && !stale()) setChangeLoading(false)
    }
  }

  const personName = (id?: number) => {
    if (id === undefined || id === null) return ''
    return assigneeOptions.find((option) => option.id === id)?.name || `用户 ${id}`
  }


  return (
    <div className='case-task-panel' data-testid='case-task-panel'>
      {loadError && (
        <Alert
          type='error'
          showIcon
          message='任务清单加载失败'
          description={
            <Space direction='vertical'>
              <span>{loadError}</span>
              <Button size='small' icon={<ReloadOutlined />} onClick={() => void loadRows()}>
                重试
              </Button>
            </Space>
          }
          style={{ marginBottom: 12 }}
        />
      )}
      {loading ? (
        <p style={{ color: '#999' }}>任务清单加载中…</p>
      ) : (
        <div className='batch-table-wrap'>
          <table>
            <thead>
              <tr>
                <th>任务</th>
                <th>负责人</th>
                <th>截止</th>
                <th>状态</th>
                <th>完成记录</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => {
                const capabilities = capabilitiesOf(row)
                const expanded = expandedTaskID === row.id
                // Highlight only the verified focus target: a focus id that
                // never validated against this case's rows stays unlit even
                // if its number coincides with a rendered row.
                const isFocusTarget = focusedRowID === row.id
                return (
                  <React.Fragment key={row.id}>
                    <tr
                      data-testid={`task-row-${row.id}`}
                      className={isFocusTarget ? 'task-row-focused' : undefined}
                    >
                      <td>
                        <strong>{row.title}</strong>
                        {row.detail && (
                          <div style={{ color: '#666', fontSize: 12 }}>{row.detail}</div>
                        )}
                      </td>
                      <td>{personName(row.assignee_id)}</td>
                      <td>{row.due_date ? dayjs(row.due_date).format('YYYY-MM-DD') : '未设期限'}</td>
                      <td>
                        <Tag
                          color={
                            row.status === 'open'
                              ? 'processing'
                              : row.status === 'completed'
                                ? 'success'
                                : 'default'
                          }
                        >
                          {statusText(row.status)}
                        </Tag>
                      </td>
                      <td>
                        {row.status === 'completed' && row.completed_at ? (
                          <Space direction='vertical' size={0}>
                            <span style={{ fontSize: 12 }}>
                              {personName(row.completed_by)} ·{' '}
                              {dayjs(row.completed_at).format('YYYY-MM-DD HH:mm')}
                            </span>
                            {row.completion_note && (
                              <span style={{ fontSize: 12, color: '#666' }}>
                                {row.completion_note}
                              </span>
                            )}
                          </Space>
                        ) : (
                          <span style={{ color: '#999', fontSize: 12 }}>—</span>
                        )}
                      </td>
                      <td>
                        {row.status === 'open' && capabilities.can_complete && (
                          <Button
                            size='small'
                            type='link'
                            icon={<CheckCircleOutlined />}
                            loading={busyID === row.id}
                            disabled={busyID === row.id}
                            title='完成后关联待办会一并关闭'
                            onClick={() => {
                              setCompletingTask(row)
                              setCompletionNote('')
                            }}
                          >
                            去完成
                          </Button>
                        )}
                        {row.status === 'open' && capabilities.can_adjust && (
                          <Button
                            size='small'
                            type='link'
                            icon={<SettingOutlined />}
                            onClick={() => openEdit(row, 'adjust')}
                          >
                            调整
                          </Button>
                        )}
                        {row.status === 'open' && capabilities.can_cancel && (
                          <Button
                            size='small'
                            type='link'
                            icon={<CloseCircleOutlined />}
                            onClick={() => openEdit(row, 'cancel')}
                          >
                            取消
                          </Button>
                        )}
                        {capabilities.can_view_changes && (
                          <Button
                            size='small'
                            type='link'
                            icon={<HistoryOutlined />}
                            onClick={() => void toggleChanges(row)}
                          >
                            {expanded ? '收起变更' : '变更记录'}
                          </Button>
                        )}
                      </td>
                    </tr>
                    {expanded && (
                      <tr>
                        <td colSpan={6} style={{ background: '#fafafa' }}>
                          {changeLoading && <span style={{ color: '#999' }}>变更记录加载中…</span>}
                          {changeError && <span style={{ color: '#cf1322' }}>{changeError}</span>}
                          {!changeLoading && !changeError && changeRows.length === 0 && (
                            <span style={{ color: '#999' }}>暂无变更记录。</span>
                          )}
                          {changeRows.length > 0 && (
                            <ul style={{ margin: 0, paddingLeft: 18 }}>
                              {changeRows.map((change) => (
                                <li key={change.id} style={{ fontSize: 12, marginBottom: 4 }}>
                                  {dayjs(change.created_at).format('YYYY-MM-DD HH:mm')} ·{' '}
                                  {changeTypeText(change.change_type)} ·{' '}
                                  {personName(change.actor_id)}
                                  {change.new_assignee_id !== undefined &&
                                    change.old_assignee_id !== undefined &&
                                    change.new_assignee_id !== change.old_assignee_id &&
                                    `（负责人：${personName(change.old_assignee_id)} → ${personName(change.new_assignee_id)}）`}
                                  {(change.old_due_date !== undefined ||
                                    change.new_due_date !== undefined) &&
                                    `（期限：${
                                      change.old_due_date
                                        ? dayjs(change.old_due_date).format('YYYY-MM-DD')
                                        : '未设'
                                    } → ${
                                      change.new_due_date
                                        ? dayjs(change.new_due_date).format('YYYY-MM-DD')
                                        : '已清除'
                                    }）`}
                                  {change.reason ? ` · 原因：${change.reason}` : ''}
                                </li>
                              ))}
                            </ul>
                          )}
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                )
              })}
              {rows.length === 0 && !loadError && (
                <tr>
                  <td colSpan={6} style={{ color: '#999' }}>
                    暂无案件任务。可在下方创建通用任务事项。
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}


      <Form layout='inline' style={{ marginTop: 12 }} component='div'>
        <Form.Item label='新任务' style={{ marginBottom: 8 }}>
          <Input
            value={title}
            disabled={pendingCreate !== null}
            onChange={(event) => setTitle(event.target.value)}
            placeholder='任务标题（必填）'
            style={{ width: 220 }}
            status={createError ? 'error' : undefined}
          />
        </Form.Item>
        <Form.Item style={{ marginBottom: 8 }}>
          <Input
            value={detail}
            disabled={pendingCreate !== null}
            onChange={(event) => setDetail(event.target.value)}
            placeholder='补充说明（可选）'
            style={{ width: 200 }}
          />
        </Form.Item>
        <Form.Item style={{ marginBottom: 8 }}>
          <Select
            value={assigneeID ?? undefined}
            disabled={pendingCreate !== null}
            onChange={(value) => setAssigneeID(value ?? null)}
            placeholder='负责人'
            style={{ width: 140 }}
          >
            {assigneeOptions.map((option) => (
              <Option key={option.id} value={option.id}>
                {option.name}
              </Option>
            ))}
          </Select>
        </Form.Item>
        <Form.Item style={{ marginBottom: 8 }}>
          <DatePicker
            value={dueDate}
            disabled={pendingCreate !== null}
            onChange={(value) => setDueDate(value)}
            placeholder='截止（律师确认）'
            style={{ width: 160 }}
          />
        </Form.Item>
        <Form.Item style={{ marginBottom: 8 }}>
          <Button
            type='primary'
            size='small'
            icon={<PlusOutlined />}
            loading={creating}
            disabled={creating || pendingCreate !== null}
            onClick={() => void createTask()}
          >
            创建任务
          </Button>
        </Form.Item>
        {createError && (
          <div style={{ width: '100%', marginBottom: 8 }}>
            <Alert type='error' showIcon message={createError} />
            {pendingCreate && pendingOutcome !== 'found' && (
              <Space style={{ marginTop: 8 }}>
                <Button
                  size='small'
                  icon={<ReloadOutlined />}
                  loading={verifying}
                  onClick={() => void checkCreationResult()}
                >
                  核对创建结果
                </Button>
                {pendingOutcome === 'not_found' && (
                  <Button size='small' type='primary' loading={creating} onClick={retryPendingCreate}>
                    用相同内容重试
                  </Button>
                )}
                {pendingOutcome === 'unknown' && (
                  <span style={{ color: '#999', fontSize: 12 }}>
                    原内容已保留，核对期间暂不能修改；不会自动重试。
                  </span>
                )}
              </Space>
            )}
          </div>
        )}
      </Form>
      <p style={{ color: '#999', fontSize: 12, marginTop: 8 }}>
        截止日期是律师确认的工作期限，不参与法定期限计算。任务创建、完成与调整会同步个人待办并留下变更记录。
      </p>
      <Modal
        title={completingTask ? `确认完成：${completingTask.title}` : '确认完成'}
        open={completingTask !== null}
        okText='确认完成'
        cancelText='取消'
        confirmLoading={busyID !== null}
        onCancel={() => {
          setCompletingTask(null)
          setCompletionNote('')
        }}
        onOk={() => void completeTask()}
      >
        <p style={{ color: '#666' }}>完成后关联待办会一并关闭。</p>
        <TextArea
          value={completionNote}
          onChange={(event) => setCompletionNote(event.target.value)}
          placeholder='完成记录（可选）：做了什么、结论或下一步'
          rows={3}
          maxLength={500}
          showCount
        />
      </Modal>
      <Modal
        title={
          editingTask
            ? editingMode === 'cancel'
              ? `取消任务：${editingTask.title}`
              : `调整任务：${editingTask.title}`
            : '任务变更'
        }
        open={editingTask !== null}
        okText={editingMode === 'cancel' ? '确认取消任务' : '提交变更'}
        cancelText='取消'
        confirmLoading={editBusy}
        onCancel={() => {
          setEditingTask(null)
          setEditError('')
        }}
        onOk={() => void submitEdit()}
      >
        <Space direction='vertical' style={{ width: '100%' }} size={12}>
          {editServerRow && (
            <Alert
              type='warning'
              showIcon
              message='任务已被他人修改，请对照最新内容后重试'
              description={(
                <span style={{ fontSize: 12 }}>
                  最新负责人：{personName(editServerRow.assignee_id)}；最新期限：
                  {editServerRow.due_date
                    ? dayjs(editServerRow.due_date).format('YYYY-MM-DD')
                    : '未设期限'}
                  。你在本窗口填写的原因与修改内容已保留，
                  下一次手动提交会基于最新任务内容处理，不会自动重提。
                </span>
              )}
            />
          )}
          {editRefreshing && (
            <p style={{ color: '#999', fontSize: 12, margin: 0 }}>正在读取最新任务内容…</p>
          )}
          {editingMode === 'adjust' && (
            <>
              <Form.Item label='负责人' style={{ marginBottom: 0 }}>
                <Select
                  value={editAssigneeID ?? undefined}
                  onChange={(value) => setEditAssigneeID(value ?? null)}
                  style={{ width: 200 }}
                >
                  {assigneeOptions.map((option) => (
                    <Option key={option.id} value={option.id}>
                      {option.name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
              <Form.Item label='截止时间' style={{ marginBottom: 0 }}>
                <Space>
                  <DatePicker
                    value={editClearDue ? null : editDueDate}
                    disabled={editClearDue}
                    onChange={(value) => setEditDueDate(value)}
                  />
                  <Checkbox
                    checked={editClearDue}
                    onChange={(event) => setEditClearDue(event.target.checked)}
                  >
                    清除截止时间
                  </Checkbox>
                </Space>
              </Form.Item>
            </>
          )}
          <Form.Item label='变更原因' style={{ marginBottom: 0 }}>
            <TextArea
              value={editReason}
              onChange={(event) => setEditReason(event.target.value)}
              placeholder={
                editingMode === 'cancel' ? '请说明取消原因（必填）' : '调整原因（可选）'
              }
              rows={3}
              maxLength={2000}
              showCount
            />
          </Form.Item>
          {editError && <Alert type='error' showIcon message={editError} />}
          <p style={{ color: '#999', fontSize: 12, margin: 0 }}>
            若任务已被他人修改，会读取最新任务内容供你对照；请核对后手动提交，不会自动重试。
          </p>
        </Space>
      </Modal>
    </div>
  )
}

export default CaseTaskPanel




export interface CaseTaskRow {
  id: number
  case_id: number
  title: string
  detail?: string
  assignee_id: number
  created_by: number
  due_date?: string
  status: string
  completed_by?: number
  completed_at?: string
  completion_note?: string
  updated_at?: string
  version?: number
  capabilities?: TaskCapabilities
}

export interface CaseTaskChangeRow {
  id: number
  created_at: string
  change_type: string
  actor_id: number
  old_assignee_id?: number
  new_assignee_id?: number
  old_due_date?: string
  new_due_date?: string
  old_version: number
  new_version: number
  reason?: string
}

export interface TaskAssigneeOption {
  id: number
  name: string
}

interface Props {
  caseID: number
  assigneeOptions: TaskAssigneeOption[]
  currentUserId?: number
  focusTaskID?: number
}

function statusText(status: string) {
  const labels: Record<string, string> = {
    open: '进行中',
    completed: '已完成',
    cancelled: '已取消',
  }
  return labels[status] || status
}

function changeTypeText(changeType: string) {
  const labels: Record<string, string> = {
    reassign: '改派',
    due_date_change: '截止调整',
    cancel: '取消',
  }
  return labels[changeType] || changeType
}

function isValidCreatedTask(task: unknown, caseID: number): boolean {
  if (!task || typeof task !== 'object') return false
  const row = task as { id?: unknown; case_id?: unknown }
  if (typeof row.id !== 'number' || !Number.isInteger(row.id) || row.id <= 0) return false
  // The creation-result endpoint is new, so there is no legacy omission to
  // excuse: the row must prove a positive integer id in this exact scope.
  if (row.case_id !== caseID) return false
  return true
}

type CreateFailureKind = 'conflict' | 'rejected' | 'unknown'

function classifyCreateFailure(error: unknown): CreateFailureKind {
  if (!(error instanceof TaskCreateError)) return 'unknown'
  const status = error.status
  if (error.code === 'IDEMPOTENCY_CONFLICT' || status === 409) return 'conflict'
  // Only an explicit server rejection (4xx) proves nothing was written.
  // 5xx/504 and anything else keep the outcome unknown.
  return status >= 400 && status < 500 ? 'rejected' : 'unknown'
}

async function taskRequest(path: string, init: RequestInit = {}): Promise<any> {
  const token = getToken()
  const response = await fetch(`/api/v1${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init.headers || {}),
    },
  })
  const body = (await response.json().catch(() => ({}))) as {
    success?: boolean
    message?: string
    error?: { message?: string; details?: string | string[]; code?: string }
    data?: any
  }
  if (!response.ok || body.success === false) {
    throw new TaskCreateError(
      body.error?.message || body.message || `任务操作失败（${response.status}）`,
      body.error?.code || '',
      response.status,
    )
  }
  return body.data ?? body
}

class TaskCreateError extends Error {
  code: string
  status: number
  constructor(messageText: string, code: string, status: number) {
    super(messageText)
    this.code = code
    this.status = status
  }
}

async function taskCreateRequest(path: string, init: RequestInit = {}): Promise<any> {
  const token = getToken()
  const response = await fetch(`/api/v1${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init.headers || {}),
    },
  })
  const body = (await response.json().catch(() => ({}))) as {
    success?: boolean
    message?: string
    error?: { message?: string; details?: string | string[]; code?: string }
    data?: any
  }
  if (!response.ok || body.success === false) {
    throw new TaskCreateError(
      body.error?.message || body.message || `任务操作失败（${response.status}）`,
      body.error?.code || (response.status === 409 ? 'IDEMPOTENCY_CONFLICT' : ''),
      response.status,
    )
  }
  return body.data ?? body
}

function generateIdempotencyKey() {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  return `task-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

type PendingOutcome = '' | 'not_found' | 'unknown' | 'found'

interface PendingCreate {
  requestId: string
  body: Record<string, unknown>
}
