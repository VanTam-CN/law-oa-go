import React from 'react';
import { useNavigate } from 'react-router-dom';
import {
  FaCalculator as CalculatorIcon,
  FaFileLines as DocumentTextIcon,
  FaMagnifyingGlass as MagnifyingGlassIcon,
  FaLanguage as LanguageIcon,
  FaCalendar as CalendarIcon,
  FaServer as ServerIcon,
  FaPrint as PrinterIcon,
  FaCloudArrowDown as CloudArrowDownIcon,
  FaGear as CogIcon,
  FaLightbulb as LightBulbIcon,
  FaChartBar as ChartBarIcon,
  FaUsers as UserGroupIcon
} from 'react-icons/fa6';

interface ToolItem {
  name: string;
  description: string;
  icon: React.ReactNode;
  action: () => void;
  status: 'available' | 'development' | 'planned';
}

interface ToolCategory {
  title: string;
  tools: ToolItem[];
}

const ToolsPage: React.FC = () => {
  const navigate = useNavigate();

  const toolCategories: ToolCategory[] = [
    {
      title: '计算工具',
      tools: [
        {
          name: '诉讼费计算器',
          description: '根据案件标的额计算诉讼费用',
          icon: <CalculatorIcon className="w-8 h-8" />,
          action: () => navigate('/tools/litigation-fee'),
          status: 'available'
        },
        {
          name: '利息计算器',
          description: '计算借款利息和违约金',
          icon: <CalculatorIcon className="w-8 h-8" />,
          action: () => navigate('/tools/interest-calculator'),
          status: 'available'
        },
        {
          name: '工期计算器',
          description: '计算法定期限和工作日',
          icon: <CalendarIcon className="w-8 h-8" />,
          action: () => navigate('/tools/deadline-calculator'),
          status: 'available'
        }
      ]
    },
    {
      title: '文档工具',
      tools: [
        {
          name: '合同模板库',
          description: '常用法律文书和合同模板',
          icon: <DocumentTextIcon className="w-8 h-8" />,
          action: () => alert('合同模板库功能开发中...'),
          status: 'development'
        },
        {
          name: '文档生成器',
          description: '自动生成标准法律文书',
          icon: <DocumentTextIcon className="w-8 h-8" />,
          action: () => alert('文档生成器功能开发中...'),
          status: 'development'
        },
        {
          name: '文档转换器',
          description: '文档格式转换工具',
          icon: <CloudArrowDownIcon className="w-8 h-8" />,
          action: () => alert('文档转换器功能开发中...'),
          status: 'planned'
        }
      ]
    },
    {
      title: '查询工具',
      tools: [
        {
          name: '法条查询',
          description: '快速查询相关法律条文',
          icon: <MagnifyingGlassIcon className="w-8 h-8" />,
          action: () => navigate('/tools/law-search'),
          status: 'available'
        },
        {
          name: '案例检索',
          description: '搜索相关判决案例',
          icon: <ServerIcon className="w-8 h-8" />,
          action: () => alert('案例检索功能开发中...'),
          status: 'development'
        },
        {
          name: '企业信息查询',
          description: '查询企业工商信息',
          icon: <MagnifyingGlassIcon className="w-8 h-8" />,
          action: () => alert('企业信息查询功能开发中...'),
          status: 'planned'
        }
      ]
    },
    {
      title: '效率工具',
      tools: [
        {
          name: '翻译工具',
          description: '法律文件翻译助手',
          icon: <LanguageIcon className="w-8 h-8" />,
          action: () => alert('翻译工具功能开发中...'),
          status: 'planned'
        },
        {
          name: '打印助手',
          description: '批量打印和格式化工具',
          icon: <PrinterIcon className="w-8 h-8" />,
          action: () => alert('打印助手功能开发中...'),
          status: 'planned'
        },
        {
          name: '数据导出',
          description: '案件数据导出和备份',
          icon: <CloudArrowDownIcon className="w-8 h-8" />,
          action: () => alert('数据导出功能开发中...'),
          status: 'development'
        }
      ]
    },
    {
      title: '分析工具',
      tools: [
        {
          name: '案件分析',
          description: '案件胜率分析和风险评估',
          icon: <ChartBarIcon className="w-8 h-8" />,
          action: () => navigate('/tools/case-analysis'),
          status: 'available'
        },
        {
          name: '客户画像',
          description: '客户行为分析和价值评估',
          icon: <UserGroupIcon className="w-8 h-8" />,
          action: () => alert('客户画像功能开发中...'),
          status: 'development'
        },
        {
          name: '智能建议',
          description: 'AI驱动的法律建议系统',
          icon: <LightBulbIcon className="w-8 h-8" />,
          action: () => alert('智能建议功能开发中...'),
          status: 'planned'
        }
      ]
    },
    {
      title: '系统工具',
      tools: [
        {
          name: '设置管理',
          description: '系统配置和偏好设置',
          icon: <CogIcon className="w-8 h-8" />,
          action: () => navigate('/settings'),
          status: 'available'
        },
        {
          name: '数据备份',
          description: '系统数据备份和恢复',
          icon: <CloudArrowDownIcon className="w-8 h-8" />,
          action: () => alert('数据备份功能开发中...'),
          status: 'development'
        },
        {
          name: '日志查看',
          description: '系统操作日志和审计',
          icon: <DocumentTextIcon className="w-8 h-8" />,
          action: () => alert('日志查看功能开发中...'),
          status: 'planned'
        }
      ]
    }
  ];

  const getStatusBadge = (status: ToolItem['status']) => {
    switch (status) {
      case 'available':
        return <span className="badge bg-success">可用</span>;
      case 'development':
        return <span className="badge bg-warning">开发中</span>;
      case 'planned':
        return <span className="badge bg-secondary">计划中</span>;
      default:
        return <span className="badge bg-light">未知</span>;
    }
  };

  const getStatusButtonVariant = (status: ToolItem['status']) => {
    switch (status) {
      case 'available':
        return 'primary';
      case 'development':
        return 'warning';
      case 'planned':
        return 'secondary';
      default:
        return 'outline-secondary';
    }
  };

  const getButtonText = (status: ToolItem['status']) => {
    switch (status) {
      case 'available':
        return '使用工具';
      case 'development':
        return '开发中';
      case 'planned':
        return '敬请期待';
      default:
        return '查看详情';
    }
  };

  return (
    <div className="tools-page p-4">
      <Card className="mb-4">
        <Card.Body>
          <div className="text-center mb-4">
            <h1 className="display-5 mb-3">
              <CogIcon className="w-12 h-12 text-primary me-3" />
              专业工具箱
            </h1>
            <p className="lead text-muted">
              律所日常工作中的实用工具集合，提升您的工作效率
            </p>
          </div>

          <Alert variant="info" className="mb-4">
            <LightBulbIcon className="w-5 h-5 me-2" />
            <strong>提示：</strong>
            点击工具卡片可以使用相应功能，标记为"开发中"的功能正在积极开发中。
          </Alert>
        </Card.Body>
      </Card>

      {toolCategories.map((category, categoryIndex) => (
        <Card key={categoryIndex} className="mb-4">
          <Card.Header>
            <h3 className="mb-0">
              {category.title}
            </h3>
          </Card.Header>
          <Card.Body>
            <div className="row g-4">
              {category.tools.map((tool, toolIndex) => (
                <div key={toolIndex} className="col-12 col-sm-6 col-md-4 col-lg-3">
                  <Card className="h-100 tool-card">
                    <Card.Body className="text-center">
                      <div className="mb-3">
                        <div className="text-primary mb-2">
                          {tool.icon}
                        </div>
                        <div className="mb-2">
                          {getStatusBadge(tool.status)}
                        </div>
                      </div>

                      <Card.Title className="h6 mb-2">
                        {tool.name}
                      </Card.Title>

                      <Card.Text className="text-muted small mb-3">
                        {tool.description}
                      </Card.Text>

                      <Button
                        variant={getStatusButtonVariant(tool.status)}
                        size="sm"
                        onClick={tool.action}
                        disabled={tool.status === 'planned'}
                        className="w-100"
                      >
                        {getButtonText(tool.status)}
                      </Button>
                    </Card.Body>
                  </Card>
                </div>
              ))}
            </div>
          </Card.Body>
        </Card>
      ))}

      <Card className="text-center bg-light">
        <Card.Body>
          <h3 className="text-primary mb-3">
            <LightBulbIcon className="w-6 h-6 me-2" />
            更多工具正在开发中
          </h3>
          <p className="text-muted mb-0">
            我们将持续添加更多实用的法律工作工具，提升您的工作效率
          </p>
        </Card.Body>
      </Card>

      <style>
        {`
        .tool-card {
          transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        .tool-card:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }
      `}
      </style>
    </div>
  );
};

export default ToolsPage;