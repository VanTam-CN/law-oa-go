# 数据库设计最佳实践指南

**版本**: v2.1.0
**更新日期**: 2025-09-30
**适用数据库**: MySQL 8.0+, PostgreSQL 13+

---

## 📋 概述

本文档基于Law OA Go项目的数据库设计经验，提供了数据库设计的最佳实践指南，包括表结构设计、索引优化、数据完整性、性能调优等方面。

---

## 🎯 设计原则

### 1. 数据完整性原则
- 确保数据的准确性和一致性
- 使用适当的数据类型和约束
- 实现合理的业务规则验证

### 2. 性能优化原则
- 优化查询性能
- 合理使用索引
- 避免不必要的复杂查询

### 3. 可扩展性原则
- 设计支持未来扩展的结构
- 预留合理的字段和表空间
- 考虑数据分区策略

### 4. 安全性原则
- 保护敏感数据
- 实现访问控制
- 记录数据变更历史

---

## 📊 表结构设计规范

### 1. 命名规范

#### 表命名
```sql
-- ✅ 好的表命名
users                 -- 用户表
cases                -- 案件表
case_lawyers         -- 案件律师关联表
user_permissions     -- 用户权限表

-- ❌ 避免的命名
user                 -- 单数形式
user_info           -- 不必要的后缀
tbl_users           -- 不需要前缀
user_data           -- 过于通用
```

#### 字段命名
```sql
-- ✅ 好的字段命名
id                   -- 主键
username             -- 用户名
email                -- 邮箱
created_at           -- 创建时间
updated_at           -- 更新时间
is_active            -- 布尔字段使用is_前缀

-- ❌ 避免的命名
user_id              -- 如果表名已表明上下文
user_name            -- user_info表中的冗余前缀
name                 -- 过于通用
timestamp           -- 不够具体
flag                 -- 含义不明确
```

#### 索引命名
```sql
-- ✅ 好的索引命名
idx_users_email                    -- 单字段索引
idx_cases_status_created_at        -- 复合索引
uk_users_email                     -- 唯一索引
pk_users                          -- 主键索引

-- ❌ 避免的命名
index_1                           -- 无意义的名称
users_email_index                 -- 冗余的index后缀
temp_index                        -- 临时索引不应进入生产
```

### 2. 数据类型选择

#### 数值类型
```sql
-- ✅ 合理的数值类型选择
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,  -- 大表使用BIGINT
    age TINYINT UNSIGNED,                            -- 年龄0-255足够
    status TINYINT UNSIGNED DEFAULT 1,               -- 枚举值使用TINYINT
    balance DECIMAL(15,2),                          -- 金额使用DECIMAL
    rating FLOAT(3,2),                              -- 评分使用FLOAT
    created_count INT UNSIGNED DEFAULT 0             -- 计数器使用INT
);

-- ❌ 不合理的类型选择
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,              -- 大表可能不够
    age INT,                                        -- 浪费空间
    status VARCHAR(20),                             -- 应使用数值类型
    balance FLOAT,                                  -- 金额不应使用FLOAT
    rating VARCHAR(10),                             -- 评分不应使用字符串
    created_count BIGINT                           -- 小计数器不需要BIGINT
);
```

#### 字符串类型
```sql
-- ✅ 合理的字符串类型选择
CREATE TABLE users (
    username VARCHAR(50) NOT NULL,                  -- 固定长度的用户名
    email VARCHAR(255) NOT NULL,                    -- 邮箱最大长度
    phone VARCHAR(20),                              -- 电话号码
    bio TEXT,                                       -- 长文本使用TEXT
    avatar_url VARCHAR(500),                        -- URL可能较长
    settings JSON                                   -- 结构化数据使用JSON
);

-- ❌ 不合理的类型选择
CREATE TABLE users (
    username CHAR(50) NOT NULL,                     -- 可变长度不应使用CHAR
    email TEXT,                                     -- 邮箱不需要TEXT
    phone VARCHAR(255),                             -- 电话号码不需要这么长
    bio VARCHAR(255),                               -- 长文本应使用TEXT
    settings VARCHAR(1000)                          -- 复杂数据应使用JSON
);
```

