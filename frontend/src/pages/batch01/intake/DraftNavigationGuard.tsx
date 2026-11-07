import React from 'react'
import { Alert, Button, Modal } from 'antd'
import { UNSAFE_DataRouterContext, useBlocker } from 'react-router'

interface Props {
  dirty: boolean
  busy: boolean
  bypass: React.MutableRefObject<boolean>
  save: () => Promise<void>
}
function RouteGuard({ dirty, busy, bypass, save }: Props) {
  const blocker = useBlocker(() => dirty && !bypass.current)
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState('')
  return (
    <Modal
      title='尚有未保存修改'
      zIndex={1050}
      open={blocker.state === 'blocked'}
      onCancel={() => {
        if (!saving && !busy) {
          setError('')
          blocker.reset?.()
        }
      }}
      closable={!saving && !busy}
      maskClosable={false}
      footer={[
        <Button
          key='stay'
          disabled={saving || busy}
          onClick={() => {
            setError('')
            blocker.reset?.()
          }}
        >
          继续编辑
        </Button>,
        <Button key='leave' disabled={saving || busy} onClick={() => blocker.proceed?.()}>
          直接离开
        </Button>,
        <Button
          key='save'
          type='primary'
          loading={saving}
          disabled={busy}
          onClick={async () => {
            setSaving(true)
            setError('')
            try {
              await save()
              blocker.proceed?.()
            } catch (failure) {
              setError(failure instanceof Error ? failure.message : '保存失败，输入仍保留')
            } finally {
              setSaving(false)
            }
          }}
        >
          保存后离开
        </Button>,
      ]}
    >
      <p>请选择保存后离开、直接离开或继续编辑。身份号码仅保留在当前页面，离开后需重新填写。</p>
      {error && <Alert type='error' message={error} />}
    </Modal>
  )
}

export default function DraftNavigationGuard(props: Props) {
  const context = React.useContext(UNSAFE_DataRouterContext)
  React.useEffect(() => {
    const warn = (event: BeforeUnloadEvent) => {
      if (props.dirty && !props.bypass.current) {
        event.preventDefault()
        event.returnValue = ''
      }
    }
    window.addEventListener('beforeunload', warn)
    return () => window.removeEventListener('beforeunload', warn)
  }, [props.dirty, props.bypass])
  return context ? <RouteGuard {...props} /> : null
}
