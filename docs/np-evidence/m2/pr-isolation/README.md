# M2 试点 PR 的 CI 隔离

工程提交 `b070323`。规则适用于 head 为 `codex/np-main-m2` 且工作流匹配目标分支的 PR。当前开发分支按 push 执行原生 PG 门禁。以 M1 为 base 的评审 PR 与面向 main 的集成 PR 有不同事件过滤条件，须分别取证。

## 核对结果

| 工作流 | 试点 PR 的处理 |
| --- | --- |
| tests | PR merge ref 也进入原生 PG 门禁。保留旧 Go、前端和 E2E 检查；Go 与前端在产生覆盖率的 job 内写入数值摘要。屏蔽 Codecov、原始报告/截图上传、对应下载和机器人评论。 |
| ci | 保留质量检查；屏蔽覆盖率外传和整个通知 job，包括失败路径的 Slack、邮件、部署报告。 |
| ci-cd | 保留原有分析、测试、基准与 fuzz 检查；屏蔽原始目录上传和未筛选的步骤摘要。docker-build、部署与 generate-report 在 PR 上原已排除。 |
| security | 保留冒烟门禁和依赖检查；依赖检查可记录退出码，详细扫描报告留在 runner。 |
| pgo | 试点在同一 runner 采集各包的 profile，合并后构建标准和 PGO 应用，分别核对本地存活探针。试点不跨 job 搬运 profile、二进制和原始报告。原有工作负载检查的失败仍能独立反映。 |
| fuzzing | 继续使用前轮修复后的路由和目标级 PASS 凭证。 |

试点上同时关闭显式 Actions cache 和 setup-go/setup-node 的自动缓存。普通 PR 与 main push 的原有路径作为对照。离线检查按 job 依赖与 step 条件求值，考虑失败和取消，并用已配置通知地址的合成环境测试。故意恢复失败截图上传、通知 job 或工具缓存后，检查器须能检出。检查器覆盖工作流中列出的 action、部署 job 与工具缓存；脚本和步骤摘要的读写另经源码复核。

## 隔离运行

- [远端正常 PG 对照](https://github.com/VanTam-CN/law-oa-go/actions/runs/45317903182) 与 [同 SHA 的五项负向](https://github.com/VanTam-CN/law-oa-go/actions/runs/45317907862) 逐 job 核对。缺 DSN 在连接前退出；部分用例通过仍返回 `M2_REQUIRED_TEST_MISSING`。真实 t.Skip、零用例和缺 PG 的既有红灯继续生效。
- 本地以两包测试剖面核对 CPU/内存合并，并验证空包清单失败。另以单包剖面构建两份 Linux/arm64 应用，在无外网、无主机端口映射的 PG 16.14 + Redis 7 临时环境分别检查 50 次存活。此处验证了构建和启动，不据此推断性能提升或 M5 真入口验收。
- 首次隔离启动发现开发模式的空库演示数据初始化会碰撞 `idx_users_username`。PGO 使用带一条不可登录合成用户的基线，业务启动代码未修改。空库演示初始化缺陷另计。
- 根目录六个工作流与父版本分别运行 actionlint 1.7.9，两者均报告十处既有 setup-go@v4 版本提示。真实 main 目标 PR、全包 PGO 性能结果、M3/M4 夹具和完整 I11 仍待验证。历史 PGO job 引用的两个工作负载源文件不在当前 Git 树中。

[验收数据](verification.json) 仅包含结构字段。重跑事件条件检查时，在仓库根目录执行以下命令。依赖安装在临时目录。

```sh
audit_dir=$(mktemp -d)
npm install --prefix "$audit_dir" --ignore-scripts --no-audit --no-fund @actions/expressions@0.3.62 yaml@2.9.1
cp docs/np-evidence/m2/pr-isolation/check-routes.mjs "$audit_dir/check-routes.mjs"
node "$audit_dir/check-routes.mjs"
```
