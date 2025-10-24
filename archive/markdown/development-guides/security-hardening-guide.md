# 律所OA系统安全加固实施指南

## 概述

本指南基于安全评估报告的发现，提供详细的安全加固实施步骤，按优先级和复杂性分类。

---

## 紧急修复 (24小时内完成)

### 1. 修复JWT密钥硬编码问题

**问题**: JWT密钥在代码中硬编码存在严重安全风险

**修复步骤**:

#### 1.1 更新环境变量配置
```bash
# 在 .env 文件中添加
JWT_SECRET=your-super-secure-random-key-at-least-32-characters-long
```

#### 1.2 修改JWT中间件
**文件**: `internal/middleware/jwt.go`

```go
// 移除这行
var jwtSecret []byte

// 修改 InitJWT 函数
func InitJWT(cfg *config.Config) {
    if cfg.JWT.Secret == "" {
        log.Fatal("JWT_SECRET environment variable is required")
    }
    jwtSecret = []byte(cfg.JWT.Secret)
}
```

#### 1.3 生成安全的JWT密钥
```bash
# 生成32字节的随机密钥
openssl rand -base64 32
```

### 2. 修复配置文件安全问题

#### 2.1 加密敏感配置
**文件**: `internal/config/config.go`

```go
// 添加配置解密功能
func (c *Config) DecryptSensitiveFields() error {
    if c.Database.Password != "" {
        decrypted, err := decryptConfigValue(c.Database.Password)
        if err != nil {
            return err
        }
        c.Database.Password = decrypted
    }
    // 其他敏感字段...
    return nil
}
```

#### 2.2 创建配置加密工具
```bash
# 创建配置加密脚本
cat > scripts/encrypt-config.sh << 'EOF'
#!/bin/bash
if [ -z "$CONFIG_KEY" ]; then
    echo "CONFIG_KEY environment variable is required"
    exit 1
fi
# 使用AES-256-GCM加密配置文件
openssl enc -aes-256-gcm -salt -in .env -out .env.enc -k $CONFIG_KEY
EOF
chmod +x scripts/encrypt-config.sh
```

---

## 高优先级修复 (1周内完成)

### 3. 实现完整SQL注入防护

#### 3.1 创建输入验证中间件
**文件**: `internal/middleware/validation.go`

