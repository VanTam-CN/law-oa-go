# 代码质量保障体系

Law OA
Go 项目前端配置了完整的代码质量保障体系，包括 ESLint、Prettier、husky 和 lint-staged，确保代码质量和一致性。

## 🔧 工具配置

### ESLint

- **配置文件**: `.eslintrc.cjs`
- **功能**: 静态代码分析，检查语法错误和代码规范
- **运行命令**:
  ```bash
  npm run lint        # 检查代码质量
  npm run lint:fix     # 自动修复可修复的问题
  ```

### Prettier

- **配置文件**: `.prettierrc`
- **功能**: 代码格式化工具，保持代码风格一致
- **运行命令**:
  ```bash
  npm run format      # 格式化代码
  npm run format:check # 检查代码格式
  ```

### TypeScript 类型检查

- **配置文件**: `tsconfig.json`
- **功能**: 静态类型检查
- **运行命令**:
  ```bash
  npm run type-check  # TypeScript 类型检查
  ```

## 🪝 Git Hooks

### Pre-commit 钩子

在每次提交前自动运行以下检查：

1. **lint-staged**: 对暂存文件运行 ESLint 和 Prettier
2. **TypeScript 类型检查**: 确保类型安全
3. **格式检查**: 确保代码格式正确

### Commit Message 钩子

使用 commitlint 检查提交信息格式，确保符合约定式提交规范。

## 📝 提交信息规范

项目使用约定式提交格式：

```
<type>(<scope>): <subject>

<body>

<footer>
```

### 类型说明

- `feat`: 新功能
- `fix`: 修复问题
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 代码重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动
- `revert`: 回滚提交
- `build`: 构建系统或外部依赖变动
- `ci`: CI配置文件和脚本变动
- `cd`: CD部署配置变动

### 示例

```bash
feat(auth): 添加JWT令牌刷新功能
fix: 修复案件列表分页问题
docs: 更新API文档
style: 调整组件样式格式
```

## 🔍 Lint-staged 配置

对暂存文件执行以下操作：

- `*.{ts,tsx}`: 运行 ESLint 修复和 Prettier 格式化
- `*.{css,less,json,md}`: 运行 Prettier 格式化
- `*.{js,jsx}`: 运行 ESLint 修复

## 📁 忽略文件

以下文件/目录不会被格式化和检查：

- `node_modules/`
- `dist/`
- `build/`
- `coverage/`
- `*.config.js`
- `*.config.ts`
- `.env*`
- `.husky/`

## 🚀 使用建议

### 开发流程

1. 开发完成后，运行代码质量检查：

   ```bash
   npm run lint:fix    # 修复 ESLint 问题
   npm run format      # 格式化代码
   npm run type-check  # 类型检查
   ```

2. 暂存文件：

   ```bash
   git add .
   ```

3. 提交代码（pre-commit 钩子会自动运行检查）：
   ```bash
   git commit -m "feat(component): 添加新功能"
   ```

### VS Code 集成

推荐安装以下 VS Code 扩展：

- ESLint
- Prettier
- TypeScript Importer

在 VS Code 设置中启用：

```json
{
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "typescript.preferences.importModuleSpecifier": "relative"
}
```

## ⚠️ 注意事项

1. **提交前检查**: 所有代码提交前都会自动运行质量检查，不通过则无法提交
2. **格式化**: 代码会自动格式化，确保团队风格一致
3. **类型安全**: TypeScript 严格模式，确保类型安全
4. **提交信息**: 必须符合约定式提交规范

## 🐛 常见问题

### ESLint 报错过多

- 大部分错误可以通过 `npm run lint:fix` 自动修复
- 剩余错误需要手动修复，通常涉及类型定义或逻辑问题

### Prettier 格式冲突

- 如果 Prettier 与 ESLint 规则冲突，以 Prettier 为主
- 可以在 `.eslintrc.cjs` 中禁用冲突的规则

### 类型检查失败

- 检查 `tsconfig.json` 配置
- 确保导入路径正确
- 检查类型定义文件

### 提交信息不通过

- 确保提交信息符合约定式提交格式
- 类型必须是预定义的类型之一
- 主题不能为空，不能以句号结尾

这套代码质量保障体系确保了项目的代码质量、一致性和可维护性，是团队协作开发的重要基础。
