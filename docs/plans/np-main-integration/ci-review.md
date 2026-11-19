# np-main-integration CI 静态核验

- 性质：GLM初稿为只读静态审查，随后主控复核修订。未运行构建、测试或安装工具，未手动触发远端Actions；准备分支发布后的自动失败记录见末节。
- 目标 main worktree：`output/np-main-integration-prep`，HEAD 固定 `d9c92fbd583a0cbbf7f9b4c071e1f2329981f4d7`，分支 `codex/np-main-integration-prep`，开始时工作树干净。
- 源首切片：`output/np-v04-first-slice`，HEAD 固定 `2d659d9dd750f49ddb1f517746e22851392231a6`，分支 `codex/np-v04-first-slice`，工作树干净。
- 本报告内 `.github/...`、`frontend/...` 等路径均指目标 worktree 内路径。

## 结论

1. **面向main的文档Draft PR会匹配现有CI，且配置了外部副作用**。6个workflow均声明面向main的pull_request（部分也面向develop），没有draft或本目录的整体排除；本轮不创建PR验证这些动作。事件依据见[GitHub官方说明](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)。主控只读确认Actions enabled=true，近期main定时运行有成功也有失败，详见remote-observation.json。
   - `.github/workflows/ci.yml:7-13`（另含 workflow_dispatch，输入允许 production）；
   - `.github/workflows/ci-cd.yml:3-13`；`.github/workflows/tests.yml:6-11`；
   - `.github/workflows/security.yml:3-11`；`.github/workflows/fuzzing.yml:3-11`；`.github/workflows/pgo.yml:3-11`。
2. Draft PR上配置的外部写入：`tests.yml:332-348`有`github.rest.issues.createComment`，test-summary为`if:always()`（:317）；权限/运行结果未核，不称实际评论成功。`ci-cd.yml:473-485`也有评论步骤，但其所属generate-report的条件（:432）限定push/schedule，PR上不会执行。
3. Draft PR 上不会登录/推送镜像或部署：`ci.yml:134,163,188,226,273,294` 与 `ci-cd.yml:315,339` 都排除 `pull_request`；部署 job 条件分别是 develop push / tag / release / 手动（`ci.yml:334,446`，`ci-cd.yml:354,393`）。但 ci.yml 的 `workflow_dispatch` 可手动选 production（`ci.yml:13-27,446`），这是独立于 PR 的部署入口。
4. PR上配置了Codecov上传（`tests.yml:105-112,163-170`，`ci.yml:120-124`）；另有ci.yml的notify job（:585-660），其`if:always()`不排除PR，Slack/邮件步骤仅按相关env/secrets非空判断（:635-660），也可能对PR尝试发送。实际token/权限/secrets与发送结果均未核，不能称已外发。
5. **缺少PG测试环境可能被Skip掩盖**：显式PG集成测试用环境开关Skip（`tests/integration/production_external_evidence_test.go:23-24`、`postgresql_elasticsearch_test.go:19-20`）；跨方言测试缺DSN会跳过（`conflict_dialects/conflict_detection_dialects_test.go:1,85`）。tests.yml虽有postgres服务（:26-39），但没有传CONFLICT_INTEGRATION_PG_DSN，也未检查关键用例是否执行，不能以绿色Go结果证明已测PG。服务自身健康失败可能让runner先失败，不能反向声称“任何PG宕机整条CI仍绿”。NP测试必须单独强制。
6. `-race` 命令真实存在于 tests.yml、ci.yml、ci-cd.yml、Makefile（`tests.yml:87,98`、`ci.yml:116`、`ci-cd.yml:137,139`、`Makefile:50,56,62,68`），但主库大量单测落在 SQLite/内存实现（例如 `internal/auth/test_utils.go:61`、`internal/models/basic_test.go:12`），所以“有 -race”不等于“在 PG 上有 -race 竞态覆盖”；security.yml 默认冒烟无 `-race`（`security.yml:47-49`），race 只在定时/手动扩展任务里（`security.yml:100,145-146`）且 `continue-on-error: true`。
7. 源首切片没有 `.github/`（`git ls-tree` 仅 backend/go.mod、frontend Dockerfile×2、frontend/package.json），无法从切片复用任何 CI workflow；它只提供“最小切片交付物”参照。