```go
package middleware

import (
    "regexp"
    "strings"
    "github.com/gin-gonic/gin"
)

// SQLInjectionProtection SQL注入防护中间件
func SQLInjectionProtection() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 检查查询参数
        for key, values := range c.Request.URL.Query() {
            for _, value := range values {
                if containsSQLInjection(value) {
                    c.JSON(400, gin.H{"error": "Invalid input detected"})
                    c.Abort()
                    return
                }
            }
        }

        // 检查POST数据
        if c.Request.Method == "POST" || c.Request.Method == "PUT" {
            var body map[string]interface{}
            if err := c.ShouldBindJSON(&body); err == nil {
                if containsSQLInjectionInMap(body) {
                    c.JSON(400, gin.H{"error": "Invalid input detected"})
                    c.Abort()
                    return
                }
            }
        }

        c.Next()
    }
}

// SQL注入检测模式
var sqlInjectionPatterns = []string{
    `(?i)(union\s+select|insert\s+into|delete\s+from|update\s+set|drop\s+table)`,
    `(?i)(or\s+1\s*=\s*1|and\s+1\s*=\s*1)`,
    `(?i)(;|--|\/\*|\*\/|@@|@)`,
    `(?i)(exec\s*\(|execute\s*\()`,
    `(?i)(xp_cmdshell|sp_executesql)`,
}

func containsSQLInjection(input string) bool {
    for _, pattern := range sqlInjectionPatterns {
        matched, _ := regexp.MatchString(pattern, input)
        if matched {
            return true
        }
    }
    return false
}

func containsSQLInjectionInMap(data map[string]interface{}) bool {
    for _, value := range data {
        if str, ok := value.(string); ok {
            if containsSQLInjection(str) {
                return true
            }
        }
    }
    return false
}
```

#### 3.2 在路由中应用验证中间件
**文件**: `internal/router/router.go`

```go
// 在路由初始化中添加
app.Use(middleware.SQLInjectionProtection())
app.Use(middleware.XSSProtection())
```

### 4. 增强XSS防护

#### 4.1 创建XSS防护中间件
**文件**: `internal/middleware/xss.go`

```go
package middleware

import (
    "html"
    "regexp"
    "github.com/gin-gonic/gin"
)

// XSSProtection XSS防护中间件
func XSSProtection() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 设置XSS防护头
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")

        // 清理输出数据
        c.Next()

        // 对响应数据进行XSS过滤
        if writer, ok := c.Writer.(*responseWriter); ok {
            writer.filterXSS()
        }
    }
}
```

### 5. 改进密码策略执行

#### 5.1 创建密码验证服务
**文件**: `internal/services/password_service.go`

```go
package services

import (
    "errors"
    "regexp"
    "unicode"
)

type PasswordService struct {
    commonPasswords []string
}

func NewPasswordService() *PasswordService {
    return &PasswordService{
        commonPasswords: []string{
            "password", "123456", "123456789", "qwerty", "abc123",
            "password123", "admin", "letmein", "welcome", "monkey",
        },
    }
}

func (s *PasswordService) ValidatePassword(password string, userInfo *UserInfo) error {
    // 长度检查
    if len(password) < 8 {
        return errors.New("密码长度至少8位")
    }
    if len(password) > 128 {
        return errors.New("密码长度不能超过128位")
    }

    // 复杂度检查
    var hasUpper, hasLower, hasNumber, hasSpecial bool
    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }

    if !hasUpper {
        return errors.New("密码必须包含大写字母")
    }
    if !hasLower {
        return errors.New("密码必须包含小写字母")
    }
    if !hasNumber {
        return errors.New("密码必须包含数字")
    }
    if !hasSpecial {
        return errors.New("密码必须包含特殊字符")
    }

    // 常见密码检查
    for _, common := range s.commonPasswords {
        if strings.ToLower(password) == common {
            return errors.New("不能使用常见密码")
        }
    }

    // 个人信息检查
    if userInfo != nil {
        if strings.Contains(strings.ToLower(password), strings.ToLower(userInfo.Username)) {
            return errors.New("密码不能包含用户名")
        }
        if strings.Contains(strings.ToLower(password), strings.ToLower(userInfo.Email)) {
            return errors.New("密码不能包含邮箱")
        }
    }

    // 模式检查
    if s.containsSequentialPatterns(password) {
        return errors.New("密码不能包含连续字符模式")
    }

    return nil
}

func (s *PasswordService) containsSequentialPatterns(password string) bool {
    // 检查键盘序列
    sequences := []string{
        "qwerty", "asdfgh", "zxcvbn", "123456", "234567", "345678",
        "qazwsx", "edcrfv", "tgbnhy", "yhnujm", "ujmik", "ikolp",
    }

    lowerPassword := strings.ToLower(password)
    for _, seq := range sequences {
        if strings.Contains(lowerPassword, seq) {
            return true
        }
        // 反向序列
        reverse := reverseString(seq)
        if strings.Contains(lowerPassword, reverse) {
            return true
        }
    }

    return false
}
```

---

## 中优先级修复 (2-4周内完成)

### 6. 完善令牌管理

#### 6.1 调整令牌有效期
**文件**: `internal/middleware/jwt.go`

```go
// 修改GenerateToken函数
func GenerateToken(userID uint, username, role string) (string, time.Time, error) {
    expiresAt := time.Now().Add(time.Hour * 2) // 改为2小时过期
    claims := JWTClaims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            ExpiresAt: jwt.NewNumericDate(expiresAt),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtSecret)
    return tokenString, expiresAt, err
}
```

#### 6.2 实现令牌刷新机制
**文件**: `internal/handlers/auth_handler.go`

```go
// RefreshToken 刷新访问令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    var req RefreshTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        _ = c.Error(errors.NewValidationError("invalid_request", "Invalid request format", err.Error(), ""))
        return
    }

    // 验证刷新令牌
    claims, err := validateRefreshToken(req.RefreshToken)
    if err != nil {
        _ = c.Error(errors.NewAuthorizationError("invalid_token", "Invalid refresh token", "refresh_token", "invalid"))
        return
    }

    // 生成新的访问令牌
    token, expiresAt, err := middleware.GenerateToken(claims.UserID, claims.Username, claims.Role)
    if err != nil {
        _ = c.Error(errors.NewInternalError("token_generation", "Failed to generate token", err))
        return
    }

    response := RefreshTokenResponse{
        Token:     token,
        ExpiresAt: expiresAt,
    }

    common.APISuccess(c, response)
}
```

### 7. 增强文件上传安全

#### 7.1 创建文件验证服务
**文件**: `internal/services/file_security_service.go`

```go
package services

import (
    "bytes"
    "errors"
    "mime"
    "path/filepath"
    "strings"
)

type FileSecurityService struct {
    allowedTypes    map[string]bool
    maxFileSize     int64
    virusScanner    VirusScanner
}

type VirusScanner interface {
    ScanFile(data []byte) error
}

func NewFileSecurityService() *FileSecurityService {
    allowedTypes := map[string]bool{
        ".jpg":  true, ".jpeg": true, ".png":  true, ".gif":  true,
        ".pdf":  true, ".doc":  true, ".docx": true, ".xls":  true,
        ".xlsx": true, ".txt":  true, ".rtf":  true,
    }

    return &FileSecurityService{
        allowedTypes: allowedTypes,
        maxFileSize:  50 * 1024 * 1024, // 50MB
    }
}

func (s *FileSecurityService) ValidateFile(filename string, data []byte) error {
    // 文件大小检查
    if int64(len(data)) > s.maxFileSize {
        return errors.New("文件大小超过限制")
    }

    // 文件扩展名检查
    ext := strings.ToLower(filepath.Ext(filename))
    if !s.allowedTypes[ext] {
        return errors.New("不支持的文件类型")
    }

    // MIME类型检查
    mimeType := mime.TypeByExtension(ext)
    if mimeType == "" {
        return errors.New("无法识别的文件类型")
    }

    // 文件头验证
    if !s.validateFileHeader(data, ext) {
        return errors.New("文件头与扩展名不匹配")
    }

    // 病毒扫描
    if s.virusScanner != nil {
        if err := s.virusScanner.ScanFile(data); err != nil {
            return errors.New("文件安全检查失败")
        }
    }

    return nil
}

func (s *FileSecurityService) validateFileHeader(data []byte, ext string) bool {
    if len(data) < 8 {
        return false
    }

    switch ext {
    case ".jpg", ".jpeg":
        return bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF})
    case ".png":
        return bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47})
    case ".pdf":
        return bytes.HasPrefix(data, []byte{0x25, 0x50, 0x44, 0x46})
    default:
        return true // 其他类型暂时允许
    }
}
```

### 8. 实现API速率限制增强

#### 8.1 创建高级速率限制器
**文件**: `internal/middleware/advanced_rate_limiter.go`

```go
package middleware

import (
    "context"
    "fmt"
    "net/http"
    "strconv"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
)

type AdvancedRateLimiter struct {
    redis    *redis.Client
    limits   map[string]RateLimitConfig
    mu       sync.RWMutex
}

type RateLimitConfig struct {
    Requests int           `json:"requests"`
    Window   time.Duration `json:"window"`
    Burst    int           `json:"burst"`
}

func NewAdvancedRateLimiter(redis *redis.Client) *AdvancedRateLimiter {
    return &AdvancedRateLimiter{
        redis:  redis,
        limits: make(map[string]RateLimitConfig),
    }
}

func (arl *AdvancedRateLimiter) SetLimit(endpoint string, config RateLimitConfig) {
    arl.mu.Lock()
    defer arl.mu.Unlock()
    arl.limits[endpoint] = config
}

func (arl *AdvancedRateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        endpoint := c.FullPath()
        userID := arl.getUserID(c)
        ip := c.ClientIP()

        // 获取限制配置
        config := arl.getLimitConfig(endpoint)

        // 检查用户级别限制
        if userID != "" {
            if !arl.checkLimit(fmt.Sprintf("user:%s:%s", userID, endpoint), config) {
                arl.handleRateLimit(c)
                return
            }
        }

        // 检查IP级别限制
        ipConfig := RateLimitConfig{
            Requests: config.Requests * 2, // IP限制更宽松
            Window:   config.Window,
            Burst:    config.Burst * 2,
        }

        if !arl.checkLimit(fmt.Sprintf("ip:%s:%s", ip, endpoint), ipConfig) {
            arl.handleRateLimit(c)
            return
        }

        c.Next()
    }
}

func (arl *AdvancedRateLimiter) checkLimit(key string, config RateLimitConfig) bool {
    ctx := context.Background()
    now := time.Now()
    window := now.Truncate(config.Window)

    // 使用滑动窗口算法
    pipe := arl.redis.Pipeline()

    // 移除过期的记录
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", window.Unix()))

    // 添加当前请求
    pipe.ZAdd(ctx, key, redis.Z{
        Score:  float64(now.Unix()),
        Member: now.UnixNano(),
    })

    // 获取当前窗口内的请求数
    pipe.ZCard(ctx, key)

    // 设置过期时间
    pipe.Expire(ctx, key, config.Window*2)

    results, err := pipe.Exec(ctx)
    if err != nil {
        return true // 出错时允许通过
    }

    count := results[2].(*redis.IntCmd).Val()
    return count <= config.Requests
}

func (arl *AdvancedRateLimiter) handleRateLimit(c *gin.Context) {
    c.JSON(http.StatusTooManyRequests, gin.H{
        "error": "请求过于频繁，请稍后再试",
        "code":  "RATE_LIMIT_EXCEEDED",
    })
    c.Abort()
}
```

---

## 长期改进项目 (1-3个月)

### 9. 实现多因素认证(MFA)

#### 9.1 创建MFA服务
**文件**: `internal/services/mfa_service.go`

```go
package services

import (
    "crypto/rand"
    "encoding/base32"
    "fmt"
    "time"

    "github.com/pquerna/otp/totp"
)

type MFAService struct {
    issuer string
}

func NewMFAService(issuer string) *MFAService {
    return &MFAService{
        issuer: issuer,
    }
}

func (s *MFAService) GenerateSecret(username string) (string, string, error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      s.issuer,
        AccountName: username,
        SecretSize:  32,
    })
    if err != nil {
        return "", "", err
    }

    return key.Secret(), key.URL(), nil
}

func (s *MFAService) ValidateCode(secret, code string) bool {
    return totp.Validate(code, secret)
}

func (s *MFAService) GenerateBackupCodes() []string {
    codes := make([]string, 10)
    for i := 0; i < 10; i++ {
        codes[i] = s.generateRandomCode()
    }
    return codes
}

func (s *MFAService) generateRandomCode() string {
    bytes := make([]byte, 5)
    rand.Read(bytes)
    return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)[:8]
}
```

### 10. 实现安全头部管理

#### 10.1 创建安全头部中间件
**文件**: `internal/middleware/security_headers.go`

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "time"
)

// SecurityHeaders 安全头部中间件
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 防止点击劫持
        c.Header("X-Frame-Options", "DENY")

        // 防止MIME类型嗅探
        c.Header("X-Content-Type-Options", "nosniff")

        // XSS保护
        c.Header("X-XSS-Protection", "1; mode=block")

        // 强制HTTPS
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

        // 内容安全策略
        csp := "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
               "style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; " +
               "font-src 'self'; connect-src 'self'; frame-ancestors 'none';"
        c.Header("Content-Security-Policy", csp)

        // 引用策略
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

        // 权限策略
        permissions := "geolocation=(), microphone=(), camera=(), " +
                      "payment=(), usb=(), magnetometer=(), gyroscope=()"
        c.Header("Permissions-Policy", permissions)

        c.Next()
    }
}
```

---

## 监控和检测

### 11. 安全事件监控

#### 11.1 创建安全监控服务
**文件**: `internal/services/security_monitoring_service.go`

```go
package services

