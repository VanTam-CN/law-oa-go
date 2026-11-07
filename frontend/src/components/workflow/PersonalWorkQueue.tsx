import React from 'react'
import { Alert, Button, Space } from 'antd'
import { useNavigate } from 'react-router'
import { apiRequest } from '@/services/collaborationRequest'
import WorkflowActions, { workflowActionsOf, type WorkflowAction } from './WorkflowActions'

interface QueueItem {
  id: number
  title: string
  content?: string
  actions?: WorkflowAction[]
  workflow_action?: WorkflowAction
}
interface QueuePage {
  items: QueueItem[]
  pagination: { total: number }
  stats?: { needs_action: number; waiting: number }
}

export default function PersonalWorkQueue() {
  const navigate = useNavigate()
  const [bucket, setBucket] = React.useState('needs_action')
  const [page, setPage] = React.useState(1)
  const [result, setResult] = React.useState<QueuePage | null>(null)
  const [error, setError] = React.useState('')
  const [loading, setLoading] = React.useState(false)
  const [refresh, setRefresh] = React.useState(0)
  React.useEffect(() => {
    const controller = new AbortController()
    setLoading(true)
    setError('')
    setResult(null)
    apiRequest<QueuePage>(`/inbox?page=${page}&page_size=6&workflow_bucket=${bucket}`, {
      signal: controller.signal,
    })
      .then((data) => {
        if (!controller.signal.aborted) setResult(data)
      })
      .catch((failure) => {
        if (!controller.signal.aborted)
          setError(failure instanceof Error ? failure.message : '待办读取失败')
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false)
      })
    return () => controller.abort()
  }, [bucket, page, refresh])
  return (
    <section className='ng-panel'>
      <header className='ng-panel-head'>
        <div className='ng-panel-title'>我的待办</div>
        <Space wrap>
          {[
            ['needs_action', '需要我处理'],
            ['waiting', '等待他人'],
          ].map(([value, label]) => (
            <Button
              key={value}
              type={bucket === value ? 'primary' : 'default'}
              onClick={() => {
                setBucket(value)
                setPage(1)
              }}
            >
              {label}
              {result?.stats
                ? `（${value === 'needs_action' ? result.stats.needs_action : result.stats.waiting}）`
                : ''}
            </Button>
          ))}
        </Space>
      </header>
      {error && (
        <Alert
          type='error'
          message={error}
          action={<Button onClick={() => setRefresh((value) => value + 1)}>重试</Button>}
        />
      )}
      <div className='ng-todo-list'>
        {result?.items?.map((item) => (
          <article key={item.id} className='ng-todo-item'>
            <div className='ng-todo-body'>
              <strong>{item.title}</strong>
              <p>{item.content}</p>
              <WorkflowActions actions={workflowActionsOf(item)} />
            </div>
          </article>
        ))}
        {!loading && !error && !result?.items?.length && <p>当前没有此类事项</p>}
      </div>
      <Space wrap style={{ padding: 16 }}>
        <Button disabled={page === 1 || loading} onClick={() => setPage((value) => value - 1)}>
          上一页
        </Button>
        <span>
          第 {page} 页 · 共 {result?.pagination?.total ?? 0} 项
        </span>
        <Button
          disabled={loading || page * 6 >= (result?.pagination?.total ?? 0)}
          onClick={() => setPage((value) => value + 1)}
        >
          下一页
        </Button>
        <Button type='link' onClick={() => navigate('/inbox')}>
          查看全部待办
        </Button>
      </Space>
    </section>
  )
}