## 现有 workflow 盘点（目标 SHA）

| 文件 | PR 触发 | PR 上副作用 | push/tag 部署或外发 |
|---|---|---|---|
| ci.yml | main/develop | 配置Codecov；notify可能按secrets发Slack/邮件，未排除PR；quality执行gofumpt并校验diff | 非PR镜像登录/推送；develop部署staging、tag/手动部署production；其他发布操作须独立授权 |
| ci-cd.yml | main/develop | 无登录/推送；PR 专用 fuzz 快跑（`ci-cd.yml:273-276`） | push main/develop/release/tag 触发 GHCR 登录与镜像推送（`ci-cd.yml:314-347`）；develop 部署 staging、release 部署 production（`ci-cd.yml:350-425`） |
| tests.yml | main/develop | PG+Redis 服务容器、`-race` 测试、PR 评论（`tests.yml:26-98,332-348`）、Codecov | 无 |
| security.yml | main/develop | 无外部写入；依赖扫描只产 artifact 且不阻塞（`security.yml:69-95`） | 无 |
| fuzzing.yml | main/develop | 无外部写入 | 无 |
| pgo.yml | main/develop | 无外部写入；依赖 MySQL runner 服务（`pgo.yml:61-68`） | 无 |

## 失真与缺口

- **runner 服务不等于被测 PG**：tests.yml 起了 postgres:15，却没有任何 job 断言 PG 可用或传 PG DSN 给跨方言测试；PG 宕机时 Go 单测走 SQLite 依然可能绿。
- **tests.yml E2E 环境变量名错**：`tests.yml:234-238` 用 `DB_USER/DB_NAME`，而配置绑定是 `DB_USERNAME/DB_DATABASE`（`internal/config/config.go:185-190`），E2E 后端大概率起不来；`curl ... || echo "Server starting..."`（`tests.yml:244`）把失败吞掉。
- **引用缺失**：ci-cd.yml/pgo.yml引用scripts/pgo_workload.go与comprehensive_pgo_workload.go，目标树未找到相应文件；若执行到对应步骤会失败，未运行，不报本轮已失败。
- **前端Docker路径不一致**：frontend/Dockerfile使用Node18、CI使用20；Dockerfile.prod复制/app/build而Vite输出dist，且只装production依赖。静态契约存在缺口，未实际构建；先核定集成制品采用哪个Dockerfile，再决定必要整改，不同时翻新两条发布链。
- **版本漂移**：根 `Dockerfile:13` 用 golang:1.25-alpine，`go.mod` 声明 go 1.25.0，ci.yml 固定 1.25.0，ci-cd.yml 环境变量却写 `'1.25'`（`ci-cd.yml:16`），且 `actions/setup-go@v4`（`ci-cd.yml:37`）比 ci.yml 的 `@v5`（`ci.yml:73`）旧。
- **Makefile 依赖本机路径**：`Makefile:44` 硬编码 `/Users/mac/go/bin/golangci-lint`，CI 不可用。
- **fuzzing.yml 在 PR 上跑但依赖点不牢**：validators 矩阵引用的 fuzz 函数与包是否存在未逐个核对（未运行）；`go mod tidy`（`fuzzing.yml:90`）会在 CI 中改写 go.mod/go.sum 而不校验 diff。
- **通知/评论权限偏大**：ci.yml `notify` 需要 issues: write 且能发 Slack/邮件（`ci.yml:587-660`）；与“PR 检查最小权限 contents: read”的目标冲突。
- **浏览器缓存缺失**：tests.yml 每次在 PR 上执行 `npx playwright install --with-deps chromium`（`tests.yml:224-226`），无 Playwright 浏览器缓存；frontend 依赖里同时有 puppeteer（`frontend/package.json:35`）与 @playwright/test（`:47`），npm ci 时 puppeteer 的 postinstall 可能下载 Chromium（未核 puppeteer 配置缓存）。
- **frontend Dockerfile npm 源改写**：`frontend/Dockerfile:19` 把 registry 固定到 npmmirror.com，供应链与 CI 一致性差。

## 可复用检查（下一阶段直接保留的素材）

