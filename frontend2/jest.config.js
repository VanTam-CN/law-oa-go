# 测试覆盖率配置
coverage:
  status:
    project:
      default:
        target: 80%
        threshold: 5%
    patch:
      default:
        target: 70%
        threshold: 5%

# 忽略的文件
ignore:
  - "src/test/**"
  - "**/*.test.ts"
  - "**/*.test.tsx"
  - "**/*.spec.ts"
  - "**/*.spec.tsx"
  - "src/main.tsx"
  - "src/vite-env.d.ts"
  - "src/assets/**"

# 包含的文件
include:
  - "src/**/*"

# 覆盖率报告格式
reporting:
  reports:
    - html
    - lcov
    - text
    - json

# 测试环境配置
testEnvironment: jsdom
setupFiles: ['./src/test/setup.ts']

# 测试匹配模式
testMatch:
  - "**/__tests__/**/*.(ts|tsx|js|jsx)"
  - "**/*.(test|spec).(ts|tsx|js|jsx)"

# 转换配置
transform:
  "^.+\\.(ts|tsx)$": "ts-jest"

# 模块名称映射
moduleNameMapping:
  "^@/(.*)$": "<rootDir>/src/$1"
  "^@components/(.*)$": "<rootDir>/src/components/$1"
  "^@pages/(.*)$": "<rootDir>/src/pages/$1"
  "^@utils/(.*)$": "<rootDir>/src/utils/$1"
  "^@api/(.*)$": "<rootDir>/src/api/$1"
  "^@assets/(.*)$": "<rootDir>/src/assets/$1"