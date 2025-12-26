ALTER TABLE capabilities ADD COLUMN maturity_value INT;

UPDATE capabilities SET maturity_value =
    CASE maturity_level
        WHEN 'Genesis' THEN 12
        WHEN 'Custom Build' THEN 37
        WHEN 'Product' THEN 62
        WHEN 'Commodity' THEN 87
        ELSE 12
    END;

ALTER TABLE capabilities ALTER COLUMN maturity_value SET NOT NULL;
ALTER TABLE capabilities ALTER COLUMN maturity_value SET DEFAULT 12;

CREATE INDEX IF NOT EXISTS idx_capabilities_maturity_value ON capabilities(tenant_id, maturity_value);
