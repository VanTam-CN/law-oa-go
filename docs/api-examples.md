# Law OA Go API 使用示例

本文档提供 Law OA Go API 的实际使用示例，帮助开发者快速集成和使用API。

## 🚀 快速开始

### 1. 环境准备

确保你的开发环境已准备就绪：

```bash
# 安装 Go 1.23+
# 推荐使用 Go 1.23 或更高版本

# 克隆项目
git clone https://github.com/law-oa-go.git
cd law-oa-go

# 安装依赖
go mod download

# 构建服务
go build -o bin/law-oa-go .

# 启动服务器
./bin/law-oa-go
```

服务器启动后，API将在 `http://localhost:8080` 提供服务。

### 2. 使用 curl 测试

#### 获取认证令牌
```bash
# 登录获取 JWT 令牌
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@lawfirm.com",
    "password": "password123"
  }'

# 响应示例
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-12-31T23:59:59Z",
    "user": {
      "id": 1,
      "name": "系统管理员",
      "email": "admin@lawfirm.com",
      "role": "admin",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z"
    }
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

#### 使用令牌访问受保护的API
```bash
# 设置令牌变量
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 获取用户列表
curl -X GET http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer $TOKEN"

# 创建客户
curl -X POST http://localhost:8080/api/v1/clients \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三",
    "email": "zhangsan@example.com",
    "phone": "13900139000",
    "address": "北京市朝阳区"
  }'
```

## 📝 JavaScript/TypeScript 示例

### 1. 基础 API 客户端

```typescript
// src/api/client.ts
class ApiClient {
  private baseUrl: string;
  private token: string | null = null;

  constructor(baseUrl: string = 'http://localhost:8080/api/v1') {
    this.baseUrl = baseUrl;
  }

  // 设置认证令牌
  setToken(token: string): void {
    this.token = token;
  }

