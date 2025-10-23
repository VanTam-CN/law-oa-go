# 在线编辑架构设计

## 概述

基于前面研究的Quill.js、Monaco Editor、Yjs、Socket.IO等最新技术文档，设计一个支持富文本编辑、代码编辑和实时协作的在线编辑架构。

## 核心功能模块

### 1. 编辑器类型支持

#### 富文本编辑器 (Rich Text Editor)
- **技术栈**: Quill.js + QuillCursors + 自定义扩展
- **核心功能**:
  - 基础文本编辑：加粗、斜体、下划线、删除线
  - 段落格式：标题、列表、引用、代码块
  - 媒体嵌入：图片、视频、链接
  - 表格支持：quill-better-table扩展
  - 协作光标：quill-cursors显示多用户光标
  - 自定义模块：评论、批注、签名

#### 代码编辑器 (Code Editor)
- **技术栈**: Monaco Editor + 自定义语言支持
- **核心功能**:
  - 语法高亮：支持50+编程语言
  - 智能提示：IntelliSense代码补全
  - 错误检测：实时语法检查
  - 代码折叠：支持函数和代码块折叠
  - 多光标编辑：支持多位置同时编辑
  - 主题定制：支持多种编辑器主题
  - 快捷键：完整的快捷键支持

#### Markdown编辑器 (Markdown Editor)
- **技术栈**: Monaco Editor + Markdown扩展 + 实时预览
- **核心功能**:
  - 实时预览：分屏显示渲染效果
  - 语法扩展：支持GitHub Flavored Markdown
  - 数学公式：KaTeX数学公式渲染
  - 图表支持：Mermaid图表渲染
  - 文档大纲：自动生成目录结构

### 2. 实时协作架构

#### CRDT同步引擎
- **技术栈**: Yjs + y-websocket + y-protocols
- **核心组件**:
  - **Y文档管理**: 每个文档对应一个Y.Doc实例
  - **共享类型**: Y.Text用于文本，Y.Map用于结构化数据
  - **同步提供者**: WebSocketProvider实时同步
  - **离线支持**: 自动处理离线编辑和冲突解决
  - **持久化**: IndexDB/LevelDB本地存储

#### WebSocket通信层
- **技术栈**: Socket.IO + Redis Adapter
- **核心功能**:
  - **房间管理**: 基于文档ID的房间隔离
  - **用户认证**: JWT令牌验证
  - **消息路由**: 高效的消息广播机制
  - **负载均衡**: Redis多实例支持
  - **连接管理**: 自动重连和心跳检测

#### 协作感知 (Awareness)
- **技术栈**: y-protocols/awareness
- **核心功能**:
  - **用户状态**: 在线状态、用户信息
  - **光标位置**: 实时显示多用户光标
  - **选择范围**: 显示文本选择区域
  - **用户标识**: 不同颜色区分用户
  - **操作日志**: 记录用户操作历史

### 3. 版本控制系统

#### 文档版本管理
- **技术栈**: Git-like版本控制 + 增量存储
- **核心功能**:
  - **自动保存**: 定时保存编辑内容
  - **版本快照**: 关键节点创建快照
  - **差异比较**: 基于算法的差异计算
  - **分支合并**: 支持多分支编辑
  - **回滚操作**: 一键回滚到历史版本

#### 操作历史 (Operation History)
- **技术栈**: Yjs UndoManager + 自定义扩展
- **核心功能**:
  - **撤销重做**: 无限级撤销/重做
  - **操作分类**: 区分用户操作类型
  - **批量操作**: 合并连续的小操作
  - **历史压缩**: 压缩历史数据节省空间

### 4. 安全控制架构

#### 权限管理
- **技术栈**: ABAC + JWT令牌验证
- **权限级别**:
  - **读取权限**: 只读访问文档
  - **评论权限**: 可以添加评论和批注
  - **编辑权限**: 可以编辑文档内容
  - **管理权限**: 可以管理文档和协作者
  - **所有者权限**: 完全控制权限

#### 数据加密
- **传输加密**: HTTPS + WSS协议
- **存储加密**: AES-256数据加密
- **敏感信息**: 端到端加密支持

