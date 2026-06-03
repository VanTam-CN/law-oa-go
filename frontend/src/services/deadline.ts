import { get, post } from './http'

export interface DeadlineCalculateRequest {
  case_id?: number
  start_date: string
  deadline_type: string
  business_days?: boolean
}

export interface DeadlineCalculateResponse {
  start_date: string
  end_date: string
  deadline_type: string
  type_name: string
  total_days: number
  business_days: number
  calendar_days: number
  holidays: string[]
  warnings: string[]
}

export interface DeadlineType {
  type: string
  name: string
  description: string
  default_days: number
  category: string
}

export interface CaseDeadline {
  id: number
  case_id: number
  deadline_type: string
  type_name: string
  start_date: string
  end_date: string
  remaining_days: number
  status: string
  created_at: string
}

export interface DeadlineReminderRequest {
  case_id: number
  deadline_type: string
  deadline_date: string
  reminder_offsets: number[]
  description?: string
}

export const deadlineAPI = {
  calculate: (data: DeadlineCalculateRequest) =>
    post<DeadlineCalculateResponse>('/deadlines/calculate', data),

  getTypes: () =>
    get<DeadlineType[]>('/deadlines/types'),

  getCaseDeadlines: (caseId: number) =>
    get<CaseDeadline[]>(`/deadlines/cases/${caseId}`),

  createReminder: (data: DeadlineReminderRequest) =>
    post('/deadlines/reminders', data),
}
