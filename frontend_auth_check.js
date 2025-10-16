// 前端认证状态检查脚本
// 在浏览器控制台中运行此脚本

console.log('🔍 前端认证状态检查');
console.log('=====================================');

// 1. 检查token存储
console.log('📋 检查认证令牌:');
const authToken = localStorage.getItem('auth_token');
const lawOaToken = localStorage.getItem('law_oa_token');
const storageToken = localStorage.getItem('token');

console.log('   auth_token:', authToken ? '✅ 已设置' : '❌ 未设置');
console.log('   law_oa_token:', lawOaToken ? '✅ 已设置' : '❌ 未设置');
console.log('   token:', storageToken ? '✅ 已设置' : '❌ 未设置');

if (authToken) {
    console.log('   auth_token内容:', authToken.substring(0, 50) + '...');
}

// 2. 设置测试token（如果需要）
if (!authToken && !lawOaToken) {
    console.log('\n⚠️ 未找到认证令牌，设置测试令牌...');

    const testToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJ1c2VybmFtZSI6ImFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzYwMjUwODQ2LCJpYXQiOjE3NjAxNjQ0NDZ9.4N-Gj2OCUQQRb_sAh1lxxdGyROfn591sFCQ_kNRSOtc';

    localStorage.setItem('auth_token', testToken);
    localStorage.setItem('law_oa_token', testToken);

    console.log('✅ 测试令牌已设置');
    console.log('   请刷新页面后重试搜索功能');
}

// 3. 测试API调用
console.log('\n🧪 测试API调用:');

// 测试客户列表API
const testClientAPI = async () => {
    const token = localStorage.getItem('auth_token') || localStorage.getItem('law_oa_token');

    if (!token) {
        console.log('❌ 没有认证令牌，无法测试API');
        return;
    }

    try {
        console.log('   📡 测试基本客户列表...');

        const response = await fetch('/api/v1/clients?page=1&page_size=10', {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });

        console.log('   📊 响应状态:', response.status);

        if (response.ok) {
            const data = await response.json();
            console.log('   ✅ API调用成功');
            console.log('   📦 数据格式:', data.success ? '正确' : '异常');
            console.log('   📋 客户数量:', data.data ? data.data.length : 0);
            console.log('   📈 总数:', data.pagination ? data.pagination.total : 0);

            // 显示前几个客户
            if (data.data && data.data.length > 0) {
                console.log('   👥 客户示例:');
                data.data.slice(0, 3).forEach((client, index) => {
                    console.log(`      ${index + 1}. ${client.name} (${client.type})`);
                });
            }

            // 测试搜索
            console.log('\n   🔍 测试搜索功能...');
            const searchResponse = await fetch('/api/v1/clients?name=张三&page=1&page_size=10', {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (searchResponse.ok) {
                const searchData = await searchResponse.json();
                console.log('   ✅ 搜索API调用成功');
                console.log('   📋 搜索结果数量:', searchData.data ? searchData.data.length : 0);
                console.log('   📈 搜索总数:', searchData.pagination ? searchData.pagination.total : 0);

                if (searchData.data && searchData.data.length > 0) {
                    console.log('   🎯 搜索结果:');
                    searchData.data.forEach((client, index) => {
                        console.log(`      ${index + 1}. ${client.name} (${client.type}) - ${client.status}`);
                    });
                }

                // 检查搜索结果是否正确
                if (searchData.data && searchData.data.length === 1 && searchData.data[0].name === '张三') {
                    console.log('   🎉 搜索功能正常！');
                } else if (searchData.data && searchData.data.length > 1) {
                    console.log('   ⚠️ 搜索结果过多，可能存在搜索逻辑问题');
                } else {
                    console.log('   ❌ 没有找到搜索结果，可能存在搜索逻辑问题');
                }
            } else {
                console.log('   ❌ 搜索API调用失败:', searchResponse.status);
            }
        } else {
            console.log('   ❌ API调用失败:', response.status);
            const text = await response.text();
            console.log('   📄 错误信息:', text);
        }
    } catch (error) {
        console.log('   ❌ API调用异常:', error);
    }
};

// 执行测试
testClientAPI();

// 4. 检查前端配置
console.log('\n⚙️ 检查前端配置:');
console.log('   API baseURL:', window.location.origin + '/api');
console.log('   当前页面:', window.location.href);

// 5. 提供修复建议
console.log('\n💡 修复建议:');
console.log('1. 如果没有认证令牌，已自动设置测试令牌');
console.log('2. 请刷新页面后重试搜索功能');
console.log('3. 如果搜索仍然返回19条记录，请查看Network标签页的详细请求');
console.log('4. 检查请求URL中的参数是否正确传递');

console.log('\n🚀 检查完成！');