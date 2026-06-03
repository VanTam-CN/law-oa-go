-- 000025_case_folders.up.sql
-- 案件文件夹（卷宗目录实例）表

CREATE TABLE IF NOT EXISTS case_folders (
    id SERIAL PRIMARY KEY,
    case_id INTEGER NOT NULL REFERENCES cases(id),
    parent_id INTEGER REFERENCES case_folders(id),
    name VARCHAR(200) NOT NULL,
    display_order INTEGER DEFAULT 0,
    description VARCHAR(500),
    template_id INTEGER REFERENCES case_folder_templates(id),
    template_path VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_case_folders_case ON case_folders(case_id);
CREATE INDEX IF NOT EXISTS idx_case_folders_parent ON case_folders(parent_id);

COMMENT ON TABLE case_folders IS '案件文件夹（卷宗目录实例）';
