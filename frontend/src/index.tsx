import React from 'react'
import ReactDOM from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import './index.css'
import './assets/styles/design-tokens.css'
import App from './App'
import { initializeApp } from './stores/useAppStore'

initializeApp()

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <ConfigProvider locale={zhCN}>
    <RouterProvider router={createBrowserRouter([{ path: '*', element: <App /> }])} />
  </ConfigProvider>,
)
