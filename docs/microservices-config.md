# 服务间通信配置

## gRPC服务定义

### 用户服务proto文件
```protobuf
// proto/user_service.proto
syntax = "proto3";

package user_service;
option go_package = "./pb";

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}

message GetUserRequest {
  uint64 user_id = 1;
}

message GetUserResponse {
  uint64 id = 1;
  string name = 2;
  string email = 3;
  string role = 4;
  string status = 5;
  string created_at = 6;
  string updated_at = 7;
}

message CreateUserRequest {
  string name = 1;
  string email = 2;
  string password = 3;
  string role = 4;
}

message CreateUserResponse {
  uint64 id = 1;
  string name = 2;
  string email = 3;
  string role = 4;
  string status = 5;
  string created_at = 6;
}

message UpdateUserRequest {
  uint64 user_id = 1;
  string name = 2;
  string email = 3;
  string role = 4;
  string status = 5;
}

message UpdateUserResponse {
  uint64 id = 1;
  string name = 2;
  string email = 3;
  string role = 4;
  string status = 5;
  string updated_at = 6;
}

message ListUsersRequest {
  int32 page = 1;
  int32 page_size = 2;
  string status = 3;
  string role = 4;
}

message ListUsersResponse {
  repeated GetUserResponse users = 1;
  int64 total = 2;
  int32 page = 3;
  int32 page_size = 4;
}
```

### 案件服务proto文件
```protobuf
// proto/case_service.proto
syntax = "proto3";

package case_service;
option go_package = "./pb";

service CaseService {
  rpc GetCase(GetCaseRequest) returns (GetCaseResponse);
  rpc CreateCase(CreateCaseRequest) returns (CreateCaseResponse);
  rpc UpdateCase(UpdateCaseRequest) returns (UpdateCaseResponse);
  rpc ListCases(ListCasesRequest) returns (ListCasesResponse);
  rpc GetCaseStats(GetCaseStatsRequest) returns (GetCaseStatsResponse);
}

message GetCaseRequest {
  uint64 case_id = 1;
}

message GetCaseResponse {
  uint64 id = 1;
  string case_no = 2;
  string case_name = 3;
  string case_type = 4;
  string status = 5;
  string priority = 6;
  uint64 client_id = 7;
  uint64 lawyer_id = 8;
  string created_at = 9;
  string updated_at = 10;
}

message CreateCaseRequest {
  string case_no = 1;
  string case_name = 2;
  string case_type = 3;
  string priority = 4;
  uint64 client_id = 5;
  uint64 lawyer_id = 6;
  string description = 7;
}

message CreateCaseResponse {
  uint64 id = 1;
  string case_no = 2;
  string case_name = 3;
  string case_type = 4;
  string status = 5;
  string priority = 6;
  uint64 client_id = 7;
  uint64 lawyer_id = 8;
  string created_at = 9;
}

message UpdateCaseRequest {
  uint64 case_id = 1;
  string case_name = 2;
  string case_type = 3;
  string status = 4;
  string priority = 5;
  uint64 lawyer_id = 6;
  string description = 7;
}

message UpdateCaseResponse {
  uint64 id = 1;
  string case_no = 2;
  string case_name = 3;
  string case_type = 4;
  string status = 5;
  string priority = 6;
  uint64 client_id = 7;
  uint64 lawyer_id = 8;
  string updated_at = 9;
}

message ListCasesRequest {
  int32 page = 1;
  int32 page_size = 2;
  string status = 3;
  string case_type = 4;
  string priority = 5;
  uint64 client_id = 6;
  uint64 lawyer_id = 7;
  string search = 8;
}

message ListCasesResponse {
  repeated GetCaseResponse cases = 1;
  int64 total = 2;
  int32 page = 3;
  int32 page_size = 4;
}

message GetCaseStatsRequest {}

message GetCaseStatsResponse {
  int64 total_cases = 1;
  int64 active_cases = 2;
  int64 pending_cases = 3;
  int64 closed_cases = 4;
  int64 urgent_cases = 5;
}
```

## 服务配置文件

### 用户服务配置
```yaml
# config/user-service.yaml
server:
  port: 8080
  mode: "production"

database:
  host: "mysql-user-service"
  port: 3306
  username: "root"
  password: "password"
  database: "law_oa_users"
  max_connections: 50
  max_idle_connections: 10
  connection_lifetime: "5m"

grpc:
  port: 9090
  max_message_size: 10485760  # 10MB

redis:
  host: "redis-user-service"
  port: 6379
  password: ""
  db: 0

jaeger:
  enabled: true
  service_name: "user-service"
  agent_host: "jaeger-agent"
  agent_port: 6831

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  port: 9091
```

### 网关服务配置
```yaml
# config/gateway-service.yaml
server:
  port: 8080
  mode: "production"

services:
  user_service:
    url: "http://user-service:8080"
    timeout: "5s"
    retries: 3
  
  case_service:
    url: "http://case-service:8080"
    timeout: "10s"
    retries: 3
  
  client_service:
    url: "http://client-service:8080"
    timeout: "5s"
    retries: 3
  
  auth_service:
    url: "http://auth-service:8080"
    timeout: "3s"
    retries: 3

rate_limit:
  enabled: true
  requests_per_minute: 100
  burst: 10

cors:
  allowed_origins:
    - "https://app.lawoa.com"
    - "https://admin.lawoa.com"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "Content-Type"
    - "Authorization"
    - "X-Request-ID"

circuit_breaker:
  enabled: true
  timeout: "30s"
  max_concurrent_requests: 100
  error_threshold_percentage: 50
  recovery_timeout: "30s"
```

## 部署脚本

### 构建和部署脚本
```bash
#!/bin/bash
# scripts/deploy-microservices.sh

set -e

echo "开始部署微服务架构..."

# 构建所有服务镜像
echo "构建服务镜像..."
docker-compose -f docker-compose.microservices.yml build

# 推送到镜像仓库
echo "推送镜像到仓库..."
docker-compose -f docker-compose.microservices.yml push

# 部署到Kubernetes
echo "部署到Kubernetes..."
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmaps/
kubectl apply -f k8s/secrets/
kubectl apply -f k8s/databases/
kubectl apply -f k8s/services/
kubectl apply -f k8s/deployments/

# 等待服务就绪
echo "等待服务就绪..."
kubectl wait --for=condition=available --timeout=300s deployment/gateway-service -n law-oa
kubectl wait --for=condition=available --timeout=300s deployment/user-service -n law-oa
kubectl wait --for=condition=available --timeout=300s deployment/case-service -n law-oa
kubectl wait --for=condition=available --timeout=300s deployment/client-service -n law-oa
kubectl wait --for=condition=available --timeout=300s deployment/auth-service -n law-oa

echo "部署完成！"
```