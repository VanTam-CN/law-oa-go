// 开发环境认证Token设置脚本
// 在浏览器控制台中运行此脚本来设置有效的JWT token

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
    console.log('🌐 现在可以正常访问审批API了！');
    console.log('💡 请刷新页面以使设置生效');

    // 可选：自动刷新页面
    if (confirm('是否立即刷新页面以应用新的认证信息？')) {
        window.location.reload();
    }
})();