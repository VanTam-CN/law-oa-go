#!/usr/bin/env node

const http = require('http');

// 测试前端页面访问和数据获取
async function testFrontendPages() {
  console.log('🔍 测试前端页面访问和数据获取...\n');
  
  // 1. 首先登录获取token
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
    
    if (loginResponse.status === 200 && loginResponse.data.success) {
      const token = loginResponse.data.data.token;
      console.log('✅ 登录成功，获取到token');
      
      // 2. 测试律师用户应该能访问的页面数据
      const lawyerAccessTests = [
        {
          name: '用户资料',
          url: '/api/v1/users/profile',
          expected: 200
        },
        {
          name: '仪表盘统计',
          url: '/api/v1/dashboard/statistics',
          expected: 200
        },
        {
          name: '客户列表',
          url: '/api/v1/clients',
          expected: 200
        },
        {
          name: '客户统计',
          url: '/api/v1/clients/stats',
          expected: 200
        },
        {
          name: '案件列表',
          url: '/api/v1/cases',
          expected: 200
        },
        {
          name: '案件统计',
          url: '/api/v1/cases/stats',
          expected: 200
        }
      ];
      
      console.log('\n📋 测试律师用户访问权限:');
      
      for (const test of lawyerAccessTests) {
        console.log(`\n测试: ${test.name}`);
        
        const options = {
          hostname: 'localhost',
          port: 8080,
          path: test.url,
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        };
        
        try {
          const response = await new Promise((resolve, reject) => {
            const req = http.request(options, (res) => {
              let data = '';
              res.on('data', (chunk) => data += chunk);
              res.on('end', () => {
                try {
                  const parsedData = data ? JSON.parse(data) : null;
                  resolve({ status: res.statusCode, data: parsedData });
                } catch (e) {
                  resolve({ status: res.statusCode, data: data });
                }
              });
            });
            req.on('error', reject);
            req.end();
          });
          
          if (response.status === test.expected) {
            console.log(`✅ ${test.name}: 访问成功 (${response.status})`);
            
            // 检查是否有实际数据
            if (response.data && response.data.data) {
              const data = response.data.data;
              if (Array.isArray(data)) {
                console.log(`   📊 数据量: ${data.length} 条记录`);
              } else if (typeof data === 'object') {
                console.log(`   📊 数据类型: ${Object.keys(data).join(', ')}`);
              }
            } else if (response.data && response.data.success === false) {
              console.log(`   ⚠️  API返回成功但无数据: ${response.data.error?.message || '未知原因'}`);
            }
          } else {
            console.log(`❌ ${test.name}: 访问失败 (${response.status})`);
            if (response.data) {
              console.log(`   错误信息: ${response.data.error?.message || response.data.message || '未知错误'}`);
            }
          }
        } catch (error) {
          console.log(`❌ ${test.name}: 请求失败 - ${error.message}`);
        }
      }
      
      // 3. 测试律师用户不应该能访问的管理员页面
      const adminAccessTests = [
        {
          name: '用户管理',
          url: '/api/v1/admin/users',
          expected: 403
        }
      ];
      
      console.log('\n📋 测试律师用户管理员权限限制:');
      
      for (const test of adminAccessTests) {
        console.log(`\n测试: ${test.name}`);
        
        const options = {
          hostname: 'localhost',
          port: 8080,
          path: test.url,
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        };
        
        try {
          const response = await new Promise((resolve, reject) => {
            const req = http.request(options, (res) => {
              let data = '';
              res.on('data', (chunk) => data += chunk);
              res.on('end', () => {
                try {
                  const parsedData = data ? JSON.parse(data) : null;
                  resolve({ status: res.statusCode, data: parsedData });
                } catch (e) {
                  resolve({ status: res.statusCode, data: data });
                }
              });
            });
            req.on('error', reject);
            req.end();
          });
          
          if (response.status === test.expected) {
            console.log(`✅ ${test.name}: 正确拒绝访问 (${response.status})`);
          } else {
            console.log(`❌ ${test.name}: 权限控制异常 (${response.status})`);
          }
        } catch (error) {
          console.log(`❌ ${test.name}: 请求失败 - ${error.message}`);
        }
      }
      
      console.log('\n🎉 前端页面访问测试完成！');
      
    } else {
      console.log('❌ 登录失败:', loginResponse.data);
    }
  } catch (error) {
    console.log('❌ 测试失败:', error.message);
  }
}

testFrontendPages();