# M2、M3 向 main 提 PR 的 CI 隔离验收

基线 `34099b7`，工程提交 `c9415b8`，Hermes 任务 `t_5cd3d63a`。glm-flash 将五套工作流里 42 处 PR 来源分支条件扩为 M2 和 M3。复核覆盖六套工作流，包括无需改动的 fuzzing。push、定时和手动触发条件沿用原配置。

## GitHub 实跑

使用 [M2 临时 Draft PR #49](https://github.com/VanTam-CN/law-oa-go/pull/49) 和 [M3 临时 Draft PR #50](https://github.com/VanTam-CN/law-oa-go/pull/50)，两者目标均为 `main@d9c92fb`。M2 head 为 `448047e`，M3 head 为 `c9415b8`。每个 PR 均触发六套工作流。通过取消这些专用探针的运行观测取消路径，并保留自然失败的质量检查。M3 质量检查第一次被取消，第二次重跑自然失败，两次分别取证。

| 真实事件 | M2 | M3 |
| --- | --- | --- |
| 质量检查失败，通知 job 跳过 | [CI](https://github.com/VanTam-CN/law-oa-go/actions/runs/45334857440) | [CI 第 2 次](https://github.com/VanTam-CN/law-oa-go/actions/runs/45334868179/attempts/2) |
| 工作流取消 | [Test Pipeline](https://github.com/VanTam-CN/law-oa-go/actions/runs/45334857311) | [Test Pipeline](https://github.com/VanTam-CN/law-oa-go/actions/runs/45334868282) |

共 12 次工作流运行、13 个执行尝试（含 M3 重跑），其中 2 次自然失败、11 次取消。所有运行的 artifact 数为 0，两个 merge ref 的远端缓存列表均为 0。按实际步骤名称识别的 30 个外部步骤全部跳过，实际保存日志扫描覆盖 30 份工具步骤日志，保存标记为 0。两份临时 PR 已关闭，均未合并。

[GitHub 取证明细](hosted.json) 记录每次运行的事件、提交、结论、失败步骤、产物计数、相关步骤的跳过情况及缓存观测。检查包括工作流配置的 Slack、邮件、PR 评论、显式缓存及工具自动缓存。GitHub 用户订阅的原生通知由平台管理。质量检查在现有 Go 格式检查处失败，该步骤提供了真实的失败路径。

## 路由复核及反例

- 在固定基线运行新检查器，M3 与 M2 的隔离规则不一致，检查器退出 1。对工程提交重跑退出 0。
- 使用 GitHub Actions 表达式解析器，在成功、失败、取消、失败后取消四种条件下分别计算六个 YAML 的分支相关字段。对 M2、M3 都与原 M2 的整份有效配置逐字段比较，包含 job 依赖、步骤、命令、摘要开关、cache 输入及 action 版本。在保守计入可先于故障完成的 job 和步骤后，两条分支均有 25 个可能运行的 job，题定出口为 0。
- `feature/example`、`codex/np-main-m3-followup` 和 `codex/np-main-m30` 的完整有效配置均与各自基线一致。普通 PR 的对照在本地执行，未实际发送测试邮件或 Slack 消息。
- 临时放开失败或取消时的截图上传、工具缓存和通知 job 后，两条分支的负对照均检出产物上传、工具缓存、Slack 和邮件。完整数值见 [路由结果](routes.json)。
- actionlint 1.7.9 对六个工作流的基线与候选均给出相同的十条旧版 setup-go 提示。诊断明细逐项相同。

### 复跑

在仓库根目录执行。依赖只装到临时目录。传入 `34099b7` 可复现修复前的红灯。

```sh
audit_dir=$(mktemp -d)
npm install --prefix "$audit_dir" --ignore-scripts --no-audit --no-fund @actions/expressions@0.3.62 yaml@2.9.1
cp docs/np-evidence/m3/pr-isolation/verify.mjs "$audit_dir/"
node "$audit_dir/verify.mjs" c9415b8
```

可用以下只读请求抽查 M3 的失败步骤及产物。缓存列表是查询时点的状态，需与当次工作流的 cache 条件、实际步骤及工具保存日志合读。

```sh
gh api repos/VanTam-CN/law-oa-go/actions/runs/45334868179/attempts/2/jobs
gh api repos/VanTam-CN/law-oa-go/actions/runs/45334868179/artifacts --jq .total_count
gh api 'repos/VanTam-CN/law-oa-go/actions/caches?ref=refs%2Fpull%2F50%2Fmerge' --jq .total_count
```

本次验收只关闭分支 CI 隔离问题。完整 I11 等待 M4 夹具。
