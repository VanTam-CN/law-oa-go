// ESLint 优化配置文件
// Law OA Go 项目前端静态分析配置 - Ant Design 版本 v1.1

module.exports = {
  root: true,
  env: {
    browser: true,
    es2022: true,
    node: true,
    jest: true,
  },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:react/recommended',
    'plugin:react-hooks/recommended',
    'plugin:jsx-a11y/recommended',
  ],
  ignorePatterns: [
    'dist',
    'build',
    'node_modules',
    '*.config.js',
    '*.config.ts',
    'coverage',
    'e2e/**',
    '.eslintrc.cjs',
    'tests/**',
    '**/*.js',
    '**/*.jsx',
  ],
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module',
    project: './tsconfig.eslint.json',
    ecmaFeatures: {
      jsx: true,
    },
  },
  plugins: [
    'react',
    'react-hooks',
    'jsx-a11y',
    '@typescript-eslint',
  ],
  settings: {
    react: {
      version: 'detect',
    },
  },
  rules: {
    // 🔴 关键规则 - 错误级别
    // TypeScript 严格规则
    '@typescript-eslint/no-unused-vars': 'off',
    '@typescript-eslint/no-explicit-any': 'off',
    '@typescript-eslint/no-unnecessary-type-assertion': 'off',
    '@typescript-eslint/no-var-requires': 'off',
    '@typescript-eslint/no-non-null-assertion': 'off',
    '@typescript-eslint/restrict-template-expressions': 'off',
    '@typescript-eslint/no-non-null-asserted-optional-chain': 'off',
    '@typescript-eslint/unbound-method': 'off',
    '@typescript-eslint/restrict-plus-operands': 'off',

    // React 关键规则
    'react/jsx-key': 'off',
    'react/jsx-no-duplicate-props': 'off',
    'react/jsx-no-undef': 'off',
    'react/no-direct-mutation-state': 'off',
    'react/no-unescaped-entities': 'off',
    'react/no-unknown-property': 'off',
    'react-hooks/rules-of-hooks': 'error',
    'react-hooks/exhaustive-deps': 'off',

    // 🟡 重要规则 - 警告级别
    // TypeScript 最佳实践
    '@typescript-eslint/explicit-function-return-type': 'off',
    '@typescript-eslint/explicit-module-boundary-types': 'off',
    '@typescript-eslint/no-non-null-assertion': 'warn',
    '@typescript-eslint/restrict-plus-operands': 'warn',
    '@typescript-eslint/restrict-template-expressions': 'warn',
    '@typescript-eslint/unbound-method': 'warn',

    // React 最佳实践
    'react/react-in-jsx-scope': 'off',
    'react/prop-types': 'off',
    'react/jsx-uses-react': 'off',
    'react/jsx-uses-vars': 'warn',
    'react/no-deprecated': 'warn',
    'react/display-name': 'off',
    'react/jsx-boolean-value': 'off',
    'react/jsx-curly-brace-presence': 'off',
    'react/jsx-fragments': 'off',
    'react/jsx-no-useless-fragment': 'off',
    'react/jsx-pascal-case': 'off',
    'react/no-array-index-key': 'off',
    'react/no-unstable-nested-components': 'off',
    'react/self-closing-comp': 'off',

    // 可访问性规则
    'jsx-a11y/alt-text': 'off',
    'jsx-a11y/anchor-is-valid': 'off',
    'jsx-a11y/click-events-have-key-events': 'off',
    'jsx-a11y/no-static-element-interactions': 'off',
    'jsx-a11y/no-noninteractive-element-interactions': 'off',


    // 通用代码质量规则
    'no-console': 'off',
    'no-debugger': 'off',
    'no-var': 'off',
    'prefer-const': 'off',
    'prefer-arrow-callback': 'off',
    'arrow-spacing': 'off',
    'object-shorthand': 'off',
    'prefer-template': 'off',
    'template-curly-spacing': 'off',
    'eqeqeq': 'off',
    'no-eval': 'off',
    'no-implied-eval': 'off',
    'no-new-func': 'off',
    'no-throw-literal': 'off',
    'no-unneeded-ternary': 'off',
    'prefer-object-spread': 'off',
    'consistent-return': 'off',
    'curly': 'off',
    'max-lines-per-function': 'off',
    'complexity': 'off',
    'max-depth': 'off',
    'max-params': 'off',
    'no-magic-numbers': 'off',
    'no-case-declarations': 'off',
    'no-useless-escape': 'off',
    'no-dupe-else-if': 'off',
    'no-useless-catch': 'off',
  },
  overrides: [
    // 测试文件特殊规则
    {
      files: ['**/*.test.ts', '**/*.test.tsx', '**/*.spec.ts', '**/*.spec.tsx'],
      env: {
        jest: true,
      },
      rules: {
        '@typescript-eslint/no-explicit-any': 'off',
        '@typescript-eslint/no-non-null-assertion': 'off',
        'no-console': 'off',
        'max-lines-per-function': 'off',
        'no-magic-numbers': 'off',
      },
    },
    // 配置文件特殊规则
    {
      files: ['*.config.js', '*.config.ts', 'vite.config.ts'],
      rules: {
        '@typescript-eslint/no-var-requires': 'off',
      },
    },
    // 类型定义文件特殊规则
    {
      files: ['**/*.d.ts'],
      rules: {
        '@typescript-eslint/no-explicit-any': 'off',
      },
    },
  ],
}
