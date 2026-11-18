import React from 'react'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { ClientMasterProfile } from '../Batch01Prototype'
import { useAppStore } from '@/stores/useAppStore'

jest.mock('@/utils/storage', () => ({
  getRoles: jest.fn(() => []),
  getToken: jest.fn(() => 'test-token'),
  getUserInfo: jest.fn(() => ({ id: 7, role: 'lawyer', roles: ['lawyer'], name: '律师甲' })),
}))

jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() },
  setAppMessage: jest.fn(),
}))

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

function jsonResponse(data: unknown, status = 200) {
  return { ok: status < 400, status, json: async () => ({ success: true, data }) } as Response
}

function failedResponse(status: number) {
  return { ok: false, status, json: async () => ({ success: false, message: '服务不可用' }) } as Response
}

const FIRST_PAGE = Array.from({ length: 20 }, (_, index) => ({
  id: index + 1,
  name: '客户' + String(index + 1).padStart(2, '0'),
  email: '',
}))

function clientName(id: number) {
  return '客户' + String(id).padStart(2, '0')
}

function masterProfile(id: number | string) {
  return {
    client: { id, name: clientName(Number(id)), type: '企业', status: 'active' },
    completeness: { score: 80 },
    primary_contact: null,
    related_parties: [],
    matter_history: [],
    conflict_history: [],
  }
}

function clientListPayload(rows: unknown[], total: number, page: number) {
  return {
    clients: rows,
    pagination: { page, page_size: 20, total, total_page: Math.max(1, Math.ceil(total / 20)) },
  }
}

function makeSearchResponse(searchTerm: string, page: number) {
  const twentyFirst = { id: 21, name: '第21条搜索客户', email: '' }
  const all = searchTerm ? [twentyFirst] : FIRST_PAGE.concat([twentyFirst])
  const start = (page - 1) * 20
  return clientListPayload(all.slice(start, start + 20), all.length, page)
}

function makeOlderListResponse(page: number) {
  return clientListPayload(FIRST_PAGE.slice(0, 20), 21, page)
}

function setupProfile(mode: 'server-search' | 'server-error') {
  fetchMock.mockReset()
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    const url = new URL(path, 'http://localhost:3000')
    if (url.pathname === '/api/v1/clients') {
      if (mode === 'server-error' && !url.searchParams.get('search')) {
        return Promise.resolve(failedResponse(500))
      }
      const search = url.searchParams.get('search') || ''
      const page = Number(url.searchParams.get('page') || '1')
      return Promise.resolve(jsonResponse(makeSearchResponse(search, page)))
    }
    if (url.pathname.endsWith('/master-profile')) {
      const id = url.pathname.split('/')[3]
      return Promise.resolve(jsonResponse(masterProfile(id)))
    }
    return Promise.resolve(jsonResponse({}))
  })
  render(
    <MemoryRouter>
      <ClientMasterProfile />
    </MemoryRouter>,
  )
}

