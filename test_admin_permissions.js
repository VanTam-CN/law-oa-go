#!/usr/bin/env node

const http = require('http');

// 测试登录admin@example.com账号
const loginData = JSON.stringify({
  email: 'admin@example.com',
  password: '123456'
});

const loginOptions = {
  hostname: 'localhost',
  port: 8080,
  path: '/v1/auth/login',
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Content-Length': Buffer.byteLength(loginData)
  }
};

const loginReq = http.request(loginOptions, (res) => {
  console.log('登录响应状态:', res.statusCode);
  console.log('登录响应头:', JSON.stringify(res.headers, null, 2));
  
  let data = '';
  res.on('data', (chunk) => {
    data += chunk;
  });
  
  res.on('end', () => {
    try {
      const response = JSON.parse(data);
      console.log('登录响应数据:', JSON.stringify(response, null, 2));
      
      if (response.data && response.data.token) {
        // 测试获取案件列表
        testCaseList(response.data.token);
        // 测试获取客户列表
        testClientList(response.data.token);
      }
    } catch (e) {
      console.error('解析响应失败:', e);
      console.log('原始响应:', data);
    }
  });
});

loginReq.on('error', (e) => {
  console.error('登录请求失败:', e);
});

loginReq.write(loginData);
loginReq.end();

function testCaseList(token) {
  console.log('\n=== 测试获取案件列表 ===');
  
  const caseOptions = {
    hostname: 'localhost',
    port: 8080,
    path: '/v1/cases',
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  };
  
  const caseReq = http.request(caseOptions, (res) => {
    console.log('案件列表响应状态:', res.statusCode);
    
    let data = '';
    res.on('data', (chunk) => {
      data += chunk;
    });
    
    res.on('end', () => {
      try {
        const response = JSON.parse(data);
        console.log('案件列表响应数据:', JSON.stringify(response, null, 2));
      } catch (e) {
        console.error('解析案件列表响应失败:', e);
        console.log('原始响应:', data);
      }
    });
  });
  
  caseReq.on('error', (e) => {
    console.error('案件列表请求失败:', e);
  });
  
  caseReq.end();
}

function testClientList(token) {
  console.log('\n=== 测试获取客户列表 ===');
  
  const clientOptions = {
    hostname: 'localhost',
    port: 8080,
    path: '/v1/clients',
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  };
  
  const clientReq = http.request(clientOptions, (res) => {
    console.log('客户列表响应状态:', res.statusCode);
    
    let data = '';
    res.on('data', (chunk) => {
      data += chunk;
    });
    
    res.on('end', () => {
      try {
        const response = JSON.parse(data);
        console.log('客户列表响应数据:', JSON.stringify(response, null, 2));
      } catch (e) {
        console.error('解析客户列表响应失败:', e);
        console.log('原始响应:', data);
      }
    });
  });
  
  clientReq.on('error', (e) => {
    console.error('客户列表请求失败:', e);
  });
  
  clientReq.end();
}