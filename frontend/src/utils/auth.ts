// 临时的认证工具，用于测试
export const setAuthToken = (token: string) => {
  localStorage.setItem('auth_token', token);
};

export const getAuthToken = () => {
  return localStorage.getItem('auth_token');
};

export const removeAuthToken = () => {
  localStorage.removeItem('auth_token');
};

// 临时设置一个测试token
export const setTestToken = () => {
  // 使用刚才获取的token
  const testToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJ1c2VybmFtZSI6ImFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzYwMDkzMjkzLCJpYXQiOjE3NjAwMDY4OTN9.868MdMFobxA9bth5oOGvXPMVnDvkdNfAE9U9Vq29I4s';
  
  // 同时设置到两个地方确保兼容性
  setAuthToken(testToken);
  localStorage.setItem('law_oa_token', testToken);
  
  console.log('测试token已设置到 auth_token 和 law_oa_token');
};