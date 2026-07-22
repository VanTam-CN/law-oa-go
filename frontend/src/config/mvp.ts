export const MVP_MENU_KEYS = [
  'dashboard',
  'case',
  'conflict',
  'client',
  'approval',
  'inbox',
  'trust',
] as const

export const MVP_UNAVAILABLE_PATHS = ['/file', '/finance', '/settings'] as const

export const MVP_UNAVAILABLE_MODULES = {
  '/file': '文档中心',
  '/finance': '财务中心',
  '/settings': '系统设置',
} as const

type MvpMenuKey = (typeof MVP_MENU_KEYS)[number]

export const isMvpMenuKey = (key: string): key is MvpMenuKey => {
  return MVP_MENU_KEYS.includes(key as MvpMenuKey)
}

export const isMvpUnavailablePath = (pathname: string): boolean => {
  return MVP_UNAVAILABLE_PATHS.some((path) => pathname === path || pathname.startsWith(`${path}/`))
}

export const getMvpUnavailableModuleName = (pathname: string): string | null => {
  const matchedPath = MVP_UNAVAILABLE_PATHS.find((path) => pathname === path || pathname.startsWith(`${path}/`))
  return matchedPath ? MVP_UNAVAILABLE_MODULES[matchedPath] : null
}
