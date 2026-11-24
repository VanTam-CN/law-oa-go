# M0 前端与CI闭包复核（t_a56dfe09）

日期2026-11-20。GLM Flash B完成源码读取并形成草稿；草稿落盘未完成，由主代理取回、逐项复核并整理本文。原生main应用基线d9c92fb，执行HEAD2481a47；源首切片2d659d9。未修改业务/工作流，本文不是运行验收。

## 1. 回执合同（MI-R08）

- 源`frontend/src/components/intake/useNarrowPilotSave.ts:5`的Receipt没有action；`:6`的Pending.action='update'属于本地恢复/放弃请求，不能与回执混淆。
- 源`backend/internal/handlers/narrow_pilot_write.go:28-34,184-186`的回执只有id/revision/state/request_id/source_revision，npResult不返回action。放弃POST请求要求action=update，但GET无请求体。M0裁决：沿用可选回执action规则，缺省允许；存在且不符才视为错配，不为此新增一套接口。
- 源Hook`:66-68`只核对PUT的state/id/revision，`:85-89`回查/放弃只按终态清理；均须补原请求request_id与对象核对，成功终态须核有效revision。主干尚无NP Hook、来源与回执接口，不是仅改一个字段即可接入。

## 2. 最小前端/缓存闭包

| 项 | 实际事实与位置 | 后续范围 / 验收 |
|---|---|---|
| 原生正文 | MAIN Batch01Prototype.tsx:1254,4191读写localStorage；仅清部分身份字段 | M4整个接案工作台正文仅内存，C01/C02 |
| 恢复标识 | SRC Hook:23,63按账号+接案单槽；人工输入/冻结体在内存 | M3/M4按请求独立键，不改正文重放，C04–C08 |
| 资格 | SRC Hook eligible以对象/版本/投影为主，角色/所有权须补；服务端仍当次鉴权 | M3本人在职律师/draft/仅摘要分流，I13 |
| 保存/重载 | SRC Batch:7361保存后直接退出，:7285起重载仅防请求后编辑；:7762摘要可编辑 | M4保存A期间输入B留页，重载确认后输入C不被晚到覆盖，C13/C14 |
| 响应缓存 | MAIN main.go:188先注册全局缓存，router.Init:200；router.go:399工作台别名 | M4在任何业务响应缓存读写前排除三类路径，I10/C10 |
| Service Worker | MAIN frontend/src/main.tsx:159-169生产构建尝试register('/sw.js')；本轮源码/构建产物未找到sw.js | 新静态发现；注册成功/当前控制者/Cache Storage均未运行验证。C10须实际核对，不等于已发生泄漏 |

## 3. CI闭包（MI-R05/06）

六套workflow声明PR事件。tests.yml:332-348有PR自动评论，ci.yml:120及:317有Codecov；notify区:635起Slack/邮件按配置条件尝试，实际发送未验证。ci-cd.yml:480虽有评论代码，但所属generate-report job在:432限定push/schedule，不能计作PR可达评论。

`tests.yml:258-272`上传浏览器报告/test-results；Playwright配置保留失败trace/截图/video。fuzzing含各job上传、汇总下载及上传（:428,:512）。M2须沿可达事件追踪完整upload/download/summary，只保留白名单脱敏结构证据，不用删除全部证据链解决问题。

fuzzing六处job级if（:73,126,175,229,287,341）使用env.TEST_TYPE；修表达式之外，必须验证输入枚举、选中job状态和证据缺失，不能以全Skip/零crasher判成功。I11正常/负向远端运行仍NOT_RUN；本轮不修workflow或触发CI。

## 4. 主代理对草稿的裁决与剩余项

已修正草稿的三处混淆：请求action必填不代表回执action必填；MAIN迁移止于79+001而非旧AGENTS的29；ci-cd评论job不在PR路径。接受Service Worker注册尝试这一新事实，并更正缓存提案旧有限检索结论。不能从代码注册语句推定当前浏览器已被控制。

I09用户已选择B：现有case_materials/关联记录保全必测，两版文件流I09-FILES DEFERRED/HOLD；不声称原完整文件能力等价。所有NP写入/缓存/双标签/回执新验收均NOT_RUN。生产main+Redis、浏览器实际SW、真实断线、CI负向及升级回滚不在本次M0通过范围。
