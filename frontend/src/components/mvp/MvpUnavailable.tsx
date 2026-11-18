import React from 'react'
import { Button, Result, Typography } from 'antd'
import { useNavigate } from 'react-router'

const { Paragraph } = Typography

interface MvpUnavailableProps {
  moduleName: string
}

const MvpUnavailable: React.FC<MvpUnavailableProps> = ({ moduleName }) => {
  const navigate = useNavigate()

  return (
    <Result
      status='info'
      title={`${moduleName}未纳入本次 MVP 试用范围`}
      subTitle={
        <Paragraph style={{ marginBottom: 0 }}>
          当前试用版聚焦主任工作台、案件、利益冲突、客户、审批和信托账户流程。
        </Paragraph>
      }
      extra={
        <Button type='primary' onClick={() => navigate('/dashboard')}>
          返回工作台
        </Button>
      }
    />
  )
}

export default MvpUnavailable
