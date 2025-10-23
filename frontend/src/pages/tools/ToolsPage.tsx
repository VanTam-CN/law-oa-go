import React from 'react';
import { useNavigate } from 'react-router';
import { Card, Row, Col, Button, Divider, message } from 'antd';
import { 
  CalculatorOutlined, 
  FileTextOutlined, 
  SearchOutlined, 
  TranslationOutlined,
  CalendarOutlined,
  DatabaseOutlined,
  PrinterOutlined,
  CloudDownloadOutlined
} from '@ant-design/icons';

const ToolsPage: React.FC = () => {
  const navigate = useNavigate();
  const toolCategories = [
    {
      title: '计算工具',
      tools: [
        {
          name: '诉讼费计算器',
          description: '根据案件标的额计算诉讼费用',
          icon: <CalculatorOutlined />,
          action: () => navigate('/tools/litigation-fee')
        },
        {
          name: '利息计算器',
          description: '计算借款利息和违约金',
          icon: <CalculatorOutlined />,
          action: () => navigate('/tools/interest-calculator')
        },
        {
          name: '工期计算器',
          description: '计算法定期限和工作日',
          icon: <CalendarOutlined />,
          action: () => navigate('/tools/deadline-calculator')
        }
      ]
    },
    {
      title: '文档工具',
      tools: [
        {
          name: '合同模板库',
          description: '常用法律文书和合同模板',
          icon: <FileTextOutlined />,
          action: () => message.info('合同模板库功能开发中...')
        },
        {
          name: '文档生成器',
          description: '自动生成标准法律文书',
          icon: <FileTextOutlined />,
          action: () => message.info('文档生成器功能开发中...')
        },
        {
          name: '文档转换器',
          description: '文档格式转换工具',
          icon: <CloudDownloadOutlined />,
          action: () => message.info('文档转换器功能开发中...')
        }
      ]
    },
    {
      title: '查询工具',
      tools: [
        {
          name: '法条查询',
          description: '快速查询相关法律条文',
          icon: <SearchOutlined />,
          action: () => navigate('/tools/law-search')
        },
        {
          name: '案例检索',
          description: '搜索相关判决案例',
          icon: <DatabaseOutlined />,
          action: () => message.info('案例检索功能开发中...')
        },
        {
          name: '企业信息查询',
          description: '查询企业工商信息',
          icon: <SearchOutlined />,
          action: () => message.info('企业信息查询功能开发中...')
        }
      ]
    },
    {
      title: '其他工具',
      tools: [
        {
          name: '翻译工具',
          description: '法律文件翻译助手',
          icon: <TranslationOutlined />,
          action: () => message.info('翻译工具功能开发中...')
        },
        {
          name: '打印助手',
          description: '批量打印和格式化工具',
          icon: <PrinterOutlined />,
          action: () => message.info('打印助手功能开发中...')
        },
        {
          name: '数据导出',
          description: '案件数据导出和备份',
          icon: <CloudDownloadOutlined />,
          action: () => message.info('数据导出功能开发中...')
        }
      ]
    }
  ];

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">工具箱</h1>
        <p>律所日常工作中的实用工具集合</p>
      </div>

      {toolCategories.map((category, categoryIndex) => (
        <div key={categoryIndex} style={{ marginBottom: 32 }}>
          <Divider orientation="left" style={{ fontSize: 16, fontWeight: 600 }}>
            {category.title}
          </Divider>
          
          <Row gutter={[16, 16]}>
            {category.tools.map((tool, toolIndex) => (
              <Col xs={24} sm={12} md={8} lg={6} key={toolIndex}>
                <Card
                  hoverable
                  style={{ height: '100%' }}
                  bodyStyle={{ padding: 20, textAlign: 'center' }}
                >
                  <div style={{ fontSize: 32, color: '#1c4e80', marginBottom: 12 }}>
                    {tool.icon}
                  </div>
                  <h3 style={{ marginBottom: 8, fontSize: 16 }}>{tool.name}</h3>
                  <p style={{ 
                    color: '#666', 
                    fontSize: 12, 
                    marginBottom: 16,
                    minHeight: 36,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center'
                  }}>
                    {tool.description}
                  </p>
                  <Button 
                    type="primary" 
                    size="small"
                    onClick={tool.action}
                    style={{ width: '100%' }}
                  >
                    使用工具
                  </Button>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      ))}

      <Card style={{ marginTop: 32, textAlign: 'center', background: '#f8f9fa' }}>
        <div style={{ padding: '20px 0' }}>
          <h3 style={{ color: '#1c4e80', marginBottom: 8 }}>更多工具正在开发中</h3>
          <p style={{ color: '#666', marginBottom: 0 }}>
            我们将持续添加更多实用的法律工作工具，提升您的工作效率
          </p>
        </div>
      </Card>
    </div>
  );
};

export default ToolsPage;