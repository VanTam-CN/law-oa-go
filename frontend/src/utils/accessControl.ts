import type { User } from '@/stores/useAppStore'

const ADMIN_ROLES = new Set(['admin', 'super_admin'])

const PERMISSION_ALIASES: Record<string, string[]> = {
  'dashboard:view': ['dashboard'],
  'user:view': ['user_management'],
  'user:manage': ['user_management', 'user:create', 'user:edit', 'user:delete'],
  'role:view': ['role_management'],
  'role:manage': ['role_management', 'role:create', 'role:edit', 'role:delete'],
  'permission:view': ['permission_management'],
  'permission:manage': ['permission_management'],
  'case:view': ['case_management'],
  'case:manage': ['case_management', 'case:create', 'case:edit', 'case:delete', 'case:assign'],
  'client:view': ['client_management'],
  'client:manage': ['client_management', 'client:create', 'client:edit', 'client:delete'],
  'approval:view': ['approval_center'],
  'approval:manage': ['approval_center'],
  'file:view': ['document_management', 'document:view'],
  'file:manage': ['document_management', 'document:upload', 'document:edit', 'document:delete'],
  'finance:view': ['finance_management', 'finance:view'],
  'finance:manage': ['finance_management', 'finance:create', 'finance:edit'],
  'system:manage': ['system_settings'],
  'tools:view': ['tools_center'],
  'reports:view': ['statistics_reports'],
}

const ROLE_FALLBACKS: Record<string, string[]> = {
  'dashboard:view': [
    'admin',
    'super_admin',
    'lawyer',
    'assistant',
    'finance',
    'intern',
    'user',
    'conflict_officer',
    'director',
    'partner',
    'management',
    'compliance',
    'risk',
    'risk_control',
  ],
  'user:view': ['admin', 'super_admin'],
  'user:manage': ['admin', 'super_admin'],
  'role:view': ['admin', 'super_admin'],
  'role:manage': ['admin', 'super_admin'],
  'permission:view': ['admin', 'super_admin'],
  'permission:manage': ['admin', 'super_admin'],
  'case:view': ['admin', 'super_admin', 'lawyer', 'assistant'],
  'case:manage': ['admin', 'super_admin', 'lawyer'],
  'client:view': ['admin', 'super_admin', 'lawyer', 'assistant'],
  'client:manage': ['admin', 'super_admin', 'lawyer'],
  'lawyer:manage': ['admin', 'super_admin'],
  'approval:view': ['admin', 'super_admin', 'lawyer', 'conflict_officer', 'compliance'],
  'approval:manage': ['admin', 'super_admin', 'lawyer'],
  'conflict:check': [
    'lawyer',
    'conflict_officer',
    'director',
    'partner',
    'compliance',
    'risk',
    'risk_control',
    'management',
  ],
  'conflict:governance': [
    'director',
    'partner',
    'management',
    'compliance',
    'risk',
    'risk_control',
    'conflict_officer',
  ],
  'file:view': ['admin', 'super_admin', 'lawyer', 'assistant'],
  'file:manage': ['admin', 'super_admin', 'lawyer'],
  'finance:view': ['admin', 'super_admin', 'finance'],
  'finance:manage': ['admin', 'super_admin', 'finance'],
  'trust:manage': ['admin', 'super_admin', 'finance'],
  'system:manage': ['admin', 'super_admin'],
  'tools:view': ['admin', 'super_admin', 'lawyer', 'assistant', 'finance', 'intern', 'user'],
  'reports:view': ['admin', 'super_admin', 'lawyer', 'finance'],
}

function toArray(value?: string | string[]): string[] {
  if (!value) {
    return []
  }
  return Array.isArray(value) ? value : [value]
}

function roleMatches(userRoles: string[], acceptedRoles: string[]): boolean {
  return acceptedRoles.some((role) => userRoles.includes(role))
}

export function hasPermission(user: User | null | undefined, permission: string): boolean {
  if (!user) {
    return false
  }

  const roles = user.roles || []
  if (
    ['conflict:check', 'conflict:governance'].includes(permission) &&
    roles.some((role) => ADMIN_ROLES.has(role))
  ) {
    return false
  }
  if (roles.some((role) => ADMIN_ROLES.has(role))) {
    return true
  }

  const permissions = user.permissions || []
  const acceptedPermissions = [permission, ...(PERMISSION_ALIASES[permission] || [])]
  if (permissions.includes('*') || acceptedPermissions.some((code) => permissions.includes(code))) {
    return true
  }

  const acceptedRoles = ROLE_FALLBACKS[permission] || []
  return roleMatches(roles, acceptedRoles)
}

export function canAccess(
  user: User | null | undefined,
  permissions?: string | string[],
  roles?: string | string[],
): boolean {
  if (!user) {
    return false
  }

  const userRoles = user.roles || []
  if (userRoles.some((role) => ADMIN_ROLES.has(role))) {
    if (toArray(permissions).some((permission) => ['conflict:check', 'conflict:governance'].includes(permission))) {
      return false
    }
    return true
  }

  const acceptedRoles = toArray(roles)
  if (acceptedRoles.length > 0 && roleMatches(userRoles, acceptedRoles)) {
    return true
  }

  const requiredPermissions = toArray(permissions)
  if (requiredPermissions.length === 0) {
    return acceptedRoles.length === 0
  }

  return requiredPermissions.some((permission) => hasPermission(user, permission))
}
