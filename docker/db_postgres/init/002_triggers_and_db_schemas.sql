-- ///////////////////
-- // Schemas
-- ///////////////////
CREATE SCHEMA IF NOT EXISTS auth;


-- ///////////////////
-- // Triggers
-- ///////////////////
CREATE OR REPLACE FUNCTION auth.update_current_timestamp()
    RETURNS trigger
    LANGUAGE plpgsql AS
$$
BEGIN
    IF NEW IS DISTINCT FROM OLD THEN
        NEW.updated := CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END
$$;
