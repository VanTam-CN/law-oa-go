import React from 'react'
import { Alert, Button, Select, Space } from 'antd'
import { getIntakeClientPage, type ClientSummary } from '@/services/client'

export function mergeClientChoices(current: ClientSummary[], incoming: ClientSummary[]) {
  const rows = new Map(current.map((row) => [row.id, row]))
  incoming.forEach((row) => {
    if (row.id > 0) rows.set(row.id, row)
  })
  return [...rows.values()]
}

export default function ClientSearchSelect({
  value,
  onChange,
  invalid,
}: {
  value?: ClientSummary
  onChange: (client: ClientSummary) => void
  invalid?: boolean
}) {
  const [query, setQuery] = React.useState('')
  const [rows, setRows] = React.useState<ClientSummary[]>([])
  const [page, setPage] = React.useState(0)
  const [total, setTotal] = React.useState(0)
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState('')
  const sequence = React.useRef(0)
  const request = React.useRef<AbortController | null>(null)
  const load = React.useCallback(async (term: string, nextPage: number) => {
    const seq = ++sequence.current
    request.current?.abort()
    const controller = new AbortController()
    request.current = controller
    setLoading(true)
    setError('')
    try {
      const result = await getIntakeClientPage(term, nextPage, controller.signal)
      if (seq !== sequence.current) return
      setRows((old) => (nextPage === 1 ? result.clients : mergeClientChoices(old, result.clients)))
      setTotal(result.pagination.total)
      setPage(nextPage)
    } catch (failure) {
      if (seq === sequence.current && !controller.signal.aborted)
        setError(failure instanceof Error ? failure.message : '客户搜索失败')
    } finally {
      if (seq === sequence.current) setLoading(false)
    }
  }, [])
  React.useEffect(() => {
    const timer = window.setTimeout(() => void load(query, 1), 250)
    return () => {
      window.clearTimeout(timer)
      ++sequence.current
      request.current?.abort()
    }
  }, [query, load])
  const options = mergeClientChoices(rows, value ? [value] : [])
  return (
    <Space direction='vertical' style={{ width: '100%' }}>
      <Select
        id='intake-client'
        aria-label='我方当事人（客户）'
        aria-invalid={invalid}
        status={invalid ? 'error' : undefined}
        style={{ width: '100%', minWidth: 180 }}
        showSearch
        filterOption={false}
        loading={loading}
        value={value?.id}
        placeholder='搜索并选择客户'
        onSearch={(term) => {
          if (term === query) return
          ++sequence.current
          request.current?.abort()
          setQuery(term)
          setRows([])
          setPage(0)
          setTotal(0)
        }}
        onChange={(id) => {
          const client = options.find((row) => row.id === id)
          if (client) onChange(client)
        }}
        options={options.map((row) => ({
          value: row.id,
          label: `${row.name} · ${row.type} · 编号 ${row.id}`,
        }))}
        popupRender={(menu) => (
          <>
            {menu}
            {rows.length < total && (
              <Button
                block
                loading={loading}
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => void load(query, page + 1)}
              >
                加载更多客户
              </Button>
            )}
          </>
        )}
      />
      {error && (
        <Alert
          type='error'
          message={error}
          action={<Button onClick={() => void load(query, 1)}>重试</Button>}
        />
      )}
    </Space>
  )
}
