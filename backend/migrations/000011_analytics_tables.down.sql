-- 用户行为分析系统数据库回滚迁移
-- 创建时间: 2025-01-01
-- 版本: 1.0.0

-- 删除视图
DROP VIEW IF EXISTS `v_user_session_stats`;
DROP VIEW IF EXISTS `v_page_view_stats`;
DROP VIEW IF EXISTS `v_event_stats`;

-- 删除表（按照外键依赖关系的相反顺序）
DROP TABLE IF EXISTS `real_time_stats`;
DROP TABLE IF EXISTS `conversion_events`;
DROP TABLE IF EXISTS `search_events`;
DROP TABLE IF EXISTS `performance_metrics`;
DROP TABLE IF EXISTS `form_interactions`;
DROP TABLE IF EXISTS `clickstream_records`;
DROP TABLE IF EXISTS `heatmap_data`;
DROP TABLE IF EXISTS `analytics_reports`;
DROP TABLE IF EXISTS `retention_analyses`;
DROP TABLE IF EXISTS `funnel_analyses`;
DROP TABLE IF EXISTS `user_behavior_records`;
DROP TABLE IF EXISTS `behavior_patterns`;
DROP TABLE IF EXISTS `user_segment_memberships`;
DROP TABLE IF EXISTS `user_segments`;
DROP TABLE IF EXISTS `user_journeys`;
DROP TABLE IF EXISTS `user_events`;
DROP TABLE IF EXISTS `page_views`;
DROP TABLE IF EXISTS `user_sessions`;