import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, createBrowserRouter, RouterProvider } from 'react-router'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import './index.css'
import './assets/styles/design-tokens.css'
import App from './App'
import { initializeApp } from './stores/useAppStore'
import { intakeMemoryOnly } from './config/intakePilot'


const container = document.getElementById('root') as HTMLElement
const root = ReactDOM.createRoot(container)

function startApp() {
  initializeApp()
  root.render(
    <ConfigProvider locale={zhCN}>
      {intakeMemoryOnly
        ? <RouterProvider router={createBrowserRouter([{ path: '*', element: <App /> }])} />
        : <BrowserRouter><App /></BrowserRouter>}
    </ConfigProvider>,
  )
}

async function startPilot() {
  let blocked = false
  function hold() {
    if (blocked) return
    blocked = true
    root.unmount()
    container.dataset.npSwHold = 'blocked'
    container.textContent = '无法确认此浏览器可安全打开接案资料。请使用经核对的新浏览器配置。'
  }
  try {
    const workers = navigator.serviceWorker
    if (!workers || workers.controller) return hold()
    workers.addEventListener('controllerchange', hold)
    // A registration can cover the intake route even when the login page is uncontrolled.
    const registrations = await Promise.all([workers.getRegistration(), workers.getRegistration('/case/create')])
    if (blocked || workers.controller || registrations.some(Boolean)) return hold()
    startApp()
  } catch {
    hold()
  }
}

if (intakeMemoryOnly) void startPilot()
else startApp()