describe('ClientMasterProfile server-backed search', () => {
  beforeEach(() => {
    window.localStorage.clear()
    useAppStore.setState({
      user: {
        id: 7,
        name: '律师甲',
        roles: ['lawyer'],
        permissions: ['client:manage'],
      },
      isAuthenticated: true,
    } as never)
  })

  it('retrieves the 21st visible client through server search', async () => {
    setupProfile('server-search')
    const search = await screen.findByPlaceholderText('搜索客户名称或邮箱')
    fireEvent.change(search, { target: { value: '第21条搜索客户' } })
    await screen.findByText('第21条搜索客户')
    expect(screen.queryByText('客户20')).not.toBeInTheDocument()
    const searchCalls = fetchMock.mock.calls.filter(([input]) => {
      const url = new URL(String(input), 'http://localhost:3000')
      return url.pathname === '/api/v1/clients' && Boolean(url.searchParams.get('search'))
    })
    expect(searchCalls.length).toBeGreaterThan(0)
    const called = new URL(String(searchCalls[0][0]), 'http://localhost:3000')
    expect(called.searchParams.get('search')).toBe('第21条搜索客户')
    expect(called.searchParams.get('page')).toBe('1')
    expect(called.searchParams.get('page_size')).toBe('20')
  })

 it('pages through the full visible list instead of clipping at 20 rows', async () => {
   setupProfile('server-search')
    await screen.findByText('客户01')
    expect(screen.queryByText('第21条搜索客户')).not.toBeInTheDocument()
   const listBody = document.querySelector('.batch-client-items')
   expect(listBody).toBeInTheDocument()
    const nextButton = document.querySelector('.ant-pagination-next') as HTMLButtonElement
    expect(nextButton).not.toBeNull()
    fireEvent.click(nextButton)
   await screen.findByText('第21条搜索客户')
   expect(screen.getByText(/共 21 条/)).toBeInTheDocument()
 })

  it('resets to page 1 when the user searches a new keyword after paging ahead', async () => {
    setupProfile('server-search')
    await screen.findByText('客户01')
    fireEvent.click(document.querySelector('.ant-pagination-next') as HTMLButtonElement)
    await screen.findByText('第21条搜索客户')
    fireEvent.change(screen.getByPlaceholderText('搜索客户名称或邮箱'), {
      target: { value: '第21条搜索客户' },
    })
    await waitFor(() => {
      const paged = fetchMock.mock.calls.filter(([input]) => {
        const url = new URL(String(input), 'http://localhost:3000')
        return (
          url.pathname === '/api/v1/clients' &&
          Boolean(url.searchParams.get('search')) &&
          url.searchParams.get('page') === '1'
        )
      })
      expect(paged.length).toBeGreaterThan(0)
    })
 })

  it(
   'refreshes the list through the same search-aware path after creating a client',
    async () => {
      setupProfile('server-search')
      await screen.findByText('客户01')
      fireEvent.change(screen.getByPlaceholderText('搜索客户名称或邮箱'), {
        target: { value: '第21条搜索客户' },
      })
      await screen.findByText('第21条搜索客户')
      const callsBeforeCreate = fetchMock.mock.calls.length
      fireEvent.click(screen.getByRole('button', { name: /新增客户/ }))
      fireEvent.change(screen.getByPlaceholderText('请输入客户名称'), {
        target: { value: '新建测试客户' },
      })
      fireEvent.change(screen.getByPlaceholderText('按营业执照填写'), {
        target: { value: '91330100TEST000001' },
      })
      fireEvent.click(screen.getByRole('button', { name: '保存客户' }))
      await waitFor(() => {
        expect(fetchMock.mock.calls.some(([, options]) => options?.method === 'POST')).toBe(true)
      })
      await waitFor(() => {
        expect(fetchMock.mock.calls.length).toBeGreaterThan(callsBeforeCreate + 1)
      })
      const refreshCalls = fetchMock.mock.calls.slice(callsBeforeCreate).filter(([input, options]) => {
        if (options?.method === 'POST') return false
        const url = new URL(String(input), 'http://localhost:3000')
        return url.pathname === '/api/v1/clients'
      })
      expect(refreshCalls.length).toBeGreaterThan(0)
      const refreshUrl = new URL(String(refreshCalls[0][0]), 'http://localhost:3000')
      expect(refreshUrl.searchParams.get('search')).toBe('第21条搜索客户')
    },
  )

 it('distinguishes load failure from an empty list', async () => {
   setupProfile('server-error')
   await screen.findByText(/客户列表加载失败/)
   expect(screen.queryByText('暂无数据库客户')).not.toBeInTheDocument()
 })

  it('offers a working retry button after a load failure', async () => {
    // First page request fails, retry succeeds.
    fetchMock.mockReset()
    let shouldFail = true
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = new URL(String(input), 'http://localhost:3000')
      if (url.pathname === '/api/v1/clients') {
        if (shouldFail) {
          return Promise.resolve(failedResponse(500))
        }
        const search = url.searchParams.get('search') || ''
        const page = Number(url.searchParams.get('page') || '1')
        return Promise.resolve(jsonResponse(makeSearchResponse(search, page)))
      }
      if (url.pathname.endsWith('/master-profile')) {
        const id = url.pathname.split('/')[3]
        return Promise.resolve(jsonResponse(masterProfile(id)))
      }
      return Promise.resolve(jsonResponse({}))
    })
    render(
      <MemoryRouter>
        <ClientMasterProfile />
      </MemoryRouter>,
    )
    await screen.findByText(/客户列表加载失败/)
    shouldFail = false
    fireEvent.click(screen.getByRole('button', { name: /重试/ }))
    await screen.findByText('客户01')
    expect(screen.queryByText(/客户列表加载失败/)).not.toBeInTheDocument()
  })

  it('drops out-of-order list responses so an older reply cannot overwrite a newer search', async () => {
    setupProfile('server-search')
    const search = await screen.findByPlaceholderText('搜索客户名称或邮箱')
    const resolvers: Array<(value: unknown) => void> = []
    fetchMock.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolvers.push(resolve)
        }),
    )
    act(() => {
      fireEvent.change(search, { target: { value: '第21条' } })
    })
    await waitFor(() => expect(resolvers.length).toBe(1))
    act(() => {
      fireEvent.change(search, { target: { value: '第21条搜索客户' } })
    })
    await screen.findByText('第21条搜索客户')
    act(() => {
      resolvers[0](jsonResponse(makeOlderListResponse(1)))
    })
    await waitFor(() => {
      expect(screen.queryByText('客户01')).not.toBeInTheDocument()
    })
    expect(screen.getByText('第21条搜索客户')).toBeInTheDocument()
  })
})
