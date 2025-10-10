# TypeScript 代码规范
## Law OA Go 项目前端编码标准

**版本**: 1.0  
**创建日期**: 2025-09-30  
**适用项目**: Law OA Go v2.1.0 Frontend  

---

## 📋 目录

1. [概述](#概述)
2. [基础规范](#基础规范)
3. [类型定义](#类型定义)
4. [React 组件规范](#react-组件规范)
5. [状态管理](#状态管理)
6. [性能优化](#性能优化)
7. [错误处理](#错误处理)
8. [测试规范](#测试规范)
9. [安全编程](#安全编程)
10. [文档规范](#文档规范)

---

## 🎯 概述

本文档定义了 Law OA Go 项目前端 TypeScript 代码编写标准，适用于 Bootstrap 版本（frontend/）和 Ant Design 版本（frontend-vue/）。

### 设计原则

- **类型安全**: 充分利用 TypeScript 的类型系统
- **组件化**: 构建可复用的 React 组件
- **性能优先**: 优化渲染性能和用户体验
- **可维护性**: 代码结构清晰，易于维护
- **一致性**: 两个前端版本保持一致的编码风格

---

## 🔧 基础规范

### 1. 文件命名

**组件文件**：使用 PascalCase
```
UserList.tsx
UserCard.tsx
LoginForm.tsx
```

**工具文件**：使用 camelCase
```
apiClient.ts
userUtils.ts
dateHelpers.ts
```

**类型文件**：使用 camelCase，以 .types.ts 结尾
```
user.types.ts
api.types.ts
common.types.ts
```

### 2. 目录结构

**推荐目录结构**：
```
src/
├── components/          # 可复用组件
│   ├── common/         # 通用组件
│   ├── forms/          # 表单组件
│   └── layout/         # 布局组件
├── pages/              # 页面组件
├── hooks/              # 自定义 Hooks
├── services/           # API 服务
├── store/              # 状态管理
├── types/              # 类型定义
├── utils/              # 工具函数
└── constants/          # 常量定义
```### 
3. 导入规范

**导入顺序**：
1. React 相关
2. 第三方库
3. 内部组件
4. 类型导入
5. 样式文件

```typescript
// ✅ 正确的导入顺序
import React, { useState, useEffect, useCallback } from 'react';

import { Button, Form, Input } from 'antd';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

import { UserCard } from '../components/UserCard';
import { useAuth } from '../hooks/useAuth';

import type { User, CreateUserRequest } from '../types/user.types';
import type { ApiResponse } from '../types/api.types';

import './UserList.less';
```

---

## 🏷️ 类型定义

### 1. 接口定义

**使用 interface 定义对象类型**：
```typescript
// ✅ 好的接口定义
interface User {
  readonly id: number;
  name: string;
  email: string;
  role: UserRole;
  isActive: boolean;
  createdAt: Date;
  updatedAt: Date;
}

interface CreateUserRequest {
  name: string;
  email: string;
  password: string;
  role: UserRole;
}

interface UserListProps {
  users: User[];
  loading?: boolean;
  onUserSelect: (user: User) => void;
  onUserEdit?: (user: User) => void;
  className?: string;
}
```

### 2. 类型别名

**使用 type 定义联合类型和复杂类型**：
```typescript
// ✅ 类型别名
type UserRole = 'admin' | 'lawyer' | 'client';
type LoadingState = 'idle' | 'loading' | 'success' | 'error';
type ApiStatus = 'pending' | 'fulfilled' | 'rejected';

type UserWithCases = User & {
  cases: Case[];
  caseCount: number;
};

type UserFormData = Omit<User, 'id' | 'createdAt' | 'updatedAt'>;
```

### 3. 泛型使用

**合理使用泛型提高代码复用性**：
```typescript
// ✅ 泛型接口
interface ApiResponse<T> {
  data: T;
  message: string;
  success: boolean;
  timestamp: string;
}

interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

// ✅ 泛型函数
async function fetchData<T>(url: string): Promise<ApiResponse<T>> {
  const response = await axios.get<ApiResponse<T>>(url);
  return response.data;
}

// ✅ 泛型组件
interface TableProps<T> {
  data: T[];
  columns: TableColumn<T>[];
  onRowClick?: (item: T) => void;
}

function Table<T>({ data, columns, onRowClick }: TableProps<T>) {
  // 实现
}
```---


## ⚛️ React 组件规范

### 1. 函数组件定义

**使用函数组件和 TypeScript**：
```typescript
// ✅ 好的函数组件
interface UserCardProps {
  user: User;
  showActions?: boolean;
  onEdit?: (user: User) => void;
  onDelete?: (userId: number) => void;
  className?: string;
}

const UserCard: React.FC<UserCardProps> = ({
  user,
  showActions = true,
  onEdit,
  onDelete,
  className = ''
}) => {
  const handleEdit = useCallback(() => {
    onEdit?.(user);
  }, [onEdit, user]);

  const handleDelete = useCallback(() => {
    onDelete?.(user.id);
  }, [onDelete, user.id]);

  return (
    <div className={`user-card ${className}`}>
      <div className="user-info">
        <h3>{user.name}</h3>
        <p>{user.email}</p>
        <span className={`role-badge role-${user.role}`}>
          {user.role}
        </span>
      </div>
      
      {showActions && (
        <div className="user-actions">
          <Button onClick={handleEdit}>编辑</Button>
          <Button danger onClick={handleDelete}>删除</Button>
        </div>
      )}
    </div>
  );
};

export default UserCard;
```

### 2. Props 验证

**使用 TypeScript 进行 Props 验证**：
```typescript
// ✅ 严格的 Props 类型
interface FormInputProps {
  label: string;
  name: string;
  type?: 'text' | 'email' | 'password' | 'number';
  value: string;
  onChange: (value: string) => void;
  onBlur?: () => void;
  error?: string;
  required?: boolean;
  disabled?: boolean;
  placeholder?: string;
}

const FormInput: React.FC<FormInputProps> = ({
  label,
  name,
  type = 'text',
  value,
  onChange,
  onBlur,
  error,
  required = false,
  disabled = false,
  placeholder
}) => {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value);
  };

  return (
    <div className="form-input">
      <label htmlFor={name} className={required ? 'required' : ''}>
        {label}
      </label>
      <input
        id={name}
        name={name}
        type={type}
        value={value}
        onChange={handleChange}
        onBlur={onBlur}
        disabled={disabled}
        placeholder={placeholder}
        className={error ? 'error' : ''}
      />
      {error && <span className="error-message">{error}</span>}
    </div>
  );
};
```

### 3. 自定义 Hooks

**创建可复用的自定义 Hooks**：
```typescript
// ✅ 自定义 Hook
interface UseApiOptions<T> {
  initialData?: T;
  onSuccess?: (data: T) => void;
  onError?: (error: Error) => void;
}

interface UseApiReturn<T> {
  data: T | null;
  loading: boolean;
  error: Error | null;
  execute: () => Promise<void>;
  reset: () => void;
}

function useApi<T>(
  apiCall: () => Promise<T>,
  options: UseApiOptions<T> = {}
): UseApiReturn<T> {
  const [data, setData] = useState<T | null>(options.initialData || null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const execute = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      
      const result = await apiCall();
      setData(result);
      options.onSuccess?.(result);
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Unknown error');
      setError(error);
      options.onError?.(error);
    } finally {
      setLoading(false);
    }
  }, [apiCall, options]);

  const reset = useCallback(() => {
    setData(options.initialData || null);
    setError(null);
    setLoading(false);
  }, [options.initialData]);

  return { data, loading, error, execute, reset };
}

// 使用示例
const UserList: React.FC = () => {
  const {
    data: users,
    loading,
    error,
    execute: fetchUsers
  } = useApi(() => userService.getUsers());

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      {users?.map(user => (
        <UserCard key={user.id} user={user} />
      ))}
    </div>
  );
};
```