# Package D3 — 五主体全链 UI 验收与成案幂等复验报告

- 日期：2026-09-20
- 候选：分支 codex/d3-multi-party-ui，HEAD 86a4a223933292d75393b924abda3d1ce0341884（报告写入时未产生新代码提交，本报告随下一提交入库）
- 环境：受控虚构库 lawoa_auditfix_3b81228，后端 127.0.0.1:18096，前端 127.0.0.1:15180，全部为受控虚构数据
- production_ready=false（维持；本包为受控虚构库验收，不构成真实律所试运行证据）

## 1. 全链 UI 验收（space13/p1 真实 UI 操作）

1. 律师甲 UI 登记接案单 INT-20260920000929-25E4ED，5 个不同身份全部入列：
   主客户 星河智联（LEGAL_PERSON）；共同委托方 宏图实业集团（CO_CLIENT，信用代码）；对方1 远东恒升（主对手方栏）；对方2 华曦国际贸易（OPPOSING_PARTY 相关方）；保证人 柳云松（GUARANTOR，自然人+身份证）；同名异证 宏图B（WITNESS，自然人+身份证）
2. 保存并冲突检测成功：规范化检索主体 5 个，命中受限历史档案 → 待人工复核
3. 核查岗 qa_conflict_officer 从接案单进入冲突复核，发起冲突审批（APR-20260920001203-9D0F173B），在冲突任务 dialog 中指定管理合伙人为独立复核人并填写回避声明
4. 管理合伙人 qa_managing_partner 在 /conflict 打开本案详情，复核结论选 no_conflict（真实下拉操作）+ 复核依据，提交成功；DB conflict_reviews 确认 no_conflict|qa_managing_partner
5. 审批页“同意并成案”，填处理依据，确认 → 审批通过，正式案件 CASE-20260920002149 生成（cases.id=37，status=active）
6. DB 落库：case_parties 6 行，5 个不同身份，6 个不同 entity——
   PLAINTIFF|CLIENT 星河智联；DEFENDANT|OPPOSING 远东恒升；PLAINTIFF|CLIENT 宏图实业集团（信用代码，digest 9ab0f8df0aba）；DEFENDANT|OPPOSING 华曦国际贸易；INTERESTED_PARTY|RELATED 柳云松（GUARANTOR 正确映射）；WITNESS|RELATED 宏图实业集团（INDIVIDUAL，身份证，digest 51ce889140b6）
   → 同名异证确认为独立 entity，身份 digest 不同，未合并
7. F09：律师甲退出重登，/case 清单直接可见 CASE-20260920002149，详情打开正常
8. 管理合伙人访问 /case/37 显示“无权访问”（权限隔离正常，非缺陷）

证据：00–12 号截图/snapshot（本目录），截图与 snapshot 已做 eyJ token 扫描，无命中。

## 2. 成案幂等/重放复验（本轮新增）

对照“重复审批单”与“成案重放”两条路径：

- 重复审批单：同一冲突任务 CCT_27696f4f 重复 POST /conflict/tasks/:id/approval 两次 → 均 409 CONFLICT_APPROVAL_FINAL，审批计数仍为 1（证据 12）
- 成案重放（决策端点）：以与 UI“同意并成案”完全相同的 POST /integration/approvals/:id/decision 请求体（decision=approve，同一处理依据，未增删任何字段）原样连续重放 2 次：
  - replay1/replay2 均 HTTP 200，data.id、request_number、status=approved、case_created=true、created_case_id=37、case_creation_status=completed、records=1 完全一致
  - cases 总数重放前=29、重放后=29，无新建；cases.id=37 仍唯一，case_parties=6、distinct entity=6 不变
  - 服务端短路：internal/services/approval_conflict_integration_service.go 中“已同意+已成案”重放短路，先校验真实案件存在后才返回同一审批，悬挂链接以 CASE_CREATION_INCOMPLETE 拒绝
  - 证据：13-decision-replay-1.json、14-decision-replay-2.json、16-decision-replay-db-verify.txt
- 成案重放（独立 /case 端点）：不带冲突复核绑定的重放 2 次 → 均 400 CONFLICT_CHECK_MISMATCH，校验在先未触达成案逻辑，不作为幂等证据（证据 15）
- 结论：合法请求（与首次完全相同）重放返回同一完整案件，无新建、无部分创建；未通过“补 conflict_check_id 才走通”的方式绕过，幂等复验通过。

## 3. 覆盖矩阵对照

本轮 D3 样本覆盖：F01 建单、F02 事实确认（保存时自动）、F03 冲突检测、F04 独立复核、F05 审批、F06 成案、F09 退出重登可见性；R 系列同名异证独立 entity、多主体不合并、审批终态防重、成案重放防重均在本样本实证。

## 4. 统计口径更正

本轮样本后基线：cases=29（原 28+1）、approval_requests=56（原 55+1）、case_intakes=88。历史 run 目录不回改。

## 5. 边界与未覆盖

- 本包为受控虚构库 UI/API 验收；不覆盖真实律师试运行、生产数据迁移、多浏览器并发
- /case 独立端点的正常建案路径由 decision 端点成案覆盖，/case 端点仅验证了校验拒绝路径
- production_ready=false 维持，待 F01–F09/R01–R12 全矩阵统一复验报告后按计划另行评定
