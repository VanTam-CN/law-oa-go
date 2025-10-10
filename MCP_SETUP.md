# MCP 服务器配置说明

## 配置文件

项目已创建 `claude_desktop_config.json` 配置文件：

```json
{
  "mcpServers": {
    "streamable-mcp-server": {
      "type": "streamable-http",
      "url": "http://127.0.0.1:12306/mcp"
    }
  }
}
```

## 使用方法

### 1. 对于 Claude Desktop 应用

将配置文件内容复制到 Claude Desktop 的配置文件中：
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

### 2. 对于其他 Claude 客户端

根据具体的Claude客户端配置要求，将MCP服务器配置添加到相应的配置文件中。

## 服务器要求

需要确保在 `http://127.0.0.1:12306/mcp` 上运行一个兼容的MCP服务器服务。

### 服务器状态检查

```bash
# 检查端口是否在监听
lsof -i :12306

# 或者使用netstat
netstat -an | grep 12306
```

## 注意事项

1. 确保MCP服务器正在运行
2. 检查防火墙设置是否允许端口12306的访问
3. 确认服务器端点 `/mcp` 正确响应MCP协议请求

## 故障排除

如果MCP服务无法连接，请检查：

1. 服务器是否正在运行
2. 端口12306是否可访问
3. 服务器是否正确实现了MCP协议
4. 网络连接是否正常

---

*配置文件已创建于项目根目录的 `claude_desktop_config.json`*