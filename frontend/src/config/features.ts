import { isOnlyOfficeEnabledFromEnv, normalizeOnlyOfficeOrigin } from './onlyoffice'

const onlyOfficeOrigin = normalizeOnlyOfficeOrigin(import.meta.env.VITE_ONLYOFFICE_URL)

export const isOnlyOfficeEnabled = (): boolean =>
  isOnlyOfficeEnabledFromEnv(
    import.meta.env.VITE_ONLYOFFICE_ENABLED,
    import.meta.env.VITE_ONLYOFFICE_URL,
  )

export const getOnlyOfficeOrigin = (): string => onlyOfficeOrigin ?? ''
