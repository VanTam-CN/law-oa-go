import React, { useState } from 'react';
import { Button, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import EnhancedCreateCaseModal from './EnhancedCreateCaseModal';

/**
 * 增强案例创建示例组件
 * 展示如何使用EnhancedCreateCaseModal组件
 */
const EnhancedCaseCreateExample: React.FC = () => {
  const [modalVisible, setModalVisible] = useState(false);

  const handleCreateSuccess = () => {
    message.success('案例创建成功！');
    // 这里可以添加其他成功后的逻辑，比如刷新列表
  };

  return (
    <div style={{ padding: '20px' }}>
      <Button
        type="primary"
        icon={<PlusOutlined />}
        onClick={() => setModalVisible(true)}
        size="large"
      >
        创建增强案例
      </Button>

      <EnhancedCreateCaseModal
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        onSuccess={handleCreateSuccess}
      />
    </div>
  );
};

export default EnhancedCaseCreateExample;