import (
    "context"
    "fmt"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type SecurityMonitoringService struct {
    loginFailures     *prometheus.CounterVec
    suspiciousIPs     *prometheus.CounterVec
    bruteForceAttempts *prometheus.CounterVec
    dataAccessEvents  *prometheus.CounterVec
}

func NewSecurityMonitoringService() *SecurityMonitoringService {
    return &SecurityMonitoringService{
        loginFailures: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "security_login_failures_total",
                Help: "Total number of login failures",
            },
            []string{"ip", "username"},
        ),
        suspiciousIPs: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "security_suspicious_ips_total",
                Help: "Total number of suspicious IP activities",
            },
            []string{"ip", "type"},
        ),
        bruteForceAttempts: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "security_brute_force_attempts_total",
                Help: "Total number of brute force attempts",
            },
            []string{"ip", "target"},
        ),
        dataAccessEvents: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "security_data_access_total",
                Help: "Total number of data access events",
            },
            []string{"user", "resource", "action"},
        ),
    }
}

func (s *SecurityMonitoringService) RecordLoginFailure(ip, username string) {
    s.loginFailures.WithLabelValues(ip, username).Inc()
}

func (s *SecurityMonitoringService) RecordSuspiciousActivity(ip, activityType string) {
    s.suspiciousIPs.WithLabelValues(ip, activityType).Inc()
}

