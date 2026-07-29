-- Remove legacy organization-specific wording from denormalized public-demo
-- work items. Match stable fixture attributes instead of retaining private
-- names in repository source.

UPDATE inbox_items
SET title = '待办：示例科技证据材料补充',
    updated_at = CURRENT_TIMESTAMP
WHERE source_type = 'case'
  AND due_date_type = 'evidence'
  AND content = '请补齐服务合同、验收记录、付款流水和往来函件。'
  AND title <> '待办：示例科技证据材料补充';
