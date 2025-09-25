import fetch from 'node-fetch';

// 测试登录和获取数据
async function test() {
  // 1. 登录
  const loginRes = await fetch('http://localhost:3002/api/auth/login', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({email: 'admin@law-oa.com', password: 'password'})
  });
  
  const loginData = await loginRes.json();
  console.log('登录成功:', loginData.success);
  
  if (loginData.success) {
    const token = loginData.data.token;
    
    // 2. 测试案件API
    const casesRes = await fetch('http://localhost:3002/api/cases', {
      headers: {Authorization: `Bearer ${token}`}
    });
    const casesData = await casesRes.json();
    console.log('案件API状态:', casesRes.status, '数据条数:', casesData.data?.length || 0);
    
    // 3. 测试客户API
    const clientsRes = await fetch('http://localhost:3002/api/clients', {
      headers: {Authorization: `Bearer ${token}`}
    });
    const clientsData = await clientsRes.json();
    console.log('客户API状态:', clientsRes.status, '数据条数:', clientsData.data?.length || 0);
  }
}

test().catch(console.error);