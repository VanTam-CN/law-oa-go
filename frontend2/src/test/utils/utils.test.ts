import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

// 测试工具函数
describe('Utility Functions', () => {
  describe('localStorage utilities', () => {
    beforeEach(() => {
      localStorage.clear()
      vi.clearAllMocks()
    })

    it('should set and get item from localStorage', () => {
      const testKey = 'test-key'
      const testValue = 'test-value'
      
      localStorage.setItem(testKey, testValue)
      const result = localStorage.getItem(testKey)
      
      expect(result).toBe(testValue)
    })

    it('should remove item from localStorage', () => {
      const testKey = 'test-key'
      const testValue = 'test-value'
      
      localStorage.setItem(testKey, testValue)
      expect(localStorage.getItem(testKey)).toBe(testValue)
      
      localStorage.removeItem(testKey)
      expect(localStorage.getItem(testKey)).toBeNull()
    })

    it('should clear localStorage', () => {
      localStorage.setItem('key1', 'value1')
      localStorage.setItem('key2', 'value2')
      
      expect(localStorage.getItem('key1')).toBe('value1')
      expect(localStorage.getItem('key2')).toBe('value2')
      
      localStorage.clear()
      
      expect(localStorage.getItem('key1')).toBeNull()
      expect(localStorage.getItem('key2')).toBeNull()
    })
  })

  describe('data validation utilities', () => {
    it('should validate email format', () => {
      const validateEmail = (email: string) => {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
        return emailRegex.test(email)
      }

      expect(validateEmail('test@example.com')).toBe(true)
      expect(validateEmail('invalid-email')).toBe(false)
      expect(validateEmail('test@')).toBe(false)
      expect(validateEmail('@example.com')).toBe(false)
    })

    it('should validate phone number format', () => {
      const validatePhone = (phone: string) => {
        const phoneRegex = /^1[3-9]\d{9}$/
        return phoneRegex.test(phone)
      }

      expect(validatePhone('13800138000')).toBe(true)
      expect(validatePhone('13900139000')).toBe(true)
      expect(validatePhone('12345678901')).toBe(false)
      expect(validatePhone('1380013800')).toBe(false)
      expect(validatePhone('138001380000')).toBe(false)
    })

    it('should validate ID card format', () => {
      const validateIdCard = (idCard: string) => {
        const idCardRegex = /^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$/
        return idCardRegex.test(idCard)
      }

      expect(validateIdCard('11010119900307888X')).toBe(true)
      expect(validateIdCard('123456789012345678')).toBe(false)
      expect(validateIdCard('11010119900307888')).toBe(false)
    })
  })

  describe('date utilities', () => {
    it('should format date correctly', () => {
      const formatDate = (date: Date, format: string = 'YYYY-MM-DD') => {
        const year = date.getFullYear()
        const month = String(date.getMonth() + 1).padStart(2, '0')
        const day = String(date.getDate()).padStart(2, '0')
        const hours = String(date.getHours()).padStart(2, '0')
        const minutes = String(date.getMinutes()).padStart(2, '0')
        const seconds = String(date.getSeconds()).padStart(2, '0')

        return format
          .replace('YYYY', String(year))
          .replace('MM', month)
          .replace('DD', day)
          .replace('HH', hours)
          .replace('mm', minutes)
          .replace('ss', seconds)
      }

      const testDate = new Date('2024-01-15T10:30:45')
      
      expect(formatDate(testDate, 'YYYY-MM-DD')).toBe('2024-01-15')
      expect(formatDate(testDate, 'YYYY-MM-DD HH:mm:ss')).toBe('2024-01-15 10:30:45')
      expect(formatDate(testDate, 'MM/DD/YYYY')).toBe('01/15/2024')
    })

    it('should calculate date difference', () => {
      const dateDiff = (date1: Date, date2: Date, unit: 'days' | 'hours' | 'minutes' = 'days') => {
        const diffMs = Math.abs(date2.getTime() - date1.getTime())
        
        switch (unit) {
          case 'days':
            return Math.floor(diffMs / (1000 * 60 * 60 * 24))
          case 'hours':
            return Math.floor(diffMs / (1000 * 60 * 60))
          case 'minutes':
            return Math.floor(diffMs / (1000 * 60))
          default:
            return diffMs
        }
      }

      const date1 = new Date('2024-01-01T00:00:00')
      const date2 = new Date('2024-01-02T12:30:00')

      expect(dateDiff(date1, date2, 'days')).toBe(1)
      expect(dateDiff(date1, date2, 'hours')).toBe(36)
      expect(dateDiff(date1, date2, 'minutes')).toBe(2190)
    })
  })

  describe('string utilities', () => {
    it('should truncate string', () => {
      const truncate = (str: string, length: number, suffix: string = '...') => {
        if (str.length <= length) return str
        return str.substring(0, length) + suffix
      }

      expect(truncate('Hello World', 5)).toBe('Hello...')
      expect(truncate('Hello World', 20)).toBe('Hello World')
      expect(truncate('Hello World', 8, '…')).toBe('Hello Wo…')
    })

    it('should capitalize string', () => {
      const capitalize = (str: string) => {
        return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase()
      }

      expect(capitalize('hello')).toBe('Hello')
      expect(capitalize('HELLO')).toBe('Hello')
      expect(capitalize('hELLO')).toBe('Hello')
    })

    it('should generate random string', () => {
      const generateRandomString = (length: number) => {
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
        let result = ''
        for (let i = 0; i < length; i++) {
          result += chars.charAt(Math.floor(Math.random() * chars.length))
        }
        return result
      }

      const randomStr = generateRandomString(10)
      expect(randomStr).toHaveLength(10)
      expect(randomStr).toMatch(/^[A-Za-z0-9]+$/)
    })
  })

  describe('array utilities', () => {
    it('should remove duplicates from array', () => {
      const removeDuplicates = <T>(arr: T[]): T[] => {
        return [...new Set(arr)]
      }

      const arrayWithDuplicates = [1, 2, 2, 3, 4, 4, 5]
      const uniqueArray = removeDuplicates(arrayWithDuplicates)
      
      expect(uniqueArray).toEqual([1, 2, 3, 4, 5])
    })

    it('should sort array by property', () => {
      const sortByProperty = <T>(arr: T[], property: keyof T, order: 'asc' | 'desc' = 'asc') => {
        return [...arr].sort((a, b) => {
          const aValue = a[property]
          const bValue = b[property]
          
          if (aValue < bValue) return order === 'asc' ? -1 : 1
          if (aValue > bValue) return order === 'asc' ? 1 : -1
          return 0
        })
      }

      const users = [
        { name: 'Alice', age: 25 },
        { name: 'Bob', age: 30 },
        { name: 'Charlie', age: 20 }
      ]

      const sortedByAgeAsc = sortByProperty(users, 'age', 'asc')
      expect(sortedByAgeAsc[0].age).toBe(20)
      expect(sortedByAgeAsc[2].age).toBe(30)

      const sortedByAgeDesc = sortByProperty(users, 'age', 'desc')
      expect(sortedByAgeDesc[0].age).toBe(30)
      expect(sortedByAgeDesc[2].age).toBe(20)
    })

    it('should group array by property', () => {
      const groupByProperty = <T>(arr: T[], property: keyof T) => {
        return arr.reduce((groups, item) => {
          const key = String(item[property])
          if (!groups[key]) {
            groups[key] = []
          }
          groups[key].push(item)
          return groups
        }, {} as Record<string, T[]>)
      }

      const users = [
        { name: 'Alice', department: 'IT' },
        { name: 'Bob', department: 'HR' },
        { name: 'Charlie', department: 'IT' },
        { name: 'David', department: 'HR' }
      ]

      const groupedByDepartment = groupByProperty(users, 'department')
      
      expect(groupedByDepartment['IT']).toHaveLength(2)
      expect(groupedByDepartment['HR']).toHaveLength(2)
      expect(groupedByDepartment['IT'][0].name).toBe('Alice')
      expect(groupedByDepartment['HR'][0].name).toBe('Bob')
    })
  })

  describe('number utilities', () => {
    it('should format currency', () => {
      const formatCurrency = (amount: number, currency: string = 'CNY', locale: string = 'zh-CN') => {
        return new Intl.NumberFormat(locale, {
          style: 'currency',
          currency: currency,
        }).format(amount)
      }

      expect(formatCurrency(1234.56)).toBe('¥1,234.56')
      expect(formatCurrency(1000)).toBe('¥1,000.00')
      expect(formatCurrency(0.99)).toBe('¥0.99')
    })

    it('should format percentage', () => {
      const formatPercentage = (value: number, decimals: number = 2) => {
        return `${(value * 100).toFixed(decimals)}%`
      }

      expect(formatPercentage(0.1234)).toBe('12.34%')
      expect(formatPercentage(0.5)).toBe('50.00%')
      expect(formatPercentage(1)).toBe('100.00%')
    })

    it('should generate random number in range', () => {
      const randomInRange = (min: number, max: number) => {
        return Math.floor(Math.random() * (max - min + 1)) + min
      }

      const randomNum = randomInRange(1, 10)
      expect(randomNum).toBeGreaterThanOrEqual(1)
      expect(randomNum).toBeLessThanOrEqual(10)
    })
  })
})