// Shared intake case-type contract. These codes must stay in sync with
// models.ValidConflictCaseTypes in internal/models/conflict.go; a UI option
// outside this list fails the formal conflict check with VALIDATION_005.
export const intakeCaseTypeContract = [
  'civil',
  'commercial',
  'criminal',
  'administrative',
  'labor',
  'intellectual',
  'financial',
  'arbitration',
  'consultation',
  'construction',
  'other',
  'civil_litigation',
  'ma',
] as const

export type IntakeCaseType = (typeof intakeCaseTypeContract)[number]

export function isIntakeCaseType(value: unknown): value is IntakeCaseType {
  return typeof value === 'string' && (intakeCaseTypeContract as readonly string[]).includes(value)
}
