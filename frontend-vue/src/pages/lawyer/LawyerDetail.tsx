import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Descriptions, 
  Tag, 
  Button, 
  Space, 
  Timeline, 
  Divider,
  List,
  Avatar,
  message,
  Spin,
  Tabs,
  Badge
} from 'antd';
import { 
  ArrowLeftOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  FileTextOutlined,
  UserOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  DollarCircleOutlined,
  TeamOutlined,
  PhoneOutlined,
  MailOutlined,
  EnvironmentOutlined,
  CommentOutlined,
  PaperClipOutlined,
  BankOutlined as GavelOutlined,
  MedicineBoxOutlined,
  TrophyOutlined,
  PauseCircleOutlined,
  UserSwitchOutlined,
  EyeOutlined
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { lawyerService, Lawyer } from '@/services/lawyer';
import dayjs from 'dayjs';


interface LawyerCase {
  id: number;
  caseNo: string;
  caseName: string;
  caseType: string;
  clientName: string;
  status: string;
  amount: number;
  createTime: string;
}

interface LawyerDocument {
  id: number;
  name: string;
  type: string;
  size: string;
  uploadTime: string;
  uploader: string;
}

interface LawyerTimeline {
  id: number;
  time: string;
  event: string;
  description: string;
}

const LawyerDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = useState(true);
  const [lawyerDetail, setLawyerDetail] = useState<Lawyer | null>(null);
  const [cases, setCases] = useState<LawyerCase[]>([]);
  const [documents, setDocuments] = useState<LawyerDocument[]>([]);
  const [timeline, setTimeline] = useState<LawyerTimeline[]>([]);

  // 模拟数据
  const mockCases: LawyerCase[] = [
    {
      id: 1,
      caseNo: '2025001',
      caseName: '张三与李四借款纠纷',
      caseType: 'CIVIL',
      clientName: '张三',
      status: '1',
      amount: 550000,
      createTime: '2025-01-15 10:30:00'
    },
    {
      id: 2,
      caseNo: '2025002',
      caseName: 'ABC公司合同纠纷',
      caseType: 'COMMERCIAL',
      clientName: 'ABC公司',
      status: '1',
      amount: 2000000,
      createTime: '2025-01-18 09:15:00'
    }
  ];

  const mockDocuments: LawyerDocument[] = [
    {
      id: 1,
      name: '律师执业证.pdf',
      type: 'PDF',
      size: '1.2MB',
      uploadTime: '2024-01-15 10:35:00',
      uploader: '管理员'
    },
    {
      id: 2,
      name: '学历证书.pdf',
      type: 'PDF',
      size: '800KB',
      uploadTime: '2024-01-15 10:40:00',
      uploader: '管理员'
    }
  ];

  const mockTimeline: LawyerTimeline[] = [
    {
      id: 1,
      time: '2024-01-15 09:00:00',
      event: '入职',
      description: '正式加入律所，成为诉讼部律师'
    },
    {
      id: 2,
      time: '2024-06-20 14:30:00',
      event: '晋升',
      description: '晋升为资深律师'
    },
    {
      id: 3,
      time: '2025-01-10 16:45:00',
      event: '培训',
      description: '参加律师专业技能培训'
    }
  ];

  useEffect(() => {
    if (id) {
      fetchLawyerDetail();
    }
  }, [id]);

  const fetchLawyerDetail = async () => {
    setLoading(true);
    try {
      // 这里应该调用API获取数据
      // const response = await lawyerService.getLawyerDetail(Number(id));
      
      // 模拟API调用
      setTimeout(() => {
        const mockLawyer: Lawyer = {
          id: Number(id),
          name: '张律师',
          gender: 'male',
          phone: '13800138001',
          email: 'zhang@lawfirm.com',
          licenseNumber: '123456789012345',
          specialty: ['民事案件', '商事案件'],
          experience: 8,
          status: 'active',
          department: '诉讼部',
          position: '资深律师',
          joinDate: '2024-01-15',
          profile: '资深律师，专注于民事诉讼和商事诉讼，拥有丰富的实战经验。'
        };
        
        setLawyerDetail(mockLawyer);
        setCases(mockCases);
        setDocuments(mockDocuments);
        setTimeline(mockTimeline);
        setLoading(false);
      }, 1000);
    } catch (error) {
      message.error('获取律师详情失败');
      setLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const statusMap = {
      'active': { text: '在职', color: 'success' },
      'inactive': { text: '离职', color: 'default' },
      'on_leave': { text: '休假', color: 'warning' }
    };
    const config = statusMap[status as keyof typeof statusMap] || { text: '未知', color: 'default' };
    return <Badge status={config.color as any} text={config.text} />;
  };

  const getCaseTypeTag = (type: string) => {
    const typeMap = {
      'CIVIL': { text: '民事案件', color: 'blue' },
      'COMMERCIAL': { text: '商事案件', color: 'orange' },
      'CRIMINAL': { text: '刑事案件', color: 'red' },
      'ADMINISTRATIVE': { text: '行政案件', color: 'purple' }
    };
    const config = typeMap[type as keyof typeof typeMap] || { text: '其他', color: 'default' };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const getSpecialtyTags = (specialties: string[]) => {
    const specialtyMap: Record<string, { icon: React.ReactNode; color: string }> = {
      '民事案件': { icon: <GavelOutlined />, color: 'blue' },
      '刑事案件': { icon: <MedicineBoxOutlined />, color: 'red' },
      '商事案件': { icon: <TrophyOutlined />, color: 'orange' },
      '行政案件': { icon: <TrophyOutlined />, color: 'purple' },
      '知识产权': { icon: <TrophyOutlined />, color: 'cyan' },
      '劳动纠纷': { icon: <MedicineBoxOutlined />, color: 'green' }
    };

    return specialties.map(spec => {
      const config = specialtyMap[spec] || { icon: <GavelOutlined />, color: 'default' };
      return (
        <Tag key={spec} color={config.color} icon={config.icon}>
          {spec}
        </Tag>
      );
    });
  };

  if (loading) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Spin size="large" />
        <div style={{ marginTop: 16 }}>加载中...</div>
      </div>
    );
  }

  if (!lawyerDetail) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <div>律师不存在</div>
      </div>
    );
  }

  return (
    <div className="lawyer-detail">
      <Card>
        <div className="detail-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
          <Space>
            <Button 
              type="text" 
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/lawyer')}
            >
              返回
            </Button>
            <h2 style={{ margin: 0 }}>{lawyerDetail.name}</h2>
          </Space>
          <Button 
            type="primary" 
            icon={<EditOutlined />}
            onClick={() => {
              // TODO: 实现编辑功能
              message.info('编辑功能待实现');
            }}
          >
            编辑
          </Button>
        </div>

        <Tabs
          defaultActiveKey="basic"
          items={[
            {
              key: 'basic',
              label: '基本信息',
              children: (
                <Card title="律师信息" loading={loading}>
                  <Descriptions bordered column={2}>
                    <Descriptions.Item label="姓名">
                      <Space>
                        <UserOutlined />
                        {lawyerDetail.name}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="性别">
                      <Tag color={lawyerDetail.gender === 'male' ? 'blue' : 'pink'}>
                        {lawyerDetail.gender === 'male' ? '男' : '女'}
                      </Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label="联系电话">
                      <Space>
                        <PhoneOutlined />
                        {lawyerDetail.phone}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="邮箱地址">
                      <Space>
                        <MailOutlined />
                        {lawyerDetail.email}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="执业证号">
                      {lawyerDetail.licenseNumber}
                    </Descriptions.Item>
                    <Descriptions.Item label="工作经验">
                      <Space>
                        <ClockCircleOutlined />
                        {lawyerDetail.experience}年
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="部门">
                      <TeamOutlined />
                      {lawyerDetail.department}
                    </Descriptions.Item>
                    <Descriptions.Item label="职位">
                      {lawyerDetail.position}
                    </Descriptions.Item>
                    <Descriptions.Item label="入职日期" span={2}>
                      <Space>
                        <CalendarOutlined />
                        {lawyerDetail.joinDate}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="状态" span={2}>
                      {getStatusBadge(lawyerDetail.status)}
                    </Descriptions.Item>
                    <Descriptions.Item label="专业领域" span={2}>
                      <Space size="small" wrap>
                        {getSpecialtyTags(lawyerDetail.specialty)}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label="个人简介" span={2}>
                      {lawyerDetail.profile || '暂无个人简介'}
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              )
            },
            {
              key: 'cases',
              label: '负责案件',
              children: (
                <Card title="负责案件列表" loading={loading}>
                  <List
                    dataSource={cases}
                    renderItem={(item) => (
                      <List.Item
                        actions={[
                          <Button type="link" icon={<EyeOutlined />}>查看</Button>
                        ]}
                      >
                        <List.Item.Meta
                          avatar={<Avatar icon={<FileTextOutlined />} />}
                          title={
                            <Space>
                              {item.caseNo}
                              {getCaseTypeTag(item.caseType)}
                            </Space>
                          }
                          description={
                            <div>
                              <div>{item.caseName}</div>
                              <div style={{ color: '#666', fontSize: '12px' }}>
                                客户: {item.clientName} · 金额: ¥{item.amount.toLocaleString()}
                              </div>
                            </div>
                          }
                        />
                      </List.Item>
                    )}
                  />
                </Card>
              )
            },
            {
              key: 'documents',
              label: '相关文档',
              children: (
                <Card title="律师文档" loading={loading}>
                  <List
                    dataSource={documents}
                    renderItem={(item) => (
                      <List.Item
                        actions={[
                          <Button type="link" icon={<PaperClipOutlined />}>下载</Button>,
                          <Button type="link" icon={<EyeOutlined />}>查看</Button>
                        ]}
                      >
                        <List.Item.Meta
                          avatar={<Avatar icon={<PaperClipOutlined />} />}
                          title={item.name}
                          description={
                            <Space>
                              <Tag>{item.type}</Tag>
                              <span>{item.size}</span>
                              <span>{item.uploadTime}</span>
                              <span>上传者: {item.uploader}</span>
                            </Space>
                          }
                        />
                      </List.Item>
                    )}
                  />
                </Card>
              )
            },
            {
              key: 'timeline',
              label: '职业发展',
              children: (
                <Card title="职业发展时间线" loading={loading}>
                  <Timeline>
                    {timeline.map((item) => (
                      <Timeline.Item key={item.id}>
                        <div>
                          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>
                            {item.event}
                          </div>
                          <div style={{ color: '#666', marginBottom: 4 }}>
                            {item.description}
                          </div>
                          <div style={{ color: '#999', fontSize: '12px' }}>
                            {item.time}
                          </div>
                        </div>
                      </Timeline.Item>
                    ))}
                  </Timeline>
                </Card>
              )
            }
          ]}
        />
      </Card>
    </div>
  );
};

export default LawyerDetail;