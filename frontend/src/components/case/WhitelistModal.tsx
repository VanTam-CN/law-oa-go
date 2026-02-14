/**
 * WhitelistModal 组件
 * 白名单管理弹窗
 * 用于管理案件隔离墙的白名单用户
 */

import React, { useCallback, useState, useMemo, useRef } from 'react'
import {
  Modal,
  List,
  Avatar,
  Tag,
  Button,
  Space,
  Typography,
  Input,
  Select,
  message,
  Popconfirm,
  Empty,
  Spin,
  Divider,
  Tooltip,
  Alert,
} from 'antd'
import {
  TeamOutlined,
  UserAddOutlined,
  DeleteOutlined,
  SearchOutlined,
  UserOutlined,
  MailOutlined,
  ApartmentOutlined,
} from '@ant-design/icons'
import { WhitelistUser, UserOption } from '../../types/ethicalWall'
import {
  useWhitelist,
  useAddWhitelist,
  useRemoveWhitelist,
  useUserSearch,
} from '../../hooks/useEthicalWall'
import './WhitelistModal.less'

const { Text, Paragraph } = Typography
const { Search } = Input
const { Option } = Select

interface WhitelistModalProps {
  open: boolean
  caseId: string
  caseName?: string
  onClose: () => void
  className?: string
}

/**
 * 用户列表项组件
 */
const UserListItem: React.FC<{
  user: WhitelistUser
  onRemove: (userId: string) => void
  removing?: boolean
}> = ({ user, onRemove, removing }) => {
  const handleRemove = useCallback(() => {
    onRemove(user.userId)
  }, [user.userId, onRemove])

  return (
    <List.Item
      actions={[
        <Tooltip key="remove" title="从白名单移除">
          <Popconfirm
            title="确定要移除该用户吗？"
            description="移除后该用户将无法访问此案件。"
            onConfirm={handleRemove}
            okText="确定"
            cancelText="取消"
            okButtonProps={{ danger: true }}
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              loading={removing}
              size="small"
            >
              移除
            </Button>
          </Popconfirm>
        </Tooltip>,
      ]}
    >
      <List.Item.Meta
        avatar={
          <Avatar size={40} icon={<UserOutlined />} src={user.userEmail ? undefined : undefined}>
            {user.userName.charAt(0).toUpperCase()}
          </Avatar>
        }
        title={
          <Space>
            <Text strong>{user.userName}</Text>
            {user.department && (
              <Tag color="blue" icon={<ApartmentOutlined />} style={{ margin: 0 }}>
                {user.department}
              </Tag>
            )}
            {user.position && <Tag style={{ margin: 0 }}>{user.position}</Tag>}
          </Space>
        }
        description={
          <Space size="middle">
            {user.userEmail && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                <MailOutlined style={{ marginRight: 4 }} />
                {user.userEmail}
              </Text>
            )}
            <Text type="secondary" style={{ fontSize: 12 }}>
              添加时间: {user.addedAt}
            </Text>
            {user.addedBy && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                操作人: {user.addedBy}
              </Text>
            )}
          </Space>
        }
      />
      {user.reason && (
        <div style={{ marginTop: 8 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            备注: {user.reason}
          </Text>
        </div>
      )}
    </List.Item>
  )
}

/**
 * 添加用户表单组件
 */
const AddUserForm: React.FC<{
  onAdd: (userId: string, reason?: string) => void
  loading: boolean
  onCancel: () => void
}> = ({ onAdd, loading, onCancel }) => {
  const [selectedUserId, setSelectedUserId] = useState<string>()
  const [reason, setReason] = useState<string>()
  const [searchKeyword, setSearchKeyword] = useState('')

  // 搜索用户
  const { data: users = [], isLoading: searchLoading } = useUserSearch(searchKeyword)

  const handleAdd = useCallback(() => {
    if (!selectedUserId) {
      message.warning('请选择要添加的用户')
      return
    }
    onAdd(selectedUserId, reason)
    setSelectedUserId(undefined)
    setReason(undefined)
    setSearchKeyword('')
  }, [selectedUserId, reason, onAdd])

  const handleCancel = useCallback(() => {
    setSelectedUserId(undefined)
    setReason(undefined)
    setSearchKeyword('')
    onCancel()
  }, [onCancel])

  const selectedUser = useMemo(
    () => users.find((u) => u.id === selectedUserId),
    [users, selectedUserId],
  )

  return (
    <div className="add-user-form">
      <Space direction="vertical" size="small" style={{ width: '100%' }}>
        <div>
          <Text strong>选择用户</Text>
          <Select
            showSearch
            value={selectedUserId}
            onChange={setSelectedUserId}
            onSearch={setSearchKeyword}
            filterOption={false}
            notFoundContent={searchLoading ? <Spin size="small" /> : '请输入姓名或邮箱搜索'}
            placeholder="输入姓名或邮箱搜索用户"
            style={{ width: '100%', marginTop: 8 }}
            allowClear
            size="middle"
          >
            {users.map((user) => (
              <Option key={user.id} value={user.id}>
                <Space>
                  <Avatar size={24} icon={<UserOutlined />}>
                    {user.name.charAt(0).toUpperCase()}
                  </Avatar>
                  <div>
                    <div>{user.name}</div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {user.email}
                    </Text>
                  </div>
                </Space>
              </Option>
            ))}
          </Select>
        </div>

        <div>
          <Text strong>备注（可选）</Text>
          <Input.TextArea
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="添加该用户到白名单的原因"
            rows={2}
            maxLength={200}
            showCount
            style={{ marginTop: 8 }}
          />
        </div>

        {selectedUser && (
          <Alert
            message={`将添加用户: ${selectedUser.name}`}
            description={selectedUser.email}
            type="info"
            showIcon
            style={{ marginTop: 8 }}
          />
        )}

        <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
          <Button size="small" onClick={handleCancel}>
            取消
          </Button>
          <Button
            type="primary"
            icon={<UserAddOutlined />}
            onClick={handleAdd}
            loading={loading}
            size="small"
            disabled={!selectedUserId}
          >
            添加到白名单
          </Button>
        </Space>
      </Space>
    </div>
  )
}

