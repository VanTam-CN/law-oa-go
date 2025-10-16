// 临时的认证工具，用于测试
export const setAuthToken = (token: string) => {
  localStorage.setItem('auth_token', token);
};

export const getAuthToken = () => {
  // 首先尝试从新的位置获取
  let token = localStorage.getItem('auth_token');

  // 如果没有，从旧的位置获取
  if (!token) {
    token = localStorage.getItem('law_oa_token');
  }

  // 如果还没有，尝试从sessionStorage获取
  if (!token) {
    token = sessionStorage.getItem('auth_token') || sessionStorage.getItem('law_oa_token');
  }

  return token;
};

export const removeAuthToken = () => {
  localStorage.removeItem('auth_token');
};

// 临时设置一个测试token
export const setTestToken = () => {
  // 使用最新的有效token
  const testToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJ1c2VybmFtZSI6ImFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzYwMjUwODQ2LCJpYXQiOjE3NjAxNjQ0NDZ9.4N-Gj2OCUQQRb_sAh1lxxdGyROfn591sFCQ_kNRSOtc';
  
  // 同时设置到两个地方确保兼容性
  setAuthToken(testToken);
  localStorage.setItem('law_oa_token', testToken);
  
  console.log('测试token已设置到 auth_token 和 law_oa_token');
};