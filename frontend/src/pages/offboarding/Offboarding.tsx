import React from 'react'
import { PageContainer } from '@ant-design/pro-layout'
import { Card, Typography } from 'antd'
import OffboardingManagement from '../../components/offboarding/OffboardingManagement'

const { Title } = Typography

const Offboarding: React.FC = () => {
  return (
    <PageContainer
      title="离职交接管理"
      content="管理员工离职交接流程，包括案件移交、待办事项移交、文档处理等"
    >
      <OffboardingManagement />
    </PageContainer>
  )
}

export default Offboarding
