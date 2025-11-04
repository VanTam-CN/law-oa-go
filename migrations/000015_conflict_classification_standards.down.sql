-- 回滚冲突分类标准数据模型

-- 删除视图
DROP VIEW IF EXISTS conflict_classification_stats;
DROP VIEW IF EXISTS conflict_classifications_active;

-- 删除表
DROP TABLE IF EXISTS conflict_classifications;