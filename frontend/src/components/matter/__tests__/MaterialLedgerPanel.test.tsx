import React from 'react'
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { message } from '@/utils/messageHelper'
import MaterialLedgerPanel from '../MaterialLedgerPanel'

jest.mock('@/utils/storage', () => ({
  getToken: () => 'test-token',
}))

jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() },
  setAppMessage: jest.fn(),
}))

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

function jsonResponse(data: unknown) {
  return { ok: true, json: async () => ({ success: true, data }) } as Response
}

function blobResponse(content: string) {
  return { ok: true, blob: async () => new Blob([content]) } as Response
}

const originalComputedStyle = window.getComputedStyle
const originalCreateObjectURL = Object.getOwnPropertyDescriptor(URL, 'createObjectURL')
const originalRevokeObjectURL = Object.getOwnPropertyDescriptor(URL, 'revokeObjectURL')

function deferredResponse() {
  let resolve!: (response: Response) => void
  let reject!: (error: Error) => void
  const promise = new Promise<Response>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

function historyMaterial(id: number, name: string) {
  return {
    id,
    name,
    status: 'uploaded',
    required: true,
    storage_url: `/api/v1/case-materials/${id}/files/${id * 10}/download`,
  }
}

function historyPayload(id: number, fileName: string, uploadedBy = 7) {
  return {
    files: [
      {
        id: id * 10,
        file_name: fileName,
        size_bytes: 1024,
        uploaded_by: uploadedBy,
        uploaded_at: '2026-10-01T10:00:00Z',
        version: 1,
        download_url: `/api/v1/case-materials/${id}/files/${id * 10}/download`,
      },
    ],
  }
}

describe('MaterialLedgerPanel (intake scope)', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    // jsdom has no pseudo-element layout or blob navigation. Keep the
    // real component and normal styles, with only those browser APIs stubbed.
    jest
      .spyOn(window, 'getComputedStyle')
      .mockImplementation((element) => originalComputedStyle(element))
    Object.defineProperty(URL, 'createObjectURL', {
      configurable: true,
      value: jest.fn(() => 'blob:material-test'),
    })
    Object.defineProperty(URL, 'revokeObjectURL', { configurable: true, value: jest.fn() })
    jest.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})
  })

  afterAll(() => {
    if (originalCreateObjectURL)
      Object.defineProperty(URL, 'createObjectURL', originalCreateObjectURL)
    else Reflect.deleteProperty(URL, 'createObjectURL')
    if (originalRevokeObjectURL)
      Object.defineProperty(URL, 'revokeObjectURL', originalRevokeObjectURL)
    else Reflect.deleteProperty(URL, 'revokeObjectURL')
  })

  it('lists real ledger rows with status and downloads via authorized fetch', async () => {
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path === '/api/v1/case-intakes/intake-9/materials') {
        return Promise.resolve(
          jsonResponse({
            materials: [
              {
                id: 1,
                intake_id: 'intake-9',
                name: '客户主体资料',
                material_type: 'identity',
                status: 'uploaded',
                required: true,
                storage_url: '/api/v1/case-materials/1/files/11/download',
                file_count: 2,
                file_version: 2,
                current_file: {
                  id: 11,
                  file_name: '主体资料-最新.pdf',
                  size_bytes: 2048,
                  uploaded_by: 7,
                  uploaded_at: '2026-10-01T10:00:00Z',
                  version: 2,
                  download_url: '/api/v1/case-materials/1/files/11/download',
                },
              },
              {
                id: 2,
                intake_id: 'intake-9',
                name: '初步证据目录',
                material_type: 'evidence',
                status: 'missing',
                required: false,
              },
            ],
          }),
        )
      }
      if (path === '/api/v1/case-materials/1/files/11/download') {
        return Promise.resolve(blobResponse('current fictional attachment'))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <MaterialLedgerPanel
        scope={{ kind: 'intake', id: 'intake-9' }}
        ownerOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )

    expect(await screen.findByText('客户主体资料')).toBeInTheDocument()
    expect(screen.getByText('初步证据目录')).toBeInTheDocument()
    expect(screen.getByText('已上传')).toBeInTheDocument()
    expect(screen.getByText('待补充')).toBeInTheDocument()
    expect(screen.getByText('主体资料-最新.pdf')).toBeInTheDocument()
    expect(screen.getByText(/律师甲 ·/)).toBeInTheDocument()
    expect(screen.getByText(/第2版（共2版）/)).toBeInTheDocument()
    // Download is an authorized fetch (Authorization header), never a raw
    // link that would 403 without the JWT cookie this project does not use.
    fireEvent.click(screen.getByRole('button', { name: /下载/ }))
    await waitFor(() => {
      const downloadCall = fetchMock.mock.calls.find(
        ([input]) => String(input) === '/api/v1/case-materials/1/files/11/download',
      )
      expect(downloadCall).toBeTruthy()
      const headers = new Headers((downloadCall?.[1] as RequestInit)?.headers as HeadersInit)
      expect(headers.get('Authorization')).toBe('Bearer test-token')
      expect(HTMLAnchorElement.prototype.click).toHaveBeenCalled()
    })
    // The missing row offers the first upload; the uploaded row offers an
    // explicit new-version upload on the same row.
    expect(screen.getByRole('button', { name: /^上传$/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /上传新版本/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /历史/ })).toBeInTheDocument()
  })

  it('opens version history and downloads a historical file through the authorized fetch', async () => {
    const downloadURLs: string[] = []
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path === '/api/v1/case-intakes/intake-9/materials') {
        return Promise.resolve(
          jsonResponse({
            materials: [
              {
                id: 1,
                intake_id: 'intake-9',
                name: '客户主体资料',
                material_type: 'identity',
                status: 'uploaded',
                required: true,
                storage_url: '/api/v1/case-materials/1/files/11/download',
                file_count: 2,
                file_version: 2,
                current_file: {
                  id: 11,
                  file_name: '主体资料-最新.pdf',
                  size_bytes: 2048,
                  uploaded_by: 7,
                  uploaded_at: '2026-10-01T10:00:00Z',
                  version: 2,
                  download_url: '/api/v1/case-materials/1/files/11/download',
                },
              },
            ],
          }),
        )
      }
      if (path === '/api/v1/case-materials/1/files') {
        return Promise.resolve(
          jsonResponse({
            files: [
              {
                id: 10,
                file_name: '主体资料-旧版.pdf',
                size_bytes: 1024,
                uploaded_by: 7,
                uploaded_at: '2026-09-30T09:00:00Z',
                version: 1,
                download_url: '/api/v1/case-materials/1/files/10/download',
              },
              {
                id: 11,
                file_name: '主体资料-最新.pdf',
                size_bytes: 2048,
                uploaded_by: 7,
                uploaded_at: '2026-10-01T10:00:00Z',
                version: 2,
                download_url: '/api/v1/case-materials/1/files/11/download',
              },
            ],
          }),
        )
      }
      if (path.includes('/download')) {
        downloadURLs.push(path)
        return Promise.resolve(blobResponse('old fictional attachment'))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <MaterialLedgerPanel
        scope={{ kind: 'intake', id: 'intake-9' }}
        ownerOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )

    fireEvent.click(await screen.findByRole('button', { name: /历史/ }))
    expect(await screen.findByText('主体资料-旧版.pdf')).toBeInTheDocument()
    // The newest file name appears twice: once in the ledger row's
    // current-file cell and once inside the history table.
    expect(screen.getAllByText('主体资料-最新.pdf').length).toBeGreaterThanOrEqual(2)
    expect(screen.getByText('当前')).toBeInTheDocument()

    const row = screen.getByText('主体资料-旧版.pdf').closest('tr')
    expect(row).toBeTruthy()
    fireEvent.click(within(row as HTMLElement).getByRole('button', { name: /下载/ }))
    await waitFor(() => {
      expect(downloadURLs).toEqual(['/api/v1/case-materials/1/files/10/download'])
      expect(HTMLAnchorElement.prototype.click).toHaveBeenCalled()
      const downloadCall = fetchMock.mock.calls.find(([input]) =>
        String(input).endsWith('/files/10/download'),
      )
      const headers = new Headers((downloadCall?.[1] as RequestInit)?.headers)
      expect(headers.get('Authorization')).toBe('Bearer test-token')
    })
  })

  it('shows the real authorization error when the ledger load fails', async () => {
    fetchMock.mockImplementation(() =>
      Promise.resolve({
        ok: false,
        status: 403,
        json: async () => ({ success: false, message: '无权访问该接案材料' }),
      } as Response),
    )

    render(
      <MaterialLedgerPanel scope={{ kind: 'intake', id: 'intake-denied' }} ownerOptions={[]} />,
    )

    expect(await screen.findByText('材料清单加载失败')).toBeInTheDocument()
    expect(screen.getByText('无权访问该接案材料')).toBeInTheDocument()
    // No fake rows are rendered on failure.
    expect(screen.queryByText('客户主体资料')).not.toBeInTheDocument()
  })

  it('refuses downloads whose storage_url points at another material row or carries query/hash', async () => {
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path === '/api/v1/case-intakes/intake-9/materials') {
        return Promise.resolve(
          jsonResponse({
            materials: [
              {
                id: 1,
                name: '被篡改行-异行ID',
                material_type: 'document',
                status: 'uploaded',
                required: true,
                storage_url: '/api/v1/case-materials/999/files/11/download',
              },
              {
                id: 2,
                name: '被篡改行-带参数',
                material_type: 'document',
                status: 'uploaded',
                required: true,
                storage_url: '/api/v1/case-materials/2/files/22/download?token=x',
              },
              {
                id: 3,
                name: '被篡改行-锚点',
                material_type: 'document',
                status: 'uploaded',
                required: true,
                storage_url: '/api/v1/case-materials/3/files/33/download#frag',
              },
            ],
          }),
        )
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(<MaterialLedgerPanel scope={{ kind: 'intake', id: 'intake-9' }} ownerOptions={[]} />)

    expect(await screen.findByText('被篡改行-异行ID')).toBeInTheDocument()

    // Every tampered URL is refused BEFORE any fetch fires: no download
    // request at all, and each click surfaces the real refusal message.
    for (const name of ['被篡改行-异行ID', '被篡改行-带参数', '被篡改行-锚点']) {
      const row = screen.getByText(name).closest('tr')
      expect(row).toBeTruthy()
      fireEvent.click(within(row as HTMLElement).getByRole('button', { name: /下载/ }))
    }
    await waitFor(() =>
      expect(
        fetchMock.mock.calls.filter(([input]) => String(input).includes('/download')),
      ).toHaveLength(0),
    )
    expect(message.error).toHaveBeenCalledWith('附件下载地址无效，请重新上传或联系管理员')
  })

  it('creates a material row with one-shot owner selection via POST', async () => {
    let created = false
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path === '/api/v1/cases/12/materials' && method === 'GET') {
        return Promise.resolve(
          jsonResponse({
            materials: created
              ? [
                  {
                    id: 5,
                    case_id: 12,
                    name: '新证据',
                    material_type: 'evidence',
                    status: 'missing',
                    required: true,
                  },
                ]
              : [],
          }),
        )
      }
      if (path === '/api/v1/cases/12/materials' && method === 'POST') {
        created = true
        expect(JSON.parse(String(init?.body))).toMatchObject({
          name: '新证据',
          material_type: 'document',
          required: true,
          owner_id: 7,
        })
        return Promise.resolve(jsonResponse({ material: { id: 5 } }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <MaterialLedgerPanel
        scope={{ kind: 'case', id: 12 }}
        ownerOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )
    await screen.findByText('暂无材料。可在下方添加材料。')

    fireEvent.change(screen.getByPlaceholderText('材料名称（必填）'), {
      target: { value: '新证据' },
    })
    // The create form renders the type select first and the owner select
    // last, so target the last combobox for the one-shot owner choice.
    const combos = screen.getAllByRole('combobox')
    fireEvent.mouseDown(combos[combos.length - 1])
    const ownerOption = await screen.findByText('律师甲', {
      selector: '.ant-select-item-option-content',
    })
    fireEvent.click(ownerOption)
    fireEvent.click(screen.getByRole('button', { name: /添加材料/ }))

    expect(await screen.findByText('新证据')).toBeInTheDocument()
  })

  it.each(['A after B', 'A before B', 'A error after B'])(
    'keeps only material B history when cancelled A settles: %s',
    async (order) => {
      const pendingA = deferredResponse()
      const pendingB = deferredResponse()
      let signalA: AbortSignal | undefined
      fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
        const path = String(input)
        if (path === '/api/v1/case-intakes/intake-9/materials') {
          return Promise.resolve(
            jsonResponse({
              materials: [historyMaterial(1, '材料A'), historyMaterial(2, '材料B')],
            }),
          )
        }
        if (path === '/api/v1/case-materials/1/files') {
          signalA = init?.signal as AbortSignal
          // Deliberately ignore abort here: the sequence guard must still
          // reject an old response even if the transport already completed.
          return pendingA.promise
        }
        if (path === '/api/v1/case-materials/2/files') return pendingB.promise
        return Promise.resolve(jsonResponse({}))
      })
      render(<MaterialLedgerPanel scope={{ kind: 'intake', id: 'intake-9' }} ownerOptions={[]} />)
      await screen.findByText('材料A')
      fireEvent.click(
        within(screen.getByTestId('material-row-1')).getByRole('button', { name: /历史/ }),
      )
      fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: '关闭' }))
      expect(signalA?.aborted).toBe(true)
      fireEvent.click(
        within(screen.getByTestId('material-row-2')).getByRole('button', { name: /历史/ }),
      )

      if (order === 'A before B') {
        await act(async () => {
          pendingA.resolve(jsonResponse(historyPayload(1, 'A旧附件.txt')))
        })
        expect(screen.getByText('附件历史加载中…')).toBeInTheDocument()
        expect(screen.queryByText('A旧附件.txt')).not.toBeInTheDocument()
      }
      await act(async () => {
        pendingB.resolve(jsonResponse(historyPayload(2, 'B附件.txt')))
      })
      expect(await screen.findByText('B附件.txt')).toBeInTheDocument()
      if (order === 'A after B') {
        await act(async () => {
          pendingA.resolve(jsonResponse(historyPayload(1, 'A旧附件.txt')))
        })
      } else if (order === 'A error after B') {
        await act(async () => {
          pendingA.reject(new Error('A迟到的错误'))
        })
      }
      const dialog = screen.getByRole('dialog')
      expect(within(dialog).getByText('附件历史：材料B')).toBeInTheDocument()
      expect(within(dialog).getByText('B附件.txt')).toBeInTheDocument()
      expect(screen.queryByText('A旧附件.txt')).not.toBeInTheDocument()
      expect(screen.queryByText('A迟到的错误')).not.toBeInTheDocument()
      expect(screen.queryByText('附件历史加载失败')).not.toBeInTheDocument()
    },
  )

  it('invalidates history on scope change and ignores the previous scope response', async () => {
    const pendingA = deferredResponse()
    let signalA: AbortSignal | undefined
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (path === '/api/v1/case-intakes/intake-9/materials') {
        return Promise.resolve(jsonResponse({ materials: [historyMaterial(1, '原接案材料')] }))
      }
      if (path === '/api/v1/cases/12/materials') {
        return Promise.resolve(jsonResponse({ materials: [historyMaterial(2, '新案件材料')] }))
      }
      if (path === '/api/v1/case-materials/1/files') {
        signalA = init?.signal as AbortSignal
        return pendingA.promise
      }
      if (path === '/api/v1/case-materials/2/files') {
        return Promise.resolve(jsonResponse(historyPayload(2, '新案件附件.txt')))
      }
      return Promise.resolve(jsonResponse({}))
    })
    const { rerender } = render(
      <MaterialLedgerPanel scope={{ kind: 'intake', id: 'intake-9' }} ownerOptions={[]} />,
    )
    fireEvent.click(await screen.findByRole('button', { name: /历史/ }))
    rerender(<MaterialLedgerPanel scope={{ kind: 'case', id: 12 }} ownerOptions={[]} />)
    expect(signalA?.aborted).toBe(true)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    await screen.findByText('新案件材料')
    expect(screen.queryByText('原接案材料')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /历史/ }))
    expect(await screen.findByText('新案件附件.txt')).toBeInTheDocument()
    await act(async () => {
      pendingA.resolve(jsonResponse(historyPayload(1, '原接案附件.txt')))
    })
    expect(within(screen.getByRole('dialog')).getByText('新案件附件.txt')).toBeInTheDocument()
    expect(screen.queryByText('原接案附件.txt')).not.toBeInTheDocument()
  })

  it('cancels pending history on unmount and ignores a response that arrives afterwards', async () => {
    const pending = deferredResponse()
    let signal: AbortSignal | undefined
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      if (String(input).endsWith('/materials')) {
        return Promise.resolve(jsonResponse({ materials: [historyMaterial(1, '材料A')] }))
      }
      signal = init?.signal as AbortSignal
      return pending.promise
    })
    const { unmount } = render(
      <MaterialLedgerPanel scope={{ kind: 'intake', id: 'intake-9' }} ownerOptions={[]} />,
    )
    fireEvent.click(await screen.findByRole('button', { name: /历史/ }))
    unmount()
    expect(signal?.aborted).toBe(true)
    await act(async () => {
      pending.resolve(jsonResponse(historyPayload(1, '过期附件.txt')))
    })
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(screen.queryByText('过期附件.txt')).not.toBeInTheDocument()
  })

  it('ignores a late ledger response from an old scope before a history row can be opened', async () => {
    const oldLedger = deferredResponse()
    let oldSignal: AbortSignal | undefined
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      if (String(input) === '/api/v1/case-intakes/intake-9/materials') {
        oldSignal = init?.signal as AbortSignal
        return oldLedger.promise
      }
      return Promise.resolve(jsonResponse({ materials: [historyMaterial(2, '新案件材料')] }))
    })
    const { rerender } = render(
      <MaterialLedgerPanel scope={{ kind: 'intake', id: 'intake-9' }} ownerOptions={[]} />,
    )
    rerender(<MaterialLedgerPanel scope={{ kind: 'case', id: 12 }} ownerOptions={[]} />)
    expect(oldSignal?.aborted).toBe(true)
    expect(await screen.findByText('新案件材料')).toBeInTheDocument()
    await act(async () => {
      oldLedger.resolve(jsonResponse({ materials: [historyMaterial(1, '原接案材料')] }))
    })
    expect(screen.queryByText('原接案材料')).not.toBeInTheDocument()
    expect(screen.getByTestId('material-row-2')).toBeInTheDocument()
    expect(screen.queryByTestId('material-row-1')).not.toBeInTheDocument()
  })

  it('offers a history retry and shows known uploader names, unknown fallback and Chinese versions', async () => {
    let attempts = 0
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      if (String(input).endsWith('/materials')) {
        return Promise.resolve(jsonResponse({ materials: [historyMaterial(1, '材料A')] }))
      }
      attempts += 1
      if (attempts === 1) {
        return Promise.resolve({
          ok: false,
          status: 403,
          json: async () => ({ success: false, message: '无权读取附件历史' }),
        } as Response)
      }
      return Promise.resolve(
        jsonResponse({
          files: [
            ...historyPayload(1, '已知上传者附件.txt').files,
            { ...historyPayload(1, '未知上传者附件.txt', 88).files[0], id: 11, version: 2 },
          ],
        }),
      )
    })
    render(
      <MaterialLedgerPanel
        scope={{ kind: 'intake', id: 'intake-9' }}
        ownerOptions={[{ id: 7, name: '律师甲' }]}
      />,
    )
    fireEvent.click(await screen.findByRole('button', { name: /历史/ }))
    expect(await screen.findByText('无权读取附件历史')).toBeInTheDocument()
    fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: '重试' }))
    expect(await screen.findByText('已知上传者附件.txt')).toBeInTheDocument()
    const dialog = screen.getByRole('dialog')
    expect(within(dialog).getByText('律师甲')).toBeInTheDocument()
    expect(within(dialog).getByText('姓名暂缺（人员编号 88）')).toBeInTheDocument()
    expect(within(dialog).getByText('第1版')).toBeInTheDocument()
    expect(within(dialog).getByText('第2版')).toBeInTheDocument()
    expect(within(dialog).queryByText('无权读取附件历史')).not.toBeInTheDocument()
    expect(attempts).toBe(2)
  })
})
