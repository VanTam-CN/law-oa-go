import React from 'react'
import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import ClientCreateDialog from '../ClientCreateDialog'
import ClientSearchSelect from '../ClientSearchSelect'
import { createIntakeClient, getIntakeClientPage } from '@/services/client'

jest.mock('@/services/client', () => ({ createIntakeClient: jest.fn(), getIntakeClientPage: jest.fn() }))
const create = createIntakeClient as jest.Mock
const search = getIntakeClientPage as jest.Mock
beforeEach(() => {
  create.mockReset(); search.mockReset()
  Object.defineProperty(crypto, 'randomUUID', { configurable: true, value: jest.fn(() => 'test-request-key') })
})
afterEach(cleanup)

it('retains fields and the same key after an uncertain response; returns the created id and clears identity', async () => {
  create.mockRejectedValueOnce(new Error('响应中断')).mockResolvedValueOnce({ id: 71, name: '新客户', type: '企业' })
  const onCreated = jest.fn()
  const view = render(<ClientCreateDialog open onCancel={jest.fn()} onCreated={onCreated} />)
  fireEvent.change(screen.getByLabelText('客户名称'), { target: { value: '新客户' } })
  fireEvent.change(screen.getByLabelText('身份号码'), { target: { value: 'FICTIONAL-IDENTITY-71' } })
  fireEvent.click(screen.getByRole('button', { name: '保存客户' }))
  await screen.findByText('响应中断')
  expect(screen.getByLabelText('身份号码')).toHaveValue('FICTIONAL-IDENTITY-71')
  fireEvent.click(screen.getByRole('button', { name: '保存客户' }))
  await waitFor(() => expect(onCreated).toHaveBeenCalledWith(expect.objectContaining({ id: 71 })))
  expect(create.mock.calls[0][1]).toBe(create.mock.calls[1][1])
  expect(create.mock.calls[0][0]).toEqual(create.mock.calls[1][0])
  view.rerender(<ClientCreateDialog open={false} onCancel={jest.fn()} onCreated={onCreated} />)
  view.rerender(<ClientCreateDialog open onCancel={jest.fn()} onCreated={onCreated} />)
  expect(screen.getByLabelText('身份号码')).toHaveValue('')
})

it('cancels without creating a client and clears the reopened identity field', async () => {
  const cancel = jest.fn()
  render(<ClientCreateDialog open onCancel={cancel} onCreated={jest.fn()} />)
  fireEvent.change(screen.getByLabelText('身份号码'), { target: { value: 'FICTIONAL-CANCEL' } })
  fireEvent.click(screen.getByRole('button', { name: '取 消' }))
  expect(cancel).toHaveBeenCalledTimes(1)
  expect(create).not.toHaveBeenCalled()
  expect(screen.getByLabelText('身份号码')).toHaveValue('')
})

it('loads page two and selects a same-name client by its own id', async () => {
  search.mockImplementation((_term: string, page: number) => Promise.resolve({
    clients: page === 1 ? [{ id: 1, name: '同名客户', type: '企业' }] : [{ id: 51, name: '同名客户', type: '个人' }],
    pagination: { page, page_size: 50, total: 2 },
  }))
  const choose = jest.fn()
  render(<ClientSearchSelect value={{ id: 99, name: '刚创建客户', type: '企业' }} onChange={choose} />)
  await waitFor(() => expect(search).toHaveBeenCalledWith('', 1, expect.any(AbortSignal)))
  fireEvent.mouseDown(screen.getByRole('combobox'))
  fireEvent.click(await screen.findByRole('button', { name: '加载更多客户' }))
  await waitFor(() => expect(search).toHaveBeenCalledWith('', 2, expect.any(AbortSignal)))
  fireEvent.click(await screen.findByText('同名客户 · 个人 · 编号 51'))
  expect(choose).toHaveBeenCalledWith(expect.objectContaining({ id: 51 }))
})

it('ignores a delayed result from an older search', async () => {
  let resolveOld: (data: unknown) => void = () => undefined
  search.mockImplementation((query: string) => query === '旧' ? new Promise((resolve) => { resolveOld = resolve }) : Promise.resolve({ clients: query ? [{ id: 2, name: '新客户', type: '企业' }] : [], pagination: { total: 1 } }))
  render(<ClientSearchSelect onChange={jest.fn()} />)
  fireEvent.change(screen.getByRole('combobox'), { target: { value: '旧' } })
  await waitFor(() => expect(search).toHaveBeenCalledWith('旧', 1, expect.any(AbortSignal)))
  fireEvent.change(screen.getByRole('combobox'), { target: { value: '新' } })
  await screen.findByText('新客户 · 企业 · 编号 2')
  await act(async () => { resolveOld({ clients: [{ id: 1, name: '旧客户', type: '企业' }], pagination: { total: 1 } }) })
  await waitFor(() => expect(screen.queryByText('旧客户 · 企业 · 编号 1')).not.toBeInTheDocument())
})

it('submits only once while a double-clicked creation is pending', async () => {
  let finish: (value: unknown) => void = () => undefined
  create.mockImplementation(() => new Promise((resolve) => { finish = resolve }))
  const created = jest.fn()
  render(<ClientCreateDialog open onCancel={jest.fn()} onCreated={created} />)
  fireEvent.change(screen.getByLabelText('客户名称'), { target: { value: '并发测试客户' } })
  fireEvent.change(screen.getByLabelText('身份号码'), { target: { value: 'FICTIONAL-DOUBLE' } })
  const save = screen.getByRole('button', { name: '保存客户' })
  fireEvent.click(save); fireEvent.click(save)
  await waitFor(() => expect(create).toHaveBeenCalledTimes(1))
  await act(async () => { finish({ id: 93, name: '并发测试客户', type: '企业' }) })
  expect(created).toHaveBeenCalledTimes(1)
})
