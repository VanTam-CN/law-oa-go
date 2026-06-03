/**
 * 导出工具函数
 */

/**
 * 下载文件
 * @param blob 文件数据
 * @param filename 文件名
 */
export const downloadFile = (blob: Blob, filename: string): void => {
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

/**
 * 从响应头中获取文件名
 * @param headers 响应头
 * @returns 文件名
 */
export const getFilenameFromHeaders = (headers: Record<string, string>): string => {
  const contentDisposition = headers['content-disposition'] || ''
  const filenameMatch = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/)
  if (filenameMatch && filenameMatch[1]) {
    let filename = filenameMatch[1].replace(/['"]/g, '')
    // 处理 UTF-8 编码的文件名
    if (filename.startsWith('UTF-8')) {
      filename = decodeURIComponent(filename.replace(/^UTF-8''/, ''))
    }
    return filename
  }
  return `export_${new Date().getTime()}.csv`
}

/**
 * 导出CSV文件
 * @param promise API请求Promise
 * @param defaultFilename 默认文件名
 */
export const exportCSV = async (
  promise: Promise<Blob>,
  defaultFilename: string
): Promise<void> => {
  try {
    const response = await promise as any
    const blob = response.data || response

    // 尝试从响应头获取文件名
    let filename = defaultFilename
    if (response.headers && typeof response.headers === 'object') {
      filename = getFilenameFromHeaders(response.headers)
    }

    downloadFile(blob, filename)
  } catch (error) {
    console.error('导出失败:', error)
    throw error
  }
}
