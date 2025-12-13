DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM tenants
        WHERE id !~ '^([a-z0-9]{3}|[a-z0-9][a-z0-9-]{2,48}[a-z0-9])$'
    ) THEN
        RAISE EXCEPTION 'Migration blocked: Existing tenant IDs would violate new pattern. Please rename tenants that start or end with hyphens.';
    END IF;
END $$;

ALTER TABLE tenants DROP CONSTRAINT IF EXISTS chk_tenant_id;

ALTER TABLE tenants ADD CONSTRAINT chk_tenant_id
    CHECK (id ~ '^([a-z0-9]{3}|[a-z0-9][a-z0-9-]{2,48}[a-z0-9])$');

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_tenant_id'
        AND conrelid = 'tenants'::regclass
    ) THEN
        RAISE EXCEPTION 'Failed to create tenant_id constraint';
    END IF;
END $$;
