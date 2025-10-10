-- 用户行为分析系统数据库迁移
-- 创建时间: 2025-01-01
-- 版本: 1.0.0

-- 创建用户会话表
CREATE TABLE IF NOT EXISTS `user_sessions` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `user_id` VARCHAR(64) NOT NULL,
    `ip_address` VARCHAR(45),
    `user_agent` VARCHAR(500),
    `start_time` DATETIME NOT NULL,
    `end_time` DATETIME NULL,
    `duration` BIGINT DEFAULT 0 COMMENT '会话持续时间（毫秒）',
    `is_active` BOOLEAN DEFAULT TRUE,
    `page_views` INT DEFAULT 0,
    `last_active` DATETIME NOT NULL,
    `referrer` VARCHAR(500),
    `source` VARCHAR(100) COMMENT '来源：direct, search, social, email等',
    `campaign` VARCHAR(100),
    `device_type` VARCHAR(50) COMMENT 'desktop, mobile, tablet',
    `platform` VARCHAR(50) COMMENT 'windows, mac, linux, ios, android',
    `browser` VARCHAR(100),
    `location` JSON COMMENT '地理位置信息',
    `metadata` JSON COMMENT '额外元数据',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX `idx_user_sessions_user_id` (`user_id`),
    INDEX `idx_user_sessions_start_time` (`start_time`),
    INDEX `idx_user_sessions_is_active` (`is_active`),
    INDEX `idx_user_sessions_last_active` (`last_active`),
    INDEX `idx_user_sessions_source` (`source`),
    INDEX `idx_user_sessions_device_type` (`device_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';

-- 创建页面浏览表
CREATE TABLE IF NOT EXISTS `page_views` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `session_id` VARCHAR(64) NOT NULL,
    `user_id` VARCHAR(64) NOT NULL,
    `url` VARCHAR(2000) NOT NULL,
    `path` VARCHAR(500) NOT NULL,
    `title` VARCHAR(500),
    `referrer` VARCHAR(500),
    `timestamp` DATETIME NOT NULL,
    `duration` BIGINT DEFAULT 0 COMMENT '页面停留时间（毫秒）',
    `scroll_depth` INT DEFAULT 0 COMMENT '滚动深度百分比',
    `viewport_size` VARCHAR(20) COMMENT '视口大小 "1024x768"',
    `screen_size` VARCHAR(20) COMMENT '屏幕大小 "1920x1080"',
    `interaction` VARCHAR(50) COMMENT '交互类型',
    `is_bounce` BOOLEAN DEFAULT FALSE COMMENT '是否跳出',
    `exit_page` BOOLEAN DEFAULT FALSE COMMENT '是否退出页',
    `entry_page` BOOLEAN DEFAULT FALSE COMMENT '是否入口页',
    `properties` JSON COMMENT '页面属性',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_page_views_session_id` (`session_id`),
    INDEX `idx_page_views_user_id` (`user_id`),
    INDEX `idx_page_views_url` (`url`(255)),
    INDEX `idx_page_views_path` (`path`),
    INDEX `idx_page_views_timestamp` (`timestamp`),
    INDEX `idx_page_views_entry_page` (`entry_page`),
    INDEX `idx_page_views_exit_page` (`exit_page`),

    FOREIGN KEY (`session_id`) REFERENCES `user_sessions`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='页面浏览记录表';

-- 创建用户事件表
CREATE TABLE IF NOT EXISTS `user_events` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `session_id` VARCHAR(64) NOT NULL,
    `user_id` VARCHAR(64) NOT NULL,
    `event_type` VARCHAR(50) NOT NULL COMMENT '事件类型：click, form_submit, download等',
    `event_category` VARCHAR(50) NOT NULL COMMENT '事件分类',
    `event_action` VARCHAR(50) NOT NULL COMMENT '事件动作',
    `event_label` VARCHAR(100),
    `event_value` DECIMAL(10,2),
    `url` VARCHAR(2000),
    `element` VARCHAR(200) COMMENT '触发事件的元素',
    `properties` JSON COMMENT '事件属性',
    `timestamp` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_user_events_session_id` (`session_id`),
    INDEX `idx_user_events_user_id` (`user_id`),
    INDEX `idx_user_events_type` (`event_type`),
    INDEX `idx_user_events_category` (`event_category`),
    INDEX `idx_user_events_action` (`event_action`),
    INDEX `idx_user_events_timestamp` (`timestamp`),

    FOREIGN KEY (`session_id`) REFERENCES `user_sessions`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户事件记录表';

-- 创建用户旅程表
CREATE TABLE IF NOT EXISTS `user_journeys` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `user_id` VARCHAR(64) NOT NULL,
    `journey_type` VARCHAR(50) NOT NULL COMMENT '旅程类型：registration, purchase, onboarding等',
    `start_time` DATETIME NOT NULL,
    `end_time` DATETIME NULL,
    `steps` JSON COMMENT '旅程步骤',
    `current_step` INT DEFAULT 0,
    `is_completed` BOOLEAN DEFAULT FALSE,
    `properties` JSON COMMENT '旅程属性',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX `idx_user_journeys_user_id` (`user_id`),
    INDEX `idx_user_journeys_type` (`journey_type`),
    INDEX `idx_user_journeys_start_time` (`start_time`),
    INDEX `idx_user_journeys_completed` (`is_completed`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户旅程表';

-- 创建用户细分表
CREATE TABLE IF NOT EXISTS `user_segments` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL,
    `description` TEXT,
    `segment_type` VARCHAR(50) NOT NULL COMMENT '细分类型：behavioral, demographic, custom等',
    `criteria` JSON COMMENT '细分条件',
    `is_active` BOOLEAN DEFAULT TRUE,
    `user_count` INT DEFAULT 0,
    `created_by` VARCHAR(64),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX `idx_user_segments_type` (`segment_type`),
    INDEX `idx_user_segments_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户细分表';

-- 创建用户细分成员关系表
CREATE TABLE IF NOT EXISTS `user_segment_memberships` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `segment_id` VARCHAR(64) NOT NULL,
    `user_id` VARCHAR(64) NOT NULL,
    `added_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `removed_at` DATETIME NULL,
    `is_active` BOOLEAN DEFAULT TRUE,

    INDEX `idx_user_segment_memberships_segment_id` (`segment_id`),
    INDEX `idx_user_segment_memberships_user_id` (`user_id`),
    INDEX `idx_user_segment_memberships_active` (`is_active`),

    UNIQUE KEY `uk_segment_user` (`segment_id`, `user_id`),
    FOREIGN KEY (`segment_id`) REFERENCES `user_segments`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户细分成员关系表';

-- 创建行为模式表
CREATE TABLE IF NOT EXISTS `behavior_patterns` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `user_id` VARCHAR(64) NOT NULL,
    `pattern_type` VARCHAR(50) NOT NULL COMMENT '模式类型：time_preference, navigation_pattern等',
    `pattern_name` VARCHAR(100) NOT NULL,
    `description` TEXT,
    `confidence` DECIMAL(3,2) DEFAULT 0.00 COMMENT '置信度 0-1',
    `frequency` INT DEFAULT 0 COMMENT '出现频次',
    `pattern_data` JSON COMMENT '模式数据',
    `last_detected` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_behavior_patterns_user_id` (`user_id`),
    INDEX `idx_behavior_patterns_type` (`pattern_type`),
    INDEX `idx_behavior_patterns_confidence` (`confidence`),
    INDEX `idx_behavior_patterns_last_detected` (`last_detected`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='行为模式表';

-- 创建用户行为记录表
CREATE TABLE IF NOT EXISTS `user_behavior_records` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `user_id` VARCHAR(64) NOT NULL,
    `session_id` VARCHAR(64) NOT NULL,
    `behavior_type` VARCHAR(50) NOT NULL,
    `behavior_data` JSON,
    `context` JSON COMMENT '行为上下文',
    `timestamp` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_user_behavior_records_user_id` (`user_id`),
    INDEX `idx_user_behavior_records_session_id` (`session_id`),
    INDEX `idx_user_behavior_records_type` (`behavior_type`),
    INDEX `idx_user_behavior_records_timestamp` (`timestamp`),

    FOREIGN KEY (`session_id`) REFERENCES `user_sessions`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户行为记录表';

-- 创建漏斗分析表
CREATE TABLE IF NOT EXISTS `funnel_analyses` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `funnel_name` VARCHAR(100) NOT NULL,
    `step_name` VARCHAR(100) NOT NULL,
    `step_order` INT NOT NULL,
    `user_count` INT DEFAULT 0,
    `conversion_rate` DECIMAL(5,2) DEFAULT 0.00 COMMENT '转化率百分比',
    `time_to_convert` INT DEFAULT 0 COMMENT '转化时间（秒）',
    `drop_off_rate` DECIMAL(5,2) DEFAULT 0.00 COMMENT '流失率百分比',
    `analysis_date` DATE NOT NULL,
    `properties` JSON,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_funnel_analyses_funnel_name` (`funnel_name`),
    INDEX `idx_funnel_analyses_step_order` (`step_order`),
    INDEX `idx_funnel_analyses_date` (`analysis_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='漏斗分析表';

-- 创建留存分析表
CREATE TABLE IF NOT EXISTS `retention_analyses` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `cohort_date` DATE NOT NULL COMMENT '队列日期',
    `period_number` INT NOT NULL COMMENT '周期编号（0,1,2...）',
    `period_type` VARCHAR(20) DEFAULT 'day' COMMENT '周期类型：day, week, month',
    `cohort_size` INT DEFAULT 0 COMMENT '队列大小',
    `retained_users` INT DEFAULT 0 COMMENT '留存用户数',
    `retention_rate` DECIMAL(5,2) DEFAULT 0.00 COMMENT '留存率百分比',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_retention_analyses_cohort_date` (`cohort_date`),
    INDEX `idx_retention_analyses_period_number` (`period_number`),
    INDEX `idx_retention_analyses_period_type` (`period_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='留存分析表';

-- 创建分析报告表
CREATE TABLE IF NOT EXISTS `analytics_reports` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `title` VARCHAR(200) NOT NULL,
    `report_type` VARCHAR(50) NOT NULL COMMENT '报告类型：daily, weekly, monthly, custom',
    `description` TEXT,
    `data` JSON COMMENT '报告数据',
    `filters` JSON COMMENT '过滤条件',
    `date_range_start` DATE,
    `date_range_end` DATE,
    `is_scheduled` BOOLEAN DEFAULT FALSE,
    `schedule_frequency` VARCHAR(20) COMMENT '调度频率：daily, weekly, monthly',
    `recipients` JSON COMMENT '收件人列表',
    `created_by` VARCHAR(64),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX `idx_analytics_reports_type` (`report_type`),
    INDEX `idx_analytics_reports_date_range` (`date_range_start`, `date_range_end`),
    INDEX `idx_analytics_reports_scheduled` (`is_scheduled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分析报告表';

-- 创建热力图数据表
CREATE TABLE IF NOT EXISTS `heatmap_data` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `page_url` VARCHAR(2000) NOT NULL,
    `viewport_size` VARCHAR(20),
    `click_data` JSON COMMENT '点击热力图数据',
    `move_data` JSON COMMENT '移动热力图数据',
    `scroll_data` JSON COMMENT '滚动热力图数据',
    `date_collected` DATE NOT NULL,
    `total_clicks` INT DEFAULT 0,
    `total_views` INT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_heatmap_data_page_url` (`page_url`(255)),
    INDEX `idx_heatmap_data_date` (`date_collected`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='热力图数据表';

-- 创建点击流记录表
CREATE TABLE IF NOT EXISTS `clickstream_records` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `session_id` VARCHAR(64) NOT NULL,
    `user_id` VARCHAR(64) NOT NULL,
    `sequence_number` INT NOT NULL,
    `action_type` VARCHAR(50) NOT NULL COMMENT '动作类型：click, view, scroll等',
    `target_element` VARCHAR(200),
    `page_url` VARCHAR(2000),
    `coordinates` JSON COMMENT '点击坐标',
    `timestamp` DATETIME NOT NULL,
    `duration` INT DEFAULT 0 COMMENT '动作持续时间（毫秒）',
    `properties` JSON,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_clickstream_records_session_id` (`session_id`),
    INDEX `idx_clickstream_records_user_id` (`user_id`),
    INDEX `idx_clickstream_records_sequence` (`session_id`, `sequence_number`),
    INDEX `idx_clickstream_records_timestamp` (`timestamp`),

    FOREIGN KEY (`session_id`) REFERENCES `user_sessions`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点击流记录表';

-- 创建表单交互表
CREATE TABLE IF NOT EXISTS `form_interactions` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `session_id` VARCHAR(64) NOT NULL,
    `user_id` VARCHAR(64) NOT NULL,
    `form_id` VARCHAR(100) NOT NULL,
    `form_name` VARCHAR(200),
    `page_url` VARCHAR(2000),
    `interaction_type` VARCHAR(50) NOT NULL COMMENT '交互类型：focus, blur, change, submit等',
    `field_name` VARCHAR(100),
    `field_value` VARCHAR(500),
    `validation_errors` JSON,
    `is_completed` BOOLEAN DEFAULT FALSE,
    `time_spent` INT DEFAULT 0 COMMENT '表单填写时间（秒）',
    `timestamp` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_form_interactions_session_id` (`session_id`),
    INDEX `idx_form_interactions_user_id` (`user_id`),
    INDEX `idx_form_interactions_form_id` (`form_id`),
    INDEX `idx_form_interactions_type` (`interaction_type`),
    INDEX `idx_form_interactions_completed` (`is_completed`),

    FOREIGN KEY (`session_id`) REFERENCES `user_sessions`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='表单交互表';

-- 创建性能指标表
CREATE TABLE IF NOT EXISTS `performance_metrics` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `page_url` VARCHAR(2000) NOT NULL,
    `session_id` VARCHAR(64) NOT NULL,
    `user_id` VARCHAR(64) NOT NULL,
    `metric_type` VARCHAR(50) NOT NULL COMMENT '指标类型：page_load, dom_content_loaded, first_paint等',
    `metric_value` BIGINT NOT NULL COMMENT '指标值（毫秒）',
    `device_type` VARCHAR(50),
    `connection_type` VARCHAR(50),
    `browser` VARCHAR(100),
    `timestamp` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_performance_metrics_page_url` (`page_url`(255)),
    INDEX `idx_performance_metrics_type` (`metric_type`),
    INDEX `idx_performance_metrics_value` (`metric_value`),
    INDEX `idx_performance_metrics_timestamp` (`timestamp`),

    FOREIGN KEY (`session_id`) REFERENCES `user_sessions`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='性能指标表';

-- 创建搜索事件表
CREATE TABLE IF NOT EXISTS `search_events` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `session_id` VARCHAR(64) NOT NULL,
    `user_id` VARCHAR(64) NOT NULL,
    `search_query` VARCHAR(500) NOT NULL,
    `search_type` VARCHAR(50) DEFAULT 'global' COMMENT '搜索类型：global, site, internal等',
    `results_count` INT DEFAULT 0,
    `clicked_result_url` VARCHAR(2000),
    `time_to_click` INT DEFAULT 0 COMMENT '点击时间（毫秒）',
    `page_url` VARCHAR(2000),
    `timestamp` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_search_events_session_id` (`session_id`),
    INDEX `idx_search_events_user_id` (`user_id`),
    INDEX `idx_search_events_query` (`search_query`),
    INDEX `idx_search_events_type` (`search_type`),
    INDEX `idx_search_events_timestamp` (`timestamp`),

    FOREIGN KEY (`session_id`) REFERENCES `user_sessions`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='搜索事件表';

-- 创建转化事件表
CREATE TABLE IF NOT EXISTS `conversion_events` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `session_id` VARCHAR(64) NOT NULL,
    `user_id` VARCHAR(64) NOT NULL,
    `event_type` VARCHAR(50) NOT NULL COMMENT '转化类型：signup, login, purchase, download等',
    `event_category` VARCHAR(50) NOT NULL,
    `event_action` VARCHAR(50) NOT NULL,
    `event_label` VARCHAR(100),
    `event_value` DECIMAL(10,2) COMMENT '转化值',
    `currency` VARCHAR(10) DEFAULT 'CNY',
    `url` VARCHAR(2000) NOT NULL,
    `referrer` VARCHAR(500),
    `source` VARCHAR(100),
    `campaign` VARCHAR(100),
    `properties` JSON,
    `timestamp` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_conversion_events_session_id` (`session_id`),
    INDEX `idx_conversion_events_user_id` (`user_id`),
    INDEX `idx_conversion_events_type` (`event_type`),
    INDEX `idx_conversion_events_category` (`event_category`),
    INDEX `idx_conversion_events_timestamp` (`timestamp`),

    FOREIGN KEY (`session_id`) REFERENCES `user_sessions`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='转化事件表';

-- 创建实时统计表
CREATE TABLE IF NOT EXISTS `real_time_stats` (
    `id` VARCHAR(64) NOT NULL PRIMARY KEY,
    `metric_name` VARCHAR(100) NOT NULL COMMENT '指标名称：active_users, page_views, events_per_minute等',
    `value` DECIMAL(15,2) NOT NULL COMMENT '指标值',
    `dimensions` JSON COMMENT '维度信息',
    `timestamp` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_real_time_stats_metric_name` (`metric_name`),
    INDEX `idx_real_time_stats_timestamp` (`timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实时统计表';

-- 创建一些视图用于常用查询

-- 用户会话统计视图
CREATE OR REPLACE VIEW `v_user_session_stats` AS
SELECT
    DATE(start_time) as session_date,
    COUNT(*) as total_sessions,
    COUNT(DISTINCT user_id) as unique_users,
    AVG(duration) as avg_duration,
    COUNT(CASE WHEN is_active = 1 THEN 1 END) as active_sessions,
    SUM(page_views) as total_page_views
FROM user_sessions
GROUP BY DATE(start_time);

-- 页面浏览统计视图
CREATE OR REPLACE VIEW `v_page_view_stats` AS
SELECT
    DATE(timestamp) as view_date,
    url,
    path,
    COUNT(*) as total_views,
    COUNT(DISTINCT session_id) as unique_sessions,
    COUNT(DISTINCT user_id) as unique_users,
    AVG(duration) as avg_duration
FROM page_views
GROUP BY DATE(timestamp), url, path;

-- 事件统计视图
CREATE OR REPLACE VIEW `v_event_stats` AS
SELECT
    DATE(timestamp) as event_date,
    event_type,
    event_category,
    COUNT(*) as total_events,
    COUNT(DISTINCT session_id) as unique_sessions,
    COUNT(DISTINCT user_id) as unique_users
FROM user_events
GROUP BY DATE(timestamp), event_type, event_category;

-- 插入初始数据

-- 插入默认用户细分
INSERT IGNORE INTO `user_segments` (`id`, `name`, `description`, `segment_type`, `criteria`, `is_active`) VALUES
('segment_new_users', '新用户', '注册后30天内的用户', 'behavioral', '{"days_since_registration": {"lte": 30}}', TRUE),
('segment_active_users', '活跃用户', '过去7天内有活动的用户', 'behavioral', '{"days_since_last_activity": {"lte": 7}}', TRUE),
('segment_inactive_users', '非活跃用户', '超过30天未活动的用户', 'behavioral', '{"days_since_last_activity": {"gt": 30}}', TRUE),
('segment_power_users', '重度用户', '月活跃天数超过20天的用户', 'behavioral', '{"monthly_active_days": {"gte": 20}}', TRUE);