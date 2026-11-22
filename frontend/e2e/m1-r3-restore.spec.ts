import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser } from './utils/test-helpers'

test('M1 R3 nullable alias from a historical party is visible for lawyer review', async ({page}) => {
  await seedAuthenticatedUser(page,'lawyer')
  await page.route(url=>url.pathname==='/api/v1/case-intakes',r=>r.fulfill({json:{success:true,data:{items:[{id:801,title:'打开合成草稿'}],total:1}}}))
  await page.route(url=>url.pathname==='/api/v1/case-intakes/801',r=>r.fulfill({json:{success:true,data:{intake:{id:801,title:'Synthetic draft',status:'draft',client_id:1},revision:'fixture-rev',parties:[{party_role:'opposing_party',entity_type:'LEGAL_PERSON',entity_name:'Sample opponent',aliases:null,metadata:{aliases:['Historic Alias']}}],materials:[]}}}))
  await page.goto('/case/create')
  await page.getByRole('button',{name:'打开服务端草稿'}).click()
  await page.getByRole('button',{name:'打开合成草稿'}).click()
  await expect(page.getByRole('textbox',{name:'法定名称或证件姓名'})).toHaveValue('Sample opponent')
  await expect(page.locator('#intake-opponent-aliases')).toHaveValue('Historic Alias')
})

test('M1 R3 existing related party survives local recovery at identical revision', async ({page}) => {
  await seedAuthenticatedUser(page,'lawyer')
  await page.route(url=>url.pathname==='/api/v1/case-intakes',r=>r.fulfill({json:{success:true,data:{items:[{id:801,title:'打开合成草稿'}],total:1}}}))
  await page.route(url=>url.pathname==='/api/v1/case-intakes/801',r=>r.fulfill({json:{success:true,data:{intake:{id:801,title:'Synthetic draft',status:'draft',client_id:1},revision:'fixture-rev',parties:[{party_role:'opposing_party',entity_type:'LEGAL_PERSON',entity_name:'Sample opponent',aliases:null,metadata:{aliases:['Historic Alias']}},{party_role:'related_party',entity_type:'LEGAL_PERSON',entity_name:'Synthetic related party',metadata:{}}],materials:[]}}}))
  await page.goto('/case/create')
  await page.getByRole('button',{name:'打开服务端草稿'}).click()
  await page.getByRole('button',{name:'打开合成草稿'}).click()
  await expect(page.getByRole('textbox',{name:'法定名称或证件姓名'})).toHaveValue('Sample opponent')
  await expect(page.locator('.batch-related-party')).toContainText('Synthetic related party')
  await page.reload()
  await page.getByRole('button',{name:'继续未完成草稿'}).click()
  await expect(page.getByRole('textbox',{name:'法定名称或证件姓名'})).toHaveValue('Sample opponent')
  await expect(page.locator('.batch-related-party')).toContainText('Synthetic related party')
})