  // 通用请求方法
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error?.message || 'Request failed');
    }

    return response.json();
  }

  // 认证相关方法
  async login(email: string, password: string): Promise<any> {
    const response = await this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
    
    this.setToken(response.data.token);
    return response;
  }

  async register(data: any): Promise<any> {
    return this.request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  // 用户管理方法
  async getUsers(params?: any): Promise<any> {
    const queryParams = new URLSearchParams(params).toString();
    const endpoint = queryParams ? `/admin/users?${queryParams}` : '/admin/users';
    return this.request(endpoint);
  }

  async getUser(id: number): Promise<any> {
    return this.request(`/admin/users/${id}`);
  }

  async createUser(data: any): Promise<any> {
    return this.request('/admin/users', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateUser(id: number, data: any): Promise<any> {
    return this.request(`/admin/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteUser(id: number): Promise<any> {
    return this.request(`/admin/users/${id}`, {
      method: 'DELETE',
    });
  }

  // 客户管理方法
  async getClients(params?: any): Promise<any> {
    const queryParams = new URLSearchParams(params).toString();
    const endpoint = queryParams ? `/clients?${queryParams}` : '/clients';
    return this.request(endpoint);
  }

  async getClient(id: number): Promise<any> {
    return this.request(`/clients/${id}`);
  }

  async createClient(data: any): Promise<any> {
    return this.request('/clients', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateClient(id: number, data: any): Promise<any> {
    return this.request(`/clients/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteClient(id: number): Promise<any> {
    return this.request(`/clients/${id}`, {
      method: 'DELETE',
    });
  }

  async getClientStats(): Promise<any> {
    return this.request('/clients/stats');
  }

  // 案件管理方法
  async getCases(params?: any): Promise<any> {
    const queryParams = new URLSearchParams(params).toString();
    const endpoint = queryParams ? `/cases?${queryParams}` : '/cases';
    return this.request(endpoint);
  }

  async getCase(id: number): Promise<any> {
    return this.request(`/cases/${id}`);
  }

  async createCase(data: any): Promise<any> {
    return this.request('/cases', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateCase(id: number, data: any): Promise<any> {
    return this.request(`/cases/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteCase(id: number): Promise<any> {
    return this.request(`/cases/${id}`, {
      method: 'DELETE',
    });
  }

  async assignLawyer(caseId: number, lawyerId: number): Promise<any> {
    return this.request(`/cases/${caseId}/assign`, {
      method: 'POST',
      body: JSON.stringify({ lawyer_id: lawyerId }),
    });
  }

  async updateCaseStatus(caseId: number, status: string): Promise<any> {
    return this.request(`/cases/${caseId}/status`, {
      method: 'POST',
      body: JSON.stringify({ status }),
    });
  }

  async getCaseStats(): Promise<any> {
    return this.request('/cases/stats');
  }
}

export default ApiClient;
```

### 2. 使用示例

```typescript
// src/examples/usage.ts
import ApiClient from './api/client';

async function main() {
  const api = new ApiClient();

  try {
    // 1. 用户登录
    console.log('=== 用户登录 ===');
    const loginResponse = await api.login('admin@lawfirm.com', 'password123');
    console.log('登录成功:', loginResponse.data.user.name);

    // 2. 获取用户列表
    console.log('\n=== 获取用户列表 ===');
    const users = await api.getUsers({ page: 1, page_size: 10 });
    console.log('用户列表:', users.data);

    // 3. 创建客户
    console.log('\n=== 创建客户 ===');
    const newClient = await api.createClient({
      name: '李四',
      email: 'lisi@example.com',
      phone: '13900139001',
      address: '北京市海淀区',
      company: '某科技公司',
      notes: '新客户'
    });
    console.log('创建客户成功:', newClient.data);

    // 4. 获取客户列表
    console.log('\n=== 获取客户列表 ===');
    const clients = await api.getClients({ page: 1, page_size: 5 });
    console.log('客户列表:', clients.data);

    // 5. 创建案件
    console.log('\n=== 创建案件 ===');
    const newCase = await api.createCase({
      title: '商标侵权案',
      description: '某科技公司商标侵权纠纷',
      client_id: newClient.data.id,
      lawyer_id: 1,
      case_type: 'commercial',
      priority: 'high',
      status: 'pending',
      start_date: new Date().toISOString()
    });
    console.log('创建案件成功:', newCase.data);

    // 6. 获取案件列表
    console.log('\n=== 获取案件列表 ===');
    const cases = await api.getCases({ page: 1, page_size: 5 });
    console.log('案件列表:', cases.data);

    // 7. 分配律师
    console.log('\n=== 分配律师 ===');
    const assignResult = await api.assignLawyer(newCase.data.id, 2);
    console.log('分配律师成功:', assignResult.data);

    // 8. 更新案件状态
    console.log('\n=== 更新案件状态 ===');
    const statusResult = await api.updateCaseStatus(newCase.data.id, 'active');
    console.log('更新状态成功:', statusResult.data);

    // 9. 获取统计信息
    console.log('\n=== 获取统计信息 ===');
    const clientStats = await api.getClientStats();
    console.log('客户统计:', clientStats.data);

    const caseStats = await api.getCaseStats();
    console.log('案件统计:', caseStats.data);

  } catch (error) {
    console.error('API 调用失败:', error);
  }
}

main();
```

## 🐍 Python 示例

### 1. 基础 API 客户端

```python
# src/api/client.py
import requests
import json
from typing import Dict, Any, Optional

class ApiClient:
    def __init__(self, base_url: str = 'http://localhost:8080/api/v1'):
        self.base_url = base_url
        self.token: Optional[str] = None
        self.session = requests.Session()
    
    def set_token(self, token: str) -> None:
        """设置认证令牌"""
        self.token = token
        self.session.headers.update({
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        })
    
    def _request(self, method: str, endpoint: str, data: Optional[Dict] = None, 
                 params: Optional[Dict] = None) -> Dict[str, Any]:
        """通用请求方法"""
        url = f"{self.base_url}{endpoint}"
        
        try:
            response = self.session.request(
                method=method,
                url=url,
                json=data,
                params=params
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            error_msg = f"API请求失败: {str(e)}"
            if hasattr(e, 'response') and e.response is not None:
                try:
                    error_data = e.response.json()
                    error_msg = error_data.get('error', {}).get('message', error_msg)
                except:
                    pass
            raise Exception(error_msg)
    
    # 认证相关方法
    def login(self, email: str, password: str) -> Dict[str, Any]:
        """用户登录"""
        response = self._request('POST', '/auth/login', {
            'email': email,
            'password': password
        })
        self.set_token(response['data']['token'])
        return response
    
    def register(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """用户注册"""
        return self._request('POST', '/auth/register', data)
    
    # 用户管理方法
    def get_users(self, params: Optional[Dict] = None) -> Dict[str, Any]:
        """获取用户列表"""
        return self._request('GET', '/admin/users', params=params)
    
    def get_user(self, user_id: int) -> Dict[str, Any]:
        """获取用户详情"""
        return self._request('GET', f'/admin/users/{user_id}')
    
    def create_user(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """创建用户"""
        return self._request('POST', '/admin/users', data)
    
    def update_user(self, user_id: int, data: Dict[str, Any]) -> Dict[str, Any]:
        """更新用户"""
        return self._request('PUT', f'/admin/users/{user_id}', data)
    
    def delete_user(self, user_id: int) -> Dict[str, Any]:
        """删除用户"""
        return self._request('DELETE', f'/admin/users/{user_id}')
    
    # 客户管理方法
    def get_clients(self, params: Optional[Dict] = None) -> Dict[str, Any]:
        """获取客户列表"""
        return self._request('GET', '/clients', params=params)
    
    def get_client(self, client_id: int) -> Dict[str, Any]:
        """获取客户详情"""
        return self._request('GET', f'/clients/{client_id}')
    
    def create_client(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """创建客户"""
        return self._request('POST', '/clients', data)
    
    def update_client(self, client_id: int, data: Dict[str, Any]) -> Dict[str, Any]:
        """更新客户"""
        return self._request('PUT', f'/clients/{client_id}', data)
    
    def delete_client(self, client_id: int) -> Dict[str, Any]:
        """删除客户"""
        return self._request('DELETE', f'/clients/{client_id}')
    
    def get_client_stats(self) -> Dict[str, Any]:
        """获取客户统计"""
        return self._request('GET', '/clients/stats')
    
    # 案件管理方法
    def get_cases(self, params: Optional[Dict] = None) -> Dict[str, Any]:
        """获取案件列表"""
        return self._request('GET', '/cases', params=params)
    
    def get_case(self, case_id: int) -> Dict[str, Any]:
        """获取案件详情"""
        return self._request('GET', f'/cases/{case_id}')
    
    def create_case(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """创建案件"""
        return self._request('POST', '/cases', data)
    
    def update_case(self, case_id: int, data: Dict[str, Any]) -> Dict[str, Any]:
        """更新案件"""
        return self._request('PUT', f'/cases/{case_id}', data)
    
    def delete_case(self, case_id: int) -> Dict[str, Any]:
        """删除案件"""
        return self._request('DELETE', f'/cases/{case_id}')
    
    def assign_lawyer(self, case_id: int, lawyer_id: int) -> Dict[str, Any]:
        """分配律师"""
        return self._request('POST', f'/cases/{case_id}/assign', {
            'lawyer_id': lawyer_id
        })
    
    def update_case_status(self, case_id: int, status: str) -> Dict[str, Any]:
        """更新案件状态"""
        return self._request('POST', f'/cases/{case_id}/status', {
            'status': status
        })
    
    def get_case_stats(self) -> Dict[str, Any]:
        """获取案件统计"""
        return self._request('GET', '/cases/stats')
```

### 2. 使用示例

```python
# src/examples/usage.py
from api.client import ApiClient
import json

def main():
    api = ApiClient()
    
    try:
        # 1. 用户登录
        print("=== 用户登录 ===")
        login_response = api.login('admin@lawfirm.com', 'password123')
        print(f"登录成功: {login_response['data']['user']['name']}")
        
        # 2. 获取用户列表
        print("\n=== 获取用户列表 ===")
        users = api.get_users({'page': 1, 'page_size': 10})
        print(f"用户数量: {users['pagination']['total']}")
        
        # 3. 创建客户
        print("\n=== 创建客户 ===")
        client_data = {
            'name': '王五',
            'email': 'wangwu@example.com',
            'phone': '13900139002',
            'address': '上海市浦东新区',
            'company': '某互联网公司',
            'notes': 'VIP客户'
        }
        new_client = api.create_client(client_data)
        print(f"创建客户成功: {new_client['data']['name']}")
        
        # 4. 获取客户列表
        print("\n=== 获取客户列表 ===")
        clients = api.get_clients({'page': 1, 'page_size': 5})
        print(f"客户数量: {clients['pagination']['total']}")
        
        # 5. 创建案件
        print("\n=== 创建案件 ===")
        case_data = {
            'title': '劳动合同纠纷',
            'description': '员工与公司之间的劳动合同纠纷',
            'client_id': new_client['data']['id'],
            'lawyer_id': 1,
            'case_type': 'civil',
            'priority': 'medium',
            'status': 'pending',
            'start_date': '2024-01-01T00:00:00Z'
        }
        new_case = api.create_case(case_data)
        print(f"创建案件成功: {new_case['data']['title']}")
        
        # 6. 获取案件列表
        print("\n=== 获取案件列表 ===")
        cases = api.get_cases({'page': 1, 'page_size': 5})
        print(f"案件数量: {cases['pagination']['total']}")
        
        # 7. 分配律师
        print("\n=== 分配律师 ===")
        assign_result = api.assign_lawyer(new_case['data']['id'], 2)
        print(f"分配律师成功: {assign_result['data']['message']}")
        
        # 8. 更新案件状态
        print("\n=== 更新案件状态 ===")
        status_result = api.update_case_status(new_case['data']['id'], 'active')
        print(f"更新状态成功: {status_result['data']['message']}")
        
        # 9. 获取统计信息
        print("\n=== 获取统计信息 ===")
        client_stats = api.get_client_stats()
        print(f"客户总数: {client_stats['data']['total']}")
        print(f"活跃客户: {client_stats['data']['active']}")
        
        case_stats = api.get_case_stats()
        print(f"案件总数: {case_stats['data']['total']}")
        print(f"进行中案件: {case_stats['data']['active']}")
        
        # 10. 分页查询示例
        print("\n=== 分页查询示例 ===")
        page_1 = api.get_cases({'page': 1, 'page_size': 2})
        print(f"第1页: {len(page_1['data'])} 条记录")
        
        if page_1['pagination']['total_pages'] > 1:
            page_2 = api.get_cases({'page': 2, 'page_size': 2})
            print(f"第2页: {len(page_2['data'])} 条记录")
        
        # 11. 搜索示例
        print("\n=== 搜索示例 ===")
        search_results = api.get_cases({'search': '合同', 'case_type': 'civil'})
        print(f"搜索到 {len(search_results['data'])} 个相关案件")
        
    except Exception as e:
        print(f"API 调用失败: {e}")

if __name__ == '__main__':
    main()
```

## 🔄 React 集成示例

### 1. API Hook

```typescript
// src/hooks/useApi.ts
import { useState, useEffect } from 'react';
import ApiClient from '../api/client';

const api = new ApiClient();

export function useApi() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const request = async <T>(
    apiCall: () => Promise<T>
  ): Promise<{ data: T | null; error: string | null }> => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await apiCall();
      return { data: result, error: null };
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      return { data: null, error: errorMessage };
    } finally {
      setLoading(false);
    }
  };

  return { loading, error, request, api };
}

// 登录 Hook
export function useAuth() {
  const { loading, error, request } = useApi();

  const login = async (email: string, password: string) => {
    return await request(() => api.login(email, password));
  };

  const register = async (data: any) => {
    return await request(() => api.register(data));
  };

  const logout = () => {
    api.setToken('');
    localStorage.removeItem('token');
  };

  return { loading, error, login, register, logout };
}

// 客户管理 Hook
export function useClients() {
  const { loading, error, request } = useApi();

  const getClients = async (params?: any) => {
    return await request(() => api.getClients(params));
  };

  const createClient = async (data: any) => {
    return await request(() => api.createClient(data));
  };

  const updateClient = async (id: number, data: any) => {
    return await request(() => api.updateClient(id, data));
  };

  const deleteClient = async (id: number) => {
    return await request(() => api.deleteClient(id));
  };

  const getClientStats = async () => {
    return await request(() => api.getClientStats());
  };

  return {
    loading,
    error,
    getClients,
    createClient,
    updateClient,
    deleteClient,
    getClientStats,
  };
}

// 案件管理 Hook
export function useCases() {
  const { loading, error, request } = useApi();

  const getCases = async (params?: any) => {
    return await request(() => api.getCases(params));
  };

  const createCase = async (data: any) => {
    return await request(() => api.createCase(data));
  };

  const updateCase = async (id: number, data: any) => {
    return await request(() => api.updateCase(id, data));
  };

  const deleteCase = async (id: number) => {
    return await request(() => api.deleteCase(id));
  };

  const assignLawyer = async (caseId: number, lawyerId: number) => {
    return await request(() => api.assignLawyer(caseId, lawyerId));
  };

  const updateCaseStatus = async (caseId: number, status: string) => {
    return await request(() => api.updateCaseStatus(caseId, status));
  };

  const getCaseStats = async () => {
    return await request(() => api.getCaseStats());
  };

  return {
    loading,
    error,
    getCases,
    createCase,
    updateCase,
    deleteCase,
    assignLawyer,
    updateCaseStatus,
    getCaseStats,
  };
}
```

### 2. 组件示例

```typescript
// src/components/ClientList.tsx
import React, { useEffect, useState } from 'react';
import { useClients } from '../hooks/useApi';
import { Client } from '../types';

interface ClientListProps {
  onEdit?: (client: Client) => void;
  onDelete?: (client: Client) => void;
}

const ClientList: React.FC<ClientListProps> = ({ onEdit, onDelete }) => {
  const { loading, error, getClients, deleteClient } = useClients();
  const [clients, setClients] = useState<Client[]>([]);
  const [pagination, setPagination] = useState({
    page: 1,
    page_size: 10,
    total: 0,
    total_pages: 0,
  });

  useEffect(() => {
    loadClients();
  }, []);

  const loadClients = async (params?: any) => {
    const result = await getClients({
      page: pagination.page,
      page_size: pagination.page_size,
      ...params,
    });
    
    if (result.data) {
      setClients(result.data.data);
      setPagination(result.data.pagination);
    }
  };

  const handleDelete = async (client: Client) => {
    if (window.confirm(`确定要删除客户 "${client.name}" 吗？`)) {
      const result = await deleteClient(client.id);
      if (result.data) {
        loadClients(); // 重新加载列表
      }
    }
  };

  const handlePageChange = (newPage: number) => {
    setPagination(prev => ({ ...prev, page: newPage }));
    loadClients({ page: newPage });
  };

  if (loading && clients.length === 0) {
    return <div className="loading">加载中...</div>;
  }

  if (error) {
    return <div className="error">错误: {error}</div>;
  }

  return (
    <div className="client-list">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>姓名</th>
            <th>邮箱</th>
            <th>电话</th>
            <th>公司</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {clients.map(client => (
            <tr key={client.id}>
              <td>{client.id}</td>
              <td>{client.name}</td>
              <td>{client.email}</td>
              <td>{client.phone}</td>
              <td>{client.company || '-'}</td>
              <td>
                <span className={`status ${client.status}`}>
                  {client.status === 'active' ? '活跃' : '非活跃'}
                </span>
              </td>
              <td>
                <button onClick={() => onEdit?.(client)}>编辑</button>
                <button onClick={() => handleDelete(client)} className="delete">
                  删除
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* 分页控件 */}
      <div className="pagination">
        <button
          disabled={pagination.page <= 1}
          onClick={() => handlePageChange(pagination.page - 1)}
        >
          上一页
        </button>
        <span>
          第 {pagination.page} 页，共 {pagination.total_pages} 页
        </span>
        <button
          disabled={pagination.page >= pagination.total_pages}
          onClick={() => handlePageChange(pagination.page + 1)}
        >
          下一页
        </button>
      </div>
    </div>
  );
};

export default ClientList;
```

## 📊 错误处理

### 1. 统一错误处理

```typescript
// src/utils/errorHandler.ts
interface ApiError {
  code: string;
  message: string;
  details?: string;
  suggestions?: string[];
}

export function handleApiError(error: any): never {
  if (error.response) {
    const apiError: ApiError = error.response.data.error;
    
    switch (apiError.code) {
      case 'VALIDATION_ERROR':
        throw new Error(`验证错误: ${apiError.details}`);
      case 'AUTHENTICATION_ERROR':
        throw new Error('认证失败，请重新登录');
      case 'AUTHORIZATION_ERROR':
        throw new Error('权限不足');
      case 'NOT_FOUND':
        throw new Error('资源不存在');
      case 'CONFLICT':
        throw new Error('资源冲突');
      case 'RATE_LIMIT_ERROR':
        throw new Error('请求过于频繁，请稍后再试');
      default:
        throw new Error(apiError.message || '服务器错误');
    }
  } else if (error.request) {
    throw new Error('网络错误，请检查网络连接');
  } else {
    throw new Error('未知错误');
  }
}
```

### 2. 重试机制

```typescript
// src/utils/retry.ts
export async function withRetry<T>(
  fn: () => Promise<T>,
  maxRetries: number = 3,
  delay: number = 1000
): Promise<T> {
  let lastError: Error;
  
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;
      
      if (i < maxRetries - 1) {
        await new Promise(resolve => setTimeout(resolve, delay * (i + 1)));
      }
    }
  }
  
  throw lastError;
}
```

## 🚀 部署建议

### 1. 环境配置

```typescript
// src/config.ts
interface Config {
  apiUrl: string;
  timeout: number;
  retryAttempts: number;
}

const config: Config = {
  apiUrl: process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1',
  timeout: 30000,
  retryAttempts: 3,
};

export default config;
```

### 2. 生产环境优化

```typescript
// src/api/client.prod.ts
import ApiClient from './client';
import config from '../config';

class ProductionApiClient extends ApiClient {
  constructor() {
    super(config.apiUrl);
    
    // 添加请求拦截器
    this.addRequestInterceptor();
    
    // 添加响应拦截器
    this.addResponseInterceptor();
  }
  
  private addRequestInterceptor() {
    // 在实际项目中添加请求拦截器逻辑
  }
  
  private addResponseInterceptor() {
    // 在实际项目中添加响应拦截器逻辑
  }
}

export default new ProductionApiClient();
```

---

## 📞 支持

如果您在使用API时遇到问题，请：

1. 查看完整的 [API 文档](API.md)
2. 检查 [OpenAPI 规范](openapi.yaml)
3. 使用 Swagger UI 进行在线测试
4. 提交 Issue 到 GitHub 仓库

**最后更新**: 2024-01-01  
**API版本**: v1.0.0
