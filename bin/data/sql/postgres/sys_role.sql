CREATE TABLE sys_role
(
    id          bigserial PRIMARY KEY,
    name        varchar(50)  DEFAULT NULL,
    description varchar(100) DEFAULT NULL,
    created_at  timestamp    DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp    DEFAULT CURRENT_TIMESTAMP,
    deleted_at  timestamp    DEFAULT NULL
);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sys_role_set_updated_at
BEFORE UPDATE ON sys_role
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
