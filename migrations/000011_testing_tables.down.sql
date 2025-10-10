-- Drop testing related tables in reverse order of creation

-- Drop alert notifications table
DROP TABLE IF EXISTS alert_notifications;

-- Drop alerts table
DROP TABLE IF EXISTS alerts;

-- Drop system metrics table
DROP TABLE IF EXISTS system_metrics;

-- Drop analytics reports table
DROP TABLE IF EXISTS analytics_reports;

-- Drop user sessions table
DROP TABLE IF EXISTS user_sessions;

-- Drop user events table
DROP TABLE IF EXISTS user_events;

-- Drop test results table
DROP TABLE IF EXISTS test_results;

-- Drop test executions table
DROP TABLE IF EXISTS test_executions;

-- Drop test suites table
DROP TABLE IF EXISTS test_suites;