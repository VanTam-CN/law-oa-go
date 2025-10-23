{{/*
Law OA Go - Helm 模板助手
基于最新Helm 3最佳实践和生产环境优化策略
*/}}

{{/*
通用标签模板
*/}}
{{- define "law-oa-go.labels" -}}
helm.sh/chart: {{ include "law-oa-go.chart" . }}
{{ include "law-oa-go.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- if .Values.global.labels }}
{{ toYaml .Values.global.labels }}
{{- end }}
{{- if .Values.commonLabels }}
{{ toYaml .Values.commonLabels }}
{{- end }}
{{- end }}

{{/*
选择器标签模板
*/}}
{{- define "law-oa-go.selectorLabels" -}}
app.kubernetes.io/name: {{ include "law-oa-go.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Chart名称模板
*/}}
{{- define "law-oa-go.name" - }}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Chart全名模板
*/}}
{{- define "law-oa-go.fullname" - }}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Chart标识模板
*/}}
{{- define "law-oa-go.chart" - }}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
镜像名称模板
*/}}
{{- define "law-oa-go.image" -}}
{{- $registry := .Values.global.imageRegistry | default .Values.image.registry -}}
{{- if .Values.image.sha }}
{{- $registry }}/{{ .Values.image.repository }}:{{ .Values.image.tag }}@{{ .Values.image.sha }}
{{- else }}
{{- $registry }}/{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}
{{- end }}
{{- end }}

{{/*
服务账户名称模板
*/}}
{{- define "law-oa-go.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "law-oa-go.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
服务名称模板
*/}}
{{- define "law-oa-go.serviceName" -}}
{{- if .Values.service.name -}}
{{- .Values.service.name -}}
{{- else -}}
{{- include "law-oa-go.fullname" . -}}
{{- end -}}
{{- end }}

{{/*
配置映射名称模板
*/}}
{{- define "law-oa-go.configMapName" -}}
{{- if .Values.configMap.nameOverride -}}
{{- .Values.configMap.nameOverride -}}
{{- else -}}
{{- include "law-oa-go.fullname" . -}}
{{- end -}}
{{- end }}

{{/*
密钥名称模板
*/}}
{{- define "law-oa-go.secretName" -}}
{{- if .Values.secret.nameOverride -}}
{{- .Values.secret.nameOverride -}}
{{- else -}}
{{- include "law-oa-go.fullname" . -}}
{{- end -}}
{{- end }}

{{/*
持久卷声明名称模板
*/}}
{{- define "law-oa-go.pvcName" -}}
{{- if .Values.persistence.existingClaim -}}
{{- .Values.persistence.existingClaim -}}
{{- else -}}
{{- include "law-oa-go.fullname" . -}}
{{- end -}}
{{- end }}

{{/*
水平Pod自动缩放名称模板
*/}}
{{- define "law-oa-go.hpaName" -}}
{{- include "law-oa-go.fullname" . -}}-hpa
{{- end }}

{{/*
Pod中断预算名称模板
*/}}
{{- define "law-oa-go.pdbName" -}}
{{- include "law-oa-go.fullname" . -}}-pdb
{{- end }}

{{/*
网络策略名称模板
*/}}
{{- define "law-oa-go.networkPolicyName" -}}
{{- include "law-oa-go.fullname" . -}}-netpol
{{- end }}

{{/*
Pod安全策略名称模板
*/}}
{{- define "law-oa-go.pspName" -}}
{{- include "law-oa-go.fullname" . -}}-psp
{{- end }}

{{/*
Ingress名称模板
*/}}
{{- define "law-oa-go.ingressName" -}}
{{- if .Values.ingress.nameOverride -}}
{{- .Values.ingress.nameOverride -}}
{{- else -}}
{{- include "law-oa-go.fullname" . -}}
{{- end -}}
{{- end }}

{{/*
服务监控名称模板
*/}}
{{- define "law-oa-go.serviceMonitorName" -}}
{{- include "law-oa-go.fullname" . -}}
{{- end }}

{{/*
Pod监控名称模板
*/}}
{{- define "law-oa-go.podMonitorName" -}}
{{- include "law-oa-go.fullname" . -}}
{{- end }}

{{/*
健康检查端口名称模板
*/}}
{{- define "law-oa-go.healthPortName" -}}
{{- if .Values.service.healthPort -}}
{{- .Values.service.healthPort -}}
{{- else -}}
health
{{- end -}}
{{- end }}

