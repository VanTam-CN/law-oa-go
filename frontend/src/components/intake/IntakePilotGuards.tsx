import React from 'react'
import { intakeLogoutEvent } from '@/config/intakePilot'
import { useBlocker } from 'react-router'
import { getToken, getTokenExpiry, getUserInfo } from '@/utils/storage'

const discardPrompt = '本页还有未保存的接案内容。离开会丢弃这些内容，确定离开吗？'


export function IntakeLeaveGuard({ dirty }: { dirty: () => boolean }) {
  const isDirty = React.useRef(dirty)
  isDirty.current = dirty
  const mayLeave = () => !getToken() || !isDirty.current() || window.confirm(discardPrompt)
  const blocker = useBlocker(() => !mayLeave())
  React.useEffect(() => {
    if (blocker.state === 'blocked') blocker.reset()
  }, [blocker])
  React.useEffect(() => {
    const logout = (e: Event) => { if (!mayLeave()) e.preventDefault() }
    const unload = (e: BeforeUnloadEvent) => {
      if (!getToken() || !isDirty.current()) return
      e.preventDefault()
      e.returnValue = ''
    }
    window.addEventListener(intakeLogoutEvent, logout)
    window.addEventListener('beforeunload', unload)
    return () => {
      window.removeEventListener(intakeLogoutEvent, logout)
      window.removeEventListener('beforeunload', unload)
    }
  }, [])
  return null
}

export function IntakeSessionBoundary({ children }: { children: React.ReactNode }) {
  const [session] = React.useState(() => ({ token: getToken(), actor: getUserInfo()?.id }))
  const [ended, setEnded] = React.useState(false)
  React.useLayoutEffect(() => {
    const verify = () => {
      if (!session.token || getToken() !== session.token || getUserInfo()?.id !== session.actor) setEnded(true)
    }
    verify()
    const deadline = getTokenExpiry(session.token)
    let timer: number
    const scheduleExpiry = () => {
      const remaining = deadline - Date.now() - 30_000
      if (remaining <= 0) setEnded(true)
      else timer = window.setTimeout(scheduleExpiry, Math.min(remaining, 2_147_483_647))
    }
    scheduleExpiry()
    window.addEventListener('storage', verify)
    window.addEventListener('pageshow', verify)
    window.addEventListener('focus', verify)
    document.addEventListener('visibilitychange', verify)
    return () => {
      window.clearTimeout(timer)
      window.removeEventListener('storage', verify)
      window.removeEventListener('pageshow', verify)
      window.removeEventListener('focus', verify)
      document.removeEventListener('visibilitychange', verify)
    }
  }, [session])
  if (ended || !session.token || !session.actor) return <p role="status">登录已到期或账号已切换。请重新进入接案工作台。</p>
  return <>{children}</>
}
