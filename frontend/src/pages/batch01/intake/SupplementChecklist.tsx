import React from 'react'
import { Alert, Button, Card, Input, Tag, Typography } from 'antd'
import type { ApprovalMaterialEvidence } from '@/services/approval'

// Structured return checklist (plan §6.1/§6.2): one immutable description
// per server-generated item, one plain-text reply field per item. The
// component never renders item content as HTML and never invents items.
export interface SupplementItemFE {
  item_id: string
  category: string
  field_key: string
  material_id?: number
  description: string
  required: boolean
  source?: string
}

export function supplementMaterialEvidenceState(item: SupplementItemFE, evidence?: ApprovalMaterialEvidence): ApprovalMaterialEvidence['state'] {
  if (!evidence || evidence.item_id !== item.item_id || evidence.material_id !== item.material_id ||
    typeof item.material_id !== 'number' || !Number.isSafeInteger(item.material_id) || item.material_id <= 0 ||
    !['complete', 'missing', 'unknown'].includes(evidence.state)) return 'unknown'
  if (evidence.state === 'complete' && (typeof evidence.current_file_id !== 'number' || !Number.isSafeInteger(evidence.current_file_id) || evidence.current_file_id <= 0)) return 'unknown'
  return evidence.state
}

const categoryLabels: Record<string, string> = {
  field_correction: '字段更正',
  identity_verification: '身份核验',
  matter_explanation: '事项说明',
  material_request: '材料补充',
}

const fieldLabels: Record<string, string> = {
  client_name: '客户名称',
  client_contact: '客户联系方式',
  case_title: '案件名称',
  case_summary: '案情摘要',
  case_type: '案件类型',
  case_area: '业务领域',
  responsible_lawyer: '负责律师',
  party: '当事人',
}

interface Props {
  items: SupplementItemFE[]
  answers: Record<string, string>
  onChange: (itemID: string, value: string) => void
  materialNames?: Record<number, string>
  // Current file name per material row from the server ledger. When a
  // material already has a file on record, the checklist names it so a
  // replacement request is recognizably about a NEW file.
  materialFileNames?: Record<number, string>
  onLocateMaterial?: (item: SupplementItemFE) => void
  // Live ledger state only describes an existing attachment. The separate
  // return-item evidence determines whether the correction was uploaded.
  materialStatuses?: Record<number, string>
  materialEvidence?: Record<string, ApprovalMaterialEvidence>
  // Pending items are the APPLICANT's to-do, not the reviewer's queue.
  // The intake page says 待补充; an approver-facing reuse can override it.
  waitingLabel?: string
  // Jump to the form field this correction item targets: switch to the
  // step that owns it, scroll to it and focus it. No fallback row is
  // invented when the field cannot be found.
  onLocateField?: (item: SupplementItemFE) => void
  onRefreshMaterialStatus?: () => void
  disabled?: boolean
  saving?: boolean
  savedAt?: string
}

