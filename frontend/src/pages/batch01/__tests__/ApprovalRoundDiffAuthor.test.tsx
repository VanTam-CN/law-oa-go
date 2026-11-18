import React from 'react'
import { render, screen } from '@testing-library/react'
import { ApprovalRoundDiff } from '../Batch01Prototype'

const frozen = { facts_baseline: 'frozen', material_baseline: 'frozen' }

it('shows frozen field changes with Chinese party labels and never displays identity material', () => {
  render(<ApprovalRoundDiff diff={{ round_diff_version: 2, baseline: frozen, material_changes: [], field_changes: [
    { field: 'description', before: '上轮摘要', after: '本轮摘要' },
    { field: 'party_1_name', before: '旧名称', after: '新名称' },
    { field: 'party_1_role', before: 'client', after: 'opposing_party' },
    { field: 'party_1_identity_changed', before: 'SECRET-DIGEST-OLD', after: 'SECRET-CIPHER-NEW' },
    { field: 'party_2_identity_digest', before: 'SECRET-DIGEST-OLD', after: 'SECRET-CIPHER-NEW' },
  ] }} />)
  expect(screen.getByTestId('approval-round-diff')).toHaveTextContent('上轮摘要 → 本轮摘要')
  expect(screen.getByText('当事人 1 · 名称：')).toBeInTheDocument()
  expect(screen.getByText('当事人 1 · 角色：')).toBeInTheDocument()
  expect(screen.getByText('身份标识已变更')).toBeInTheDocument()
  expect(document.body.textContent).not.toMatch(/SECRET-|party_2_identity_digest/)
})

it('shows attachment names, carried files and counts without internal pointer numbers', () => {
  render(<ApprovalRoundDiff diff={{ round_diff_version: 2, baseline: frozen, field_changes: [], material_changes: [
    { material_id: 90, name: '委托协议', material_type: 'contract', status: 'uploaded', required: true,
      change_type: 'files_changed', new_files: ['盖章协议.pdf'], carried_over: ['旧协议.pdf'],
      file_count_before: 1, file_count_after: 2, current_file_id_before: 10, current_file_id_after: 11 },
  ] }} />)
  const summary = screen.getByTestId('approval-round-diff')
  expect(summary).toHaveTextContent('当前附件已更换')
  expect(summary).toHaveTextContent('本轮新附件：盖章协议.pdf')
  expect(summary).toHaveTextContent('沿用附件：旧协议.pdf')
  expect(summary).toHaveTextContent('附件数量：1 → 2')
  expect(summary).not.toHaveTextContent(/10\s*→\s*11|材料 #90/)
})

it('does not invent an unchanged or replaced conclusion when the previous baselines are unknown', () => {
  render(<ApprovalRoundDiff diff={{ round_diff_version: 2,
    baseline: { facts_baseline: 'unknown', material_baseline: 'unknown' }, field_changes: [], material_changes: [
      { material_id: 1, name: '历史合同', change_type: 'unchanged', new_files: ['新.pdf'], carried_over: ['旧.pdf'],
        current_file_id_before: 10, current_file_id_after: 11 },
    ] }} />)
  const summary = screen.getByTestId('approval-round-diff')
  expect(summary).toHaveTextContent('字段基线未知')
  expect(summary).toHaveTextContent('材料基线未知')
  expect(summary).not.toHaveTextContent(/未记录字段修改|当前附件已更换|沿用上轮材料|本轮新附件/)
})

it('keeps an unsupported diff version unknown without echoing its values', () => {
  render(<ApprovalRoundDiff diff={{ round_diff_version: 1, baseline: frozen,
    field_changes: [{ field: 'description', before: '未知格式内容', after: '不得猜测' }], material_changes: [],
  }} />)
  expect(screen.getByTestId('approval-round-diff')).toHaveTextContent('字段基线未知')
  expect(document.body.textContent).not.toMatch(/未知格式内容|不得猜测|未记录材料修改/)
})