{{/*
HTTP端口名称模板
*/}}
{{- define "law-oa-go.httpPortName" -}}
{{- if .Values.service.httpPort -}}
{{- .Values.service.httpPort -}}
{{- else -}}
http
{{- end -}}
{{- end }}

{{/*
gRPC端口名称模板
*/}}
{{- define "law-oa-go.grpcPortName" -}}
{{- if .Values.service.grpcPort -}}
{{- .Values.service.grpcPort -}}
{{- else -}}
grpc
{{- end -}}
{{- end }}

{{/*
指标端口名称模板
*/}}
{{- define "law-oa-go.metricsPortName" -}}
{{- if .Values.service.metricsPort -}}
{{- .Values.service.metricsPort -}}
{{- else -}}
metrics
{{- end -}}
{{- end }}

{{/*
数据库连接字符串模板
*/}}
{{- define "law-oa-go.databaseConnectionString" -}}
{{- if .Values.mysql.enabled -}}
{{- $host := printf "%s-mysql" .Release.Name -}}
{{- printf "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local" .Values.mysql.auth.username .Values.mysql.auth.password $host (.Values.mysql.primary.service.port | default 3306) .Values.mysql.auth.database -}}
{{- else -}}
{{- .Values.externalDatabase.connectionString -}}
{{- end -}}
{{- end }}

{{/*
Redis连接字符串模板
*/}}
{{- define "law-oa-go.redisConnectionString" -}}
{{- if .Values.redis.enabled -}}
{{- $host := printf "%s-redis-master" .Release.Name -}}
{{- printf "%s:%d" $host (.Values.redis.master.service.port | default 6379) -}}
{{- else -}}
{{- .Values.externalRedis.connectionString -}}
{{- end -}}
{{- end }}

{{/*
Elasticsearch连接字符串模板
*/}}
{{- define "law-oa-go.elasticsearchConnectionString" -}}
{{- if .Values.elasticsearch.enabled -}}
{{- $host := printf "%s-elasticsearch" .Release.Name -}}
{{- printf "http://%s:9200" $host -}}
{{- else -}}
{{- .Values.externalElasticsearch.connectionString -}}
{{- end -}}
{{- end }}

{{/*
Jaeger连接字符串模板
*/}}
{{- define "law-oa-go.jaegerConnectionString" -}}
{{- if .Values.jaeger.enabled -}}
{{- $host := printf "%s-jaeger-collector" .Release.Name -}}
{{- printf "http://%s:14268/api/traces" $host -}}
{{- else -}}
{{- .Values.externalJaeger.connectionString -}}
{{- end -}}
{{- end }}

{{/*
资源限制模板
*/}}
{{- define "law-oa-go.resources" -}}
limits:
  cpu: {{ .Values.resources.limits.cpu | quote }}
  memory: {{ .Values.resources.limits.memory | quote }}
  {{- if .Values.resources.limits.ephemeralStorage }}
  ephemeral-storage: {{ .Values.resources.limits.ephemeralStorage | quote }}
  {{- end }}
requests:
  cpu: {{ .Values.resources.requests.cpu | quote }}
  memory: {{ .Values.resources.requests.memory | quote }}
  {{- if .Values.resources.requests.ephemeralStorage }}
  ephemeral-storage: {{ .Values.resources.requests.ephemeralStorage | quote }}
  {{- end }}
{{- end }}

{{/*
节点选择器模板
*/}}
{{- define "law-oa-go.nodeSelector" -}}
{{- if .Values.nodeSelector }}
{{ toYaml .Values.nodeSelector }}
{{- else }}
kubernetes.io/os: linux
{{- end }}
{{- end }}

{{/*
容忍度模板
*/}}
{{- define "law-oa-go.tolerations" -}}
{{- if .Values.tolerations }}
{{ toYaml .Values.tolerations }}
{{- else }}
{{- if .Values.pod.tolerations }}
{{ toYaml .Values.pod.tolerations }}
{{- end }}
{{- end }}
{{- end }}

{{/*
亲和性模板
*/}}
{{- define "law-oa-go.affinity" -}}
{{- if .Values.affinity }}
{{ toYaml .Values.affinity }}
{{- else if .Values.pod.affinity }}
{{ toYaml .Values.pod.affinity }}
{{- end }}
{{- end }}

{{/*
Pod注释模板
*/}}
{{- define "law-oa-go.podAnnotations" -}}
{{- if .Values.pod.annotations }}
{{ toYaml .Values.pod.annotations }}
{{- end }}
{{- if .Values.podInfo.annotations }}
{{ toYaml .Values.podInfo.annotations }}
{{- end }}
{{- if .Values.commonAnnotations }}
{{ toYaml .Values.commonAnnotations }}
{{- end }}
{{- end }}

