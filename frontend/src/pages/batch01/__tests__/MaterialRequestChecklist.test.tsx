import React from 'react'
import { fireEvent, render, screen } from '@testing-library/react'
import SupplementChecklist, { type SupplementItemFE } from '../intake/SupplementChecklist'

const items: SupplementItemFE[] = [
  {
    item_id: 'reply-material-3',
    category: 'material_request',
    field_key: 'material',
    material_id: 3,
    description: '请补传客户主体资料。',
    required: true,
  },
  {
    item_id: 'reply-title',
    category: 'field_correction',
    field_key: 'case_title',
    description: '请更正案件名称。',
    required: false,
  },
]

describe('SupplementChecklist material_request rendering', () => {
  it('renders the ledger material name and a locate button', () => {
    const onLocateMaterial = jest.fn()
    render(
      <SupplementChecklist
        items={items}
        answers={{}}
        onChange={() => undefined}
        materialNames={{ 3: '客户主体资料' }}
        onLocateMaterial={onLocateMaterial}
      />,
    )

    expect(screen.getByText('材料补充')).toBeInTheDocument()
    expect(screen.getByText('客户主体资料')).toBeInTheDocument()
    const locate = screen.getByRole('button', { name: '去第 4 步上传该材料' })
    fireEvent.click(locate)
    expect(onLocateMaterial).toHaveBeenCalledTimes(1)
    expect(onLocateMaterial).toHaveBeenCalledWith(items[0])

    // Reply field labels the material by its real ledger name.
    expect(screen.getByRole('textbox', { name: '回复：客户主体资料' })).toBeInTheDocument()
    expect(screen.getByRole('textbox', { name: '回复：案件名称' })).toBeInTheDocument()
  })

  it('keeps reply and upload states distinct: a reply never marks a material complete', () => {
    render(
      <SupplementChecklist
        items={items}
        answers={{ 'reply-material-3': '已传给项目组，请查收。' }}
        onChange={() => undefined}
        materialNames={{ 3: '客户主体资料' }}
        materialStatuses={{ 3: 'missing' }}
        materialEvidence={{ 'reply-material-3': { item_id: 'reply-material-3', material_id: 3, state: 'missing', message: '' } }}
      />,
    )

    // Replied material item without a server-confirmed upload: waiting, not done.
    expect(screen.getByTestId('supplement-status-reply-material-3')).toHaveTextContent(
      '尚未补交补正附件',
    )
    expect(screen.getByText('已回复')).toBeInTheDocument()
    expect(screen.getByTestId('supplement-status-reply-material-3')).not.toHaveTextContent('补正附件已上传')
    // The untouched field item stays in the applicant's own to-do state
    // (recheck #4: 待补充, not 待复核 which implied the reviewer's queue).
    expect(screen.getByTestId('supplement-status-reply-title')).toHaveTextContent('待补充')
  })

  it('marks correction uploaded only from bound server evidence with a valid current file', () => {
    render(
      <SupplementChecklist
        items={items}
        answers={{}}
        onChange={() => undefined}
        materialNames={{ 3: '客户主体资料' }}
        materialStatuses={{ 3: 'uploaded' }}
        materialEvidence={{ 'reply-material-3': { item_id: 'reply-material-3', material_id: 3, state: 'complete', current_file_id: 11, message: '' } }}
      />,
    )

    expect(screen.getByTestId('supplement-status-reply-material-3')).toHaveTextContent(
      '补正附件已上传，待审批复核',
    )
  })

  it('names the currently registered file so a replacement request is recognizable (recheck #4)', () => {
    render(
      <SupplementChecklist
        items={items}
        answers={{}}
        onChange={() => undefined}
        materialNames={{ 3: '客户主体资料' }}
        materialStatuses={{ 3: 'uploaded' }}
        materialFileNames={{ 3: '原营业执照.pdf' }}
        materialEvidence={{ 'reply-material-3': { item_id: 'reply-material-3', material_id: 3, state: 'missing', current_file_id: 10, message: '' } }}
      />,
    )

    expect(screen.getByTestId('supplement-status-reply-material-3')).toHaveTextContent(
      '已有旧附件，尚未补交新附件',
    )
    expect(screen.getByText(/当前已登记文件：原营业执照.pdf/)).toBeInTheDocument()
    expect(screen.getByText(/请上传新文件替换/)).toBeInTheDocument()
  })

  it('shows a replied non-material item as awaiting review, not complete', () => {
    render(
      <SupplementChecklist
        items={items}
        answers={{ 'reply-title': '已更正案件名称。' }}
        onChange={() => undefined}
      />,
    )

    expect(screen.getByTestId('supplement-status-reply-title')).toHaveTextContent(
      '已回复，待审批复核',
    )
  })

  it('renders the refresh action only when the parent provides the callback', () => {
    const onRefreshMaterialStatus = jest.fn()
    const { rerender } = render(
      <SupplementChecklist
        items={items}
        answers={{}}
        onChange={() => undefined}
        onRefreshMaterialStatus={onRefreshMaterialStatus}
      />,
    )
    const refresh = screen.getByRole('button', { name: '刷新材料状态' })
    fireEvent.click(refresh)
    expect(onRefreshMaterialStatus).toHaveBeenCalledTimes(1)

    rerender(
      <SupplementChecklist
        items={items}
        answers={{}}
        onChange={() => undefined}
      />,
    )
    expect(screen.queryByRole('button', { name: '刷新材料状态' })).not.toBeInTheDocument()
  })

  it('falls back to the numeric anchor when the name lookup failed', () => {
    render(<SupplementChecklist items={items} answers={{}} onChange={() => undefined} />)
    expect(screen.getByText('材料 #3')).toBeInTheDocument()
    // No locate callback provided: button still renders (disabled flow is
    // handled by the parent), no crash.
    expect(screen.getByRole('button', { name: '去第 4 步上传该材料' })).toBeInTheDocument()
  })

  it('offers the field locate action only for non-material items when the callback is provided', () => {
    const onLocateField = jest.fn()
    render(
      <SupplementChecklist
        items={items}
        answers={{}}
        onChange={() => undefined}
        onLocateField={onLocateField}
      />,
    )

    const fieldLocate = screen.getByRole('button', { name: '定位并填写该字段' })
    expect(screen.queryAllByRole('button', { name: '定位并填写该字段' })).toHaveLength(1)
    fireEvent.click(fieldLocate)
    expect(onLocateField).toHaveBeenCalledTimes(1)
    expect(onLocateField).toHaveBeenCalledWith(items[1])
  })
})
