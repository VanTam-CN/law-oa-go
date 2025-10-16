#!/usr/bin/env node

/**
 * 简单的客户搜索功能测试脚本
 * 不依赖认证，直接测试搜索接口
 */

const axios = require('axios');

const API_BASE_URL = 'http://localhost:8080';

// 测试不同的客户搜索端点
async function testClientEndpoints() {
  console.log('🔍 开始测试客户搜索端点');
  console.log('=' .repeat(50));

  const endpoints = [
    { url: '/api/v1/clients', desc: 'v1版本的客户列表（带认证）' },
    { url: '/api/clients', desc: '旧版本的客户列表（带认证）' },
  ];

  const testParams = [
    { search: '张三' },
    { name: '张三' },
    { search: '张' },
    { company: '科技' },
    {}
  ];

  for (const endpoint of endpoints) {
    console.log(`\n📍 测试端点: ${endpoint.desc}`);
    console.log(`   URL: ${endpoint.url}`);

    for (const params of testParams) {
      console.log(`\n🔧 测试参数:`, JSON.stringify(params));

      try {
        const response = await axios.get(`${API_BASE_URL}${endpoint.url}`, {
          params: params,
          timeout: 5000
        });

        console.log('✅ 请求成功');
        console.log(`   状态码: ${response.status}`);

        let clientList = [];
        let total = 0;

        // 处理不同的响应格式
        if (response.data) {
          if (Array.isArray(response.data)) {
            clientList = response.data;
            total = response.data.length;
          } else if (response.data.data) {
            if (Array.isArray(response.data.data)) {
              clientList = response.data.data;
              total = response.data.pagination?.total || response.data.data.length;
            } else if (response.data.data.list) {
              clientList = response.data.data.list;
              total = response.data.data.total || response.data.data.list.length;
            }
          } else if (response.data.list) {
            clientList = response.data.list;
            total = response.data.total || response.data.list.length;
          }
        }

        console.log(`   返回记录数: ${total}`);

        if (clientList.length > 0) {
          console.log('   客户列表:');
          clientList.forEach((client, index) => {
            console.log(`     ${index + 1}. ID:${client.id} | 姓名:${client.name} | 类型:${client.type} | 状态:${client.status}`);
          });
        } else {
          console.log('   (无记录)');
        }

        // 分析搜索结果
        if (params.search === '张三' || params.name === '张三') {
          if (total === 1 && clientList[0]?.name === '张三') {
            console.log('✅ 搜索"张三"正确返回1条记录');
          } else if (total > 1) {
            console.log(`❌ 搜索"张三"返回${total}条记录，可能存在搜索问题`);
          } else {
            console.log('❌ 搜索"张三"没有返回记录');
          }
        }

      } catch (error) {
        if (error.response) {
          console.log(`❌ 请求失败: ${error.response.status} ${error.response.statusText}`);
          if (error.response.status === 401) {
            console.log('   原因: 需要认证');
          } else if (error.response.status === 404) {
            console.log('   原因: 端点不存在');
          }
        } else if (error.code === 'ECONNREFUSED') {
          console.log('❌ 连接被拒绝，请检查服务器是否运行');
        } else {
          console.log('❌ 请求失败:', error.message);
        }
      }
    }

    console.log('-'.repeat(40));
  }
}

// 测试健康检查端点
async function testHealthEndpoint() {
  console.log('\n🏥 测试健康检查端点');

  try {
    const response = await axios.get(`${API_BASE_URL}/health`);
    console.log('✅ 健康检查通过');
    console.log('响应:', response.data);
  } catch (error) {
    console.log('❌ 健康检查失败:', error.message);
  }
}

// 主函数
async function main() {
  console.log('🚀 客户搜索功能简单测试开始');
  console.log('API地址:', API_BASE_URL);

  // 测试健康检查
  await testHealthEndpoint();

  // 测试客户端点
  await testClientEndpoints();

  console.log('\n🎯 测试完成');
  console.log('如果需要认证，请先创建有效的用户并获取token');
}

// 运行测试
main().catch(console.error);