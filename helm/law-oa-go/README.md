# Helm Chart 状态

该 Chart 已弃用，不是当前生产发布入口。它保留在仓库中仅用于历史兼容，
不会被 CI/CD 生产作业调用，也不应通过默认 values 创建数据库或监控凭据。

当前生产部署必须使用 `k8s/` 下的 canonical PostgreSQL manifests，并按
`k8s/README.md` 先准备 PostgreSQL、Redis、Secret Manager、持久化存储和
入口路由。Elasticsearch 不属于默认部署；只有在独立评估搜索扩展时才显式
启用。生产 Secret 不得写入 values 文件或 Git。

Chart 默认只渲染应用兼容负载，不安装数据库、Redis、Elasticsearch、
Kibana、Jaeger 或监控组件。后端启动/就绪门禁使用 `/health/live` 和
`/health/ready`，发布验收以 `/health/ready` 返回 `ready: true` 为准。
调用方仍必须预置 ServiceAccount、ConfigMap、Secret、PVC 和外部服务；
该 Chart 不是开箱即用的独立部署方案。
