import React from 'react';
import { useContext } from 'react';
import { AuthContext } from '@/context/AuthContext';

export interface PermissionGuardProps {
  /**
   * 需要的权限代码
   */
  permission: string;
  /**
   * 子组件
   */
  children: React.ReactNode;
  /**
   * 无权限时显示的内容
   */
  fallback?: React.ReactNode;
}

/**
 * 权限守卫组件
 * 只有拥有指定权限的用户才能看到子组件
 */
export const PermissionGuard: React.FC<PermissionGuardProps> = ({
  permission,
  children,
  fallback = null
}) => {
  const { hasPermission } = useContext(AuthContext);

  if (!hasPermission(permission)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
};

export interface RoleGuardProps {
  /**
   * 需要的角色代码
   */
  role: string;
  /**
   * 子组件
   */
  children: React.ReactNode;
  /**
   * 无权限时显示的内容
   */
  fallback?: React.ReactNode;
}

/**
 * 角色守卫组件
 * 只有拥有指定角色的用户才能看到子组件
 */
export const RoleGuard: React.FC<RoleGuardProps> = ({
  role,
  children,
  fallback = null
}) => {
  const { hasRole } = useContext(AuthContext);

  if (!hasRole(role)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
};

export interface AnyRoleGuardProps {
  /**
   * 需要的任意一个角色代码
   */
  roles: string[];
  /**
   * 子组件
   */
  children: React.ReactNode;
  /**
   * 无权限时显示的内容
   */
  fallback?: React.ReactNode;
}

/**
 * 任意角色守卫组件
 * 拥有任意一个指定角色的用户就能看到子组件
 */
export const AnyRoleGuard: React.FC<AnyRoleGuardProps> = ({
  roles,
  children,
  fallback = null
}) => {
  const { hasAnyRole } = useContext(AuthContext);

  if (!hasAnyRole(roles)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
};

export interface AllRolesGuardProps {
  /**
   * 需要的所有角色代码
   */
  roles: string[];
  /**
   * 子组件
   */
  children: React.ReactNode;
  /**
   * 无权限时显示的内容
   */
  fallback?: React.ReactNode;
}

/**
 * 所有角色守卫组件
 * 必须拥有所有指定角色的用户才能看到子组件
 */
export const AllRolesGuard: React.FC<AllRolesGuardProps> = ({
  roles,
  children,
  fallback = null
}) => {
  const { hasAllRoles } = useContext(AuthContext);

  if (!hasAllRoles(roles)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
};