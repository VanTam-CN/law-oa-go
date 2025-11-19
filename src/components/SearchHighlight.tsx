import React from 'react'
import { Tag } from 'antd'

interface HighlightTextProps {
  text: string
  highlight: string
  color?: string
}

const HighlightText: React.FC<HighlightTextProps> = ({ text, highlight, color = '#f50' }) => {
  if (!highlight || !text) {
    return <span>{text}</span>
  }

  // 转义特殊字符以避免正则表达式错误
  const escapedHighlight = highlight.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const regex = new RegExp(`(${escapedHighlight})`, 'gi')
  const parts = text.split(regex)

  return (
    <span>
      {parts.map((part, index) => {
        if (part.toLowerCase() === highlight.toLowerCase()) {
          return (
            <span
              key={index}
              style={{
                backgroundColor: color,
                color: 'white',
                padding: '2px 4px',
                borderRadius: '3px',
                fontWeight: 'bold',
                margin: '0 1px',
              }}
            >
              {part}
            </span>
          )
        }
        return <span key={index}>{part}</span>
      })}
    </span>
  )
}

interface SearchHighlightProps {
  content: string
  searchTerms: string[]
  maxLength?: number
}

const SearchHighlight: React.FC<SearchHighlightProps> = ({
  content,
  searchTerms,
  maxLength = 200,
}) => {
  if (!searchTerms || searchTerms.length === 0) {
    return <span>{content}</span>
  }

  let highlightedContent = content

  // 如果内容太长，截取并高亮相关部分
  if (content.length > maxLength) {
    const lowerContent = content.toLowerCase()
    const firstMatch = searchTerms.reduce((foundIndex, term) => {
      const index = lowerContent.indexOf(term.toLowerCase())
      return index !== -1 && (foundIndex === -1 || index < foundIndex) ? index : foundIndex
    }, -1)

    if (firstMatch !== -1) {
      const start = Math.max(0, firstMatch - 50)
      const end = Math.min(content.length, firstMatch + maxLength)
      highlightedContent =
        (start > 0 ? '...' : '') +
        content.substring(start, end) +
        (end < content.length ? '...' : '')
    } else {
      highlightedContent = `${content.substring(0, maxLength)}...`
    }
  }

  // 应用所有搜索词高亮
  const colors = ['#f50', '#2db7f5', '#87d068', '#108ee9', '#ff6b6b', '#722ed1', '#fa8c16']
  let result: any = highlightedContent

  searchTerms.forEach((term, index) => {
    if (!term || !term.trim()) {
      return
    }

    const color = colors[index % colors.length]
    const regex = new RegExp(`(${term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')

    // 如果result已经是JSX元素，需要特殊处理
    if (typeof result === 'string') {
      const parts = result.split(regex)
      result = (
        <span>
          {parts.map((part, partIndex) => {
            if (part.toLowerCase() === term.toLowerCase()) {
              return (
                <span
                  key={`${index}-${partIndex}`}
                  style={{
                    backgroundColor: color,
                    color: 'white',
                    padding: '2px 4px',
                    borderRadius: '3px',
                    fontWeight: 'bold',
                    margin: '0 1px',
                  }}
                >
                  {part}
                </span>
              )
            }
            return <span key={`${index}-${partIndex}`}>{part}</span>
          })}
        </span>
      )
    } else {
      // 如果已经是JSX元素，需要递归处理文本节点
      result = (
        <HighlightText
          text={typeof result === 'string' ? result : ''}
          highlight={term}
          color={color}
        />
      )
    }
  })

  return <span>{result}</span>
}

export { SearchHighlight, HighlightText }
