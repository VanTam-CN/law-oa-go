// 前端API请求调试脚本
// 在浏览器控制台中运行此脚本来调试前端请求

console.log('🔍 前端API请求调试工具');
console.log('=====================================');

// 保存原始的fetch方法
const originalFetch = window.fetch;

// 重写fetch方法以拦截所有请求
window.fetch = function(...args) {
  const [url, options] = args;

  // 只拦截API请求
  if (url.includes('/clients')) {
    console.log('📡 客户API请求:');
    console.log('   URL:', url);

    if (options && options.method === 'GET') {
      // 解析URL参数
      const urlObj = new URL(url);
      const params = {};
      urlObj.searchParams.forEach((value, key) => {
        params[key] = value;
      });
      console.log('   参数:', params);
    }

    if (options && options.body) {
      console.log('   请求体:', options.body);
    }

    // 执行原始请求并记录响应
    return originalFetch.apply(this, args)
      .then(response => {
        console.log('📊 响应状态:', response.status);

        // 克隆响应以便读取body
        const clonedResponse = response.clone();

        return clonedResponse.json().then(data => {
          console.log('📦 响应数据:', data);

          // 特别分析客户列表响应
          if (data.data && Array.isArray(data.data)) {
            console.log('📋 客户列表分析:');
            console.log(`   - 返回记录数: ${data.data.length}`);
            if (data.pagination) {
              console.log(`   - 总记录数: ${data.pagination.total}`);
            }

            // 显示前3条客户记录
            data.data.slice(0, 3).forEach((client, index) => {
              console.log(`   - 记录${index + 1}: ${client.name} (${client.type}) - ${client.status}`);
            });
          }

          return response; // 返回原始响应
        }).catch(() => {
          console.log('   响应不是JSON格式');
          return response;
        });
      })
      .catch(error => {
        console.log('❌ 请求失败:', error);
        throw error;
      });
  }

  // 非客户API请求，直接执行
  return originalFetch.apply(this, args);
};

console.log('✅ API拦截器已启用');
console.log('💡 现在可以在客户管理页面进行搜索测试');
console.log('📋 测试建议:');
console.log('1. 搜索"张三"');
console.log('2. 筛选客户类型"个人"');
console.log('3. 筛选客户状态"活跃"');
console.log('4. 分页操作');
console.log('');
console.log('🔧 查看控制台输出以了解详细的请求和响应信息');