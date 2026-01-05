ALTER TABLE architecture_views
  ADD COLUMN is_private BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN owner_user_id VARCHAR(255),
  ADD COLUMN owner_email VARCHAR(500);

CREATE INDEX idx_architecture_views_owner ON architecture_views(tenant_id, owner_user_id);