#### 时间类型
```sql
-- ✅ 合理的时间类型选择
CREATE TABLE cases (
    id BIGINT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- 创建时间
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    start_date DATE,                                -- 只需要日期
    deadline_time TIME,                             -- 只需要时间
    archived_at TIMESTAMP NULL                      -- 可以为空的归档时间
);

-- ❌ 不合理的时间类型选择
CREATE TABLE cases (
    created_at VARCHAR(20),                         -- 不应使用字符串存储时间
    updated_at INT,                                 -- 不应使用时间戳
    start_date TIMESTAMP,                           -- 只需要日期
    deadline_time VARCHAR(10),                      -- 不应使用字符串
    archived_at VARCHAR(20) DEFAULT '0000-00-00 00:00:00'  -- 不应使用无效值
);
```

### 3. 表结构设计

#### 主键设计
```sql
-- ✅ 好的主键设计
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,  -- 自增主键
    uuid CHAR(36) UNIQUE NOT NULL,                  -- UUID作为对外标识
    -- 其他字段...
);

-- ✅ 复合主键设计（关联表）
CREATE TABLE case_lawyers (
    case_id BIGINT UNSIGNED NOT NULL,
    lawyer_id BIGINT UNSIGNED NOT NULL,
    role ENUM('primary', 'assistant') NOT NULL,
    PRIMARY KEY (case_id, lawyer_id, role),          -- 复合主键
    FOREIGN KEY (case_id) REFERENCES cases(id),
    FOREIGN KEY (lawyer_id) REFERENCES lawyers(id)
);

-- ❌ 避免的主键设计
CREATE TABLE users (
    email VARCHAR(255) PRIMARY KEY,                 -- 邮箱可能变更
    -- 其他字段...
);

CREATE TABLE case_lawyers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,           -- 不必要的自增ID
    case_id BIGINT NOT NULL,
    lawyer_id BIGINT NOT NULL,
    -- 其他字段...
);
```

#### 外键设计
```sql
-- ✅ 好的外键设计
CREATE TABLE cases (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    client_id BIGINT UNSIGNED NOT NULL,
    assigned_lawyer_id BIGINT UNSIGNED,
    created_by BIGINT UNSIGNED NOT NULL,

    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE RESTRICT,
    FOREIGN KEY (assigned_lawyer_id) REFERENCES lawyers(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,

    INDEX idx_client_id (client_id),
    INDEX idx_assigned_lawyer_id (assigned_lawyer_id),
    INDEX idx_created_by (created_by)
);

-- ❌ 避免的外键设计
CREATE TABLE cases (
    id BIGINT PRIMARY KEY,
    client_id BIGINT,                               -- 缺少UNSIGNED
    assigned_lawyer_id BIGINT,

    -- 缺少外键约束
    -- 缺少索引
    -- 缺少删除策略
);
```

---

## 📈 索引优化策略

### 1. 索引设计原则

#### 选择性原则
```sql
-- ✅ 高选择性字段的索引
CREATE INDEX idx_users_email ON users(email);        -- 邮箱唯一性高
CREATE INDEX idx_cases_status ON cases(status);      -- 状态枚举值选择性中等
CREATE INDEX idx_cases_created_at ON cases(created_at); -- 时间范围查询

-- ❌ 低选择性字段不建议单独索引
CREATE INDEX idx_users_gender ON users(gender);      -- 性别只有2-3个值，选择性低
CREATE INDEX idx_cases_is_deleted ON cases(is_deleted); -- 布尔值选择性很低
```

#### 最左前缀原则
```sql
-- ✅ 合理的复合索引设计
CREATE INDEX idx_cases_lawyer_status_created ON cases(lawyer_id, status, created_at);

-- 查询1：能使用索引
SELECT * FROM cases WHERE lawyer_id = 1 AND status = 'active';

-- 查询2：能使用索引
SELECT * FROM cases WHERE lawyer_id = 1;

-- 查询3：不能使用索引（跳过了lawyer_id）
SELECT * FROM cases WHERE status = 'active';

-- 查询4：不能使用索引（跳过了lawyer_id和status）
SELECT * FROM cases WHERE created_at > '2023-01-01';

-- ❌ 不合理的复合索引
CREATE INDEX idx_cases_created_lawyer_status ON cases(created_at, lawyer_id, status);
-- 如果经常按lawyer_id查询，这个索引效率很低
```

