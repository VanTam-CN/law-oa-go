import React, { useState, useEffect } from 'react'
import {
  Modal,
  Descriptions,
  Tag,
  Space,
  Button,
  Divider,
  Typography,
  message,
  Spin,
  Tooltip,
  Card,
} from 'antd'
import {
  StarOutlined,
  StarFilled,
  ShareAltOutlined,
  DownloadOutlined,
  CopyOutlined,
  FileTextOutlined,
  TagsOutlined,
} from '@ant-design/icons'
import {
  LawItem,
  LawCategory,
  getRelatedLaws,
  addToFavorites,
  removeFromFavorites,
} from '@/services/tools'

const { Title, Text, Paragraph } = Typography

interface LawDetailModalProps {
  visible: boolean
  law: LawItem | null
  onClose: () => void
}

const LawDetailModal: React.FC<LawDetailModalProps> = ({ visible, law, onClose }) => {
  const [loading, setLoading] = useState<boolean>(false)
  const [relatedLaws, setRelatedLaws] = useState<LawItem[]>([])
  const [copySuccess, setCopySuccess] = useState<boolean>(false)

  useEffect(() => {
    if (visible && law) {
      loadRelatedLaws()
    }
  }, [visible, law])

  const loadRelatedLaws = async () => {
    if (!law) {
      return
    }

    try {
      setLoading(true)
      const response = await getRelatedLaws(law.id, 5)
      setRelatedLaws(response.data?.statutes || [])
    } catch (error) {
      console.error('Failed to load related laws:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleFavorite = async () => {
    if (!law) {
      return
    }

    try {
      if (law.isFavorited) {
        await removeFromFavorites(law.id)
        message.success('已取消收藏')
        // 这里需要通知父组件更新状态
      } else {
        await addToFavorites(law.id)
        message.success('已添加到收藏')
      }
    } catch (error) {
      console.error('Failed to update favorite status:', error)
      message.error('操作失败，请稍后重试')
    }
  }

  const handleCopyContent = async () => {
    if (!law) {
      return
    }

    const content = `${law.statuteNumber}\n${law.title}\n${law.lawName}\n\n${law.content}`

    try {
      await navigator.clipboard.writeText(content)
      setCopySuccess(true)
      message.success('内容已复制到剪贴板')
      setTimeout(() => setCopySuccess(false), 2000)
    } catch (error) {
      console.error('Failed to copy content:', error)
      message.error('复制失败，请手动复制')
    }
  }

  const handleShare = () => {
    if (!law) {
      return
    }

    const shareUrl = `${window.location.origin}/laws/${law.id}`

    if (navigator.share) {
      navigator.share({
        title: law.title,
        text: `${law.statuteNumber} - ${law.title}`,
        url: shareUrl,
      })
    } else {
      navigator.clipboard.writeText(shareUrl)
      message.success('链接已复制到剪贴板')
    }
  }

  const handleDownload = () => {
    if (!law) {
      return
    }

    const content = `${law.statuteNumber}\n${law.title}\n${law.lawName}\n\n${law.content}`
    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${law.statuteNumber}_${law.title}.txt`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    message.success('法条已下载')
  }

  const getCategoryColor = (code: string) => {
    const colors: { [key: string]: string } = {
      CIVIL_LAW: 'blue',
      CRIMINAL_LAW: 'red',
      ADMINISTRATIVE_LAW: 'orange',
      ECONOMIC_LAW: 'green',
      LABOR_LAW: 'purple',
      COMMERCIAL_LAW: 'cyan',
      OTHER: 'default',
    }
    return colors[code] || 'default'
  }

  const getStatusText = (status: string) => {
    const statusMap: { [key: string]: string } = {
      active: '生效',
      expired: '失效',
      repealed: '废止',
    }
    return statusMap[status] || status
  }

  const getStatusColor = (status: string) => {
    const colorMap: { [key: string]: string } = {
      active: 'success',
      expired: 'warning',
      repealed: 'error',
    }
    return colorMap[status] || 'default'
  }

  if (!law) {
    return null
  }

  return (
    <Modal
      title={
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <FileTextOutlined />
          <span>{law.title}</span>
          <Tag color={getStatusColor(law.status)}>{getStatusText(law.status)}</Tag>
        </div>
      }
      open={visible}
      onCancel={onClose}
      width={900}
      footer={[
        <Button key='favorite' onClick={handleFavorite}>
          {law.isFavorited ? <StarFilled style={{ color: '#faad14' }} /> : <StarOutlined />}
          {law.isFavorited ? '取消收藏' : '添加收藏'}
        </Button>,
        <Button key='copy' onClick={handleCopyContent}>
          <CopyOutlined />
          {copySuccess ? '已复制' : '复制内容'}
        </Button>,
        <Button key='share' onClick={handleShare}>
          <ShareAltOutlined />
          分享
        </Button>,
        <Button key='download' onClick={handleDownload}>
          <DownloadOutlined />
          下载
        </Button>,
        <Button key='close' onClick={onClose}>
          关闭
        </Button>,
      ]}
    >
      <div className='law-detail-modal'>
        {/* 基本信息 */}
        <Card title='基本信息' size='small' style={{ marginBottom: 16 }}>
          <Descriptions column={2} size='small'>
            <Descriptions.Item label='法条编号' span={1}>
              <Text code style={{ fontSize: '14px' }}>
                {law.statuteNumber}
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label='法律名称' span={1}>
              {law.lawName}
            </Descriptions.Item>
            <Descriptions.Item label='所属分类' span={1}>
              {law.category && (
                <Tag color={getCategoryColor(law.category.code)}>{law.category.name}</Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label='状态' span={1}>
              <Tag color={getStatusColor(law.status)}>{getStatusText(law.status)}</Tag>
            </Descriptions.Item>
            {law.publishingAuthority && (
              <Descriptions.Item label='发布机关' span={2}>
                {law.publishingAuthority}
              </Descriptions.Item>
            )}
            {law.effectiveDate && (
              <Descriptions.Item label='生效日期' span={1}>
                {new Date(law.effectiveDate).toLocaleDateString()}
              </Descriptions.Item>
            )}
            {law.expiryDate && (
              <Descriptions.Item label='失效日期' span={1}>
                {new Date(law.expiryDate).toLocaleDateString()}
              </Descriptions.Item>
            )}
            {law.chapter && (
              <Descriptions.Item label='章节' span={1}>
                {law.chapter}
              </Descriptions.Item>
            )}
            {law.section && (
              <Descriptions.Item label='节' span={1}>
                {law.section}
              </Descriptions.Item>
            )}
            {law.part && (
              <Descriptions.Item label='篇' span={1}>
                {law.part}
              </Descriptions.Item>
            )}
          </Descriptions>

          {/* 标签 */}
          {law.tags && law.tags.length > 0 && (
            <div style={{ marginTop: 16 }}>
              <Text strong>标签：</Text>
              <div style={{ marginTop: 8 }}>
                {law.tags.map((tag) => (
                  <Tag key={tag} color='blue' style={{ marginBottom: 4 }}>
                    {tag}
                  </Tag>
                ))}
              </div>
            </div>
          )}
        </Card>

        {/* 法条内容 */}
        <Card title='法条内容' size='small' style={{ marginBottom: 16 }}>
          <Paragraph
            style={{
              lineHeight: 1.8,
              fontSize: '14px',
              whiteSpace: 'pre-wrap',
              background: '#f8f9fa',
              padding: '16px',
              borderRadius: '4px',
              border: '1px solid #e9ecef',
            }}
          >
            {law.content}
          </Paragraph>

          {/* 关键词 */}
          {law.keywords && law.keywords.length > 0 && (
            <div style={{ marginTop: 16 }}>
              <Text strong>
                <TagsOutlined style={{ marginRight: 8 }} />
                关键词：
              </Text>
              <div style={{ marginTop: 8 }}>
                {law.keywords.map((keyword) => (
                  <Tag key={keyword} color='purple' style={{ marginBottom: 4 }}>
                    {keyword}
                  </Tag>
                ))}
              </div>
            </div>
          )}
        </Card>

        {/* 相关法条 */}
        {relatedLaws.length > 0 && (
          <Card
            title={
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <span>相关法条</span>
                <Text type='secondary' style={{ fontSize: '12px' }}>
                  (基于分类和内容推荐)
                </Text>
              </div>
            }
            size='small'
          >
            <Spin spinning={loading}>
              <Space direction='vertical' style={{ width: '100%' }} size='small'>
                {relatedLaws.map((relatedLaw) => (
                  <div
                    key={relatedLaw.id}
                    style={{
                      padding: '12px',
                      border: '1px solid #f0f0f0',
                      borderRadius: '6px',
                      cursor: 'pointer',
                      transition: 'all 0.2s',
                    }}
                    onClick={() => {
                      // TODO: 切换到新的法条详情
                      console.log('查看相关法条:', relatedLaw)
                    }}
                  >
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'flex-start',
                      }}
                    >
                      <div style={{ flex: 1 }}>
                        <Text strong style={{ fontSize: '13px' }}>
                          {relatedLaw.statuteNumber}
                        </Text>
                        <div style={{ fontSize: '12px', color: '#666', marginTop: 2 }}>
                          {relatedLaw.title}
                        </div>
                        <div style={{ fontSize: '12px', color: '#999', marginTop: 2 }}>
                          {relatedLaw.lawName}
                        </div>
                      </div>
                      <Button type='link' size='small' style={{ padding: '0 8px' }}>
                        查看详情
                      </Button>
                    </div>
                  </div>
                ))}
              </Space>
            </Spin>
          </Card>
        )}

        {/* 元数据 */}
        <Card title='元数据' size='small'>
          <Descriptions column={3} size='small'>
            <Descriptions.Item label='创建时间'>
              {new Date(law.createdAt).toLocaleString()}
            </Descriptions.Item>
            <Descriptions.Item label='更新时间'>
              {new Date(law.updatedAt).toLocaleString()}
            </Descriptions.Item>
            <Descriptions.Item label='层级级别'>第 {law.hierarchyLevel} 级</Descriptions.Item>
            <Descriptions.Item label='浏览次数' span={1}>
              {law.viewCount || 0}
            </Descriptions.Item>
            <Descriptions.Item label='收藏次数' span={2}>
              {law.favoriteCount || 0}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      </div>
    </Modal>
  )
}

export default LawDetailModal
