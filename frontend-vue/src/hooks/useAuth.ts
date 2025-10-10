import { useContext } from 'react';
import { AuthContext } from '@/context/AuthContext';

/**
 * 认证钩子，用于在组件中获取和操作认证状态
 * @returns 认证上下文对象，包含用户信息、token、登录状态和相关方法
 */
export const useAuth = () => {
  const context = useContext(AuthContext);
  
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  
  return context;
};

export default useAuth;