{{/*
Pod标签模板
*/}}
{{- define "law-oa-go.podLabels" -}}
{{ include "law-oa-go.selectorLabels" . }}
{{- if .Values.pod.labels }}
{{ toYaml .Values.pod.labels }}
{{- end }}
{{- if .Values.commonLabels }}
{{ toYaml .Values.commonLabels }}
{{- end }}
{{- end }}

{{/*
服务账户模板
*/}}
{{- define "law-oa-go.serviceAccount" -}}
{{- if .Values.serviceAccount.create -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "law-oa-go.serviceAccountName" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "law-oa-go.labels" . | nindent 4 }}
  {{- if .Values.serviceAccount.annotations }}
  annotations:
    {{- toYaml .Values.serviceAccount.annotations | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
RBAC模板
*/}}
{{- define "law-oa-go.rbac" -}}
{{- if .Values.rbac.create -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ include "law-oa-go.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "law-oa-go.labels" . | nindent 4 }}
rules:
  {{- toYaml .Values.rbac.rules | nindent 2 }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ include "law-oa-go.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "law-oa-go.labels" . | nindent 4 }}
subjects:
  - kind: ServiceAccount
    name: {{ include "law-oa-go.serviceAccountName" . }}
    namespace: {{ .Release.Namespace }}
roleRef:
  kind: Role
  name: {{ include "law-oa-go.fullname" . }}
  apiGroup: rbac.authorization.k8s.io
{{- end }}
{{- end }}

{{/*
环境变量模板
*/}}
{{- define "law-oa-go.env" -}}
- name: GIN_MODE
  value: {{ .Values.env.ginMode | default "release" | quote }}
- name: ENVIRONMENT
  value: {{ .Values.global.environment | default "production" | quote }}
- name: TZ
  value: {{ .Values.env.timezone | default "Asia/Shanghai" | quote }}
- name: LOG_LEVEL
  value: {{ .Values.env.logLevel | default "info" | quote }}
- name: VERSION
  value: {{ .Chart.AppVersion | quote }}
- name: METRICS_ENABLED
  value: {{ .Values.env.metricsEnabled | default "true" | quote }}
- name: TRACING_ENABLED
  value: {{ .Values.env.tracingEnabled | default "true" | quote }}
- name: HEALTH_CHECK_ENABLED
  value: {{ .Values.env.healthCheckEnabled | default "true" | quote }}
- name: POD_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: POD_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: POD_IP
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
- name: NODE_NAME
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
{{- if .Values.mysql.enabled }}
- name: DB_HOST
  value: {{ printf "%s-mysql" .Release.Name | quote }}
- name: DB_PORT
  value: {{ .Values.mysql.primary.service.port | default 3306 | quote }}
- name: DB_USER
  value: {{ .Values.mysql.auth.username | quote }}
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ printf "%s-mysql" .Release.Name }}
      key: mysql-password
- name: DB_NAME
  value: {{ .Values.mysql.auth.database | quote }}
{{- end }}
{{- if .Values.redis.enabled }}
- name: REDIS_HOST
  value: {{ printf "%s-redis-master" .Release.Name | quote }}
- name: REDIS_PORT
  value: {{ .Values.redis.master.service.port | default 6379 | quote }}
{{- if .Values.redis.auth.enabled }}
- name: REDIS_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ printf "%s-redis" .Release.Name }}
      key: redis-password
{{- end }}
{{- end }}
{{- if .Values.elasticsearch.enabled }}
- name: ES_HOST
  value: {{ include "law-oa-go.elasticsearchConnectionString" . }}
{{- end }}
{{- if .Values.jaeger.enabled }}
- name: JAEGER_ENDPOINT
  value: {{ include "law-oa-go.jaegerConnectionString" . }}
{{- end }}
{{- range .Values.env }}
- name: {{ .name }}
  {{- if .value }}
  value: {{ .value | quote }}
  {{- end }}
  {{- if .valueFrom }}
  valueFrom:
    {{- toYaml .valueFrom | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
卷挂载模板
*/}}
{{- define "law-oa-go.volumeMounts" -}}
{{- if .Values.persistence.enabled }}
- name: uploads
  mountPath: /app/uploads
  subPath: uploads
- name: logs
  mountPath: /app/logs
  subPath: logs
{{- end }}
- name: config
  mountPath: /app/config
  readOnly: true