/**
 * 主弹窗组件
 */
const WhitelistModal: React.FC<WhitelistModalProps> = ({
  open,
  caseId,
  caseName,
  onClose,
  className = '',
}) => {
  const [showAddForm, setShowAddForm] = useState(false)
  const [removingUserId, setRemovingUserId] = useState<string>()

  // 查询白名单
  const { data: whitelistData, isLoading, refetch } = useWhitelist(caseId, open)

  // 添加/移除白名单
  const addMutation = useAddWhitelist()
  const removeMutation = useRemoveWhitelist()

  const users = whitelistData?.users ?? []
  const total = whitelistData?.total ?? 0

  /**
   * 处理添加用户
   */
  const handleAddUser = useCallback(
    (userId: string, reason?: string) => {
      addMutation.mutate(
        { caseId, userId, reason },
        {
          onSuccess: () => {
            message.success('已添加到白名单')
            setShowAddForm(false)
            refetch()
          },
          onError: (err: Error) => {
            message.error(`添加失败: ${err.message}`)
          },
        },
      )
    },
    [caseId, addMutation, refetch],
  )

  /**
   * 处理移除用户
   */
  const handleRemoveUser = useCallback(
    (userId: string) => {
      setRemovingUserId(userId)
      removeMutation.mutate(
        { caseId, userId },
        {
          onSuccess: () => {
            message.success('已从白名单移除')
            setRemovingUserId(undefined)
            refetch()
          },
          onError: (err: Error) => {
            message.error(`移除失败: ${err.message}`)
            setRemovingUserId(undefined)
          },
        },
      )
    },
    [caseId, removeMutation, refetch],
  )

  /**
   * 处理关闭弹窗
   */
  const handleClose = useCallback(() => {
    setShowAddForm(false)
    onClose()
  }, [onClose])

  /**
   * 弹窗标题
   */
  const modalTitle = (
    <Space>
      <TeamOutlined />
      <span>白名单管理</span>
      {caseName && <Text type="secondary">- {caseName}</Text>}
      <Tag color="blue">{total} 人</Tag>
    </Space>
  )

  return (
    <Modal
      title={modalTitle}
      open={open}
      onCancel={handleClose}
      footer={null}
      width={700}
      className={`whitelist-modal ${className}`}
      destroyOnClose
    >
      <div className="whitelist-content">
        {/* 添加用户区域 */}
        <div className="whitelist-actions">
          {!showAddForm ? (
            <Button
              type="primary"
              icon={<UserAddOutlined />}
              onClick={() => setShowAddForm(true)}
              block
            >
              添加用户到白名单
            </Button>
          ) : (
            <div className="add-user-wrapper">
              <AddUserForm
                onAdd={handleAddUser}
                loading={addMutation.isPending}
                onCancel={() => setShowAddForm(false)}
              />
            </div>
          )}
        </div>

        <Divider style={{ margin: '16px 0' }} />

        {/* 白名单列表 */}
        <div className="whitelist-list">
          {isLoading ? (
            <div style={{ textAlign: 'center', padding: 40 }}>
              <Spin />
              <div style={{ marginTop: 16 }}>
                <Text type="secondary">加载中...</Text>
              </div>
            </div>
          ) : users.length === 0 ? (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                <div>
                  <Paragraph>暂无白名单用户</Paragraph>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    点击上方按钮添加用户到白名单
                  </Text>
                </div>
              }
            />
          ) : (
            <List
              dataSource={users}
              renderItem={(user) => (
                <UserListItem
                  key={user.userId}
                  user={user}
                  onRemove={handleRemoveUser}
                  removing={removingUserId === user.userId && removeMutation.isPending}
                />
              )}
              style={{ maxHeight: 400, overflowY: 'auto' }}
            />
          )}
        </div>

        {/* 底部提示 */}
        {users.length > 0 && (
          <Alert
            message="白名单说明"
            description="白名单中的用户可以访问当前案件，即使设置了隔离墙保护。移除用户后，该用户将立即失去访问权限。"
            type="info"
            showIcon
            style={{ marginTop: 16 }}
          />
        )}
      </div>
    </Modal>
  )
}

export default WhitelistModal
