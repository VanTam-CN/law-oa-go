# 微服务架构设计

## 服务拆分策略

### 1. 网关服务 (Gateway Service)
```yaml
# k8s/gateway/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway-service
  namespace: law-oa
spec:
  replicas: 2
  selector:
    matchLabels:
      app: gateway-service
  template:
    metadata:
      labels:
        app: gateway-service
    spec:
      containers:
      - name: gateway-service
        image: gateway-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: USER_SERVICE_URL
          value: "http://user-service:8080"
        - name: CASE_SERVICE_URL
          value: "http://case-service:8080"
        - name: CLIENT_SERVICE_URL
          value: "http://client-service:8080"
        - name: AUTH_SERVICE_URL
          value: "http://auth-service:8080"
```

### 2. 用户服务 (User Service)
```yaml
# k8s/user-service/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
  namespace: law-oa
spec:
  replicas: 2
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service
        image: user-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "mysql-user-service"
        - name: DB_NAME
          value: "law_oa_users"
```

### 3. 认证服务 (Auth Service)
```yaml
# k8s/auth-service/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: law-oa
spec:
  replicas: 2
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: auth-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: REDIS_HOST
          value: "redis-auth-service"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-secrets
              key: jwt-secret
```

### 4. 案件服务 (Case Service)
```yaml
# k8s/case-service/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: case-service
  namespace: law-oa
spec:
  replicas: 3
  selector:
    matchLabels:
      app: case-service
  template:
    metadata:
      labels:
        app: case-service
    spec:
      containers:
      - name: case-service
        image: case-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "mysql-case-service"
        - name: DB_NAME
          value: "law_oa_cases"
        - name: ES_HOST
          value: "http://elasticsearch:9200"
```

### 5. 客户服务 (Client Service)
```yaml
# k8s/client-service/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: client-service
  namespace: law-oa
spec:
  replicas: 2
  selector:
    matchLabels:
      app: client-service
  template:
    metadata:
      labels:
        app: client-service
    spec:
      containers:
      - name: client-service
        image: client-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "mysql-client-service"
        - name: DB_NAME
          value: "law_oa_clients"
```

## 服务发现与通信

### Istio Service Mesh配置
```yaml
# k8s/istio/gateway.yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: law-oa-gateway
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "api.lawoa.com"
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: law-oa-vs
spec:
  hosts:
  - "api.lawoa.com"
  gateways:
  - law-oa-gateway
  http:
  - match:
    - uri:
        prefix: /api/v1/auth
    route:
    - destination:
        host: auth-service
        port:
          number: 8080
  - match:
    - uri:
        prefix: /api/v1/users
    route:
    - destination:
        host: user-service
        port:
          number: 8080
  - match:
    - uri:
        prefix: /api/v1/cases
    route:
    - destination:
        host: case-service
        port:
          number: 8080
  - match:
    - uri:
        prefix: /api/v1/clients
    route:
    - destination:
        host: client-service
        port:
          number: 8080
```

## 消息队列配置

### RabbitMQ配置
```yaml
# k8s/rabbitmq/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rabbitmq
  namespace: law-oa
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rabbitmq
  template:
    metadata:
      labels:
        app: rabbitmq
    spec:
      containers:
      - name: rabbitmq
        image: rabbitmq:3.12-management
        ports:
        - containerPort: 5672
        - containerPort: 15672
        env:
        - name: RABBITMQ_DEFAULT_USER
          value: "admin"
        - name: RABBITMQ_DEFAULT_PASS
          value: "password"
        volumeMounts:
        - name: rabbitmq-data
          mountPath: /var/lib/rabbitmq
      volumes:
      - name: rabbitmq-data
        persistentVolumeClaim:
          claimName: rabbitmq-pvc
```

## 数据库分片策略

### MySQL分片配置
```yaml
# k8s/mysql-sharded/deployment.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mysql-sharded
  namespace: law-oa
spec:
  serviceName: mysql-sharded
  replicas: 3
  selector:
    matchLabels:
      app: mysql-sharded
  template:
    metadata:
      labels:
        app: mysql-sharded
    spec:
      containers:
      - name: mysql
        image: mysql:8.0
        ports:
        - containerPort: 3306
        env:
        - name: MYSQL_ROOT_PASSWORD
          value: "password"
        - name: MYSQL_DATABASE
          value: "law_oa"
        volumeMounts:
        - name: mysql-data
          mountPath: /var/lib/mysql
  volumeClaimTemplates:
  - metadata:
      name: mysql-data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 10Gi
```

## 监控与日志聚合

### ELK Stack配置
```yaml
# k8s/monitoring/elasticsearch.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: law-oa
spec:
  serviceName: elasticsearch
  replicas: 3
  selector:
    matchLabels:
      app: elasticsearch
  template:
    metadata:
      labels:
        app: elasticsearch
    spec:
      containers:
      - name: elasticsearch
        image: docker.elastic.co/elasticsearch/elasticsearch:8.8.0
        ports:
        - containerPort: 9200
        - containerPort: 9300
        env:
        - name: discovery.type
          value: "single-node"
        - name: "ES_JAVA_OPTS"
          value: "-Xms512m -Xmx512m"
```

### Prometheus Operator配置
```yaml
# k8s/monitoring/prometheus-operator.yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: prometheus
  namespace: law-oa
spec:
  channel: beta
  name: prometheus
  source: community-operators
  sourceNamespace: openshift-marketplace
```