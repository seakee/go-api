CREATE TABLE sys_menu
(
    id            bigserial PRIMARY KEY,
    name          varchar(50) DEFAULT NULL,
    path          text,
    permission_id bigint      DEFAULT NULL,
    parent_id     bigint      DEFAULT NULL,
    icon          text,
    sort          integer     DEFAULT NULL,
    created_at    timestamp   DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp   DEFAULT CURRENT_TIMESTAMP,
    deleted_at    timestamp   DEFAULT NULL
);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sys_menu_set_updated_at
BEFORE UPDATE ON sys_menu
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
