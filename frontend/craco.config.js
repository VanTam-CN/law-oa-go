module.exports = {
  style: {
    postcss: {
      plugins: [
        require('tailwindcss'),
        require('autoprefixer'),
      ],
    },
  },
  devServer: {
    client: {
      overlay: {
        runtimeErrors: (error) => {
          // 过滤掉已知的开发环境警告
          const suppressedErrors = [
            'React Router Future Flag Warning',
            'Download the React DevTools',
            'Resource isn\'t a valid image'
          ];
          return !suppressedErrors.some(suppressed =>
            error.message.includes(suppressed)
          );
        }
      }
    }
  }
}