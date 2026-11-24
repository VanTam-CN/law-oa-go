# 路线 B 夹具准备验收

M3 夹具基于 PR #48 的 `448047e`，代码提交 `c8b3162`。glm-flash 参与四场景移植，主调度按原生接口校正了请求键、回查路径、UUID 版本和可选 `action=update`，并把候选生成接回真实服务。运行摘要见 [verification.json](verification.json)。

## PR #48 准入依据

6.2 Astra 独立只读复审，主调度复查 GitHub 日志并重跑六套工作流的表达式检查。当前 PR 的 base 是 M1，真实 PR 事件运行数为 0。M2 head 以 main 为目标的离线检查覆盖成功、失败、取消，25 个可达 job 中题定出口为 0。撤去保护后检查器报出工具缓存、报告、失败截图、Slack 和邮件。三次实跑的产物数都是 0。[五项负向运行](https://github.com/VanTam-CN/law-oa-go/actions/runs/45317907862) · [固定 head 正向运行](https://github.com/VanTam-CN/law-oa-go/actions/runs/45318673293)

当前隔离条件精确比较 M2 分支名。以 M3 为 head 向 main 发 PR 会走普通路径，Astra 的模拟发现 45 个可能的外部步骤或缓存点。M3→main 的远端运行保持 HOLD，先适配实际 head 的出口条件，再重复失败、取消和负对照检查。既有 PGO 工作负载缺少两份源文件的问题另行跟踪。

## 隔离夹具如何运行

使用独占的本机 PG 16、端口 18790、数据库 `np_first_slice_dev`、账号 `np_m1`，先做 v14 bootstrap 和受控追加 86。将该库的合成密码注入 `PGPASSWORD`，将一次性合成律师密码注入 `NP_BROWSER_PASSWORD`。下列命令从仓库根目录运行，manifest 由 seed 排他创建。每个负向用例使用新窗口。

```sh
test_dir=$(mktemp -d)
export NP_BROWSER_SEED="$test_dir/seed.json"
cmp=scripts/np-first-slice/compare-main-browser-db.py
python3 "$cmp" seed "$NP_BROWSER_SEED"
python3 "$cmp" before "$NP_BROWSER_SEED" "$test_dir/before.json"
python3 scripts/np-first-slice/test-main-browser-comparator.py
export VITE_API_PROXY_TARGET=http://127.0.0.1:18789
(cd frontend && npx --no-install playwright test -c playwright.np.config.ts --list)
```

M4 的合成服务须让 seed 返回的 actor 使用助手，并在测试入口支持提交后断开 TCP。服务就绪后，运行同一命令去掉 `--list`，再运行 `python3 "$cmp" after "$NP_BROWSER_SEED" "$test_dir/before.json"`。浏览器退出码和 PG 比较器退出码分别验收。专用配置用系统 Chrome、单 worker、零重试，关闭 trace、截图、视频，清理测试输出。通用和 M1 配置分别发现 109 个原有测试，均排除了专用 spec。

## 本轮实际运行

- 本机一次性 PG 16.14，v14+86。目标草稿初始有 2 条材料，对照另有 1 条，另有客户、两条主体及未知大整数 9007199254740993。PG 生成规范文本，Python 在内存做 SHA-256。比较器自测 12/12，覆盖 2 个正常窗口和 10 个反例。历史手工回执用空串表示无来源，比较器同时识别空串和 SQL NULL。
- 对照运行使用固定历史产品 `2d659d9` 的真实前端和带标签的后端，配合当前分支的四场景 spec/config。为匹配该历史程序写死的 actor 白名单，只在新建的专用库里将用户序列起点设为 98001。目标/对照等业务 ID 全部来自 seed。最终执行的 spec/config 与代码提交仅有文件末尾空行差异。
- Chrome 实跑四项。多轮保存并重开来源、并发拒绝且保留输入、生成期间保留输入及人工摘要均通过。断线场景在真实提交并断开后，准确停在“缺失按请求存储的 v2 恢复键”。历史实现仅写旧键，后续 v2 重登回查与清键断言仍待 M4 产品执行。
- 同一浏览器窗口的 PG 读回显示来源从 0 到 3、已提交回执从 0 到 4。其中 3 条可关联到对应版本的来源，1 条为最终手工保存。目标受保护字段、客户、主体、材料和对照整行的前后散列一致。另行运行的比较器自测使用独立数据。

本轮结论为 M3 路线 B **夹具准备通过**。目标分支的四场景业务验收和完整 I11 待 M4，I09-FILES 两版材料后置，M5 真入口仍待执行。
