/**
 * 格式化工具函数
 */

/**
 * 格式化日期
 * @param date 日期对象或日期字符串
 * @param format 格式化模式，默认为 'YYYY-MM-DD'
 * @returns 格式化后的日期字符串
 */
export function formatDate(date: Date | string | number, format: string = 'YYYY-MM-DD'): string {
  const d = new Date(date)

  if (isNaN(d.getTime())) {
    return ''
  }

  const year = d.getFullYear()
  const month = d.getMonth() + 1
  const day = d.getDate()
  const hours = d.getHours()
  const minutes = d.getMinutes()
  const seconds = d.getSeconds()

  const pad = (n: number): string => (n < 10 ? `0${n}` : `${n}`)

  return format
    .replace('YYYY', `${year}`)
    .replace('MM', pad(month))
    .replace('DD', pad(day))
    .replace('HH', pad(hours))
    .replace('mm', pad(minutes))
    .replace('ss', pad(seconds))
}

/**
 * 格式化金额
 * @param amount 金额数值
 * @param decimals 小数位数，默认为2
 * @param decimalSeparator 小数点分隔符，默认为'.'
 * @param thousandSeparator 千位分隔符，默认为','
 * @returns 格式化后的金额字符串
 */
export function formatAmount(
  amount: number | string,
  decimals: number = 2,
  decimalSeparator: string = '.',
  thousandSeparator: string = ',',
): string {
  if (amount === null || amount === undefined) {
    return ''
  }

  const num = Number(amount)

  if (isNaN(num)) {
    return ''
  }

  const fixed = num.toFixed(decimals)
  const parts = fixed.split('.')
  const integerPart = parts[0]
  const decimalPart = parts.length > 1 ? parts[1] : ''

  // 添加千位分隔符
  const formattedIntegerPart = integerPart.replace(/\B(?=(\d{3})+(?!\d))/g, thousandSeparator)

  return decimalPart
    ? `${formattedIntegerPart}${decimalSeparator}${decimalPart}`
    : formattedIntegerPart
}

/**
 * 格式化手机号码
 * @param phone 手机号码
 * @returns 格式化后的手机号码，格式为：188 **** 8888
 */
export function formatPhone(phone: string): string {
  if (!phone || phone.length !== 11) {
    return phone
  }

  return `${phone.substring(0, 3)} **** ${phone.substring(7)}`
}

/**
 * 格式化文件大小
 * @param bytes 文件大小（字节）
 * @returns 格式化后的文件大小，如：1.5 MB
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) {
    return '0 Bytes'
  }

  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}