#### 覆盖索引原则
```sql
-- ✅ 覆盖索引设计
CREATE INDEX idx_cases_list ON cases(status, lawyer_id, created_at)
INCLUDE (id, title, client_id);

-- 查询只需要索引中的字段，不需要回表
SELECT id, title, client_id
FROM cases
WHERE status = 'active' AND lawyer_id = 1
ORDER BY created_at DESC;

-- ✅ 避免SELECT *
-- ❌ 不好的查询
SELECT * FROM cases WHERE status = 'active';  -- 需要回表获取所有字段
```

### 2. 索引类型选择

#### B-Tree索引（默认）
```sql
-- 适用于等值查询、范围查询、排序
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_cases_created_at ON cases(created_at);
CREATE INDEX idx_cases_status_lawyer ON cases(status, lawyer_id);
```

#### 哈希索引
```sql
-- 适用于等值查询，不支持范围查询（PostgreSQL）
CREATE INDEX idx_users_id_hash ON users USING HASH(id);
```

#### 全文索引
```sql
-- 适用于文本搜索（MySQL）
CREATE FULLTEXT INDEX idx_cases_content ON cases(title, description);

-- 查询示例
SELECT * FROM cases
WHERE MATCH(title, description) AGAINST('合同纠纷' IN NATURAL LANGUAGE MODE);
```

### 3. 索引维护

#### 索引使用情况分析
```sql
-- 查看索引使用情况（MySQL）
SELECT
    object_schema,
    object_name,
    index_name,
    count_read,
    count_fetch,
    sum_timer_fetch / 1000000000 AS fetch_time_ms
FROM performance_schema.table_io_waits_summary_by_index_usage
WHERE index_name IS NOT NULL
ORDER BY sum_timer_fetch DESC;

-- 查看未使用的索引
SELECT
    s.table_schema,
    s.table_name,
    s.index_name,
    s.column_name
FROM information_schema.statistics s
LEFT JOIN performance_schema.table_io_waits_summary_by_index_usage i
    ON s.table_schema = i.object_schema
    AND s.table_name = i.object_name
    AND s.index_name = i.index_name
WHERE i.index_name IS NULL
    AND s.table_schema NOT IN ('mysql', 'performance_schema', 'information_schema')
    AND s.index_name != 'PRIMARY';
```

---

## 🔧 数据完整性约束

### 1. 主键约束
```sql
-- ✅ 主键约束
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT,
    uuid CHAR(36) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_uuid (uuid)
);

-- ✅ 复合主键
CREATE TABLE user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);
```

### 2. 外键约束
```sql
-- ✅ 外键约束与级联操作
CREATE TABLE cases (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    client_id BIGINT UNSIGNED NOT NULL,
    assigned_lawyer_id BIGINT UNSIGNED,

    FOREIGN KEY (client_id) REFERENCES clients(id)
        ON DELETE RESTRICT ON UPDATE CASCADE,
    FOREIGN KEY (assigned_lawyer_id) REFERENCES lawyers(id)
        ON DELETE SET NULL ON UPDATE CASCADE
);

-- ✅ 自引用外键
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    manager_id BIGINT UNSIGNED,

    FOREIGN KEY (manager_id) REFERENCES users(id)
        ON DELETE SET NULL ON UPDATE CASCADE
);
```

### 3. 唯一约束
```sql
-- ✅ 唯一约束
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,

    UNIQUE KEY uk_username (username),
    UNIQUE KEY uk_email (email)
);

-- ✅ 复合唯一约束
CREATE TABLE case_numbers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    year INT NOT NULL,
    sequence INT NOT NULL,
    case_id BIGINT UNSIGNED NOT NULL,

    UNIQUE KEY uk_year_sequence (year, sequence)
);
```