## 技术架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    前端编辑器界面                              │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ 富文本编辑器   │  │ 代码编辑器    │  │    Markdown编辑器    │  │
│  │  (Quill.js) │  │ (Monaco)    │  │   (Monaco + 预览)   │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   协作同步层                                 │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Yjs CRDT   │  │  Socket.IO  │  │    Awareness        │  │
│  │   引擎      │  │  WebSocket  │  │    用户状态          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   后端服务层                                 │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ 编辑服务      │  │ 权限服务      │  │    版本控制服务       │  │
│  │ (EditService)│  │(AuthService) │  │ (VersionService)    │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ 协作服务      │  │ 文档服务      │  │    通知服务          │  │
│  │(CollabService)│  │(DocService)  │  │ (NotifyService)     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   数据存储层                                 │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ PostgreSQL  │  │    Redis    │  │     MinIO/S3        │  │
│  │   元数据      │  │   缓存      │  │    文件存储          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 数据模型设计

### 编辑会话表 (edit_sessions)
```sql
CREATE TABLE edit_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id),
    session_token VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    editor_type VARCHAR(50) NOT NULL, -- 'rich-text', 'code', 'markdown'
    cursor_position JSONB,
    selection_range JSONB,
    activity_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,

    INDEX idx_edit_sessions_document_id (document_id),
    INDEX idx_edit_sessions_user_id (user_id),
    INDEX idx_edit_sessions_token (session_token),
    INDEX idx_edit_sessions_activity (activity_at)
);
```

### 编辑操作表 (edit_operations)
```sql
CREATE TABLE edit_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id),
    user_id UUID NOT NULL REFERENCES users(id),
    session_id UUID NOT NULL REFERENCES edit_sessions(id),
    operation_type VARCHAR(50) NOT NULL, -- 'insert', 'delete', 'format', 'cursor'
    operation_data JSONB NOT NULL,
    yjs_state_vector JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    applied BOOLEAN DEFAULT FALSE,

    INDEX idx_edit_operations_document (document_id, timestamp),
    INDEX idx_edit_operations_user (user_id, timestamp),
    INDEX idx_edit_operations_applied (applied)
);
```

### 协作会话表 (collaboration_sessions)
```sql
CREATE TABLE collaboration_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id),
    room_name VARCHAR(255) UNIQUE NOT NULL,
    active_users INTEGER DEFAULT 0,
    max_users INTEGER DEFAULT 10,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    INDEX idx_collaboration_document (document_id),
    INDEX idx_collaboration_room (room_name),
    INDEX idx_collaboration_activity (last_activity)
);
```

### 文档版本表 (document_versions)
```sql
CREATE TABLE document_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id),
    version_number INTEGER NOT NULL,
    title VARCHAR(255),
    content_hash VARCHAR(64),
    content_delta JSONB, -- Yjs增量数据
    snapshot_path VARCHAR(500), -- 完整快照路径
    editor_id UUID REFERENCES users(id),
    edit_summary TEXT,
    is_major_version BOOLEAN DEFAULT FALSE,
    is_published BOOLEAN DEFAULT FALSE,
    file_size BIGINT,
    character_count INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE (document_id, version_number),
    INDEX idx_versions_document (document_id, version_number),
    INDEX idx_versions_editor (editor_id),
    INDEX idx_versions_created (created_at)
);
```

## API接口设计

### 编辑器API
```go
// 创建编辑会话
POST /api/v1/documents/{documentId}/sessions
{
    "editor_type": "rich-text|code|markdown",
    "permissions": ["read", "write", "comment"]
}

// 获取编辑会话信息
GET /api/v1/sessions/{sessionId}

// 更新编辑状态
PUT /api/v1/sessions/{sessionId}/status
{
    "cursor_position": {"line": 1, "column": 10},
    "selection_range": {"start": {"line": 1, "column": 5}, "end": {"line": 1, "column": 15}},
    "activity": "editing"
}

// 关闭编辑会话
DELETE /api/v1/sessions/{sessionId}
```

### 协作API
```go
// 加入协作
POST /api/v1/documents/{documentId}/collaborate
{
    "session_id": "uuid",
    "user_info": {
        "name": "用户名",
        "avatar": "头像URL",
        "color": "#FF5722"
    }
}

// 获取协作用户列表
GET /api/v1/documents/{documentId}/collaborators

// 发送协作消息
POST /api/v1/documents/{documentId}/collaborate/message
{
    "type": "operation|cursor|selection|awareness",
    "data": {...}
}
```