func (s *SecurityMonitoringService) RecordBruteForceAttempt(ip, target string) {
    s.bruteForceAttempts.WithLabelValues(ip, target).Inc()
}
```

---

## 测试和验证

### 12. 安全测试脚本

#### 12.1 创建安全测试脚本
**文件**: `scripts/security-test.sh`

```bash
#!/bin/bash

echo "=== 律所OA系统安全测试 ==="

# 测试SQL注入
echo "1. 测试SQL注入防护..."
curl -X POST "http://localhost:8080/api/auth/login" \
     -H "Content-Type: application/json" \
     -d '{"email": "admin'; DROP TABLE users; --", "password": "password"}' \
     -w "HTTP Status: %{http_code}\n"

# 测试XSS
echo "2. 测试XSS防护..."
curl -X POST "http://localhost:8080/api/users/profile" \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -d '{"name": "<script>alert(1)</script>"}' \
     -w "HTTP Status: %{http_code}\n"

# 测试速率限制
echo "3. 测试速率限制..."
for i in {1..10}; do
    curl -X POST "http://localhost:8080/api/auth/login" \
         -H "Content-Type: application/json" \
         -d '{"email": "test@test.com", "password": "wrongpassword"}' \
         -w "Request $i: %{http_code}\n"
done

# 测试文件上传安全
echo "4. 测试文件上传安全..."
curl -X POST "http://localhost:8080/api/documents" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -F "file=@malicious.php" \
     -w "HTTP Status: %{http_code}\n"

