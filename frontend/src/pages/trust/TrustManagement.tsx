/**
 * 代管款综合管理页面
 * 包含账户管理、交易管理两个子模块
 */

import React, { useState } from 'react'
import { Card, Tabs } from 'antd'
import {
  WalletOutlined,
  TransactionOutlined,
} from '@ant-design/icons'
import TrustAccountManagement from './TrustAccountManagement'
import TrustTransactionManagement from './TrustTransactionManagement'
import './TrustManagement.less'

const TrustManagement: React.FC = () => {
  const [activeTab, setActiveTab] = useState('accounts')

  const tabItems = [
    {
      key: 'accounts',
      label: (
        <span>
          <WalletOutlined />
          账户管理
        </span>
      ),
      children: <TrustAccountManagement />,
    },
    {
      key: 'transactions',
      label: (
        <span>
          <TransactionOutlined />
          交易管理
        </span>
      ),
      children: <TrustTransactionManagement />,
    },
  ]

  return (
    <div className="trust-management">
      <Card
        title={
          <span>
            <WalletOutlined style={{ marginRight: 8 }} />
            代管款管理
          </span>
        }
        bordered={false}
      >
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          size="large"
        />
      </Card>
    </div>
  )
}

export default TrustManagement
