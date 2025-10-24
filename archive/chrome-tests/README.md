# Chrome测试文件归档

**归档日期**: 2025-10-23
**归档原因**: 这些是Chrome测试浏览器和DevTools验证相关的文件，占用了170MB空间，但对核心系统功能不必要。

## 归档内容

### chrome/ 目录 (40MB)
- Chrome测试浏览器文件
- 用于前端自动化测试
- 包含Chrome for Testing二进制文件

### chrome-devtools-validation/ 目录 (130MB)
- Chrome DevTools验证项目
- 独立的测试项目
- 包含完整的node_modules和测试文件

## 恢复说明

如果需要重新运行Chrome相关测试，可以从以下位置恢复：
- 从git历史恢复这些文件
- 重新下载Chrome for Testing
- 重新构建DevTools验证项目

## 清理效果

- 节省磁盘空间: ~170MB
- 减少项目扫描时间
- 简化项目结构