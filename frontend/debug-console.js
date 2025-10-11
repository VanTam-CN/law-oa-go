// 测试脚本 - 在浏览器控制台中运行
console.log('=== 开始测试 ===');

// 测试1: 检查React是否正常
try {
  console.log('1. React版本:', React.version);
} catch (e) {
  console.error('React未加载:', e);
}

// 测试2: 检查路由
try {
  console.log('2. 当前路径:', window.location.pathname);
} catch (e) {
  console.error('无法获取路径:', e);
}

// 测试3: 测试API调用
fetch('/api/lawfirm/lawyers')
  .then(response => response.json())
  .then(data => {
    console.log('3. API测试成功:', data);
  })
  .catch(error => {
    console.error('3. API测试失败:', error);
  });

// 测试4: 检查DOM元素
try {
  const root = document.getElementById('root');
  console.log('4. Root元素:', root ? '存在' : '不存在');
  console.log('   Root内容:', root?.innerHTML?.substring(0, 100));
} catch (e) {
  console.error('无法检查DOM:', e);
}

// 测试5: 检查localStorage
try {
  const token = localStorage.getItem('law_oa_token');
  console.log('5. Token状态:', token ? '存在' : '不存在');
  
  const userInfo = localStorage.getItem('law_oa_user_info');
  console.log('6. 用户信息:', userInfo ? '存在' : '不存在');
} catch (e) {
  console.error('无法检查localStorage:', e);
}

console.log('=== 测试完成 ===');