const SupplementChecklist: React.FC<Props> = ({
  items,
  answers,
  onChange,
  materialNames,
  materialFileNames,
  onLocateMaterial,
  materialStatuses,
  materialEvidence,
  waitingLabel = '待补充',
  onLocateField,
  onRefreshMaterialStatus,
  disabled,
  saving,
  savedAt,
}) => {
  const answered = items.filter((it) => (answers[it.item_id] || '').trim() !== '').length
  if (items.length === 0) {
    return null
  }
  return (
    <Card
      size='small'
      title='退回补充清单'
      extra={
        <Typography.Text type={answered === items.length ? 'success' : 'secondary'}>
          已回复 {answered}/{items.length}
          {saving ? ' · 保存中…' : savedAt ? ` · 已自动保存 ${savedAt}` : ''}
        </Typography.Text>
      }
      style={{ marginBottom: 16 }}
    >
      <Alert
        type='info'
        showIcon
        message='请逐项回复退回意见；回复会自动保存，提交时作为本轮回复，审批人将结合修改后的表单复核。'
        style={{ marginBottom: 12 }}
      />
      {items.map((item) => {
        const value = answers[item.item_id] || ''
        const missing = item.required && value.trim() === ''
        const isMaterialItem = item.category === 'material_request'
        const materialName =
          isMaterialItem && item.material_id
            ? materialNames?.[item.material_id] || `材料 #${item.material_id}`
            : ''
        // The ledger describes the current file, not whether THIS return
        // requirement was satisfied after its frozen baseline. Until trusted
        // return-item evidence is supplied, uploaded + a reply stays unknown.
        const rawMaterialStatus = (
          isMaterialItem && item.material_id
            ? materialStatuses?.[item.material_id] || ''
            : ''
        ).toLowerCase()
        const materialUploaded = rawMaterialStatus === 'uploaded'
        const currentFileName = isMaterialItem && item.material_id
          ? materialFileNames?.[item.material_id] || ''
          : ''
        const replied = value.trim() !== ''
        const evidence = materialEvidence?.[item.item_id]
        const evidenceState = supplementMaterialEvidenceState(item, evidence)
        const statusLabel = isMaterialItem
          ? evidenceState === 'complete'
            ? '补正附件已上传，待审批复核'
            : evidenceState === 'missing'
              ? materialUploaded || evidence?.current_file_id
                ? '已有旧附件，尚未补交新附件'
                : '尚未补交补正附件'
              : '未能确认补正，请刷新材料状态或联系项目组'
          : replied
            ? '已回复，待审批复核'
            : waitingLabel
        const statusColor =
          isMaterialItem ? evidenceState === 'complete' ? 'blue' : 'orange' : replied ? 'blue' : 'default'
        return (
          <div key={item.item_id} style={{ marginBottom: 16 }}>
            <div style={{ marginBottom: 4 }}>
              <Tag color={isMaterialItem ? 'orange' : 'blue'}>
                {categoryLabels[item.category] || item.category}
              </Tag>
              {isMaterialItem ? (
                <Tag color='geekblue'>{materialName}</Tag>
              ) : (
                <Tag>{fieldLabels[item.field_key] || item.field_key}</Tag>
              )}
              {item.source === 'legacy_text' && <Tag color='default'>历史补充说明</Tag>}
              {item.required && <Tag color='red'>必答</Tag>}
              <Tag
                color={statusColor}
                data-testid={`supplement-status-${item.item_id}`}
              >
                {statusLabel}
              </Tag>
              {isMaterialItem && replied && <Tag color='blue'>已回复</Tag>}
            </div>
            <Typography.Paragraph type='secondary' style={{ marginBottom: 6 }}>
              {item.description}
            </Typography.Paragraph>
            {isMaterialItem && materialUploaded && currentFileName && (
              <Typography.Paragraph type='secondary' style={{ marginBottom: 6 }}>
                当前已登记文件：{currentFileName}。如审批要求换件，请上传新文件替换，勿重复提交原文件。
              </Typography.Paragraph>
            )}
            {!isMaterialItem && onLocateField && (
              <div style={{ marginBottom: 6 }}>
                <Button
                  size='small'
                  type='primary'
                  ghost
                  onClick={() => onLocateField(item)}
                  disabled={disabled}
                >
                  定位并填写该字段
                </Button>
              </div>
            )}
            {isMaterialItem && (
              <div style={{ marginBottom: 6 }}>
                <Button
                  size='small'
                  type='primary'
                  ghost
                  onClick={() => onLocateMaterial?.(item)}
                  disabled={disabled}
                >
                  去第 4 步上传该材料
                </Button>
                {onRefreshMaterialStatus && (
                  <Button size='small' onClick={onRefreshMaterialStatus} disabled={disabled}>
                    刷新材料状态
                  </Button>
                )}
                <Typography.Text
                  type='secondary'
                  style={{ marginLeft: 8, fontSize: 12 }}
                >
                  定位到“文档与材料”步骤，找到对应材料行后上传文件，再在下方填写回复并重新提交。
                </Typography.Text>
              </div>
            )}
            <Input.TextArea
              aria-label={`回复：${isMaterialItem ? materialName : fieldLabels[item.field_key] || item.field_key}`}
              value={value}
              disabled={disabled}
              rows={2}
              maxLength={2000}
              status={missing ? 'error' : undefined}
              placeholder={missing ? '该项必答，请填写回复' : '填写回复'}
              onChange={(e) => onChange(item.item_id, e.target.value)}
            />
          </div>
        )
      })}
    </Card>
  )
}

export default SupplementChecklist