- Go 工具链固定：setup-go + `go-version: 1.25.0` + `cache-dependency-path: go.sum`（`ci.yml:72-77` 的写法）。
- Go 依赖校验：`go mod download && go mod verify`（`security.yml:42-45`）。
- PG/Redis服务容器定义可参考tests.yml:26-49，**当前实际是postgres:15与redis:7**；下一阶段NP测试拟固定至已验证的PG16.14与相应Redis版本，不把旧配置写成已升级。
- Go 测试命令骨架：`go test -race -count=1 -covermode=atomic ./...` 与 `-tags=integration` 变体（`tests.yml:87,98`）。
- 前端最小链：setup-node 20 + npm 缓存 + `npm ci` + `npm run lint` + `npm run type-check` + `npm run test:ci`（`ci.yml:239-259`、`frontend/package.json:18-21`）。
- Docker 构建校验（不推送）：`docker build --target production --build-arg BUILD_COMMIT=<sha> .`，路径与参数来自 `docker-compose.yml:7-18` 与 `Dockerfile:41-103`。
- 只读安全基线：`security-smoke`（无 race，可改加 `-race`）与 `govulncheck`（`security.yml:47-49,69-73`）。

## 下一阶段最小CI范围（主控复核后）

1. 先局部隔离现有PR路径中的评论、Codecov、Slack/邮件及发布动作；保留必要的只读检查。**本阶段不整体替换六套工作流、不重建release流水线、不改真实部署权限或设置**。本轮只准备方案，不创建PR。
2. 必须有NP专用强制检查job（可新增或复用），基于正确根目录执行go build；相关Go测试显式-race -count=1，原生/上下文/NP实际PG用例缺失或Skip均失败。
3. NP fixture保持127.0.0.1:18790、库np_first_slice_dev与NP_TEST_DSN隔离保护。CI固定PG16.14、按主干最小依赖闭包显式初始化，再验证实际数据库身份和NP两表/唯一索引。另组conflict_dialects测试不能代替它。变量用DB_USERNAME/DB_DATABASE、REDIS_HOST/REDIS_PORT，不推荐未绑定的REDIS_ADDR。
4. npm ci、类型、关键Jest、build必须实际执行。不得使用passWithNoTests或允许错误来放行；已有test:ci脚本含该参数时，在后续实施中收口脚本或采用明确的关键文件命令，并核对必需测试已执行。Playwright单独job使用playwright.np.config.ts，不复用旧前端。
5. Docker仅构建并核对制品SHA，不login、不push。root main+Redis联调与fixture入口分别记录，不把cmd/server或测试故障入口当成真实main。
6. 每项记录head/base/实际checkout或test_merge SHA、argv、UTC、显式退出码、测试事件及run链接；包括缺PG/缺DSN的负向红灯证据。未来创建Draft PR后核实只运行检查、没有外部通知/上传或部署；此演练本轮NOT_RUN。

## 未运行 / 未核界限

- 未运行任何 go/npm/docker 命令、未安装工具、未触发远端 Actions、未做 gh 查询；“job 会红/绿”的判断均为静态推断。
- fuzzing.yml 引用的各 Fuzz 函数、Playwright 配置文件是否存在、puppeteer postinstall 行为、`scripts/ci-static-analysis.sh` 与 `scripts/fuzzing_test.sh` 内部逻辑未逐行核。
- codecov 是否配置了 token/项目、law-registry.com 是否真实可达、K8s manifests 与 secrets 是否存在，均未核。
- 本文件只描述“下一阶段组合”，不含实现；创建 PR 本身仍未发生。

## 主控复核更正

GLM Flash提供初稿；主代理核对并纠正了PG版本、REDIS变量、NP与跨方言测试的区别、空测试放行退路、未运行不等于已红，以及notify在PR上的条件。缓存提案由主代理复核，未将本代理未完成的缓存复核计为独立通过。CI矩阵和未来实际运行均NOT_RUN。

## 发布后只读观察

准备分支26355fc发布后，既有fuzzing.yml形成[push失败记录44508335288](https://github.com/VanTam-CN/law-oa-go/actions/runs/44508335288)。jobs接口返回0，commit check-runs接口返回0；`gh run view`仅提示likely workflow file issue，具体校验错误未取到。不得声称“本轮没有任何Actions记录”，也不得将其称作已运行的NP/PG测试失败。文件与主干相同，本轮不修、不手动触发；M2必须在创建PR前查明配置问题。实际记录见remote-observation.json。
