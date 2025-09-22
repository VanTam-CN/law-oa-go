import React, { useState } from 'react';
import { Card, Button, Space, Typography, Alert, Divider } from 'antd';
import { PlusOutlined, FileTextOutlined } from '@ant-design/icons';
import CreateCaseWizard from '@/components/CreateCaseWizard';

const { Title, Paragraph } = Typography;

const CaseWizardDemo: React.FC = () => {
  const [wizardVisible, setWizardVisible] = useState(false);

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <Title level={2}>
          <FileTextOutlined style={{ marginRight: 8 }} />
          分步骤案件创建演示
        </Title>
        
        <Alert
          message="功能特色"
          description={
            <div>
              <p>• <strong>分步骤引导</strong>: 将复杂的案件创建过程分解为5个清晰的步骤</p>
              <p>• <strong>智能利益冲突检查</strong>: 自动分析潜在的利益冲突并提供详细报告</p>
              <p>• <strong>实时进度显示</strong>: 清晰的进度指示器和状态反馈</p>
              <p>• <strong>风险评估分析</strong>: 多维度风险评分和处理建议</p>
              <p>• <strong>数据验证保护</strong>: 每步数据验证，确保信息完整性</p>
            </div>
          }
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Divider />

        <Title level={3}>创建步骤说明</Title>
        
        <div style={{ marginBottom: 24 }}>
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <Card size="small" title="第一步：基本信息">
              <Paragraph>
                填写案件的基本信息，包括案件名称、编号、类型、项目类型、时间安排和案件描述等。
              </Paragraph>
            </Card>
            
            <Card size="small" title="第二步：当事人信息">
              <Paragraph>
                选择委托人，输入合同金额，填写对方当事人信息和案由等关键信息。
              </Paragraph>
            </Card>
            
            <Card size="small" title="第三步：团队分配">
              <Paragraph>
                分配主办律师和协办律师，设置团队成员，选择收费方式，进行风险评估标记。
              </Paragraph>
            </Card>
            
            <Card size="small" title="第四步：利益冲突检查" type="inner">
              <Paragraph>
                <strong>智能冲突分析系统</strong>会自动检查以下方面：
              </Paragraph>
              <ul>
                <li><strong>客户冲突</strong>: 检查历史案件中的对立关系</li>
                <li><strong>律师冲突</strong>: 分析律师工作负载和历史代理情况</li>
                <li><strong>案件冲突</strong>: 识别相似或关联案件</li>
                <li><strong>对方冲突</strong>: 检查对方是否曾是本所客户</li>
              </ul>
              <Alert
                message="风险评分机制"
                description="系统会给出0-100分的风险评分，并根据冲突级别提供相应的处理建议：通过(绿色)、需要注意(橙色)、存在冲突(红色)"
                type="warning"
                showIcon
              />
            </Card>
            
            <Card size="small" title="第五步：确认创建">
              <Paragraph>
                最终确认所有信息，查看冲突检查结果，确认无误后创建案件。
              </Paragraph>
            </Card>
          </Space>
        </div>

        <Divider />

        <div style={{ textAlign: 'center', padding: '40px 0' }}>
          <Button
            type="primary"
            size="large"
            icon={<PlusOutlined />}
            onClick={() => setWizardVisible(true)}
            style={{ height: '50px', fontSize: '16px', paddingLeft: '32px', paddingRight: '32px' }}
          >
            体验分步骤创建案件
          </Button>
        </div>

        <CreateCaseWizard
          visible={wizardVisible}
          onCancel={() => setWizardVisible(false)}
          onSuccess={() => {
            setWizardVisible(false);
            // 这里可以添加成功后的处理逻辑
          }}
        />
      </Card>
    </div>
  );
};

export default CaseWizardDemo;