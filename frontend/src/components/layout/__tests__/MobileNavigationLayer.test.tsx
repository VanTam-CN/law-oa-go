/** @jest-environment ./src/test/fetchJsdomEnvironment.cjs */
import React from 'react'
import fs from 'node:fs'
import path from 'node:path'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { createMemoryRouter, Outlet, RouterProvider } from 'react-router'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import Sidebar from '../Sidebar'
import DraftNavigationGuard from '@/pages/batch01/intake/DraftNavigationGuard'

jest.mock('@/stores/useAppStore', () => ({ useAppStore: () => ({ user: { roles: ['lawyer'], permissions: ['dashboard:view', 'case:manage'] } }) }))
jest.mock('@/utils/messageHelper', () => ({ message: { warning: jest.fn() } }))

let style: HTMLStyleElement
beforeAll(async () => {
  const less = require('less')
  const css = await Promise.all(['header.less', 'sidebar.less'].map(file => less.render(fs.readFileSync(path.resolve(__dirname, '..', file), 'utf8'))))
  style = document.createElement('style')
  style.textContent = fs.readFileSync(path.resolve(__dirname, '../../../assets/styles/design-tokens.css'), 'utf8') + css.map(result => result.css).join('\n')
  document.head.append(style)
})
afterAll(() => style.remove())
afterEach(cleanup)

function Shell() {
  const [open, setOpen] = React.useState(false)
  return <>
    <header className='app-header'><button onClick={() => setOpen(true)}>打开导航菜单</button></header>
    <Sidebar collapsed mobileOpen={open} onMobileClose={() => setOpen(false)} setCollapsed={() => undefined} />
    <Outlet />
  </>
}
function UnsavedIntake() {
  const [title, setTitle] = React.useState('')
  const bypass = React.useRef(false)
  return <>
    <input aria-label='案件名称' value={title} onChange={event => setTitle(event.target.value)} />
    <DraftNavigationGuard dirty={Boolean(title)} busy={false} bypass={bypass} save={async () => undefined} />
  </>
}
async function openBlockedNavigation() {
  const router = createMemoryRouter([{ element: <Shell />, children: [
    { path: '/case/create', element: <UnsavedIntake /> },
    { path: '/dashboard', element: <p>工作台目标</p> },
  ] }], { initialEntries: ['/case/create'] })
  render(<ConfigProvider locale={zhCN}><RouterProvider router={router} /></ConfigProvider>)
  fireEvent.change(screen.getByRole('textbox', { name: '案件名称' }), { target: { value: '未保存的事项' } })
  fireEvent.click(screen.getByRole('button', { name: '打开导航菜单' }))
  const drawer = await screen.findByRole('dialog')
  fireEvent.click(within(drawer).getByRole('menuitem', { name: /工作台/ }))
  const prompt = await screen.findByRole('dialog', { name: '尚有未保存修改' })
  return { router, drawer, prompt }
}

it('keeps the drawer above the fixed header and the unsaved prompt above the drawer, then allows closing after staying', async () => {
  const { router, drawer, prompt } = await openBlockedNavigation()
  const fixedZ = Number(getComputedStyle(document.documentElement).getPropertyValue('--z-index-fixed'))
  const modalZ = Number(getComputedStyle(document.documentElement).getPropertyValue('--z-index-modal'))
  const drawerZ = Number(getComputedStyle(drawer.closest('.ant-drawer')!).zIndex)
  const promptZ = Number(getComputedStyle(prompt.closest('.ant-modal-wrap')!).zIndex)
  expect(drawerZ).toBeGreaterThan(fixedZ)
  expect(promptZ).toBeGreaterThan(drawerZ)
  expect(promptZ).toBeLessThanOrEqual(modalZ)
  fireEvent.click(within(prompt).getByRole('button', { name: '继续编辑' }))
  await waitFor(() => expect(prompt).not.toBeVisible())
  expect(drawer).toBeVisible()
  fireEvent.click(within(drawer).getByRole('button', { name: '关闭' }))
  await waitFor(() => expect(drawer).not.toBeVisible())
  expect(screen.getByRole('textbox', { name: '案件名称' })).toHaveValue('未保存的事项')
  expect(router.state.location.pathname).toBe('/case/create')
})

it('still closes navigation only after the user permits leaving the unsaved form', async () => {
  const { router, drawer, prompt } = await openBlockedNavigation()
  fireEvent.click(within(prompt).getByRole('button', { name: '直接离开' }))
  await screen.findByText('工作台目标')
  await waitFor(() => expect(drawer).not.toBeVisible())
  expect(router.state.location.pathname).toBe('/dashboard')
})
