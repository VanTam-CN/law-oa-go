# P0 冲突检查与受控试用五角色 QA 夹具

## 用途

`cmd/qa-fixture` 是非生产环境的可重复验收数据入口，用于验证律师 A、律师 B、独立冲突核查人 C、助理和财务的真实操作边界。它不是数据库迁移，也不会由应用启动或生产 schema bootstrap 自动执行。

夹具中的账号、客户、案件、主体和审批单全部为虚构数据，邮箱使用 `qa.invalid` 保留域名，不得替换为真实客户、真实身份证号、统一社会信用代码或正式业务账号。

## 安全门禁

命令必须同时满足以下条件：

- `ENVIRONMENT` 为 `development`、`test` 或 `qa`；`staging` 只有显式设置 `QA_FIXTURE_ALLOW_STAGING=1` 才允许。
- `ENVIRONMENT=production` 永远拒绝。
- `QA_FIXTURE_CONFIRM=I_UNDERSTAND`。
- `QA_PASSWORD` 由执行环境注入，长度至少 12 个字符；命令运行时生成 bcrypt 哈希，不打印密码、不写入仓库。
- `SUBJECT_DATA_KEY` 必须由本地安全环境或 Secret Manager 注入，长度至少 32 字节；不得把真实密钥或测试密钥写入仓库、日志或浏览器脚本。
- 数据库驱动必须是 PostgreSQL，并且目标数据库已完成 schema bootstrap/迁移。

示例：

```bash
QA_FIXTURE_CONFIRM=I_UNDERSTAND \
QA_PASSWORD="$QA_PASSWORD_FROM_SECRET_MANAGER" \
SUBJECT_DATA_KEY="$SUBJECT_DATA_KEY_FROM_SECRET_MANAGER" \
make qa-seed-conflict-p0

QA_FIXTURE_CONFIRM=I_UNDERSTAND \
SUBJECT_DATA_KEY="$SUBJECT_DATA_KEY_FROM_SECRET_MANAGER" \
make qa-verify-conflict-p0
```

不要在 shell 历史、CI 日志或工单中写入真实密码。`QA_PASSWORD_FROM_SECRET_MANAGER` 只是示意变量名，应由实际 Secret Manager 注入。

## 夹具内容

| 对象 | 验收数据 |
|---|---|
| 律师 A | `qa.lawyer.a@qa.invalid`，拥有 A 案和 A 的冲突任务 |
| 律师 B | `qa.lawyer.b@qa.invalid`，拥有 B 案；不能读取 A 案 |
| 核查人 C | `qa.conflict.officer@qa.invalid`，可以复核隔离历史证据 |
| 助理 | `qa.assistant@qa.invalid`，拥有一条虚构冲突材料整理待办，用于验证协作与越权边界 |
| 财务 | `qa.finance@qa.invalid`，用于验证财务/信托入口与反向越权边界 |
| A 当前客户 | 星河智联科技有限公司 |
| B 历史客户 | 云杉数据服务有限公司 |
| A 案 | `QA-P0-A-2026-001`，对方名称与历史登记主体候选一致，需独立复核 |
| B 案 | `QA-P0-B-2026-001`，已开启隔离墙 |
| 主体信息 | 云杉数据的别名、曾用名和虚构关联主体 |
| 检测单 | `QA-P0-A-CHECK-20260719`，状态为已完成检索、待人工复核 |
| 审批单 | `APR-QA-P0-20260719`，状态为 `submitted`，成案状态为阻断 |

这里的名称候选只表示“需要核实”，不是自动确认高风险冲突。实际验收应检查页面是否明确提示：缺少唯一主体标识时不能得出“无冲突”结论，也不能直接成案。

## 推荐验收顺序

1. A 登录，打开利益冲突清单，查看 `QA-P0-A-2026-001` 的检测详情。
2. A 只能看到受限命中提示、当前主体和下一步联系核查人的说明，不能看到 B 案的案号、案情、团队或原始证据。
3. A 打开审批单，确认“同意并成案”在信息不足或未完成独立复核时不可用。
4. B 登录，确认案件列表、冲突清单、客户列表和直接访问 URL 都不能读取 A 的资源。
5. C 登录，确认可以查看执行复核所需的历史命中和证据，并看到复核指定与回避声明。
6. 助理登录待办中心，确认可看到“协助整理星河智联冲突复核材料”，但不能代替律师或独立核查人作出结论。
7. 财务登录财务中心，确认显示当前 MVP 不可用说明，直接 API 返回 `503/MVP_MODULE_UNAVAILABLE`。
8. 以 C 完成一次“信息不足”复核，确认审批仍被阻断，后端不能通过直接 API 绕过前端按钮。
9. 保存浏览器、HTTP 状态码、审计日志和数据库核验结果，形成验收报告。

## 重复执行规则

夹具使用固定的 `qa.invalid` 账号、案件编号和检测编号，重复执行会更新这些虚构对象并保留审计/索引证据，不会清空数据库，也不会删除正式业务数据。若 QA 数据库中混入真实业务数据，应先换用新的隔离数据库，不要用夹具命令清理。

运行 `-mode verify` 只读核验五角色、隔离案件、冲突检测、审批门禁和四类冲突范围；它不会修复数据。