### 4. 检查约束
```sql
-- ✅ 检查约束（MySQL 8.0+, PostgreSQL）
CREATE TABLE cases (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    status ENUM('pending', 'active', 'completed', 'archived') NOT NULL,
    priority TINYINT UNSIGNED NOT NULL,
    estimated_value DECIMAL(15,2),

    CHECK (priority BETWEEN 1 AND 5),
    CHECK (estimated_value >= 0),
    CHECK (title != '')
);

-- ✅ 复杂检查约束
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    birth_date DATE,

    CONSTRAINT chk_email_format CHECK (email REGEXP '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$'),
    CONSTRAINT chk_phone_format CHECK (phone IS NULL OR phone REGEXP '^1[3-9][0-9]{9}$'),
    CONSTRAINT chk_birth_date CHECK (birth_date IS NULL OR birth_date < CURDATE())
);
```

---

## ⚡ 性能优化策略

### 1. 查询优化

#### 避免 SELECT *
```sql
-- ✅ 指定需要的字段
SELECT id, username, email, created_at
FROM users
WHERE id = 1;

-- ❌ 避免 SELECT *
SELECT * FROM users WHERE id = 1;
```

#### 合理使用 JOIN
```sql
-- ✅ 使用 INNER JOIN 并明确字段
SELECT
    c.id, c.title, c.status,
    cl.name as client_name,
    l.name as lawyer_name
FROM cases c
INNER JOIN clients cl ON c.client_id = cl.id
LEFT JOIN lawyers l ON c.assigned_lawyer_id = l.id
WHERE c.status = 'active'
ORDER BY c.created_at DESC;

-- ❌ 避免不必要的 JOIN
SELECT c.*, cl.*, l.*
FROM cases c
LEFT JOIN clients cl ON c.client_id = cl.id
LEFT JOIN lawyers l ON c.assigned_lawyer_id = l.id
LEFT JOIN users u ON c.created_by = u.id  -- 不需要的连接
WHERE 1=1;
```

#### 使用 LIMIT 分页
```sql
-- ✅ 高效的分页查询
SELECT id, title, status, created_at
FROM cases
WHERE status = 'active'
ORDER BY created_at DESC
LIMIT 20 OFFSET 0;  -- 第一页

-- ✅ 使用游标分页（适合大数据量）
SELECT id, title, status, created_at
FROM cases
WHERE status = 'active' AND created_at < '2023-12-01 10:00:00'
ORDER BY created_at DESC
LIMIT 20;

-- ❌ 大偏移量的分页
SELECT id, title, status, created_at
FROM cases
ORDER BY created_at DESC
LIMIT 20 OFFSET 100000;  -- 性能很差
```

### 2. 数据库配置优化

#### MySQL 配置优化
```ini
# my.cnf

[mysqld]
# 基础配置
datadir = /var/lib/mysql
socket = /var/lib/mysql/mysql.sock
port = 3306

# 内存配置
innodb_buffer_pool_size = 2G          # 设置为可用内存的70-80%
innodb_log_file_size = 256M           # 增大日志文件
innodb_log_buffer_size = 16M          # 日志缓冲区
key_buffer_size = 32M                 # MyISAM索引缓冲
tmp_table_size = 64M                  # 临时表大小
max_heap_table_size = 64M             # 内存表大小

# 连接配置
max_connections = 200                 # 最大连接数
max_connect_errors = 1000             # 最大连接错误数
wait_timeout = 28800                  # 等待超时时间

# 查询缓存（MySQL 8.0已移除）
# query_cache_type = 1
# query_cache_size = 64M

# InnoDB 配置
innodb_file_per_table = 1             # 每表一个文件
innodb_flush_method = O_DIRECT        # 直接I/O
innodb_flush_log_at_trx_commit = 2    # 日志刷新策略
innodb_lock_wait_timeout = 50         # 锁等待超时

# 字符集配置
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci

# 日志配置
log_error = /var/log/mysql/error.log
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2
```

