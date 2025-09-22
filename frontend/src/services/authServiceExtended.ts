export const forgotPassword = async (email: string) => {
  // 简单实现，实际应该调用API
  console.log('Password reset requested for:', email);
  return { success: true, message: 'Password reset email sent' };
};

export const resetPassword = async (token: string, password: string) => {
  // 简单实现，实际应该调用API
  console.log('Password reset with token:', token);
  return { success: true, message: 'Password reset successfully' };
};