const mysql = require('mysql2/promise');
const bcrypt = require('bcrypt');

// 数据库配置
const dbConfig = {
  host: 'localhost',
  port: 3306,
  user: 'law_oa',
  password: 'law_oa_password',
  database: 'law_oa',
  charset: 'utf8mb4'
};

async function createTestUser() {
  let connection;

  try {
    // 连接数据库
    connection = await mysql.createConnection(dbConfig);
    console.log('✅ 数据库连接成功');

    // 检查用户表是否存在
    const [tables] = await connection.execute("SHOW TABLES LIKE 'users'");
    if (tables.length === 0) {
      console.log('❌ 用户表不存在，请先运行数据库迁移');
      return;
    }

    // 查询现有用户
    const [users] = await connection.execute('SELECT id, email, name, role FROM users LIMIT 5');
    console.log('\n📋 现有用户列表:');
    if (users.length === 0) {
      console.log('   暂无用户');
    } else {
      users.forEach(user => {
        console.log(`   ID: ${user.id}, Email: ${user.email}, Name: ${user.name}, Role: ${user.role}`);
      });
    }

    // 检查是否有admin@example.com用户
    const [existingUsers] = await connection.execute(
      'SELECT id, email FROM users WHERE email = ?',
      ['admin@example.com']
    );

    if (existingUsers.length > 0) {
      console.log('\n✅ admin@example.com 用户已存在');

      // 重置密码为admin123
      const hashedPassword = await bcrypt.hash('admin123', 10);
      await connection.execute(
        'UPDATE users SET password = ? WHERE email = ?',
        [hashedPassword, 'admin@example.com']
      );
      console.log('🔑 密码已重置为: admin123');
    } else {
      // 创建测试用户
      console.log('\n🔨 创建测试用户...');

      const hashedPassword = await bcrypt.hash('admin123', 10);

      const [result] = await connection.execute(
        'INSERT INTO users (email, password, name, role, phone, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())',
        ['admin@example.com', hashedPassword, '系统管理员', 'admin', '13800138000', 'active']
      );

      console.log(`✅ 测试用户创建成功，ID: ${result.insertId}`);
      console.log('📧 邮箱: admin@example.com');
      console.log('🔑 密码: admin123');
    }

    // 创建律师用户
    const [existingLawyers] = await connection.execute(
      'SELECT id, email FROM users WHERE email = ?',
      ['lawyer@example.com']
    );

    if (existingLawyers.length === 0) {
      console.log('\n🔨 创建律师用户...');

      const hashedPassword = await bcrypt.hash('lawyer123', 10);

      const [lawyerResult] = await connection.execute(
        'INSERT INTO users (username, email, password, name, role, phone, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())',
        ['lawyer', 'lawyer@example.com', hashedPassword, '张律师', 'lawyer', '13800138001', 'active']
      );

      console.log(`✅ 律师用户创建成功，ID: ${lawyerResult.insertId}`);
      console.log('📧 邮箱: lawyer@example.com');
      console.log('🔑 密码: lawyer123');
    }

    // 创建测试客户
    const [existingClients] = await connection.execute(
      'SELECT id, name FROM clients WHERE name = ?',
      ['测试客户']
    );

    if (existingClients.length === 0) {
      console.log('\n🔨 创建测试客户...');

      const [clientResult] = await connection.execute(
        'INSERT INTO clients (name, email, phone, address, company, notes, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())',
        ['测试客户', 'testclient@example.com', '13800138000', '测试地址', '测试公司', '这是一个测试客户', 'active']
      );

      console.log(`✅ 测试客户创建成功，ID: ${clientResult.insertId}`);
    }

    console.log('\n🎉 测试数据准备完成！');
    console.log('\n📝 登录信息:');
    console.log('   管理员: admin@example.com / admin123');
    console.log('   律师: lawyer@example.com / lawyer123');

  } catch (error) {
    console.error('❌ 操作失败:', error.message);
  } finally {
    if (connection) {
      await connection.end();
      console.log('\n🔌 数据库连接已关闭');
    }
  }
}

// 运行
createTestUser();