#### PostgreSQL 配置优化
```ini
# postgresql.conf

# 内存配置
shared_buffers = 256MB                 # 共享缓冲区
effective_cache_size = 1GB            # 有效缓存大小
work_mem = 4MB                        # 工作内存
maintenance_work_mem = 64MB           # 维护工作内存

# 连接配置
max_connections = 100                  # 最大连接数
shared_preload_libraries = 'pg_stat_statements'  # 预加载库

# WAL 配置
wal_buffers = 16MB                    # WAL缓冲区
checkpoint_completion_target = 0.9    # 检查点完成目标
wal_writer_delay = 200ms              # WAL写入延迟

# 查询规划
random_page_cost = 1.1                # 随机页面成本（SSD优化）
effective_io_concurrency = 200        # 有效I/O并发

# 日志配置
logging_collector = on
log_directory = 'pg_log'
log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
log_statement = 'all'                 # 记录所有SQL
log_min_duration_statement = 1000     # 记录慢查询（1秒）
```

### 3. 分区策略

#### 按时间分区
```sql
-- ✅ 按月分区（MySQL 8.0+）
CREATE TABLE cases (
    id BIGINT UNSIGNED AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    status ENUM('pending', 'active', 'completed', 'archived') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id, created_at),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
)
PARTITION BY RANGE (YEAR(created_at) * 100 + MONTH(created_at)) (
    PARTITION p202301 VALUES LESS THAN (202302),
    PARTITION p202302 VALUES LESS THAN (202303),
    PARTITION p202303 VALUES LESS THAN (202304),
    -- ... 更多分区
    PARTITION p_future VALUES LESS THAN MAXVALUE
);

-- ✅ 按时间分区的查询优化
SELECT * FROM cases
WHERE created_at >= '2023-01-01' AND created_at < '2023-02-01'
AND status = 'active';  -- 可以利用分区裁剪
```

#### 按业务分区
```sql
-- ✅ 按状态分区
CREATE TABLE cases (
    id BIGINT UNSIGNED AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    status ENUM('pending', 'active', 'completed', 'archived') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id, status),
    INDEX idx_created_at (created_at)
)
PARTITION BY LIST COLUMNS(status) (
    PARTITION p_pending VALUES IN ('pending'),
    PARTITION p_active VALUES IN ('active'),
    PARTITION p_completed VALUES IN ('completed'),
    PARTITION p_archived VALUES IN ('archived')
);
```

---

## 🛡️ 数据安全策略

### 1. 敏感数据加密

#### 字段级加密
```sql
-- ✅ 使用AES加密敏感数据
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone_encrypted VARBINARY(255),     -- 加密存储
    id_card_encrypted VARBINARY(255),   -- 加密存储
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 加密函数（MySQL）
DELIMITER $$
CREATE FUNCTION encrypt_sensitive(data VARCHAR(255))
RETURNS VARBINARY(255)
DETERMINISTIC
BEGIN
    RETURN AES_ENCRYPT(data, 'your-encryption-key');
END$$

CREATE FUNCTION decrypt_sensitive(encrypted_data VARBINARY(255))
RETURNS VARCHAR(255)
DETERMINISTIC
BEGIN
    RETURN AES_DECRYPT(encrypted_data, 'your-encryption-key');
END$$
DELIMITER ;

-- 使用示例
INSERT INTO users (username, email, phone_encrypted, id_card_encrypted)
VALUES ('testuser', 'test@example.com',
        encrypt_sensitive('13800138000'),
        encrypt_sensitive('110101199001011234'));
```

#### 数据脱敏
```sql
-- ✅ 创建脱敏视图
CREATE VIEW users_public AS
SELECT
    id,
    username,
    LEFT(email, 2) + '***' + RIGHT(email, 3) as email_mask,
    CONCAT(SUBSTRING(phone_encrypted, 1, 3), '****', SUBSTRING(phone_encrypted, 8, 4)) as phone_mask,
    created_at
FROM users;

-- 查询脱敏数据
SELECT * FROM users_public WHERE id = 1;
```

### 2. 访问控制

