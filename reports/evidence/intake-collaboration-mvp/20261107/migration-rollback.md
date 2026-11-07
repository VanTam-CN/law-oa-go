# 迁移与回退边界

本轮仅操作 `/private/tmp/lawoa-intake-collaboration-runtime` 下的隔离 PostgreSQL 16 集群、虚构数据和专用测试库。未连接正式数据库，未执行真实资料迁移或部署。

## 已执行

- `000082`：提交快照指针、轮次唯一性、不可变快照及决定保护。
- `000083`：客户创建幂等记录及唯一约束。
- `000084`：待办来源、完成方式、幂等任务键及契约约束。
- 新空库 bootstrap，`84 → 83 → 82` down / `82 → 83 → 84` up 两轮，重复 bootstrap 和目录查询，共 16 条执行结果均为 exit 0。详见 [schema/results.json](schema/results.json) 和 [schema/README.md](schema/README.md)。
- bootstrap 首次曾遇 PostgreSQL prepared statement 不允许一条语句含多命令；已拆分为独立 DDL，失败记录与后续通过日志分开保留。
- 真实浏览器首提暴露初始化缺少000080的claim字段，随后v16补齐000080/000081契约：可空VARCHAR(64) claim、单列非NULL接案唯一槽、实际tracking.creation_status。空随机schema调用真实bootstrap创建119表，已有v15形状经新连接升级后保留原数据并完成真实HTTP首提；无手工补表。具体正负约束和首红见 [bootstrap/HANDOFF.md](bootstrap/HANDOFF.md)。
- 浏览器专用虚构库以新候选正常执行bootstrap升级v16，exit0，见 [schema/bootstrap-final.log](schema/bootstrap-final.log)。两个R5完整主链随后通过；这仍不是正式环境迁移。

空库可逆只证明 schema 演练。已有新协议记录时，`000082` down 的保护会拒绝删除协议所需结构；不能把此拒绝当作需要绕过的障碍。新快照与审批决定有不可变约束，真实数据升级后的回退不等于直接运行 down。

## 候选回退

本轮未部署，代码可以在独立分支切回基线 `a806831a7cef839c2b971cd4ab260cce77f42b8b` 供比较；不要切换用户的脏主目录或用 reset 覆盖其修改。

若以后获准在有业务数据的环境升级，应先建立独立备份和恢复证据。出现问题时停止新协议写入，保留新轮次、决定、关联与审计记录，使用兼容读取版本或前向修复。不能让旧程序继续对新协议记录写入，也不能删新轮次、清指针或修改历史快照来取得表面上的回退成功。

正式环境备份恢复、异机灾备、数据升级和回退均为 **NOT_RUN**；本文件不构成执行这些操作的授权。`production_ready=false`。
