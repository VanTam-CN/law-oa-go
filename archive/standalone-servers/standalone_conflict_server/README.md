# 独立冲突检测服务

这是一个独立的冲突检测服务，用于验证修复后的利益冲突检测功能。

## 功能特性

✅ **核心修复验证**
- 路由器中冲突检测服务正确初始化
- User模型新增Department和Seniority字段
- 仓储方法调用修复（GetByID → FindByID）
- 类型转换处理（string → uint）

✅ **完整的冲突检测功能**
- 商业竞争冲突检测
- 法律对立关系冲突检测
- 行业竞争分析
- 风险评估和建议生成
- 检测历史记录

## 快速开始

### 1. 启动服务器

```bash
cd standalone_conflict_server
go run main.go
```

服务器将在 `http://localhost:8080` 启动。

### 2. 运行API测试

在另一个终端中运行：

```bash
cd standalone_conflict_server
go run test_client.go
```

## API端点

| 方法 | 端点 | 描述 |
|------|------|------|
| GET | `/` | 服务信息 |
| POST | `/api/v1/conflict/check` | 执行冲突检测 |
| GET | `/api/v1/conflict/health` | 健康检查 |
| GET | `/api/v1/conflict/history` | 获取检测历史 |
| GET | `/api/v1/conflict/stats` | 获取统计数据 |

## 测试数据

服务启动时会自动创建以下测试数据：

### 律师用户
- 张律师（诉讼部，中级）
- 李律师（合规部，高级）
- 王律师（公司部，初级）

### 客户
- 腾讯科技（互联网科技）
- 字节跳动（互联网科技）
- 阿里巴巴（电子商务）
- 美团（本地生活）

### 测试案件
- 腾讯诉字节跳动不正当竞争案
- 阿里巴巴商业秘密保护案
- 美团平台合作协议纠纷

## 冲突检测示例

### 请求示例

```json
{
  "clientId": "1",
  "clientName": "字节跳动",
  "clientType": "COMPANY",
  "otherParties": ["腾讯科技"],
  "caseName": "短视频平台版权纠纷",
  "caseType": "知识产权",
  "searchYears": 5,
  "includeCorporateRelations": true,
  "searchDepth": "STANDARD",
  "userId": 1,
  "requestTime": "2025-10-22T13:50:00Z"
}
```

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "checkId": "CC_1724321400123456789",
    "hasConflict": true,
    "conflictCases": [
      {
        "id": "industry_1",
        "caseName": "腾讯诉字节跳动不正当竞争案",
        "conflictType": "行业竞争冲突",
        "riskLevel": "HIGH",
        "description": "同行业竞争: '字节跳动' 与 '腾讯科技' 存在业务竞争"
      }
    ],
    "riskAssessment": {
      "overallRisk": "HIGH",
      "riskScore": 75.5,
      "riskReason": "HIGH风险：行业竞争冲突类型冲突1个",
      "requiresApproval": true,
      "riskFactors": ["高风险冲突1个"]
    },
    "recommendations": [
      "建议立即回避相关案件",
      "咨询专业律师或律协意见",
      "向客户充分披露潜在冲突"
    ]
  }
}
```

## 技术架构

- **Web框架**: Gin (Go)
- **数据库**: SQLite (本地文件)
- **ORM**: GORM
- **依赖注入**: 手动注入
- **CORS**: 全局支持

## 文件结构

```
standalone_conflict_server/
├── main.go          # 主服务器文件
├── test_client.go   # API测试客户端
├── go.mod          # Go模块依赖
├── README.md       # 本文档
└── conflict_detection.db  # SQLite数据库（自动创建）
```

## 验证修复内容

此独立服务验证了以下核心修复：

1. ✅ **路由器服务初始化** - 冲突检测服务不再为nil
2. ✅ **User模型字段补全** - Department和Seniority字段正常工作
3. ✅ **仓储方法调用** - FindByID方法调用正确
4. ✅ **类型转换** - string到uint转换处理正确
5. ✅ **风险评估** - 完整的风险评估算法
6. ✅ **API响应格式** - 标准化的JSON响应

## 故障排除

如果遇到问题：

1. **端口占用**: 修改 `main.go` 中的端口号
2. **数据库问题**: 删除 `conflict_detection.db` 文件重启
3. **依赖问题**: 运行 `go mod tidy` 更新依赖

## 下一步

1. 验证独立服务正常工作
2. 集成到主项目中
3. 修复其他服务的编译问题
4. 进行完整的前后端集成测试