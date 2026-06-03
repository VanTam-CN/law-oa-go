import { MVP_DASHBOARD_QUICK_ACTIONS } from '../Dashboard'

describe('Dashboard MVP quick actions', () => {
  it('contains only MVP quick actions with correct destinations', () => {
    expect(MVP_DASHBOARD_QUICK_ACTIONS.map((item) => item.label)).toEqual([
      '案件管理',
      '新建立案',
      '冲突检测',
      '客户管理',
      '审批管理',
      '信托账户',
    ])

    expect(MVP_DASHBOARD_QUICK_ACTIONS.map((item) => item.path)).toEqual([
      '/case',
      '/case/create',
      '/conflict',
      '/client',
      '/approval',
      '/trust',
    ])
  })
})
