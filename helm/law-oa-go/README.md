# Helm Chart 状态

该 Chart 已弃用，不是当前生产发布入口。它保留在仓库中仅用于历史兼容，
不会被 CI/CD 生产作业调用，也不应通过默认 values 创建数据库或监控凭据。

当前生产部署必须使用 `k8s/` 下的 canonical PostgreSQL manifests，并按
`k8s/README.md` 先准备 PostgreSQL、Redis、Elasticsearch、Secret Manager、
持久化存储和入口路由。生产 Secret 不得写入 values 文件或 Git。
