import React, { useState } from 'react';
import { Card, Form, Input, Button, Result, Divider, Space } from 'antd';
import { CalculatorOutlined } from '@ant-design/icons';
import { calculateLitigationFee, LitigationFeeParams, LitigationFeeResult } from '@/services/tools';
import './LitigationFeeCalculator.less';

const { TextArea } = Input;

const LitigationFeeCalculator: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState<boolean>(false);
  const [result, setResult] = useState<LitigationFeeResult | null>(null);

  const handleSubmit = async (values: LitigationFeeParams) => {
    try {
      setLoading(true);
      const response = await calculateLitigationFee(values);
      setResult(response.data);
    } catch (error) {
      console.error('Failed to calculate litigation fee:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('zh-CN', {
      style: 'currency',
      currency: 'CNY'
    }).format(amount);
  };

  return (
    <div className="litigation-fee-calculator">
      <Card title="诉讼费计算器" className="calculator-card">
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            amount: 100000
          }}
        >
          <Form.Item
            name="amount"
            label="标的额（元）"
            rules={[
              { required: true, message: '请输入标的额' },
              { type: 'number', min: 0, message: '标的额必须大于0' }
            ]}
          >
            <Input
              type="number"
              placeholder="请输入诉讼标的额"
              addonAfter="元"
            />
          </Form.Item>

          <Form.Item>
            <Button 
              type="primary" 
              htmlType="submit" 
              loading={loading}
              icon={<CalculatorOutlined />}
            >
              计算诉讼费
            </Button>
          </Form.Item>
        </Form>

        {result && (
          <>
            <Divider />
            <Result
              status="success"
              title="计算完成"
              subTitle={
                <div className="result-content">
                  <p><strong>标的额：</strong>{formatCurrency(result.amount)}</p>
                  <p><strong>诉讼费：</strong>{formatCurrency(result.fee)}</p>
                  <p className="fee-note">
                    注：根据《诉讼费用交纳办法》计算，具体以法院核算为准
                  </p>
                </div>
              }
            />
          </>
        )}

        <Divider orientation="left">收费标准参考</Divider>
        <div className="fee-standard">
          <TextArea
            value={`财产案件根据诉讼请求的金额或价额，按照下列比例分段累计交纳：
1. 不超过1万元的：交纳50元
2. 超过1万元至10万元的部分：按照2.5%交纳
3. 超过10万元至20万元的部分：按照2%交纳
4. 超过20万元至50万元的部分：按照1.5%交纳
5. 超过50万元至100万元的部分：按照1%交纳
6. 超过100万元至200万元的部分：按照0.9%交纳
7. 超过200万元至500万元的部分：按照0.8%交纳
8. 超过500万元至1000万元的部分：按照0.7%交纳
9. 超过1000万元至2000万元的部分：按照0.6%交纳
10. 超过2000万元的部分：按照0.5%交纳`}
            rows={10}
            readOnly
          />
        </div>
      </Card>
    </div>
  );
};

export default LitigationFeeCalculator;