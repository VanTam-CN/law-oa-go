// 测试前端API配置
import fetch from 'node-fetch';

async function testFrontendAPI() {
  console.log('测试前端API配置...\n');
  
  // 测试通过前端代理的登录API
  try {
    console.log('1. 测试登录API...');
    const loginResponse = await fetch('http://localhost:3002/api/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email: 'admin@law-oa.com',
        password: 'password'
      })
    });
    
    console.log('状态码:', loginResponse.status);
    console.log('响应头:', JSON.stringify(Object.fromEntries(loginResponse.headers), null, 2));
    
    if (loginResponse.ok) {
      const data = await loginResponse.json();
      console.log('登录成功:', JSON.stringify(data, null, 2));
      
      // 测试获取案件数据
      if (data.token) {
        console.log('\n2. 测试案件管理API...');
        const casesResponse = await fetch('http://localhost:3002/api/cases', {
          headers: {
            'Authorization': `Bearer ${data.token}`,
            'Content-Type': 'application/json'
          }
        });
        
        console.log('案件API状态码:', casesResponse.status);
        if (casesResponse.ok) {
          const casesData = await casesResponse.json();
          console.log('案件数据:', JSON.stringify(casesData, null, 2));
        } else {
          console.log('案件API失败:', await casesResponse.text());
        }
        
        // 测试获取客户数据
        console.log('\n3. 测试客户管理API...');
        const clientsResponse = await fetch('http://localhost:3002/api/clients', {
          headers: {
            'Authorization': `Bearer ${data.token}`,
            'Content-Type': 'application/json'
          }
        });
        
        console.log('客户API状态码:', clientsResponse.status);
        if (clientsResponse.ok) {
          const clientsData = await clientsResponse.json();
          console.log('客户数据:', JSON.stringify(clientsData, null, 2));
        } else {
          console.log('客户API失败:', await clientsResponse.text());
        }
      }
    } else {
      console.log('登录失败:', await loginResponse.text());
    }
  } catch (error) {
    console.error('测试失败:', error.message);
  }
}

testFrontendAPI();