import React, { useState } from 'react'
import { Button, Card, Space, message, Spin } from 'antd'
import { caseAPI, clientAPI, lawyerAPI, documentAPI } from '../services/lawfirm'

const ApiTest: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<any>(null)

  const testAPI = async (apiCall: () => Promise<any>, apiName: string) => {
    setLoading(true)
    try {
      const response = await apiCall()
      setResult({ success: true, data: response, apiName })
      message.success(`${apiName} 测试成功`)
    } catch (error: any) {
      setResult({ success: false, error: error.message, apiName })
      message.error(`${apiName} 测试失败: ${error.message}`)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ padding: '20px' }}>
      <Card title='API 集成测试' style={{ marginBottom: '20px' }}>
        <Space wrap>
          <Button
            type='primary'
            onClick={() => testAPI(() => caseAPI.getList(), '案件列表')}
            loading={loading}
          >
            测试案件列表API
          </Button>

          <Button onClick={() => testAPI(() => clientAPI.getList(), '客户列表')} loading={loading}>
            测试客户列表API
          </Button>

          <Button onClick={() => testAPI(() => lawyerAPI.getList(), '律师列表')} loading={loading}>
            测试律师列表API
          </Button>

          <Button
            onClick={() => testAPI(() => documentAPI.getList(), '文档列表')}
            loading={loading}
          >
            测试文档列表API
          </Button>

          <Button
            onClick={() =>
              testAPI(
                () =>
                  clientAPI.create({
                    clientName: 'API测试客户',
                    phone: '13999999999',
                    email: 'apitest@test.com',
                    clientType: 'PERSONAL',
                    address: 'API测试地址',
                  }),
                '创建客户',
              )
            }
            loading={loading}
          >
            测试创建客户API
          </Button>
        </Space>
      </Card>

      {result && (
        <Card title={`${result.apiName} - 测试结果`}>
          <Spin spinning={loading}>
            <pre
              style={{
                background: result.success ? '#f6ffed' : '#fff2f0',
                padding: '10px',
                borderRadius: '4px',
                overflow: 'auto',
                maxHeight: '400px',
              }}
            >
              {JSON.stringify(result, null, 2)}
            </pre>
          </Spin>
        </Card>
      )}
    </div>
  )
}

export default ApiTest
