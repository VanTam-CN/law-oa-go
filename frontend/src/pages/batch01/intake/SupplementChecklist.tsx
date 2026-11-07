import React from 'react'
import { Alert, Button, Input, Space } from 'antd'
import {
  supplementFields,
  type SupplementReturn,
  type SupplementResponse,
  type SupplementItem,
} from '@/services/intakeSupplement'

export default function SupplementChecklist({
  returned,
  responses,
  onChange,
  onLocate,
  onSave,
  saving,
}: {
  returned: SupplementReturn
  responses: SupplementResponse[]
  onChange: (responses: SupplementResponse[]) => void
  onLocate: (item: SupplementItem) => void
  onSave: () => void
  saving: boolean
}) {
  return (
    <section className='batch-supplement-checklist' aria-label='本轮补正要求'>
      <h2>本轮补正要求</h2>
      <p style={{ whiteSpace: 'pre-wrap' }}>{returned.legacyText}</p>
      {!returned.items.length && (
        <Alert
          type='info'
          message='历史补充说明'
          description='此记录只有文字意见，请在修改说明中回复。不会补造历史条目。'
        />
      )}
      <Space direction='vertical' style={{ width: '100%' }}>
        {returned.items.map((item) => (
          <article key={item.item_id}>
            <p style={{ whiteSpace: 'pre-wrap' }}>{item.description}</p>
            {(item.field_key || item.party_local_id) && (
              <Button
                disabled={
                  Boolean(item.field_key && !supplementFields[item.field_key]) &&
                  !item.party_local_id
                }
                onClick={() => onLocate(item)}
              >
                定位
                {item.field_key
                  ? supplementFields[item.field_key]?.label || '字段（暂不支持）'
                  : '当事人'}
              </Button>
            )}
            <label htmlFor={`reply-${item.item_id}`}>
              补正回复{item.required === false ? '（选填）' : '（必填）'}
            </label>
            <Input.TextArea
              id={`reply-${item.item_id}`}
              value={responses.find((reply) => reply.item_id === item.item_id)?.response || ''}
              rows={3}
              maxLength={2000}
              onChange={(event) =>
                onChange([
                  ...responses.filter((reply) => reply.item_id !== item.item_id),
                  { item_id: item.item_id, response: event.target.value },
                ])
              }
            />
          </article>
        ))}
        {!!returned.items.length && (
          <Button loading={saving} onClick={onSave}>
            保存逐项回复
          </Button>
        )}
      </Space>
      <p>回复不代表身份已核验或冲突复核已通过；是否需要重新检测由服务端核验。</p>
    </section>
  )
}
