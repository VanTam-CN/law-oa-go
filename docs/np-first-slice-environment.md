# NP v0.4 首切片本地隔离环境

工作副本：`output/np-v04-first-slice`（分支 `codex/np-v04-first-slice`，冻结 `bcd1b899`）。本文件只覆盖本地隔离基础准备；Go 基线测试和业务契约由主代理负责。

## 文件与用途

- `scripts/np-first-slice/compose.yml`：PG 16.14 + Redis 隔离实例，项目名 `np-first-slice`，数据卷 `np-first-slice_db-data` / `np-first-slice_redis-data` 全新独立。
- `scripts/np-first-slice/init-db.sh`：启动 db/redis -> 空库核对 -> 构建 `backend/cmd/migrate` -> bootstrap -> 按序追加 000082-000085 up；首失败即停。
- `scripts/np-first-slice/probe-db.sh`：host Go 测试前的只读探针，核对 127.0.0.1:18790 确为本隔离 PG。
- `scripts/np-first-slice/.env`：私有凭据（仅全新虚构值，0600，不入库）。

## 端口与隔离边界

| 端口 | 用途 | 绑定 |
|---|---|---|
| 18790 | host Go / 实际 service 联调 PG 探针 | 127.0.0.1（仅回环，不向 LAN 开放） |
| 18791 | host 联调 Redis 探针 | 127.0.0.1 |
| 18789 | `np_local_check` synthetic runner（主代理另写，复用实际 router，不 AutoMigrate） | 127.0.0.1 |
| 3003 | 前端保留（本切片不改 frontend） | 127.0.0.1 |

端口调整理由（2026-11-19 明确授权）：原计划 PG/Redis 无宿主映射；为 host Go 基线与实际 service 联调提供明确的临时回环探针，绑定 127.0.0.1 限定本机，与既有服务不重叠（启动前已核对 18789/18790/18791 空闲）。

## 初始化路径（对齐冻结包 scripts/fresh-db-checklist.sh）

bootstrap 使用候选内 `backend/cmd/migrate -command bootstrap`（PG 专用、事务式，`schema_bootstrap_state` 记录 `postgres-mvp-2026-08-27-v15`），随后仅按序执行四个追加 up：`000082_intake_collaboration_mvp`、`000083_approval_supplement_replies`、`000084_matter_collaboration`、`000085_case_task_change_audit`。明确跳过全量历史迁移；`init-db.sh` 在非全新库上直接失败。

## 用法

仅本机 Docker Desktop 的 `desktop-linux` context；脚本拒绝 DOCKER_HOST/DOCKER_CONTEXT 覆盖，避免误连远程。

私有 `.env` 必需变量：`POSTGRES_DB=np_first_slice_dev`、`POSTGRES_USER`、`POSTGRES_PASSWORD`、`REDIS_PASSWORD`、`NP_FIXTURE_PASSWORD`。密码仅为本轮隔离合成夹具专用，不提交；已有夹具重启时使用原虚构密码，不自动重置账号。

~~~sh
cd scripts/np-first-slice
sh init-db.sh        # 启动并初始化；重复执行会在非空库上失败
sh probe-db.sh       # host Go 测试前核对（断言 PG 16.14 且库为 np_first_slice_dev）
# 初始化只执行一次；之后显式添加本切片表，不运行混合历史 migrate up：
docker compose --env-file .env exec -T db sh -c 'exec psql -v ON_ERROR_STOP=1 --single-transaction -U "$POSTGRES_USER" -d np_first_slice_dev' < ../../backend/migrations/000086_narrow_pilot.up.sql
sh run-local-check.sh  # 前台启动回环 18789，普通生产构建不包含该程序
~~~

另开终端从项目根运行 `NP_FIXTURE_PASSWORD=<本轮虚构密码> /usr/bin/python3 scripts/np-first-slice/seed-browser.py`，通过实际原生 API 建两版虚构材料；再在 `frontend` 运行 `NP_FIXTURE_PASSWORD=<同一虚构密码> VITE_API_PROXY_TARGET=http://127.0.0.1:18789 npx playwright test --config playwright.np.config.ts`。

运行前后可用 `compare-browser-db.py before` / `after` 校验同一轮四条浏览器用例对 PG 的实际影响；两次之间不得插入其他针对该夹具的写操作。敏感 trace 仅本地保留且被 Git 忽略。

host runner 联调环境（主代理使用）：127.0.0.1:18789，复用实际 router 且不 AutoMigrate；DB_HOST=127.0.0.1 DB_PORT=18790，DB_USERNAME/DB_PASSWORD/DB_DATABASE、REDIS_PASSWORD 取 `.env` 对应值，REDIS_HOST=127.0.0.1 REDIS_PORT=18791。root `main.go` 默认监听所有地址且 development 会 AutoMigrate，不作为本地 runner 入口。

## 已知边界

- root `backend/main.go` 在 ENVIRONMENT=development 启动时会对 RBAC、集成、认证会话、财务等模型执行 GORM AutoMigrate（`backend/main.go` 约 205 行起）。这是既有代码行为，本任务只允许改脚本/文档，无法关闭；若首切片要求 schema 严格等于 bootstrap+000085，主代理启动 service 前需决策。当前脚本只在 init 阶段建库，不启动 service。
- migrate 配置加载顺序为进程环境变量优先，其次进程 CWD 的 `.env`，再次 `config/config.yaml` 与 `/etc/law-oa/config.yaml`；已核对目标副本无 backend 根 `.env`、仓库根 `.env`，`/etc/law-oa` 不存在，连接全部来自脚本注入的环境变量。

## 验证记录

2026-11-19 14:36-14:38（Asia/Shanghai）：`sh init-db.sh` 实际执行成功。postgres:16.14 镜像按 manifest 拉取；`np-first-slice-db-1`、`np-first-slice-redis-1` 创建并健康；全新卷 `np-first-slice_db-data`、`np-first-slice_redis-data`；空库核对通过后 bootstrap 完成（`schema_bootstrap_state` = `postgres-mvp-2026-08-27-v15`），随后按序应用 000082/000083/000084/000085 up，`INIT_OK public_tables=122`。最终库名为 `np_first_slice_dev`（初始化时的实际名；主审确认统一保留此名，不重建不重命名）。2026-11-19 14:40 只读复核：库列表仅 `np_first_slice_dev` 与系统库 `postgres`，bootstrap 版本与 122 表保持。此后不再执行 init/建库；容器保留运行供集成核验。

首失败保留：主审侧 `probe-db.sh` 曾报 database `np_first_slice` 不存在——原因是探针断言短暂指向未创建的固定名（脚本与实际库名不一致，库本身未受影响）；已改回 `np_first_slice_dev` 断言。

## 已有隔离库的后续复核

本轮验证完成后停止专用容器并保留数据卷。重新检查已有数据时，在已核对的本机 context 执行 `docker compose --env-file .env up -d db redis`，再运行 `probe-db.sh` 和 `run-local-check.sh`；不要重新执行空库 init 或已执行的 000086。共用保存和来源回执由 PG 保留，重启只丢弃未保存整理。

本次独立执行显式追加 000086 后 public 表数为 124；原 bootstrap/四个追加阶段的 122 是历史基线数。检查入口按原生 API 补入虚构材料两版本，数据库前后取证见 `np-evidence/browser-pg-comparison.json`。Redis 已准备，但检查入口实际使用持久 PG 认证会话与无 Redis 的路由构造，本报告不声称完成 Redis 功能回归。