### 版本控制API
```go
// 创建新版本
POST /api/v1/documents/{documentId}/versions
{
    "title": "版本标题",
    "edit_summary": "编辑说明",
    "is_major_version": false,
    "is_published": true
}

// 获取版本列表
GET /api/v1/documents/{documentId}/versions?page=1&limit=20

// 获取版本详情
GET /api/v1/documents/{documentId}/versions/{versionId}

// 比较版本差异
GET /api/v1/documents/{documentId}/versions/compare?from={fromId}&to={toId}

// 恢复到指定版本
POST /api/v1/documents/{documentId}/versions/{versionId}/restore
```

## 实时协作消息格式

### 操作消息 (Operation Message)
```json
{
    "type": "operation",
    "user_id": "uuid",
    "session_id": "uuid",
    "timestamp": 1640995200000,
    "data": {
        "origin": "user",
        "type": "insert|delete|retain|format",
        "position": 100,
        "content": "新增文本",
        "attributes": {"bold": true},
        "length": 4
    }
}
```

### 光标消息 (Cursor Message)
```json
{
    "type": "cursor",
    "user_id": "uuid",
    "session_id": "uuid",
    "timestamp": 1640995200000,
    "data": {
        "position": {"line": 5, "column": 12},
        "selection": {
            "start": {"line": 5, "column": 8},
            "end": {"line": 5, "column": 20}
        }
    }
}
```

### 用户状态消息 (Awareness Message)
```json
{
    "type": "awareness",
    "user_id": "uuid",
    "session_id": "uuid",
    "timestamp": 1640995200000,
    "data": {
        "name": "张三",
        "avatar": "https://example.com/avatar.jpg",
        "color": "#2196F3",
        "status": "online|away|busy",
        "last_activity": 1640995200000
    }
}
```

## 性能优化策略

### 1. 前端优化
- **虚拟滚动**: 大文档分页渲染
- **防抖节流**: 输入事件防抖处理
- **懒加载**: 按需加载编辑器模块
- **缓存策略**: 本地缓存常用数据

### 2. 网络优化
- **操作合并**: 合并连续的小操作
- **增量同步**: 只同步变更部分
- **压缩传输**: gzip消息压缩
- **连接池**: 复用WebSocket连接

### 3. 后端优化
- **消息队列**: 异步处理操作消息
- **批量写入**: 批量保存操作记录
- **缓存热点**: Redis缓存热点数据
- **水平扩展**: 支持多实例部署

## 安全措施

### 1. 认证授权
- JWT令牌验证
- 细粒度权限控制
- 会话超时管理
- 设备绑定验证

### 2. 数据保护
- 端到端加密
- 敏感数据脱敏
- 操作日志记录
- 数据备份恢复

### 3. 攻击防护
- XSS攻击防护
- CSRF攻击防护
- SQL注入防护
- DDoS攻击防护

## 监控指标

### 1. 性能指标
- 编辑响应时间 < 100ms
- 协作同步延迟 < 200ms
- 文档加载时间 < 2s
- 内存使用率 < 80%

### 2. 业务指标
- 并发编辑用户数
- 文档编辑频率
- 协作会话时长
- 版本创建频率

### 3. 系统指标
- WebSocket连接数
- 消息吞吐量
- 错误率统计
- 资源使用率

## 部署架构

### 容器化部署
```yaml
version: '3.8'
services:
  edit-service:
    image: law-oa-go/edit-service:latest
    replicas: 3
    environment:
      - DATABASE_URL=postgresql://...
      - REDIS_URL=redis://...
      - MINIO_ENDPOINT=...
    ports:
      - "8080:8080"

  websocket-server:
    image: law-oa-go/websocket-server:latest
    replicas: 2
    environment:
      - REDIS_URL=redis://...
    ports:
      - "8081:8081"
```

这个架构设计提供了完整的在线编辑解决方案，支持多种编辑器类型、实时协作、版本控制和安全控制，为律师事务所文档管理系统提供强大的编辑协作能力。