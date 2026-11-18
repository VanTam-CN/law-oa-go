import React, { useState, useEffect } from 'react'
import { Card, Table, Button, message, Spin, Empty } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

interface Lawyer {
  id: number
  name: string
  phone: string
  email: string
  licenseNumber: string
  department: string
  position: string
}

const SimpleLawyerManagement: React.FC = () => {
  const [lawyers, setLawyers] = useState<Lawyer[]>([])
  const [loading, setLoading] = useState<boolean>(false)

  // 获取律师列表
  const fetchLawyers = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/lawfirm/lawyers')
      const data = await response.json()

      if (data.code === 0 && data.data) {
        const convertedLawyers = data.data.list.map((lawyer: any) => ({
          id: lawyer.lawyerId,
          name: lawyer.lawyerName,
          phone: lawyer.phone,
          email: lawyer.email,
          licenseNumber: lawyer.licenseNo,
          department: lawyer.department || '',
          position: lawyer.position || '',
        }))
        setLawyers(convertedLawyers)
      } else {
        message.error('获取数据失败')
      }
    } catch (error) {
      console.error('获取律师列表失败:', error)
      message.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchLawyers()
  }, [])

  const columns: ColumnsType<Lawyer> = [
    {
      title: '姓名',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '联系电话',
      dataIndex: 'phone',
      key: 'phone',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '执业证号',
      dataIndex: 'licenseNumber',
      key: 'licenseNumber',
    },
    {
      title: '部门',
      dataIndex: 'department',
      key: 'department',
    },
    {
      title: '职位',
      dataIndex: 'position',
      key: 'position',
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      <Card
        title='律师管理 (简化版)'
        extra={
          <Button type='primary' icon={<PlusOutlined />}>
            新增律师
          </Button>
        }
      >
        {loading ? (
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <Spin size='large' tip='正在加载...' />
          </div>
        ) : lawyers.length === 0 ? (
          <Empty description='暂无律师数据' />
        ) : (
          <Table
            columns={columns}
            dataSource={lawyers}
            rowKey='id'
            pagination={{
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            }}
          />
        )}
      </Card>
    </div>
  )
}

export default SimpleLawyerManagement
