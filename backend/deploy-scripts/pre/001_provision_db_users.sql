DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        CREATE USER easi_app WITH PASSWORD '${EASI_APP_PASSWORD}';
    ELSE
        ALTER USER easi_app WITH PASSWORD '${EASI_APP_PASSWORD}';
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        CREATE USER easi_admin WITH PASSWORD '${EASI_ADMIN_PASSWORD}' BYPASSRLS;
    ELSE
        ALTER USER easi_admin WITH PASSWORD '${EASI_ADMIN_PASSWORD}';
    END IF;
END
$$;

GRANT CONNECT ON DATABASE easi TO easi_app;
GRANT USAGE ON SCHEMA public TO easi_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO easi_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO easi_app;

GRANT ALL PRIVILEGES ON DATABASE easi TO easi_admin;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO easi_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO easi_admin;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO easi_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT USAGE, SELECT ON SEQUENCES TO easi_app;

DO $$
DECLARE
    schema_name TEXT;
BEGIN
    FOREACH schema_name IN ARRAY ARRAY[
        'infrastructure', 'shared', 'architecturemodeling', 'architectureviews',
        'capabilitymapping', 'enterprisearchitecture', 'viewlayouts', 'importing',
        'platform', 'auth', 'accessdelegation', 'metamodel', 'releases', 'valuestreams',
        'archassistant'
    ]
    LOOP
        EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);
        EXECUTE format('GRANT USAGE ON SCHEMA %I TO easi_app', schema_name);
        EXECUTE format('GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %I TO easi_app', schema_name);
        EXECUTE format('GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA %I TO easi_app', schema_name);

        EXECUTE format('GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA %I TO easi_admin', schema_name);
        EXECUTE format('GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA %I TO easi_admin', schema_name);

        EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO easi_app', schema_name);
        EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT USAGE, SELECT ON SEQUENCES TO easi_app', schema_name);
    END LOOP;
END
$$;
