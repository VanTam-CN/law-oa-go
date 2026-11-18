const ONLYOFFICE_SCRIPT_PATH = '/web-apps/apps/api/documents/api.js'

const SUPPORTED_PROTOCOLS = new Set(['http:', 'https:'])

export const normalizeOnlyOfficeOrigin = (raw: string | undefined): string | null => {
  const value = raw?.trim()
  if (!value) {
    return null
  }

  try {
    const parsed = new URL(value)
    if (!SUPPORTED_PROTOCOLS.has(parsed.protocol)) {
      return null
    }
    if (!parsed.host || parsed.username || parsed.password) {
      return null
    }
    if (parsed.search || parsed.hash) {
      return null
    }
    if (parsed.pathname !== '' && parsed.pathname !== '/') {
      return null
    }

    return parsed.origin
  } catch {
    return null
  }
}

const toBoolean = (value: string | undefined): boolean => value === 'true'

export const isOnlyOfficeEnabledFromEnv = (
  onlyOfficeEnabledRaw: string | undefined,
  onlyOfficeUrlRaw: string | undefined,
): boolean => toBoolean(onlyOfficeEnabledRaw) && normalizeOnlyOfficeOrigin(onlyOfficeUrlRaw) !== null

export const buildOnlyOfficeScriptUrl = (origin: string): string => {
  return new URL(ONLYOFFICE_SCRIPT_PATH, origin).toString()
}

export const encodeJsonForScript = (value: unknown): string =>
  JSON.stringify(value)
    .replace(/</g, '\\u003c')
    .replace(/>/g, '\\u003e')
    .replace(/&/g, '\\u0026')
    .replace(/\u2028/g, '\\u2028')
    .replace(/\u2029/g, '\\u2029')

export const getOnlyOfficeAllowedOrigins = (
  onlyOfficeOrigin: string | null,
  parentOrigin: string,
): string[] => {
  return Array.from(
    new Set([parentOrigin, onlyOfficeOrigin].filter((origin): origin is string => Boolean(origin))),
  )
}
