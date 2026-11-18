import React, { useState, useEffect, useRef } from 'react'
import { Select, Input, Button, Tag, Space, Spin } from 'antd'
import { PlusOutlined, SearchOutlined, LoadingOutlined } from '@ant-design/icons'
import debounce from 'lodash/debounce'
import { searchEntities as searchEntitiesAPI } from '@/services/conflict'
import type { EntitySearchResult } from '@/types/conflict'
import { message } from '@/utils/messageHelper'
import './EntitySelector.less'

const { Option } = Select

interface EntityOption {
  id: string
  name: string
  type: 'person' | 'company'
  taxId?: string
  phone?: string
  address?: string
}

interface EntitySelectorProps {
  value?: string[]
  onChange?: (value: string[]) => void
  placeholder?: string
  disabled?: boolean
  allowMultiple?: boolean
  entityType?: 'person' | 'company' | 'all'
  showCreateButton?: boolean
  onCreateNew?: (name: string) => Promise<void>
  maxLength?: number
  style?: React.CSSProperties
}

const EntitySelector: React.FC<EntitySelectorProps> = ({
  value = [],
  onChange,
  placeholder = '请输入当事人名称',
  disabled = false,
  allowMultiple = true,
  entityType = 'all',
  showCreateButton = true,
  onCreateNew,
  maxLength = 10,
  style,
}) => {
  const [searchText, setSearchText] = useState('')
  const [searching, setSearching] = useState(false)
  const [options, setOptions] = useState<EntityOption[]>([])
  const [creating, setCreating] = useState(false)
  const [dropdownVisible, setDropdownVisible] = useState(false)
  const searchInputRef = useRef<any>(null)

  // 搜索实体
  const handleEntitySearch = debounce(async (text: string) => {
    if (!text || text.length < 2) {
      setOptions([])
      setSearching(false)
      return
    }

    try {
      setSearching(true)
      // 使用实际的实体搜索 API
      const results = await searchEntitiesAPI(text, {
        entityType: entityType === 'all' ? undefined : entityType === 'company' ? 'LEGAL_PERSON' : 'INDIVIDUAL',
      })

      // 转换为选项格式
      const entityOptions: EntityOption[] = results.map((result: EntitySearchResult) => ({
        id: result.id.toString(),
        name: result.name,
        type: ['LEGAL_ENTITY', 'LEGAL_PERSON', 'ORGANIZATION'].includes(result.entity_type)
          ? 'company'
          : 'person',
        taxId: result.identity_number,
      }))
      setOptions(entityOptions)
    } catch (error) {
      console.error('搜索实体失败:', error)
      setOptions([])
      message.error('实体搜索失败，请稍后重试')
    } finally {
      setSearching(false)
    }
  }, 300)

  // 处理搜索输入
  const handleSearch = (text: string) => {
    setSearchText(text)
    if (text) {
      handleEntitySearch(text)
    } else {
      setOptions([])
    }
  }

  // 处理选择
  const handleChange = (newValue: string[]) => {
    if (maxLength && newValue.length > maxLength) {
      message.warning(`最多只能选择 ${maxLength} 个当事人`)
      return
    }
    onChange?.(newValue)
    setSearchText('')
    setOptions([])
  }

  // 处理创建新实体
  const handleCreateNew = async () => {
    if (!searchText || searchText.trim().length === 0) {
      message.warning('请输入当事人名称')
      return
    }

    if (value.includes(searchText.trim())) {
      message.warning('该当事人已存在')
      return
    }

    try {
      setCreating(true)
      if (onCreateNew) {
        await onCreateNew(searchText.trim())
      } else {
        // 默认行为：直接添加到列表
        const newValue = [...value, searchText.trim()]
        onChange?.(newValue)
      }
      setSearchText('')
      setOptions([])
      message.success('添加成功')
    } catch (error) {
      console.error('创建实体失败:', error)
      message.error('添加失败')
    } finally {
      setCreating(false)
    }
  }

  // 移除标签
  const handleRemove = (removedItem: string) => {
    const newValue = value.filter((item) => item !== removedItem)
    onChange?.(newValue)
  }

  // 渲染选项
  const renderOption = (option: EntityOption) => (
    <Option key={option.id} value={option.name}>
      <div className="entity-option">
        <span className="entity-name">{option.name}</span>
        {option.type && (
          <Tag color={option.type === 'company' ? 'blue' : 'green'}>
            {option.type === 'company' ? '企业' : '个人'}
          </Tag>
        )}
        {option.taxId && <span className="entity-taxId">{option.taxId}</span>}
      </div>
    </Option>
  )

  // 渲染下拉菜单底部
  const dropdownRender = (menu: React.ReactNode) => (
    <div className="entity-selector-dropdown">
      {menu}
      {showCreateButton && searchText && !options.some((o) => o.name === searchText) && (
        <div className="create-new-section">
          <Button
            type="link"
            icon={<PlusOutlined />}
            onClick={handleCreateNew}
            loading={creating}
            disabled={value.length >= maxLength}
            style={{ width: '100%', textAlign: 'left' }}
          >
            创建 "{searchText}"
          </Button>
        </div>
      )}
      {searching && (
        <div className="searching-section">
          <Spin size="small" indicator={<LoadingOutlined spin />} />
          <span>搜索中...</span>
        </div>
      )}
    </div>
  )

  // 渲染标签
  const tagRender = (props: any) => {
    const { label, value, onClose } = props
    const onPreventMouseDown = (e: React.MouseEvent) => {
      e.preventDefault()
      e.stopPropagation()
    }

    return (
      <Tag
        color="blue"
        onMouseDown={onPreventMouseDown}
        closable
        onClose={onClose}
        style={{ marginRight: 4, marginBottom: 4 }}
      >
        {label}
      </Tag>
    )
  }

  return (
    <div className="entity-selector" style={style}>
      {allowMultiple ? (
        <Select
          mode="tags"
          value={value}
          onChange={handleChange}
          onSearch={handleSearch}
          onSelect={() => setSearchText('')}
          placeholder={placeholder}
          disabled={disabled}
          open={dropdownVisible}
          onDropdownVisibleChange={setDropdownVisible}
          filterOption={false}
          tagRender={tagRender}
          dropdownRender={dropdownRender}
          notFoundContent={searchText && searching ? null : '请输入名称搜索'}
          maxTagCount="responsive"
          style={{ width: '100%' }}
          getPopupContainer={(trigger) => trigger.parentNode as HTMLElement}
        >
          {options.map(renderOption)}
        </Select>
      ) : (
        <Select
          value={value?.[0]}
          onChange={(newValue) => handleChange(newValue ? [newValue as string] : [])}
          onSearch={handleSearch}
          placeholder={placeholder}
          disabled={disabled}
          showSearch
          filterOption={false}
          notFoundContent={searchText && searching ? null : '请输入名称搜索'}
          style={{ width: '100%' }}
          getPopupContainer={(trigger) => trigger.parentNode as HTMLElement}
          dropdownRender={dropdownRender}
        >
          {options.map(renderOption)}
        </Select>
      )}
      {value.length > 0 && (
        <div className="selected-count">
          已选择 {value.length}/{maxLength}
        </div>
      )}
    </div>
  )
}

export default EntitySelector
