import React from 'react'
import { render, screen } from '@testing-library/react'
import ConflictWorkbench from '../ConflictWorkbench'

jest.mock('@/pages/batch01/Batch01Prototype', () => ({
  ConflictCheckResults: () => <div>真实API利益冲突检测结果页</div>,
}))

describe('ConflictWorkbench', () => {
  it('keeps the MVP conflict entry backed by the real API result page', () => {
    render(<ConflictWorkbench />)

    expect(screen.getByText('真实API利益冲突检测结果页')).toBeInTheDocument()
  })
})
