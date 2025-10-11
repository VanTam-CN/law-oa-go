import React, { useState, useEffect } from 'react';
import { Card, Input, Select, Table, Tag, Space, Button } from 'antd';
import { SearchOutlined, FileTextOutlined } from '@ant-design/icons';
import { getLaws, LawItem } from '@/services/tools';
import './LawSearch.less';

const { Search } = Input;
const { Option } = Select;

const LawSearch: React.FC = () => {
  const [loading, setLoading] = useState<boolean>(false);
  const [laws, setLaws] = useState<LawItem[]>([]);
  const [filteredLaws, setFilteredLaws] = useState<LawItem[]>([]);
  const [searchText, setSearchText] = useState<string>('');
  const [selectedCategory, setSelectedCategory] = useState<string>('');

  useEffect(() => {
    fetchLaws();
  }, []);

  useEffect(() => {
    filterLaws();
  }, [laws, searchText, selectedCategory]);

  const fetchLaws = async () => {
    try {
      setLoading(true);
      const response = await getLaws();
      // 根据API测试结果，数据结构是 response.rows
      setLaws(response.rows || []);
    } catch (error) {
      console.error('Failed to fetch laws:', error);
      setLaws([]);
    } finally {
      setLoading(false);
    }
  };

  const filterLaws = () => {
    let filtered = laws;
    
    if (searchText) {
      filtered = filtered.filter(law => 
        law.title.toLowerCase().includes(searchText.toLowerCase()) ||
        law.content.toLowerCase().includes(searchText.toLowerCase())
      );
    }
    
    if (selectedCategory) {
      filtered = filtered.filter(law => law.category === selectedCategory);
    }
    
    setFilteredLaws(filtered);
  };

  const categories = Array.from(new Set(laws.map(law => law.category)));

  const columns = [
    {
      title: '法规名称',
      dataIndex: 'title',
      key: 'title',
      render: (text: string) => (
        <span style={{ fontWeight: 500 }}>{text}</span>
      ),
    },
    {
      title: '类别',
      dataIndex: 'category',
      key: 'category',
      render: (category: string) => (
        <Tag color={getCategoryColor(category)}>{category}</Tag>
      ),
    },
    {
      title: '生效日期',
      dataIndex: 'effectiveDate',
      key: 'effectiveDate',
      render: (date: string) => date,
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record: LawItem) => (
        <Space size="middle">
          <Button type="link" size="small" icon={<FileTextOutlined />}>
            查看全文
          </Button>
        </Space>
      ),
    },
  ];

  const getCategoryColor = (category: string) => {
    const colors: { [key: string]: string } = {
      '民事': 'blue',
      '刑事': 'red',
      '行政': 'orange',
      '经济': 'green',
      '劳动': 'purple',
    };
    return colors[category] || 'default';
  };

  return (
    <div className="law-search">
      <Card title="法条查询" className="search-card">
        <div className="search-filters">
          <Search
            placeholder="搜索法规名称或内容关键词"
            allowClear
            enterButton={<SearchOutlined />}
            size="large"
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ marginBottom: 16 }}
          />
          
          <Select
            placeholder="选择法规类别"
            allowClear
            style={{ width: 200, marginBottom: 16 }}
            value={selectedCategory}
            onChange={setSelectedCategory}
          >
            {categories.map(category => (
              <Option key={category} value={category}>
                {category}
              </Option>
            ))}
          </Select>
        </div>

        <Table
          columns={columns}
          dataSource={filteredLaws}
          rowKey="id"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          }}
          expandable={{
            expandedRowRender: (record) => (
              <div className="law-content">
                <p>{record.content}</p>
              </div>
            ),
            rowExpandable: (record) => true,
          }}
        />
      </Card>
    </div>
  );
};

export default LawSearch;