DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'set_tenant_context') THEN
        GRANT EXECUTE ON FUNCTION set_tenant_context(VARCHAR) TO easi_app;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'get_current_tenant') THEN
        GRANT EXECUTE ON FUNCTION get_current_tenant() TO easi_app;
    END IF;
END
$$;
