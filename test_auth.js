#!/usr/bin/env node

const http = require('http');

// 测试登录和JWT token功能
async function testAuth() {
  console.log('🔍 测试认证流程...');
  
  // 1. 测试登录
  const loginData = {
    email: 'lawyer1@lawoa.com',
    password: 'password'
  };
  
  const postData = JSON.stringify(loginData);
  
  const loginOptions = {
    hostname: 'localhost',
    port: 8080,
    path: '/api/v1/auth/login',
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Content-Length': Buffer.byteLength(postData)
    }
  };
  
  try {
    // 登录获取token
    const loginResponse = await new Promise((resolve, reject) => {
      const req = http.request(loginOptions, (res) => {
        let data = '';
        res.on('data', (chunk) => data += chunk);
        res.on('end', () => resolve({ status: res.statusCode, data: JSON.parse(data) }));
      });
      req.on('error', reject);
      req.write(postData);
      req.end();
    });
    
    console.log('登录响应:', JSON.stringify(loginResponse, null, 2));
    
    if (loginResponse.status === 200 && loginResponse.data.success) {
      const token = loginResponse.data.data.token;
      console.log('✅ 登录成功，获取到token:', token.substring(0, 50) + '...');
      
      // 2. 测试使用token访问受保护的路由
      const protectedRoutes = [
        '/api/v1/users/profile',
        '/api/v1/dashboard/statistics',
        '/api/v1/clients',
        '/api/v1/cases'
      ];
      
      for (const route of protectedRoutes) {
        console.log(`\n测试路由: ${route}`);
        
        const routeOptions = {
          hostname: 'localhost',
          port: 8080,
          path: route,
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        };
        
        try {
          const routeResponse = await new Promise((resolve, reject) => {
            const req = http.request(routeOptions, (res) => {
              let data = '';
              res.on('data', (chunk) => data += chunk);
              res.on('end', () => resolve({ status: res.statusCode, data: data ? JSON.parse(data) : null }));
            });
            req.on('error', reject);
            req.end();
          });
          
          console.log(`状态: ${routeResponse.status}`);
          if (routeResponse.status === 200) {
            console.log('✅ 访问成功');
          } else {
            console.log('❌ 访问失败:', routeResponse.data);
          }
        } catch (error) {
          console.log('❌ 请求失败:', error.message);
        }
      }
    } else {
      console.log('❌ 登录失败:', loginResponse.data);
    }
  } catch (error) {
    console.log('❌ 测试失败:', error.message);
  }
}

testAuth();