echo "=== 安全测试完成 ==="
```

---

## 部署清单

### 部署前检查清单

- [ ] JWT密钥已移至环境变量
- [ ] 配置文件已加密
- [ ] SQL注入防护已实施
- [ ] XSS防护已启用
- [ ] 密码策略已强制执行
- [ ] 文件上传验证已实施
- [ ] 安全头部已配置
- [ ] 速率限制已测试
- [ ] 审计日志已启用
- [ ] 监控指标已配置
- [ ] 备份策略已验证
- [ ] 渗透测试已完成

### 部署后验证

1. **功能测试**: 确保所有功能正常工作
2. **安全测试**: 验证安全措施有效
3. **性能测试**: 确保安全措施不影响性能
4. **监控验证**: 确认监控和告警正常工作

---

## 总结

本实施指南提供了全面的安全加固步骤，按优先级和时间要求分类。建议严格按照时间表执行，并在每个阶段完成后进行充分测试。

**关键成功因素**:
1. 及时修复高风险问题
2. 充分测试每个安全改进
3. 持续监控安全状态
4. 定期进行安全评估

**后续维护**:
- 每月进行安全扫描
- 季度进行渗透测试
- 年度进行全面安全评估
- 及时应用安全补丁

通过实施这些安全措施，律所OA系统的安全性将得到显著提升，能够有效应对常见的网络威胁。