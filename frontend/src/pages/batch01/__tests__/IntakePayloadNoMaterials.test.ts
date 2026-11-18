// Contract guard (2026-09-30): the intake save payload must NOT carry a
// `materials` field. The server-side update path deletes and recreates
// intake materials when the field is present, which would wipe uploaded
// file metadata; materials are now managed exclusively through the
// matter-collaboration ledger endpoints.

// buildIntakePayload is internal to CaseIntakeWorkbench, so this guard
// asserts the observable contract: the compiled source of the workbench
// must not embed `materials:` into request payloads.
import { readFileSync } from 'fs'
import { join } from 'path'

describe('intake save payload omits materials (delete-recreate guard)', () => {
  it('buildIntakePayload source contains no materials field', () => {
    const source = readFileSync(
      join(__dirname, '..', 'Batch01Prototype.tsx'),
      'utf8',
    )
    // No payload assignment of materials remains anywhere in the file.
    expect(source).not.toMatch(/materials:\s*defaultMaterials/)
    expect(source).not.toMatch(/^\s*materials:\s*\[/m)
  })

  it('step 4 renders the real ledger panel instead of the static table', () => {
    const source = readFileSync(
      join(__dirname, '..', 'Batch01Prototype.tsx'),
      'utf8',
    )
    expect(source).toContain('<MaterialLedgerPanel')
    expect(source).not.toContain('文件管理暂未开放')
  })
})
