CREATE TABLE sys_permission
(
    id          bigserial PRIMARY KEY,
    name        varchar(50) DEFAULT NULL,
    perm_type   varchar(10) DEFAULT NULL,
    method      varchar(10) DEFAULT NULL,
    path        text,
    description text,
    perm_group  text,
    created_at  timestamp   DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp   DEFAULT CURRENT_TIMESTAMP,
    deleted_at  timestamp   DEFAULT NULL
);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sys_permission_set_updated_at
BEFORE UPDATE ON sys_permission
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
