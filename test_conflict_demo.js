console.log('🚀 利益冲突检测算法演示...\n');

// 模拟历史案例
const cases = [
  { id: 1, clientName: '张三', title: '张三诉李四合同纠纷案' },
  { id: 2, clientName: '某科技有限公司', title: '某科技公司诉某企业侵权案' }
];

function checkConflict(clientName) {
  console.log(`检查客户: ${clientName}`);

  const matches = cases.filter(c =>
    c.clientName.toLowerCase().includes(clientName.toLowerCase())
  );

  console.log(`找到 ${matches.length} 个冲突案例`);
  matches.forEach(c => console.log(`- ${c.title}`));

  return matches.length > 0;
}

// 测试
checkConflict('张三');
checkConflict('新客户');