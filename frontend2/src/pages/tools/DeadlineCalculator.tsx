import React, { useState } from 'react';
import { Card, Form, DatePicker, Button, Result, Divider, Checkbox, Space, Input, message } from 'antd';
import { CalendarOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;

const DeadlineCalculator: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState<boolean>(false);
  const [result, setResult] = useState<any>(null);

  const handleSubmit = async (values: any) => {
    try {
      setLoading(true);
      
      // 简单的前端计算逻辑
      const startDate = dayjs(values.startDate);
      const days = values.days;
      const excludeWeekends = values.excludeWeekends || false;
      
      let endDate = startDate;
      let workDays = 0;
      let totalDays = 0;
      
      // 简单的日期计算
      while (workDays < days) {
        endDate = endDate.add(1, 'day');
        totalDays++;
        
        const dayOfWeek = endDate.day();
        if (!excludeWeekends || (dayOfWeek !== 0 && dayOfWeek !== 6)) {
          workDays++;
        }
      }
      
      const resultData = {
        startDate: values.startDate,
        days: days,
        excludeWeekends: excludeWeekends,
        excludeHolidays: values.excludeHolidays || false,
        endDate: endDate.format('YYYY-MM-DD'),
        workDays: workDays
      };
      
      setResult(resultData);
      message.success('计算完成！');
    } catch (error) {
      console.error('Failed to calculate deadline:', error);
      message.error('计算失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (date: string) => {
    return dayjs(date).format('YYYY年MM月DD日');
  };

  return (
    <div style={{ padding: '24px' }}>
      <Card title="工期计算器" style={{ maxWidth: 800, margin: '0 auto' }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            days: 15,
            excludeWeekends: true,
            excludeHolidays: false
          }}
        >
          <Form.Item
            name="startDate"
            label="开始日期"
            rules={[{ required: true, message: '请选择开始日期' }]}
          >
            <DatePicker 
              style={{ width: '100%' }}
              placeholder="请选择开始日期"
            />
          </Form.Item>

          <Form.Item
            name="days"
            label="计算天数"
            rules={[
              { required: true, message: '请输入天数' },
              { type: 'number', min: 1, message: '天数必须大于0' }
            ]}
          >
            <Input
              type="number"
              placeholder="请输入计算天数"
              addonAfter="天"
            />
          </Form.Item>

          <Form.Item name="excludeWeekends" valuePropName="checked">
            <Checkbox>排除周末（周六、周日）</Checkbox>
          </Form.Item>

          <Form.Item name="excludeHolidays" valuePropName="checked">
            <Checkbox>排除法定节假日</Checkbox>
          </Form.Item>

          <Form.Item>
            <Button 
              type="primary" 
              htmlType="submit" 
              loading={loading}
              icon={<CalendarOutlined />}
            >
              计算截止日期
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
                <div style={{ textAlign: 'left', marginTop: 20 }}>
                  <div style={{ marginBottom: 20 }}>
                    <p><strong>开始日期：</strong>{formatDate(result.startDate)}</p>
                    <p><strong>总天数：</strong>{result.days}天</p>
                    <p><strong>工作日：</strong>{result.workDays}天</p>
                    <p><strong>截止日期：</strong>{formatDate(result.endDate)}</p>
                  </div>
                  
                  <div style={{ padding: 15, background: '#f5f5f5', borderRadius: 4 }}>
                    <p><strong>计算设置：</strong></p>
                    <p>排除周末：{result.excludeWeekends ? '是' : '否'}</p>
                    <p>排除节假日：{result.excludeHolidays ? '是' : '否'}</p>
                  </div>
                </div>
              }
            />
          </>
        )}

        <Divider orientation="left">常见期限参考</Divider>
        <div>
          <Space direction="vertical" style={{ width: '100%' }}>
            <div style={{ padding: 10, marginBottom: 8, background: '#f8f9fa', borderLeft: '3px solid #1890ff', borderRadius: 4 }}>
              <strong style={{ color: '#1890ff' }}>民事诉讼时效：</strong>3年（自知道或应当知道权利受损之日起计算）
            </div>
            <div style={{ padding: 10, marginBottom: 8, background: '#f8f9fa', borderLeft: '3px solid #1890ff', borderRadius: 4 }}>
              <strong style={{ color: '#1890ff' }}>劳动争议申请仲裁：</strong>1年（自知道或应当知道权利被侵害之日起计算）
            </div>
            <div style={{ padding: 10, marginBottom: 8, background: '#f8f9fa', borderLeft: '3px solid #1890ff', borderRadius: 4 }}>
              <strong style={{ color: '#1890ff' }}>行政复议申请：</strong>60日（自知道该具体行政行为之日起计算）
            </div>
            <div style={{ padding: 10, marginBottom: 8, background: '#f8f9fa', borderLeft: '3px solid #1890ff', borderRadius: 4 }}>
              <strong style={{ color: '#1890ff' }}>提起行政诉讼：</strong>6个月（自知道或应当知道作出行政行为之日起计算）
            </div>
            <div style={{ padding: 10, marginBottom: 8, background: '#f8f9fa', borderLeft: '3px solid #1890ff', borderRadius: 4 }}>
              <strong style={{ color: '#1890ff' }}>上诉期：</strong>15日（自判决书送达之日起计算）
            </div>
          </Space>
        </div>
      </Card>
    </div>
  );
};

export default DeadlineCalculator;