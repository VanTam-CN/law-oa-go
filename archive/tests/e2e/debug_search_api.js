// 调试客户搜索API的测试脚本
const http = require('http');

// 测试不同的API调用方式
async function testSearchAPI() {
  console.log('🔍 开始测试客户搜索API');
  console.log('=====================================');

  const baseUrl = 'http://localhost:8080'; // 假设后端运行在8080端口

  // 测试1: 使用name参数（前端当前方式）
  console.log('\n📋 测试1: 使用name参数搜索"张三"');
  await makeRequest('/clients?page=1&page_size=10&name=张三');

  // 测试2: 使用search参数
  console.log('\n📋 测试2: 使用search参数搜索"张三"');
  await makeRequest('/clients?page=1&page_size=10&search=张三');

  // 测试3: 查看所有客户（无搜索条件）
  console.log('\n📋 测试3: 获取所有客户（无搜索条件）');
  await makeRequest('/clients?page=1&page_size=10');

  // 测试4: 测试类型筛选
  console.log('\n📋 测试4: 筛选个人客户');
  await makeRequest('/clients?page=1&page_size=10&type=个人');
}

function makeRequest(path) {
  return new Promise((resolve, reject) => {
    const url = `${baseUrl}${path}`;
    console.log(`🌐 请求: ${url}`);

    const options = {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        // 如果需要认证，添加Authorization头
        // 'Authorization': 'Bearer your-token-here'
      }
    };

    const req = http.request(url, options, (res) => {
      let data = '';

      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        console.log(`📊 状态码: ${res.statusCode}`);
        try {
          const jsonData = JSON.parse(data);
          console.log('📦 响应数据:');
          if (jsonData.data && Array.isArray(jsonData.data)) {
            console.log(`   - 返回记录数: ${jsonData.data.length}`);
            if (jsonData.pagination) {
              console.log(`   - 总记录数: ${jsonData.pagination.total}`);
            }
            // 显示前几条记录的名称
            jsonData.data.slice(0, 3).forEach((client, index) => {
              console.log(`   - 记录${index + 1}: ${client.name} (${client.type})`);
            });
          } else {
            console.log('   - 数据格式:', JSON.stringify(jsonData, null, 2));
          }
        } catch (e) {
          console.log('   - 原始响应:', data);
        }
        console.log('');
        resolve();
      });
    });

    req.on('error', (err) => {
      console.log(`❌ 请求失败: ${err.message}`);
      reject(err);
    });

    req.end();
  });
}

// 运行测试
testSearchAPI().catch(console.error);