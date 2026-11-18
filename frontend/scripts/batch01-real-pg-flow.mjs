import puppeteer from 'puppeteer'

const baseURL = process.env.FRONTEND_BASE_URL || 'http://127.0.0.1:13080'
const email = process.env.BATCH01_EMAIL || 'batch01.admin@example.test'
const password = process.env.BATCH01_PASSWORD || 'password'

const browser = await puppeteer.launch({
  headless: 'new',
  args: ['--no-sandbox', '--disable-setuid-sandbox'],
})

const page = await browser.newPage()
page.setDefaultTimeout(30000)

const apiFailures = []
page.on('response', (response) => {
  const url = response.url()
  if (url.includes('/api/v1/') && response.status() >= 400) {
    apiFailures.push(`${response.status()} ${url}`)
  }
})

function assert(condition, message) {
  if (!condition) {
    throw new Error(message)
  }
}

async function waitForText(text, label = text) {
  await page.waitForFunction(
    (expected) => document.body.innerText.includes(expected),
    {},
    text,
  )
  console.log(`ok - ${label}`)
}

async function waitForAnyText(texts, label) {
  await page.waitForFunction(
    (expectedTexts) => expectedTexts.some((text) => document.body.innerText.includes(text)),
    {},
    texts,
  )
  console.log(`ok - ${label}`)
}

async function clickByText(text) {
  const clicked = await page.evaluate((expected) => {
    const normalize = (value) => value.replace(/\s+/g, '')
    const normalizedExpected = normalize(expected)
    const elements = Array.from(document.querySelectorAll('button, a, [role="button"]'))
    const element = elements.find((item) => normalize(item.textContent || '').includes(normalizedExpected))
    if (!element) {
      return false
    }
    element.click()
    return true
  }, text)
  assert(clicked, `未找到可点击元素：${text}`)
}

async function goto(path) {
  await page.goto(`${baseURL}${path}`, { waitUntil: 'networkidle2' })
}

try {
  await goto('/login')
  await page.waitForSelector('input[placeholder="邮箱"]')
  await page.type('input[placeholder="邮箱"]', email)
  await page.type('input[placeholder="密码"]', password)
  await clickByText('登录')
  await page.waitForFunction(() => window.location.pathname === '/dashboard')
  await waitForText('当前数据源', '登录后进入 dashboard')
  await waitForText('正式 API', 'dashboard 使用真实 API')

  await goto('/client')
  await waitForText('客户主档案', '客户管理页面')
  await waitForText('红杉资本投资管理集团', '客户主档案读取 PG 客户')
  await waitForText('上海华信建设集团有限公司', '客户关联方读取 PG 接案关系')

  await goto('/case')
  await waitForText('案件管理', '案件管理页面')
  await waitForText('B01-CASE-001', '案件列表读取 PG 案件编号')
  await waitForText('红杉资本投资管理咨询合同纠纷案', '案件列表读取 PG 案件标题')

  await goto('/conflict')
  await waitForText('利益冲突检测结果', '利益冲突页面')
  await waitForText('conflict_check_records', '冲突页面读取 PG 检测记录')
  await waitForText('高风险', '冲突页面读取 PG 风险等级')

  await goto('/case/create')
  await waitForText('新建案件立案工作台', '新建案件页面')
  await clickByText('保存并提交审批')
  await page.waitForFunction(() => window.location.pathname.startsWith('/approval/'))
  await waitForText('审批中心', '提交后进入审批详情')
  await waitForText('新建案件审批', '审批详情读取新建审批')
  await waitForText('审批状态', '审批详情读取真实状态')

  await clickByText('同意并成案')
  await waitForAnyText(['审批已通过', '已成案'], '审批通过操作完成')

  await goto('/approval')
  await waitForText('审批工作台', '审核管理工作台')
  await waitForText('审批队列', '审批队列读取 PG')

  assert(apiFailures.length === 0, `存在失败 API：\n${apiFailures.join('\n')}`)
  console.log('batch01 real PG frontend flow passed')
} finally {
  await browser.close()
}
