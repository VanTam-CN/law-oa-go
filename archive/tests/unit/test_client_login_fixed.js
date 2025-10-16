const jwt = require('jsonwebtoken');

// 测试登录并获取有效令牌
async function testLoginAndGetToken() {
  const loginData = {
    email: 'admin@law-oa.com',
    password: 'secret'
  };

  console.log('尝试登录获取有效令牌...');
  const loginResponse = await fetch('http://localhost:8080/api/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(loginData)
  });

  console.log('登录响应状态:', loginResponse.status);

  if (loginResponse.ok) {
    const responseText = await loginResponse.text();
    console.log('登录响应文本:', responseText);

    try {
      const loginResult = JSON.parse(responseText);
      console.log('登录响应JSON:', JSON.stringify(loginResult, null, 2));

      // 提取token
      const token = loginResult.token || loginResult.data?.token;
      if (token) {
        console.log('获取到的有效令牌:', token);

        // 使用这个令牌测试客户API
        await testClientAPIsWithToken(token);
      } else {
        console.log('登录响应中没有找到token');
      }
    } catch (parseError) {
      console.log('解析JSON失败:', parseError.message);
    }
  } else {
    const errorText = await loginResponse.text();
    console.log('登录失败:', errorText);
  }
}

// 使用有效令牌测试客户API
async function testClientAPIsWithToken(token) {
  console.log('\n=== 测试客户列表API ===');
  const clientListResponse = await fetch('http://localhost:8080/api/v1/clients', {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  console.log('客户列表响应状态:', clientListResponse.status);
  if (clientListResponse.ok) {
    const clientData = await clientListResponse.json();
    console.log('客户列表响应:', JSON.stringify(clientData, null, 2));
  } else {
    const errorText = await clientListResponse.text();
    console.log('客户列表错误:', errorText);
  }

  console.log('\n=== 测试客户统计API ===');
  const statsResponse = await fetch('http://localhost:8080/api/v1/clients/stats', {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  console.log('统计响应状态:', statsResponse.status);
  if (statsResponse.ok) {
    const statsData = await statsResponse.json();
    console.log('统计响应:', JSON.stringify(statsData, null, 2));
  } else {
    const errorText = await statsResponse.text();
    console.log('统计错误:', errorText);
  }

  console.log('\n=== 测试创建客户 ===');
  const newClient = {
    client_name: '测试客户公司',
    email: 'test@example.com',
    phone: '13800138000',
    address: '北京市朝阳区测试地址',
    company: '测试客户有限公司',
    notes: '这是一个测试客户',
    status: 'active'
  };

  const createResponse = await fetch('http://localhost:8080/api/v1/clients', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(newClient)
  });

  console.log('创建客户响应状态:', createResponse.status);
  if (createResponse.ok) {
    const createData = await createResponse.json();
    console.log('创建客户响应:', JSON.stringify(createData, null, 2));

    // 如果创建成功，测试获取详情
    if (createData.data?.id) {
      await testGetClientDetail(token, createData.data.id);
      await testUpdateClient(token, createData.data.id);
      await testDeleteClient(token, createData.data.id);
    }
  } else {
    const errorText = await createResponse.text();
    console.log('创建客户错误:', errorText);
  }
}

// 测试获取客户详情
async function testGetClientDetail(token, clientId) {
  console.log(`\n=== 测试获取客户详情 (ID: ${clientId}) ===`);
  const detailResponse = await fetch(`http://localhost:8080/api/v1/clients/${clientId}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  console.log('获取详情响应状态:', detailResponse.status);
  if (detailResponse.ok) {
    const detailData = await detailResponse.json();
    console.log('客户详情:', JSON.stringify(detailData, null, 2));
  } else {
    const errorText = await detailResponse.text();
    console.log('获取详情错误:', errorText);
  }
}

// 测试更新客户
async function testUpdateClient(token, clientId) {
  console.log(`\n=== 测试更新客户 (ID: ${clientId}) ===`);
  const updateData = {
    client_name: '更新后的测试客户',
    email: 'updated@example.com',
    phone: '13900139000',
    notes: '这是一个更新后的测试客户'
  };

  const updateResponse = await fetch(`http://localhost:8080/api/v1/clients/${clientId}`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(updateData)
  });

  console.log('更新客户响应状态:', updateResponse.status);
  if (updateResponse.ok) {
    const updateResult = await updateResponse.json();
    console.log('更新客户响应:', JSON.stringify(updateResult, null, 2));
  } else {
    const errorText = await updateResponse.text();
    console.log('更新客户错误:', errorText);
  }
}

// 测试删除客户
async function testDeleteClient(token, clientId) {
  console.log(`\n=== 测试删除客户 (ID: ${clientId}) ===`);
  const deleteResponse = await fetch(`http://localhost:8080/api/v1/clients/${clientId}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  console.log('删除客户响应状态:', deleteResponse.status);
  if (deleteResponse.ok) {
    const deleteResult = await deleteResponse.json();
    console.log('删除客户响应:', JSON.stringify(deleteResult, null, 2));
  } else {
    const errorText = await deleteResponse.text();
    console.log('删除客户错误:', errorText);
  }
}

// 执行测试
testLoginAndGetToken().catch(console.error);