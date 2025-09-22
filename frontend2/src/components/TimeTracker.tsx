import React, { useState, useEffect, useRef } from 'react';
import { 
  Card, 
  Button, 
  Table, 
  Tag, 
  Select, 
  DatePicker, 
  Input, 
  Form, 
  Modal, 
  message, 
  Statistic, 
  Row, 
  Col,
  Tooltip,
  Popconfirm
} from 'antd';
import { 
  PlayCircleOutlined, 
  PauseCircleOutlined, 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined,
  ClockCircleOutlined,
  DollarCircleOutlined
} from '@ant-design/icons';
import moment from 'moment';

const { Option } = Select;
const { TextArea } = Input;
const { RangePicker } = DatePicker;

interface TimeEntry {
  id: number;
  caseId: number;
  caseName: string;
  workType: string;
  description: string;
  startTime: string;
  endTime: string;
  duration: number;
  isBillable: boolean;
  amount: number;
  status: 'DRAFT' | 'SUBMITTED' | 'APPROVED' | 'BILLED';
  createTime: string;
}

interface CaseInfo {
  id: number;
  caseName: string;
  caseNo: string;
}

const TimeTracker: React.FC = () => {
  const [form] = Form.useForm();
  const [timeEntries, setTimeEntries] = useState<TimeEntry[]>([]);
  const [cases, setCases] = useState<CaseInfo[]>([]);
  const [isTracking, setIsTracking] = useState(false);
  const [currentTimer, setCurrentTimer] = useState<{
    caseId: number;
    startTime: Date;
    description: string;
  } | null>(null);
  const [elapsedTime, setElapsedTime] = useState(0);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingEntry, setEditingEntry] = useState<TimeEntry | null>(null);
  const [loading, setLoading] = useState(false);
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  // 工作类型选项
  const workTypes = [
    '法律咨询', '文书起草', '庭审准备', '案件研究', 
    '客户沟通', '文件审查', '谈判调解', '其他'
  ];

  // 状态标签配置
  const statusConfig = {
    DRAFT: { color: 'default', text: '草稿' },
    SUBMITTED: { color: 'processing', text: '已提交' },
    APPROVED: { color: 'success', text: '已批准' },
    BILLED: { color: 'purple', text: '已计费' }
  };

  // 模拟数据
  useEffect(() => {
    // 模拟API调用获取案件列表
    const mockCases: CaseInfo[] = [
      { id: 1, caseName: '张三诉李四合同纠纷案', caseNo: 'CASE-2024-001' },
      { id: 2, caseName: '某公司劳动争议案', caseNo: 'CASE-2024-002' },
      { id: 3, caseName: '知识产权侵权案', caseNo: 'CASE-2024-003' }
    ];
    setCases(mockCases);

    // 模拟时间记录数据
    const mockEntries: TimeEntry[] = [
      {
        id: 1,
        caseId: 1,
        caseName: '张三诉李四合同纠纷案',
        workType: '法律咨询',
        description: '与客户电话沟通案情',
        startTime: '2024-01-15 09:00:00',
        endTime: '2024-01-15 09:30:00',
        duration: 30,
        isBillable: true,
        amount: 150.00,
        status: 'APPROVED',
        createTime: '2024-01-15 09:35:00'
      },
      {
        id: 2,
        caseId: 2,
        caseName: '某公司劳动争议案',
        workType: '文书起草',
        description: '起草起诉状',
        startTime: '2024-01-15 10:00:00',
        endTime: '2024-01-15 12:00:00',
        duration: 120,
        isBillable: true,
        amount: 600.00,
        status: 'SUBMITTED',
        createTime: '2024-01-15 12:05:00'
      }
    ];
    setTimeEntries(mockEntries);
  }, []);

  // 计时器逻辑
  useEffect(() => {
    if (isTracking && currentTimer) {
      timerRef.current = setInterval(() => {
        setElapsedTime(prev => prev + 1);
      }, 1000);
    } else {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    }

    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, [isTracking, currentTimer]);

  // 开始计时
  const startTimer = () => {
    if (!form.getFieldValue('caseId')) {
      message.warning('请选择案件');
      return;
    }
    if (!form.getFieldValue('description')) {
      message.warning('请输入工作描述');
      return;
    }

    setIsTracking(true);
    setCurrentTimer({
      caseId: form.getFieldValue('caseId'),
      startTime: new Date(),
      description: form.getFieldValue('description')
    });
    setElapsedTime(0);
    message.success('计时已开始');
  };

  // 停止计时
  const stopTimer = () => {
    if (!currentTimer) return;

    setIsTracking(false);
    
    // 自动填充表单
    form.setFieldsValue({
      caseId: currentTimer.caseId,
      description: currentTimer.description,
      startTime: moment(currentTimer.startTime),
      endTime: moment(new Date()),
      duration: Math.floor(elapsedTime / 60)
    });

    setModalVisible(true);
    setCurrentTimer(null);
    setElapsedTime(0);
  };

  // 格式化时间显示
  const formatTime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  // 保存时间记录
  const saveTimeEntry = async (values: any) => {
    setLoading(true);
    try {
      // 模拟API调用
      const newEntry: TimeEntry = {
        id: Date.now(),
        caseId: values.caseId,
        caseName: cases.find(c => c.id === values.caseId)?.caseName || '',
        workType: values.workType,
        description: values.description,
        startTime: values.startTime.format('YYYY-MM-DD HH:mm:ss'),
        endTime: values.endTime.format('YYYY-MM-DD HH:mm:ss'),
        duration: values.duration,
        isBillable: values.isBillable,
        amount: values.amount || 0,
        status: 'DRAFT',
        createTime: moment().format('YYYY-MM-DD HH:mm:ss')
      };

      if (editingEntry) {
        // 更新记录
        setTimeEntries(prev => 
          prev.map(entry => entry.id === editingEntry.id ? newEntry : entry)
        );
        message.success('更新成功');
      } else {
        // 新增记录
        setTimeEntries(prev => [newEntry, ...prev]);
        message.success('保存成功');
      }

      setModalVisible(false);
      form.resetFields();
      setEditingEntry(null);
    } catch (error) {
      message.error('保存失败');
    }
    setLoading(false);
  };

  // 编辑时间记录
  const editEntry = (entry: TimeEntry) => {
    setEditingEntry(entry);
    form.setFieldsValue({
      ...entry,
      startTime: moment(entry.startTime),
      endTime: moment(entry.endTime)
    });
    setModalVisible(true);
  };

  // 删除时间记录
  const deleteEntry = (id: number) => {
    setTimeEntries(prev => prev.filter(entry => entry.id !== id));
    message.success('删除成功');
  };

  // 提交时间记录
  const submitEntry = (id: number) => {
    setTimeEntries(prev => 
      prev.map(entry => 
        entry.id === id ? { ...entry, status: 'SUBMITTED' } : entry
      )
    );
    message.success('提交成功');
  };

  // 计算统计数据
  const statistics = {
    totalHours: timeEntries.reduce((sum, entry) => sum + entry.duration, 0),
    billableHours: timeEntries.filter(entry => entry.isBillable).reduce((sum, entry) => sum + entry.duration, 0),
    totalAmount: timeEntries.reduce((sum, entry) => sum + entry.amount, 0),
    todayHours: timeEntries
      .filter(entry => moment(entry.createTime).isSame(moment(), 'day'))
      .reduce((sum, entry) => sum + entry.duration, 0)
  };

  const columns = [
    {
      title: '案件名称',
      dataIndex: 'caseName',
      key: 'caseName',
      width: 200
    },
    {
      title: '工作类型',
      dataIndex: 'workType',
      key: 'workType',
      width: 120
    },
    {
      title: '工作描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      key: 'startTime',
      width: 150,
      render: (text: string) => moment(text).format('MM-DD HH:mm')
    },
    {
      title: '时长',
      dataIndex: 'duration',
      key: 'duration',
      width: 100,
      render: (minutes: number) => `${minutes}分钟`
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 100,
      render: (amount: number) => `¥${amount.toFixed(2)}`
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const config = statusConfig[status as keyof typeof statusConfig];
        return <Tag color={config.color}>{config.text}</Tag>;
      }
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (text: any, record: TimeEntry) => (
        <div style={{ display: 'flex', gap: '8px' }}>
          {record.status === 'DRAFT' && (
            <>
              <Button 
                type="link" 
                size="small" 
                icon={<EditOutlined />}
                onClick={() => editEntry(record)}
              >
                编辑
              </Button>
              <Button 
                type="link" 
                size="small" 
                onClick={() => submitEntry(record.id)}
              >
                提交
              </Button>
              <Popconfirm
                title="确定删除吗？"
                onConfirm={() => deleteEntry(record.id)}
              >
                <Button 
                  type="link" 
                  size="small" 
                  danger
                  icon={<DeleteOutlined />}
                >
                  删除
                </Button>
              </Popconfirm>
            </>
          )}
        </div>
      )
    }
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        {/* 统计信息 */}
        <Row gutter={16} style={{ marginBottom: '24px' }}>
          <Col span={6}>
            <Statistic
              title="今日工作时长"
              value={statistics.todayHours}
              suffix="分钟"
              prefix={<ClockCircleOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="总工作时长"
              value={statistics.totalHours}
              suffix="分钟"
              prefix={<ClockCircleOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="可计费时长"
              value={statistics.billableHours}
              suffix="分钟"
              prefix={<ClockCircleOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="总金额"
              value={statistics.totalAmount}
              suffix="元"
              prefix={<DollarCircleOutlined />}
            />
          </Col>
        </Row>

        {/* 计时器 */}
        <Card title="快速计时" style={{ marginBottom: '24px' }}>
          <Form form={form} layout="inline">
            <Form.Item name="caseId" label="案件" rules={[{ required: true, message: '请选择案件' }]}>
              <Select style={{ width: 300 }} placeholder="选择案件">
                {cases.map(caseInfo => (
                  <Option key={caseInfo.id} value={caseInfo.id}>
                    {caseInfo.caseName}
                  </Option>
                ))}
              </Select>
            </Form.Item>
            <Form.Item name="description" label="工作描述" rules={[{ required: true, message: '请输入工作描述' }]}>
              <Input style={{ width: 300 }} placeholder="输入工作描述" />
            </Form.Item>
            <Form.Item>
              {isTracking ? (
                <Button 
                  type="primary" 
                  danger 
                  icon={<PauseCircleOutlined />}
                  onClick={stopTimer}
                >
                  停止计时 ({formatTime(elapsedTime)})
                </Button>
              ) : (
                <Button 
                  type="primary" 
                  icon={<PlayCircleOutlined />}
                  onClick={startTimer}
                >
                  开始计时
                </Button>
              )}
            </Form.Item>
            <Form.Item>
              <Button 
                type="default" 
                icon={<PlusOutlined />}
                onClick={() => setModalVisible(true)}
              >
                手动添加
              </Button>
            </Form.Item>
          </Form>
        </Card>

        {/* 时间记录列表 */}
        <Card title="时间记录">
          <Table
            columns={columns}
            dataSource={timeEntries}
            rowKey="id"
            pagination={{
              total: timeEntries.length,
              pageSize: 10,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`
            }}
          />
        </Card>
      </Card>

      {/* 添加/编辑时间记录弹窗 */}
      <Modal
        title={editingEntry ? '编辑时间记录' : '添加时间记录'}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
          setEditingEntry(null);
        }}
        onOk={() => form.submit()}
        confirmLoading={loading}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={saveTimeEntry}>
          <Form.Item 
            name="caseId" 
            label="案件" 
            rules={[{ required: true, message: '请选择案件' }]}
          >
            <Select placeholder="选择案件">
              {cases.map(caseInfo => (
                <Option key={caseInfo.id} value={caseInfo.id}>
                  {caseInfo.caseName}
                </Option>
              ))}
            </Select>
          </Form.Item>
          
          <Form.Item 
            name="workType" 
            label="工作类型" 
            rules={[{ required: true, message: '请选择工作类型' }]}
          >
            <Select placeholder="选择工作类型">
              {workTypes.map(type => (
                <Option key={type} value={type}>{type}</Option>
              ))}
            </Select>
          </Form.Item>
          
          <Form.Item 
            name="description" 
            label="工作描述" 
            rules={[{ required: true, message: '请输入工作描述' }]}
          >
            <TextArea rows={3} placeholder="输入工作描述" />
          </Form.Item>
          
          <Form.Item 
            name="startTime" 
            label="开始时间" 
            rules={[{ required: true, message: '请选择开始时间' }]}
          >
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          
          <Form.Item 
            name="endTime" 
            label="结束时间" 
            rules={[{ required: true, message: '请选择结束时间' }]}
          >
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          
          <Form.Item 
            name="duration" 
            label="工作时长(分钟)" 
            rules={[{ required: true, message: '请输入工作时长' }]}
          >
            <Input type="number" placeholder="输入工作时长(分钟)" />
          </Form.Item>
          
          <Form.Item name="isBillable" label="是否可计费" valuePropName="checked">
            <Select>
              <Option value={true}>是</Option>
              <Option value={false}>否</Option>
            </Select>
          </Form.Item>
          
          <Form.Item name="amount" label="金额">
            <Input type="number" placeholder="输入金额" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TimeTracker;