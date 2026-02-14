# 隔离墙管理界面组件使用说明

## 概述

隔离墙管理界面组件用于在案件详情页管理隔离墙和白名单功能。

## 组件文件

- `/frontend/src/types/ethicalWall.ts` - 类型定义
- `/frontend/src/hooks/useEthicalWall.ts` - API hooks
- `/frontend/src/components/case/EthicalWallSection.tsx` - 隔离墙状态显示组件
- `/frontend/src/components/case/WhitelistModal.tsx` - 白名单管理弹窗

## 使用方式

### 1. 在案件详情页中使用 EthicalWallSection

```tsx
import React from 'react'
import { EthicalWallSection } from '@/components/case'
import { useParams } from 'react-router'

const CaseDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()

  return (
    <div>
      {/* 其他案件信息 */}

      {/* 隔离墙管理组件 */}
      <EthicalWallSection
        caseId={id!}
        caseName="案件名称"
        compact={false}
      />
    </div>
  )
}

export default CaseDetail
```

### 2. Props 说明

#### EthicalWallSection

| 属性 | 类型 | 必填 | 说明 |
|------|------|------|------|
| caseId | string | 是 | 案件ID |
| caseName | string | 否 | 案件名称（用于弹窗显示） |
| className | string | 否 | 自定义类名 |
| style | React.CSSProperties | 否 | 自定义样式 |
| compact | boolean | 否 | 是否使用紧凑模式，默认 false |

### 3. API Hooks

#### useEthicalWall

获取案件隔离墙状态。

```tsx
import { useEthicalWall } from '@/components/case'

const { data, isLoading, error } = useEthicalWall(caseId)
```

#### useWhitelist

获取白名单列表。

```tsx
import { useWhitelist } from '@/components/case'

const { data: whitelistData, isLoading } = useWhitelist(caseId)
```

#### useEnableEthicalWall / useDisableEthicalWall

启用/禁用隔离墙。

```tsx
import { useEnableEthicalWall, useDisableEthicalWall } from '@/components/case'

const enableMutation = useEnableEthicalWall()
const disableMutation = useDisableEthicalWall()

// 启用
enableMutation.mutate(caseId, {
  onSuccess: () => console.log('已启用')
})

// 禁用
disableMutation.mutate(caseId, {
  onSuccess: () => console.log('已禁用')
})
```

#### useAddWhitelist / useRemoveWhitelist

添加/移除白名单用户。

```tsx
import { useAddWhitelist, useRemoveWhitelist } from '@/components/case'

const addMutation = useAddWhitelist()
const removeMutation = useRemoveWhitelist()

// 添加用户
addMutation.mutate({
  caseId,
  userId: 'user-123',
  reason: '需要访问该案件'
})

// 移除用户
removeMutation.mutate({
  caseId,
  userId: 'user-123'
})
```

## 后端 API 端点

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/cases/:caseId/ethical-wall` | 获取隔离墙状态 |
| POST | `/api/v1/cases/:caseId/ethical-wall` | 启用隔离墙 |
| DELETE | `/api/v1/cases/:caseId/ethical-wall` | 禁用隔离墙 |
| GET | `/api/v1/cases/:caseId/ethical-wall/whitelist` | 获取白名单 |
| POST | `/api/v1/cases/:caseId/ethical-wall/whitelist` | 添加白名单 |
| DELETE | `/api/v1/cases/:caseId/ethical-wall/whitelist/:userId` | 移除白名单 |

## 样式定制

组件使用 Less 样式文件，可以通过覆盖 CSS 类来定制样式：

```less
// 覆盖隔离墙卡片样式
.ethical-wall-section {
  &.enabled {
    border-color: #your-color;
  }
}

// 覆盖白名单弹窗样式
.whitelist-modal {
  .ant-modal-content {
    border-radius: 12px;
  }
}
```

## 响应式设计

组件支持以下断点：
- 1200px 以下：中等布局
- 768px 以下：平板布局
- 576px 以下：手机布局

## 无障碍支持

组件包含以下无障碍功能：
- ARIA 标签
- 键盘导航支持
- 焦点可见性
- 屏幕阅读器支持
