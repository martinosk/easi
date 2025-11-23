CREATE TABLE IF NOT EXISTS view_preferences (
    tenant_id VARCHAR(50) NOT NULL,
    view_id VARCHAR(255) NOT NULL,
    edge_type VARCHAR(20),
    layout_direction VARCHAR(2),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, view_id),
    FOREIGN KEY (view_id) REFERENCES architecture_views(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_view_preferences_view_id ON view_preferences(view_id);

INSERT INTO view_preferences (tenant_id, view_id, edge_type, layout_direction, updated_at)
SELECT tenant_id, id, edge_type, layout_direction, CURRENT_TIMESTAMP
FROM architecture_views
WHERE edge_type IS NOT NULL OR layout_direction IS NOT NULL
ON CONFLICT (tenant_id, view_id) DO NOTHING;
