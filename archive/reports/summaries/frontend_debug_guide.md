# 前端搜索功能调试指南

## 🔍 步骤1: 打开浏览器开发者工具

1. **打开客户管理页面**
2. **按F12** 或右键 → "检查" 打开开发者工具
3. **切换到Network（网络）标签页**

## 🔍 步骤2: 清除网络记录

1. 点击 **Clear** 按钮清除现有的网络记录
2. 勾选 **Preserve log** 保留日志
3. 在Filter过滤框中输入 `clients` 只显示客户相关请求

## 🔍 步骤3: 执行搜索操作

1. 在客户管理页面的搜索框中输入"张三"
2. 点击**搜索按钮**
3. 观察Network标签页中出现的请求

## 🔍 步骤4: 分析请求详情

点击出现的请求，查看以下信息：

### 📡 请求信息
- **Request URL**: 检查URL和参数是否正确
- **Request Method**: 应该是GET
- **Status Code**: 应该是200

### 📋 查询参数
在Headers标签页的Query String Parameters部分查看：
```
page: 1
page_size: 10
name: 张三
type: (如果设置了类型筛选)
status: (如果设置了状态筛选)
```

### 📦 响应信息
切换到Response或Preview标签页查看：
- **数据格式**: 应该是JSON格式
- **success字段**: 应该是true
- **data数组**: 查看返回的客户数量
- **pagination.total**: 查看总记录数

## 🔍 步骤5: 对比期望vs实际

### 期望的请求URL:
```
http://localhost:8080/api/v1/clients?page=1&page_size=10&name=张三
```

### 期望的响应格式:
```json
{
  "success": true,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "name": "张三",
      "type": "个人",
      "email": "zhangsan@example.com",
      "phone": "13800138000",
      "status": "active"
    }
  ],
  "pagination": {
    "total": 1,
    "page": 1,
    "size": 10
  }
}
```

## 🔍 步骤6: 常见问题排查

### 问题1: 参数名称错误
- **症状**: URL中包含`pageNum`而不是`page`
- **解决**: 检查`frontend/src/services/client.ts`中的参数映射

### 问题2: 响应格式不正确
- **症状**: 响应不是JSON格式或格式异常
- **解决**: 检查后端API是否正常运行

### 问题3: 搜索参数被忽略
- **症状**: 返回所有客户而不是搜索结果
- **解决**: 检查后端搜索逻辑是否正确实现

### 问题4: CORS错误
- **症状**: 控制台显示CORS相关错误
- **解决**: 检查后端CORS配置

## 🔍 步骤7: 控制台调试

在Console标签页中运行以下代码进行快速测试：

```javascript
// 测试API调用
fetch('http://localhost:8080/api/v1/clients?name=张三&page=1&page_size=10')
  .then(response => response.json())
  .then(data => {
    console.log('搜索结果:', data);
    console.log('记录数:', data.data?.length || 0);
    console.log('总数:', data.pagination?.total || 0);
  })
  .catch(error => console.error('请求失败:', error));
```

## 🔍 步骤8: 实时监控

在Console标签页中运行以下代码来监控所有API调用：

```javascript
// 拦截fetch请求
const originalFetch = window.fetch;
window.fetch = function(...args) {
  const [url] = args;
  if (url.includes('/clients')) {
    console.log('🔍 客户API请求:', url);
    return originalFetch.apply(this, args)
      .then(response => {
        console.log('📊 响应状态:', response.status);
        return response.json().then(data => {
          console.log('📦 响应数据:', data);
          return response;
        });
      })
      .catch(error => {
        console.log('❌ 请求失败:', error);
        throw error;
      });
  }
  return originalFetch.apply(this, args);
};
```

## 🎯 关键检查点

1. **URL参数**: 确认参数名称和值正确传递
2. **响应状态**: 确认API返回200状态码
3. **数据格式**: 确认返回正确的JSON格式
4. **搜索逻辑**: 确认返回的是搜索结果而非全部数据
5. **分页信息**: 确认pagination.total正确反映搜索结果总数

## 📞 需要提供的信息

如果问题仍然存在，请提供以下信息：

1. **完整的请求URL**
2. **完整的响应数据**
3. **浏览器控制台的错误信息**
4. **Network标签页中请求的Headers信息**

---

**调试完成后，请将观察到的结果与上述期望值进行对比，找出差异所在。**