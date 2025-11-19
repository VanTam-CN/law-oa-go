# 🔐 认证设置说明

## 当前问题
用户在提交利益冲突审批申请时遇到 401 Unauthorized 错误，因为浏览器中缺少有效的JWT认证token。

## 🛠️ 解决方案

### 方法1：使用自动设置脚本（推荐）

1. 打开浏览器访问 `http://localhost:3003`
2. 按 `F12` 打开开发者工具，切换到 **Console** 标签页
3. 复制以下代码并粘贴到控制台中，然后按回车执行：

```javascript
(function() {
    const validToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6InRlc3QtdXNlci0wMDEiLCJyb2xlIjoibGF3eWVyIiwiZXhwIjoxNzYyNTAyNDg0LCJpYXQiOjE3NjI0MTYwODR9.uGUolYrXvG3Tx3BbyGuAoMkBXQvHsFfAKVqTB8p1oNQ';

    // 设置JWT token
    localStorage.setItem('auth_token', validToken);

    // 设置用户信息
    const userInfo = {
        id: 1,
        username: 'test-user-001',
        role: 'lawyer',
        roles: ['lawyer'],
        realName: '测试律师',
        email: 'test-user-001@law-oa.com'
    };
    localStorage.setItem('law_oa_user_info', JSON.stringify(userInfo));

    console.log('✅ JWT Token已设置！');
    console.log('👤 用户信息已更新！');
    console.log('🌐 现在可以正常提交审批申请了！');

    // 可选：自动刷新页面
    if (confirm('是否立即刷新页面以应用新的认证信息？')) {
        window.location.reload();
    }
})();
```

### 方法2：手动设置

如果自动脚本不工作，可以在控制台逐行执行：

```javascript
// 设置JWT token
localStorage.setItem('auth_token', 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6InRlc3QtdXNlci0wMDEiLCJyb2xlIjoibGF3eWVyIiwiZXhwIjoxNzYyNTAyNDg0LCJpYXQiOjE3NjI0MTYwODR9.uGUolYrXvG3Tx3BbyGuAoMkBXQvHsFfAKVqTB8p1oNQ');

// 设置用户信息
localStorage.setItem('law_oa_user_info', JSON.stringify({
    id: 1,
    username: 'test-user-001',
    role: 'lawyer',
    roles: ['lawyer'],
    realName: '测试律师',
    email: 'test-user-001@law-oa.com'
}));
```

## ✅ 验证设置成功

设置完成后，你应该能够：

1. ✅ 正常访问审批中心页面，无401错误
2. ✅ 看到审批统计数据
3. ✅ 成功提交利益冲突审批申请
4. ✅ 不再看到 "开发模式认证错误" 的提示

## 🔄 Token有效期

- 生成的token有效期：24小时
- 过期后需要重新设置或重新登录

## 🔍 故障排除

如果设置后仍然遇到问题：

1. **清除浏览器缓存**：按 `Ctrl+Shift+R` (Windows) 或 `Cmd+Shift+R` (Mac) 强制刷新
2. **检查localStorage**：在控制台输入 `localStorage.getItem('auth_token')` 确认token已设置
3. **重新设置token**：删除现有token后重新执行设置脚本

## 📞 技术支持

如果问题持续存在，请检查浏览器控制台是否有其他错误信息。