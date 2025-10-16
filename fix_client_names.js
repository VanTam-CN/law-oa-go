const axios = require('axios');

class ClientNameFixer {
  constructor() {
    this.baseURL = 'http://localhost:8080';
    this.token = null;
  }

  async login() {
    console.log('🔐 登录获取token...');
    try {
      const response = await axios.post(`${this.baseURL}/api/auth/login`, {
        email: 'admin@example.com',
        password: 'admin123'
      });
      
      if (response.data.success) {
        this.token = response.data.data.token;
        console.log('✅ 登录成功');
        return true;
      }
    } catch (error) {
      console.log('❌ 登录失败:', error.message);
      return false;
    }
  }

  async fixClientNames() {
    console.log('🔧 修复客户名称字段...');
    
    if (!this.token) {
      console.log('❌ 无有效token');
      return false;
    }

    const headers = { Authorization: `Bearer ${this.token}` };

    try {
      // 获取所有客户
      const response = await axios.get(`${this.baseURL}/api/clients?pageNum=1&pageSize=100`, { headers });
      const clients = response.data.data || [];

      console.log(`📊 找到${clients.length}个客户`);

      let fixedCount = 0;
      for (const client of clients) {
        // 如果name字段为空但company字段有值，则更新name字段
        if (!client.name && client.company) {
          try {
            // 只更新name字段，其他字段保持不变
            const updateData = {
              name: client.company
            };

            await axios.put(`${this.baseURL}/api/clients/${client.id}`, updateData, { headers });
            console.log(`✅ 修复客户${client.id}: ${client.company}`);
            fixedCount++;
          } catch (error) {
            console.log(`❌ 修复客户${client.id}失败:`, error.response?.data?.message || error.message);
          }
        }
      }

      console.log(`🎉 修复完成，共修复${fixedCount}个客户`);
      return fixedCount;

    } catch (error) {
      console.log('❌ 修复过程出错:', error.message);
      return false;
    }
  }

  async verifyFix() {
    console.log('🔍 验证修复结果...');
    
    if (!this.token) {
      console.log('❌ 无有效token');
      return false;
    }

    const headers = { Authorization: `Bearer ${this.token}` };

    try {
      const response = await axios.get(`${this.baseURL}/api/clients?pageNum=1&pageSize=100`, { headers });
      const clients = response.data.data || [];

      let emptyNameCount = 0;
      let validNameCount = 0;

      for (const client of clients) {
        if (!client.name) {
          emptyNameCount++;
          console.log(`⚠️ 客户${client.id}仍然缺少name字段`);
        } else {
          validNameCount++;
        }
      }

      console.log(`📊 验证结果:`);
      console.log(`   有效name字段: ${validNameCount}个`);
      console.log(`   空name字段: ${emptyNameCount}个`);

      return emptyNameCount === 0;

    } catch (error) {
      console.log('❌ 验证过程出错:', error.message);
      return false;
    }
  }

  async run() {
    console.log('🚀 开始修复客户名称字段...\n');

    const loginSuccess = await this.login();
    if (!loginSuccess) {
      console.log('❌ 登录失败，无法继续');
      return false;
    }

    const fixedCount = await this.fixClientNames();
    if (fixedCount === false) {
      console.log('❌ 修复失败');
      return false;
    }

    const verifySuccess = await this.verifyFix();
    
    console.log('\n📋 修复总结:');
    console.log(`   修复客户数量: ${fixedCount}`);
    console.log(`   验证结果: ${verifySuccess ? '✅ 成功' : '❌ 仍有问题'}`);

    return verifySuccess;
  }
}

// 运行修复
async function main() {
  const fixer = new ClientNameFixer();
  try {
    const success = await fixer.run();
    if (success) {
      console.log('\n🎉 客户名称修复完成！');
    } else {
      console.log('\n❌ 修复过程中遇到问题');
    }
  } catch (error) {
    console.error('❌ 修复失败:', error);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = ClientNameFixer;