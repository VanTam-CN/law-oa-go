// Signatures are used only in memory. They must never enter browser storage.
export function intakeEditSignature(
  form: Record<string, unknown>,
  parties: unknown[],
  tags: string[],
) {
  const { revision, intakeId, intakeCode, idempotencyKey, ...facts } = form
  return JSON.stringify({ facts, parties, tags })
}

export function purgeIntakeIdentityCaches(storage: Storage) {
  const keys = Array.from({ length: storage.length }, (_, index) => storage.key(index))
  keys.forEach((key) => {
    if (key?.startsWith('law-oa-case-intake-draft-') && key.endsWith(':conflictFingerprint'))
      storage.removeItem(key)
  })
}

export function intakeSaveLabel(state: 'unsaved' | 'saving' | 'saved' | 'error', at?: string) {
  return {
    unsaved: '未保存修改',
    saving: '正在保存',
    saved: at ? `已保存于 ${at}` : '已保存（服务端）',
    error: '保存失败，输入仍保留',
  }[state]
}
