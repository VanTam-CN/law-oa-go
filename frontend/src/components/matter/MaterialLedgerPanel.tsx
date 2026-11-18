import React from 'react'
import {
  Alert,
  Button,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Spin,
  Switch,
  Upload,
} from 'antd'
import {
  CloudUploadOutlined,
  DownloadOutlined,
  HistoryOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import { message } from '@/utils/messageHelper'
import { getToken } from '@/utils/storage'

const { Option } = Select

// One case_materials row as returned by the matter-collaboration
// contract (2026-09-30). intake_id is a UUID string, case_id a uint;
// this component never joins them and never guesses a scope.
export interface CaseMaterialRow {
  id: number
  intake_id?: string
  case_id?: number
  name: string
  material_type?: string
  status: string
  required: boolean
  storage_url?: string
  owner_id?: number
  uploaded_by?: number
  uploaded_at?: string
  file_count?: number
  file_version?: number
  current_file?: MaterialFileVersion
}

export interface MaterialFileVersion {
  id: number
  file_name: string
  size_bytes: number
  content_type?: string
  content_sha256?: string
  uploaded_by: number
  uploaded_at: string
  version: number
  download_url: string
}

export interface MaterialOwnerOption {
  id: number
  name: string
}

export type MaterialScope = { kind: 'intake'; id: string } | { kind: 'case'; id: number }

interface Props {
  scope: MaterialScope
  ownerOptions: MaterialOwnerOption[]
  currentUserId?: number
}

function statusText(status: string) {
  const labels: Record<string, string> = {
    missing: '待补充',
    uploaded: '已上传',
    not_required: '无需提交',
  }
  return labels[status] || status
}

function typeText(value: unknown) {
  const labels: Record<string, string> = {
    identity: '主体身份证明',
    contract: '合同材料',
    evidence: '证据材料',
    application: '申请材料',
    conflict_report: '冲突检测报告',
    proof: '证明材料',
    document: '其他文档',
  }
  const key = String(value || '').toLowerCase()
  return labels[key] || String(value || '其他材料')
}

function listPath(scope: MaterialScope) {
  return scope.kind === 'intake'
    ? `/case-intakes/${encodeURIComponent(scope.id)}/materials`
    : `/cases/${scope.id}/materials`
}

function formatDateTime(value?: string) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function formatBytes(size: number) {
  if (!Number.isFinite(size) || size < 0) return '—'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

async function materialRequest(path: string, init: RequestInit = {}): Promise<any> {
  const token = getToken()
  const isFormData = init.body instanceof FormData
  const response = await fetch(`/api/v1${path}`, {
    ...init,
    headers: {
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init.headers || {}),
    },
  })
  const body = (await response.json().catch(() => ({}))) as {
    success?: boolean
    message?: string
    error?: { message?: string; details?: string | string[] }
    data?: any
  }
  if (!response.ok || body.success === false) {
    throw new Error(body.error?.message || body.message || `材料操作失败（${response.status}）`)
  }
  return body.data ?? body
}

// Downloads must never open storage_url directly: auth is a localStorage
// JWT (no session cookie), so a plain link would 403, and appending the
// token to the URL would leak it. We fetch the blob with the Authorization
// header instead — but only for the panel's own internal download path
// (/api/v1/case-materials/<id>/files/<fileId>/download). Any other URL is
// refused rather than forwarded, so a tampered list response cannot turn
// this component into a token exfiltration channel.
function internalMaterialDownloadPath(storageURL: string, expectedRowID: number): string | null {
  try {
    const parsed = new URL(storageURL, window.location.origin)
    if (parsed.origin !== window.location.origin) return null
    // The download API takes no query/hash; anything extra is suspicious.
    if (parsed.search || parsed.hash) return null
    const match = parsed.pathname.match(
      /^\/api\/v1\/case-materials\/(\d+)\/files\/(\d+)\/download$/,
    )
    // The material segment must reference THIS row: a tampered or misrouted
    // list row must not download another material's attachment.
    if (!match || Number(match[1]) !== expectedRowID) return null
    return parsed.pathname
  } catch {
    return null
  }
}

const MaterialLedgerPanel: React.FC<Props> = ({ scope, ownerOptions, currentUserId }) => {
  const [rows, setRows] = React.useState<CaseMaterialRow[]>([])
  const [loading, setLoading] = React.useState(false)
  const [loadError, setLoadError] = React.useState('')
  const [busyID, setBusyID] = React.useState<number | null>(null)
  const [creating, setCreating] = React.useState(false)
  const [newName, setNewName] = React.useState('')
  const [newType, setNewType] = React.useState('document')
  const [newRequired, setNewRequired] = React.useState(true)
  const [newOwnerID, setNewOwnerID] = React.useState<number | undefined>(undefined)
  const [createError, setCreateError] = React.useState('')
  const [historyRow, setHistoryRow] = React.useState<CaseMaterialRow | null>(null)
  const [historyFiles, setHistoryFiles] = React.useState<MaterialFileVersion[]>([])
  const [historyLoading, setHistoryLoading] = React.useState(false)
  const [historyError, setHistoryError] = React.useState('')
  const [historyDownloadingID, setHistoryDownloadingID] = React.useState<number | null>(null)
  const scopeKey = listPath(scope)
  const hasScope = Boolean(scope.id)
  const mountedRef = React.useRef(false)
  const activeScopeRef = React.useRef(scopeKey)
  const listRequestSeq = React.useRef(0)
  const listAbortRef = React.useRef<AbortController | null>(null)
  const historyRequestSeq = React.useRef(0)
  const historyAbortRef = React.useRef<AbortController | null>(null)
  const historyDownloadAbortRef = React.useRef<AbortController | null>(null)

  const invalidateHistory = React.useCallback(() => {
    historyRequestSeq.current += 1
    historyAbortRef.current?.abort()
    historyAbortRef.current = null
    historyDownloadAbortRef.current?.abort()
    historyDownloadAbortRef.current = null
  }, [])

  const closeHistory = React.useCallback(() => {
    invalidateHistory()
    setHistoryRow(null)
    setHistoryFiles([])
    setHistoryError('')
    setHistoryLoading(false)
    setHistoryDownloadingID(null)
  }, [invalidateHistory])

  // Invalidate before the new scope can be used. Abort saves work, while
  // sequence/scope checks also reject a response that already completed.
  React.useLayoutEffect(() => {
    mountedRef.current = true
    activeScopeRef.current = scopeKey
    setRows([])
    setLoadError('')
    setLoading(false)
    closeHistory()
    return () => {
      mountedRef.current = false
      listRequestSeq.current += 1
      listAbortRef.current?.abort()
      listAbortRef.current = null
      invalidateHistory()
    }
  }, [scopeKey, closeHistory, invalidateHistory])

  const loadRows = React.useCallback(async () => {
    if (!hasScope || !mountedRef.current || activeScopeRef.current !== scopeKey) return
    listAbortRef.current?.abort()
    const controller = new AbortController()
    listAbortRef.current = controller
    const requestSeq = ++listRequestSeq.current
    const isCurrent = () =>
      mountedRef.current &&
      activeScopeRef.current === scopeKey &&
      listRequestSeq.current === requestSeq &&
      !controller.signal.aborted
    setLoading(true)
    setLoadError('')
    try {
      const payload = await materialRequest(scopeKey, { signal: controller.signal })
      if (isCurrent()) setRows(Array.isArray(payload?.materials) ? payload.materials : [])
    } catch (error) {
      if (isCurrent()) setLoadError(error instanceof Error ? error.message : '材料清单加载失败')
    } finally {
      if (isCurrent()) {
        setLoading(false)
        listAbortRef.current = null
      }
    }
  }, [scopeKey, hasScope])

  React.useEffect(() => {
    loadRows()
  }, [loadRows])

  const canWrite = scope.kind === 'intake' || Boolean(scope.id > 0)
  const uploaderName = (id: number) => {
    const name = ownerOptions.find((option) => option.id === id)?.name.trim()
    if (name) return name
    if (id > 0 && id === currentUserId) return '本人'
    return id > 0 ? `姓名暂缺（人员编号 ${id}）` : '上传者信息暂缺'
  }

  const createMaterial = async () => {
    const name = newName.trim()
    if (!name) {
      setCreateError('材料名称必填')
      return
    }
    setCreating(true)
    setCreateError('')
    try {
      await materialRequest(listPath(scope), {
        method: 'POST',
        body: JSON.stringify({
          name,
          material_type: newType,
          required: newRequired,
          ...(newOwnerID ? { owner_id: newOwnerID } : {}),
        }),
      })
      setNewName('')
      setNewOwnerID(undefined)
      message.success('材料已添加')
      await loadRows()
    } catch (error) {
      const text = error instanceof Error ? error.message : '新增材料失败'
      setCreateError(text)
      message.error(text)
    } finally {
      setCreating(false)
    }
  }

  const updateMaterial = async (materialID: number, body: Record<string, unknown>) => {
    setBusyID(materialID)
    try {
      await materialRequest(`/case-materials/${materialID}`, {
        method: 'PUT',
        body: JSON.stringify(body),
      })
      await loadRows()
    } catch (error) {
      message.error(error instanceof Error ? error.message : '更新材料失败')
    } finally {
      setBusyID(null)
    }
  }

  const uploadProps = (row: CaseMaterialRow) => ({
    name: 'file',
    showUploadList: false,
    beforeUpload: (file: File) => {
      if (file.size > 50 * 1024 * 1024) {
        message.error('单个文件不能超过 50MB')
        return Upload.LIST_IGNORE
      }
      void (async () => {
        setBusyID(row.id)
        try {
          const formData = new FormData()
          formData.append('file', file)
          await materialRequest(`/case-materials/${row.id}/files`, {
            method: 'POST',
            body: formData,
          })
          message.success(`已上传：${file.name}`)
        } catch (error) {
          message.error(error instanceof Error ? error.message : '上传失败')
        } finally {
          setBusyID(null)
          await loadRows()
        }
      })()
      return Upload.LIST_IGNORE
    },
  })

  const downloadMaterial = async (row: CaseMaterialRow) => {
    const path = row.storage_url ? internalMaterialDownloadPath(row.storage_url, row.id) : null
    if (!path) {
      message.error('附件下载地址无效，请重新上传或联系管理员')
      return
    }
    setBusyID(row.id)
    try {
      const token = getToken()
      const response = await fetch(path, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      })
      if (!response.ok) {
        let detail = `下载失败（${response.status}）`
        try {
          const body = await response.json()
          detail = body?.error?.message || body?.message || detail
        } catch {
          // keep status fallback
        }
        throw new Error(detail)
      }
      const blob = await response.blob()
      const url = URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = row.current_file?.file_name || row.name || 'material-file'
      document.body.appendChild(anchor)
      anchor.click()
      anchor.remove()
      URL.revokeObjectURL(url)
    } catch (error) {
      message.error(error instanceof Error ? error.message : '下载失败')
    } finally {
      setBusyID(null)
    }
  }

  const openHistory = async (row: CaseMaterialRow) => {
    if (!mountedRef.current || activeScopeRef.current !== scopeKey) return
    invalidateHistory()
    const requestSeq = historyRequestSeq.current
    const controller = new AbortController()
    historyAbortRef.current = controller
    const isCurrent = () =>
      mountedRef.current &&
      activeScopeRef.current === scopeKey &&
      historyRequestSeq.current === requestSeq &&
      !controller.signal.aborted
    setHistoryRow(row)
    setHistoryFiles([])
    setHistoryError('')
    setHistoryLoading(true)
    setHistoryDownloadingID(null)
    try {
      const payload = await materialRequest(`/case-materials/${row.id}/files`, {
        signal: controller.signal,
      })
      if (isCurrent()) setHistoryFiles(Array.isArray(payload?.files) ? payload.files : [])
    } catch (error) {
      if (isCurrent()) setHistoryError(error instanceof Error ? error.message : '附件历史加载失败')
    } finally {
      if (isCurrent()) {
        setHistoryLoading(false)
        historyAbortRef.current = null
      }
    }
  }

  // History downloads reuse the strict same-row URL validator: the
  // history API supplies the path, but the row binding check still runs
  // so a tampered history payload cannot fetch another material's file.
  const downloadHistoryFile = async (row: CaseMaterialRow, file: MaterialFileVersion) => {
    if (!mountedRef.current || activeScopeRef.current !== scopeKey) return
    const path = internalMaterialDownloadPath(file.download_url, row.id)
    if (!path) {
      message.error('附件下载地址无效，请重新上传或联系管理员')
      return
    }
    historyDownloadAbortRef.current?.abort()
    const controller = new AbortController()
    historyDownloadAbortRef.current = controller
    const requestSeq = historyRequestSeq.current
    const isCurrent = () =>
      mountedRef.current &&
      activeScopeRef.current === scopeKey &&
      historyRequestSeq.current === requestSeq &&
      !controller.signal.aborted
    setHistoryDownloadingID(file.id)
    try {
      const token = getToken()
      const response = await fetch(path, {
        signal: controller.signal,
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      })
      if (!isCurrent()) return
      if (!response.ok) {
        let detail = `下载失败（${response.status}）`
        try {
          const body = await response.json()
          detail = body?.error?.message || body?.message || detail
        } catch {
          // keep status fallback
        }
        throw new Error(detail)
      }
      const blob = await response.blob()
      if (!isCurrent()) return
      const url = URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = file.file_name || `material-${file.id}`
      document.body.appendChild(anchor)
      anchor.click()
      anchor.remove()
      URL.revokeObjectURL(url)
    } catch (error) {
      if (isCurrent()) message.error(error instanceof Error ? error.message : '下载失败')
    } finally {
      if (isCurrent()) {
        setHistoryDownloadingID(null)
        historyDownloadAbortRef.current = null
      }
    }
  }

  return (
    <div className='material-ledger-panel' data-testid='material-ledger-panel'>
      {loadError && (
        <Alert
          type='error'
          showIcon
          message='材料清单加载失败'
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
        <p style={{ color: '#999' }}>材料清单加载中…</p>
      ) : (
        <div className='batch-table-wrap'>
          <table>
            <thead>
              <tr>
                <th>材料名称</th>
                <th>类型</th>
                <th>必交</th>
                <th>状态</th>
                <th>责任人</th>
                <th>当前文件</th>
                <th>上传/下载</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.id} data-testid={`material-row-${row.id}`}>
                  <td>{row.name}</td>
                  <td>{typeText(row.material_type)}</td>
                  <td>{row.required ? '必交' : '选交'}</td>
                  <td>
                    <span className={row.status === 'uploaded' ? 'green-text' : 'orange-text'}>
                      {statusText(row.status)}
                    </span>
                  </td>
                  <td>
                    <Select
                      size='small'
                      style={{ minWidth: 120 }}
                      allowClear
                      placeholder='未指定'
                      value={row.owner_id || undefined}
                      disabled={!canWrite || busyID === row.id}
                      onChange={(value) => void updateMaterial(row.id, { owner_id: value ?? 0 })}
                    >
                      {ownerOptions.map((option) => (
                        <Option key={option.id} value={option.id}>
                          {option.name}
                        </Option>
                      ))}
                    </Select>
                  </td>
                  <td>
                    {row.current_file ? (
                      <div style={{ fontSize: 12, lineHeight: 1.6 }}>
                        <div
                          style={{ maxWidth: 220, overflow: 'hidden', textOverflow: 'ellipsis' }}
                          title={row.current_file.file_name}
                        >
                          {row.current_file.file_name}
                        </div>
                        <div style={{ color: '#999' }}>
                          {uploaderName(row.current_file.uploaded_by)} ·{' '}
                          {formatDateTime(row.current_file.uploaded_at)} · 第
                          {row.file_version || row.current_file.version}版
                          {row.file_count && row.file_count > 1 ? `（共${row.file_count}版）` : ''}
                        </div>
                      </div>
                    ) : (
                      <span style={{ color: '#999', fontSize: 12 }}>
                        {row.file_count && row.storage_url
                          ? '当前附件信息异常，请查看历史或联系项目组'
                          : '—'}
                      </span>
                    )}
                  </td>
                  <td>
                    <Space size='small'>
                      {row.status !== 'uploaded' && canWrite && (
                        <Upload {...uploadProps(row)}>
                          <Button
                            aria-label='上传'
                            size='small'
                            icon={<CloudUploadOutlined />}
                            loading={busyID === row.id}
                            disabled={busyID === row.id}
                          >
                            上传
                          </Button>
                        </Upload>
                      )}
                      {canWrite && row.status === 'uploaded' && (
                        <Upload {...uploadProps(row)}>
                          <Button
                            aria-label='上传新版本'
                            size='small'
                            icon={<CloudUploadOutlined />}
                            loading={busyID === row.id}
                            disabled={busyID === row.id}
                          >
                            上传新版本
                          </Button>
                        </Upload>
                      )}
                      {row.storage_url && (
                        <>
                          <Button
                            aria-label='下载'
                            size='small'
                            type='link'
                            icon={<DownloadOutlined />}
                            onClick={() => void downloadMaterial(row)}
                          >
                            下载
                          </Button>
                        </>
                      )}
                      {(row.storage_url || Boolean(row.file_count)) && (
                        <Button
                          aria-label='历史'
                          size='small'
                          type='link'
                          icon={<HistoryOutlined />}
                          onClick={() => void openHistory(row)}
                        >
                          历史
                        </Button>
                      )}
                      {row.status === 'uploaded' && !row.storage_url && (
                        <span style={{ color: '#999', fontSize: 12 }}>附件信息缺失</span>
                      )}
                    </Space>
                  </td>
                </tr>
              ))}
              {rows.length === 0 && !loadError && (
                <tr>
                  <td colSpan={7} style={{ color: '#999' }}>
                    暂无材料。可在下方添加材料。
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
      {canWrite && (
        <Form layout='inline' style={{ marginTop: 12 }} component='div'>
          <Form.Item label='新增材料' style={{ marginBottom: 8 }}>
            <Input
              value={newName}
              onChange={(event) => setNewName(event.target.value)}
              placeholder='材料名称（必填）'
              style={{ width: 220 }}
              status={createError ? 'error' : undefined}
            />
          </Form.Item>
          <Form.Item style={{ marginBottom: 8 }}>
            <Select value={newType} onChange={setNewType} style={{ width: 140 }}>
              {['document', 'identity', 'contract', 'evidence', 'application', 'proof'].map(
                (type) => (
                  <Option key={type} value={type}>
                    {typeText(type)}
                  </Option>
                ),
              )}
            </Select>
          </Form.Item>
          <Form.Item label='必交' style={{ marginBottom: 8 }}>
            <Switch checked={newRequired} onChange={setNewRequired} size='small' />
          </Form.Item>
          <Form.Item label='责任人' style={{ marginBottom: 8 }}>
            <Select
              value={newOwnerID}
              onChange={(value: number | undefined) => setNewOwnerID(value)}
              style={{ width: 160 }}
              allowClear
              placeholder='可稍后指定'
            >
              {ownerOptions.map((option) => (
                <Option key={option.id} value={option.id}>
                  {option.name}
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item style={{ marginBottom: 8 }}>
            <Space>
              <Button
                aria-label='添加材料'
                type='primary'
                size='small'
                icon={<PlusOutlined />}
                loading={creating}
                onClick={() => void createMaterial()}
              >
                添加材料
              </Button>
              <Popconfirm
                title='责任人下拉未列出需要的人？'
                description='下拉只显示已知案件人员。选择不在列表中的人时请先联系管理员；后端会拒绝非法责任人。'
                okText='知道了'
                cancelText='关闭'
              >
                <Button type='text' size='small'>
                  找不到责任人？
                </Button>
              </Popconfirm>
            </Space>
          </Form.Item>
          {createError && (
            <Alert type='error' showIcon message={createError} style={{ marginBottom: 8 }} />
          )}
        </Form>
      )}
      <p style={{ color: '#999', fontSize: 12, marginTop: 8 }}>
        上传后状态自动变为“已上传”；“已上传”只能通过实际上传获得，不能手动设置。
      </p>
      <Modal
        open={Boolean(historyRow)}
        title={historyRow ? `附件历史：${historyRow.name}` : '附件历史'}
        width={760}
        footer={
          <Button aria-label='关闭' onClick={closeHistory}>
            关闭
          </Button>
        }
        onCancel={closeHistory}
      >
        {historyLoading ? (
          <div style={{ textAlign: 'center', padding: '32px 0' }}>
            <Spin />
            <p style={{ marginTop: 12 }}>附件历史加载中…</p>
          </div>
        ) : historyError ? (
          <Alert
            type='error'
            showIcon
            message='附件历史加载失败'
            description={historyError}
            action={
              <Button
                aria-label='重试'
                size='small'
                icon={<ReloadOutlined />}
                onClick={() => historyRow && void openHistory(historyRow)}
              >
                重试
              </Button>
            }
          />
        ) : (
          <div className='batch-table-wrap'>
            <table>
              <thead>
                <tr>
                  <th>版本</th>
                  <th>文件名</th>
                  <th>大小</th>
                  <th>上传者</th>
                  <th>上传时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {historyFiles.map((file) => (
                  <tr key={file.id}>
                    <td>
                      第{file.version}版
                      {historyRow?.file_version === file.version && (
                        <span style={{ marginLeft: 6, color: '#389e0d' }}>当前</span>
                      )}
                    </td>
                    <td>{file.file_name}</td>
                    <td>{formatBytes(file.size_bytes)}</td>
                    <td>{uploaderName(file.uploaded_by)}</td>
                    <td>{formatDateTime(file.uploaded_at)}</td>
                    <td>
                      <Button
                        aria-label='下载'
                        size='small'
                        type='link'
                        icon={<DownloadOutlined />}
                        loading={historyDownloadingID === file.id}
                        onClick={() => historyRow && void downloadHistoryFile(historyRow, file)}
                      >
                        下载
                      </Button>
                    </td>
                  </tr>
                ))}
                {historyFiles.length === 0 && (
                  <tr>
                    <td colSpan={6} style={{ color: '#999' }}>
                      暂无历史附件。
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}
      </Modal>
    </div>
  )
}

export default MaterialLedgerPanel