#### 用户权限管理
```sql
-- ✅ 创建专用数据库用户
-- 应用用户（读写权限）
CREATE USER 'law_oa_app'@'%' IDENTIFIED BY 'strong_password';
GRANT SELECT, INSERT, UPDATE, DELETE ON law_oa.* TO 'law_oa_app'@'%';

-- 只读用户（报表查询）
CREATE USER 'law_oa_readonly'@'%' IDENTIFIED BY 'readonly_password';
GRANT SELECT ON law_oa.* TO 'law_oa_readonly'@'%';

-- 备份用户（备份权限）
CREATE USER 'law_oa_backup'@'localhost' IDENTIFIED BY 'backup_password';
GRANT SELECT, LOCK TABLES, SHOW VIEW ON law_oa.* TO 'law_oa_backup'@'localhost';

-- ✅ 限制敏感表访问
REVOKE SELECT ON law_oa.users FROM 'law_oa_readonly'@'%';
GRANT SELECT (id, username, created_at) ON law_oa.users TO 'law_oa_readonly'@'%';
```

#### 审计日志
```sql
-- ✅ 创建审计表
CREATE TABLE audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    table_name VARCHAR(64) NOT NULL,
    operation ENUM('INSERT', 'UPDATE', 'DELETE') NOT NULL,
    record_id BIGINT UNSIGNED NOT NULL,
    old_data JSON,
    new_data JSON,
    user_id BIGINT UNSIGNED,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_table_record (table_name, record_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);

-- ✅ 创建触发器记录变更
DELIMITER $$
CREATE TRIGGER cases_audit_insert
AFTER INSERT ON cases
FOR EACH ROW
BEGIN
    INSERT INTO audit_logs (table_name, operation, record_id, new_data)
    VALUES ('cases', 'INSERT', NEW.id, JSON_OBJECT(
        'title', NEW.title,
        'status', NEW.status,
        'client_id', NEW.client_id
    ));
END$$

CREATE TRIGGER cases_audit_update
AFTER UPDATE ON cases
FOR EACH ROW
BEGIN
    INSERT INTO audit_logs (table_name, operation, record_id, old_data, new_data)
    VALUES ('cases', 'UPDATE', NEW.id,
            JSON_OBJECT('title', OLD.title, 'status', OLD.status),
            JSON_OBJECT('title', NEW.title, 'status', NEW.status));
END$$
DELIMITER ;
```

---

## 📊 监控和维护

### 1. 性能监控

#### 慢查询监控
```sql
-- 启用慢查询日志
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 2;  -- 2秒以上的查询记录
SET GLOBAL log_queries_not_using_indexes = 'ON';

-- 查看慢查询
SELECT
    start_time,
    query_time,
    lock_time,
    rows_sent,
    rows_examined,
    sql_text
FROM mysql.slow_log
WHERE start_time > DATE_SUB(NOW(), INTERVAL 1 DAY)
ORDER BY query_time DESC
LIMIT 10;
```

#### 性能指标监控
```sql
-- 查看数据库性能指标（MySQL）
SELECT
    VARIABLE_NAME,
    VARIABLE_VALUE
FROM performance_schema.global_status
WHERE VARIABLE_NAME IN (
    'Connections',
    'Max_used_connections',
    'Questions',
    'Slow_queries',
    'Uptime',
    'Bytes_received',
    'Bytes_sent'
);

-- 查看表大小和行数
SELECT
    table_name,
    ROUND(((data_length + index_length) / 1024 / 1024), 2) AS table_size_mb,
    table_rows
FROM information_schema.tables
WHERE table_schema = 'law_oa'
ORDER BY (data_length + index_length) DESC;
```

### 2. 定期维护

#### 数据清理策略
```sql
-- ✅ 创建数据清理存储过程
DELIMITER $$
CREATE PROCEDURE cleanup_old_data()
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    START TRANSACTION;

    -- 删除6个月前的审计日志
    DELETE FROM audit_logs
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 6 MONTH);

    -- 归档1年前的已完成案件
    UPDATE cases
    SET status = 'archived', archived_at = NOW()
    WHERE status = 'completed'
    AND updated_at < DATE_SUB(NOW(), INTERVAL 1 YEAR);

    -- 删除已归档案件的附件（超过2年）
    DELETE FROM case_attachments
    WHERE case_id IN (
        SELECT id FROM cases
        WHERE status = 'archived'
        AND archived_at < DATE_SUB(NOW(), INTERVAL 2 YEAR)
    );

    COMMIT;
END$$
DELIMITER ;

-- 定期执行清理
-- 添加到crontab：0 2 * * 0 /usr/bin/mysql -u root -p law_oa -e "CALL cleanup_old_data();"
```

