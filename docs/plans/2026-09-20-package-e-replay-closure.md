# Package E 收口补充：approve 预检链修复 + 幂等重放 digest-only 匹配（HEAD 69045ff）

- 日期：2026-09-20；分支 codex/d3-multi-party-ui；前置 54deeed（86a4a22 冻结候选矩阵）、a1b15b1（延迟交错修复）
- 证据目录：law-oa-qa-evidence/auditfix-3b81228-codex/run-20260920-multi-role-and-audit/（package-e.md + raw/ + SHA256SUMS + fr-rerun 核对）

## 修复内容（提交 69045ff）
1. approve 预检链三连败（均发生在成案写入前，案件从未落库）：
   - config 分支缺主办律师回退 500 → 补 approval.ApplicantID 回退
   - intake 冻结 client_id 字符串反序列化 500 → coerceCaseNumericFields 统一规范
   - 薄 config 缺 case_type/priority 409 → 从 case_intakes 行回填
2. 幂等重放误报 409 CASE_CREATION_INCOMPLETE：partyKey 在 digest 非空时仍含名称，与写路径「同 digest 合并到规范实体」冲突。修复为 digest 非空两侧 digest-only 参与键；digest 为空保持小写名称匹配，同名异证主体仍分离，多主体要求未缩减。回归测试 TestReplaySnapshotMatchesSameDigestDifferentName 已验证修复前失败/修复后通过（stash 对照）。

## 真实链路验证（受控虚构库，非 mock）
- 样本 38（同主体多角色 OPPOSING_PARTY+GUARANTOR 同 digest）：复核人指定 409 回避契约→200、核查结论 200、approve 首次 200 成案 3 当事人行、完全相同请求重放 200 同一案件，无第二案。
- 样本 37（D3 五主体全 UI 链：CO_CLIENT 共同委托方+两对方+GUARANTOR+WITNESS 同名异证）：修复后二进制下原样重放 200，6 当事人/6 实体不变，无第二案 → D3 共同委托/两对方缺口以 package-d3-full-ui.md 全链样本闭环，D 包不以 related-only 收口。
- go build ./... 与 services/handlers 全量测试 -count=1 通过。
- run-20260917-fr-rerun 项3核对：R03 v2、R06 v4 文件与内容在册有效，无需复跑。

## 统一复验口径
F01-F09/R01-R12 矩阵维持 54deeed（86a4a22）判定；69045ff 仅改审批集成链（F07/F08/F09 路径），该路径已用两个真实样本端到端重验。历史 run 目录零改动；production_ready=false 维持；未 push/merge/部署。
