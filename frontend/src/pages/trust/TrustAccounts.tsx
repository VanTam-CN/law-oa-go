import React from 'react'
import { PageContainer } from '@ant-design/pro-layout'
import { Card, Typography } from 'antd'
import TrustAccountList from '../../components/trust/TrustAccountList'

const { Title } = Typography

const TrustAccounts: React.FC = () => {
  return (
    <PageContainer
      title="代管款管理"
      content="管理客户代管款账户，包括存款、取款、转账等操作"
    >
      <TrustAccountList />
    </PageContainer>
  )
}

export default TrustAccounts