#### 索引维护
```sql
-- 分析表统计信息
ANALYZE TABLE users, cases, clients;

-- 重建索引（MySQL）
ALTER TABLE cases ENGINE=InnoDB;

-- 优化表
OPTIMIZE TABLE cases;

-- 检查表
CHECK TABLE cases;
```

---

## 🔄 数据迁移策略

### 1. 结构变更

#### 在线DDL操作
```sql
-- ✅ 大表添加索引（MySQL 5.6+支持在线DDL）
ALTER TABLE cases ADD INDEX idx_status_created (status, created_at), ALGORITHM=INPLACE, LOCK=NONE;

-- ✅ 添加字段（避免锁表）
ALTER TABLE cases ADD COLUMN priority TINYINT UNSIGNED DEFAULT 1, ALGORITHM=INPLACE, LOCK=NONE;

-- ❌ 避免的DDL操作
ALTER TABLE cases ADD COLUMN description TEXT;  -- 可能锁表很久
```

#### 分区表维护
```sql
-- ✅ 添加新分区
ALTER TABLE cases ADD PARTITION (
    PARTITION p202401 VALUES LESS THAN (202402)
);

-- ✅ 删除旧分区
ALTER TABLE cases DROP PARTITION p202301;

-- ✅ 合并分区
ALTER TABLE cases REORGANIZE PARTITION p202301, p202302 INTO (
    PARTITION p2023_q1 VALUES LESS THAN (202304)
);
```

### 2. 数据迁移

#### 批量数据迁移
```sql
-- ✅ 使用批量操作迁移数据
DELIMITER $$
CREATE PROCEDURE migrate_user_data()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE batch_size INT DEFAULT 1000;
    DECLARE offset_val INT DEFAULT 0;
    DECLARE affected_rows INT DEFAULT 1;

    WHILE affected_rows > 0 DO
        -- 迁移一批数据
        UPDATE users_old uo
        INNER JOIN users_new un ON uo.id = un.id
        SET un.profile_data = uo.profile_data
        WHERE uo.profile_data IS NOT NULL
        AND un.profile_data IS NULL
        LIMIT batch_size;

        SET affected_rows = ROW_COUNT();
        SET offset_val = offset_val + batch_size;

        -- 记录进度
        DO SLEEP(0.1);  -- 避免过度占用资源
    END WHILE;

END$$
DELIMITER ;

-- 执行迁移
CALL migrate_user_data();
```

---

## 📝 开发规范检查清单

### 表设计
- [ ] 表名使用小写复数形式
- [ ] 字段命名清晰、一致
- [ ] 使用合适的数据类型
- [ ] 设置必要的约束
- [ ] 添加外键关联
- [ ] 包含创建和更新时间字段

### 索引设计
- [ ] 主键自动创建索引
- [ ] 外键字段创建索引
- [ ] 查询条件字段创建索引
- [ ] 避免过度索引
- [ ] 定期分析索引使用情况
- [ ] 删除未使用的索引

### 性能优化
- [ ] 避免 SELECT *
- [ ] 合理使用 JOIN
- [ ] 实现 分页查询
- [ ] 优化慢查询
- [ ] 使用合适的存储引擎
- [ ] 配置合适的数据库参数

### 数据安全
- [ ] 敏感数据加密存储
- [ ] 实现访问权限控制
- [ ] 记录数据变更审计
- [ ] 定期备份数据
- [ ] 使用参数化查询
- [ ] 限制数据库用户权限

### 监控维护
- [ ] 启用慢查询日志
- [ ] 监控数据库性能指标
- [ ] 定期清理历史数据
- [ ] 维护表统计信息
- [ ] 监控磁盘空间使用
- [ ] 制定备份恢复策略

---

**文档版本**: v2.1.0
**最后更新**: 2025-09-30
**下次审查**: 2025-12-30
**维护团队**: 数据库团队