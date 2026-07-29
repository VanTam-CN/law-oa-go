# Kubernetes 生产清单

本目录的生产入口是拆分后的 canonical manifests。根目录的
`deployment.yaml` 仅是弃用标记，不要把它当作应用部署清单。

## 部署前必须完成

1. 准备 PostgreSQL、Redis、Elasticsearch，以及 `law-oa` 命名空间中与
   Service 名称一致的网络入口。数据库不是本目录创建的 StatefulSet。
2. 把 `k8s/deployments/backend.yaml` 和 `k8s/deployments/frontend.yaml` 中的
   镜像替换为镜像仓库中的同一版本 immutable tag。迁移 initContainer 与后端容器
   必须使用同一个镜像版本；前端镜像必须使用同一提交构建出的 immutable tag。
3. 用 Secret Manager 或 External Secrets 创建名为 `law-oa-secrets` 的 Secret，至少填充
   `db-user`、`db-password`、`jwt-secret`、`app-secret`、`subject-data-key`、
   `onlyoffice-url`、`onlyoffice-secret` 和 `cors-allowed-origins`。仓库模板
   不进入发布包；缺少或为空会让后端启动门禁或 readiness 失败。
4. 为 `law-oa-uploads-pvc` 选择已演练备份/恢复的 StorageClass。上传文件与
   数据库同样属于业务证据，不能依赖临时卷。
5. 为 Elasticsearch、Redis、PostgreSQL 配置网络策略、TLS、备份和监控。

## 应用顺序

```bash
kubectl apply -f k8s/namespaces/law-oa.yaml
kubectl apply -f k8s/configmaps/law-oa-config.yaml
kubectl apply -f k8s/persistentvolumeclaims/uploads.yaml
kubectl apply -f k8s/services/backend.yaml
kubectl apply -f k8s/deployments/backend.yaml
kubectl apply -f k8s/deployments/frontend.yaml
kubectl rollout status deployment/law-oa-backend -n law-oa --timeout=10m
kubectl rollout status deployment/law-oa-frontend -n law-oa --timeout=10m
kubectl get pods,svc,pvc -n law-oa
```

## 发布门禁

部署成功不等于可以接案。必须同时确认前端 `/health`、后端 `/health/ready` 均可访问，且
P0 冲突档案覆盖、主体身份加密回填、三角色隔离和律所 PD-01 至 PD-07
决策物均已完成。详细放行条件见
`docs/利益冲突/conflict-p0-law-firm-trial-spec.md`。

入口层（Ingress、网关或 CDN）必须把正式前端域名路由到
`law-oa-frontend-service:80`。前端容器只代理同源 `/api/` 到
`law-oa-backend-service:8080`，不能把浏览器 API 地址写成 localhost。