- name: tmp
  mountPath: /tmp
{{- if .Values.sidecars.logCollector.enabled }}
- name: fluent-bit-config
  mountPath: /fluent-bit/etc
  readOnly: true
{{- end }}
{{- end }}

{{/*
卷模板
*/}}
{{- define "law-oa-go.volumes" -}}
- name: config
  configMap:
    name: {{ include "law-oa-go.configMapName" . }}
{{- if .Values.persistence.enabled }}
- name: uploads
  persistentVolumeClaim:
    claimName: {{ include "law-oa-go.fullname" . }}-uploads
- name: logs
  persistentVolumeClaim:
    claimName: {{ include "law-oa-go.fullname" . }}-logs
{{- end }}
- name: tmp
  emptyDir:
    sizeLimit: {{ .Values.persistence.tmpSize | default "100Mi" }}
{{- if .Values.sidecars.logCollector.enabled }}
- name: fluent-bit-config
  configMap:
    name: {{ include "law-oa-go.fullname" . }}-fluent-bit
{{- end }}
{{- end }}

{{/*
初始化容器模板
*/}}
{{- define "law-oa-go.initContainers" -}}
{{- if .Values.initContainers.waitForDatabase.enabled }}
- name: wait-for-database
  image: {{ .Values.initContainers.waitForDatabase.image | default "busybox:1.36" }}
  imagePullPolicy: IfNotPresent
  command:
    - sh
    - -c
    - |
      echo "Waiting for database to be ready..."
      {{- if .Values.mysql.enabled }}
      until nc -z {{ printf "%s-mysql" .Release.Name }} {{ .Values.mysql.primary.service.port | default 3306 }}; do
        echo "Database not ready, waiting..."
        sleep 2
      done
      {{- end }}
      echo "Database is ready!"
  {{- with .Values.initContainers.waitForDatabase.resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
{{- if .Values.initContainers.waitForRedis.enabled }}
- name: wait-for-redis
  image: {{ .Values.initContainers.waitForRedis.image | default "busybox:1.36" }}
  imagePullPolicy: IfNotPresent
  command:
    - sh
    - -c
    - |
      echo "Waiting for Redis to be ready..."
      {{- if .Values.redis.enabled }}
      until nc -z {{ printf "%s-redis-master" .Release.Name }} {{ .Values.redis.master.service.port | default 6379 }}; do
        echo "Redis not ready, waiting..."
        sleep 2
      done
      {{- end }}
      echo "Redis is ready!"
  {{- with .Values.initContainers.waitForRedis.resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
{{- if .Values.initContainers.waitForElasticsearch.enabled }}
- name: wait-for-elasticsearch
  image: {{ .Values.initContainers.waitForElasticsearch.image | default "busybox:1.36" }}
  imagePullPolicy: IfNotPresent
  command:
    - sh
    - -c
    - |
      echo "Waiting for Elasticsearch to be ready..."
      {{- if .Values.elasticsearch.enabled }}
      until curl -f {{ include "law-oa-go.elasticsearchConnectionString" . }}/_cluster/health; do
        echo "Elasticsearch not ready, waiting..."
        sleep 5
      done
      {{- end }}
      echo "Elasticsearch is ready!"
  {{- with .Values.initContainers.waitForElasticsearch.resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
{{- if .Values.initContainers.migration.enabled }}
- name: database-migration
  image: {{ include "law-oa-go.image" . }}
  imagePullPolicy: {{ .Values.image.pullPolicy }}
  command: {{ .Values.initContainers.migration.command }}
  envFrom:
    - configMapRef:
        name: {{ include "law-oa-go.configMapName" . }}
    - secretRef:
        name: {{ include "law-oa-go.secretName" . }}
  {{- with .Values.initContainers.migration.resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Sidecar容器模板
*/}}
{{- define "law-oa-go.sidecars" -}}
{{- if .Values.sidecars.logCollector.enabled }}
- name: log-collector
  image: {{ .Values.sidecars.logCollector.image | default "fluent/fluent-bit:2.2" }}
  imagePullPolicy: IfNotPresent
  env:
    - name: FLUENT_BIT_SERVICE_NAME
      value: {{ include "law-oa-go.fullname" . }}
    - name: FLUENT_BIT_NAMESPACE
      valueFrom:
        fieldRef:
          fieldPath: metadata.namespace
    - name: FLUENT_BIT_POD_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.name
  volumeMounts:
    - name: logs
      mountPath: /app/logs
      readOnly: true
    - name: fluent-bit-config
      mountPath: /fluent-bit/etc
      readOnly: true
  {{- with .Values.sidecars.logCollector.resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}