CREATE TABLE sys_user
(
    id           bigserial PRIMARY KEY,
    account      varchar(200) DEFAULT NULL,
    password     char(32)     DEFAULT NULL,
    salt         char(32)     DEFAULT NULL,
    totp_key     char(32)     DEFAULT NULL,
    totp_enabled boolean      DEFAULT NULL,
    feishu_id    varchar(200) DEFAULT NULL,
    wechat_id    varchar(200) DEFAULT NULL,
    github_id    varchar(200) DEFAULT NULL,
    user_name    varchar(50)  DEFAULT NULL,
    status       smallint     DEFAULT NULL,
    avatar       text,
    created_at   timestamp    DEFAULT CURRENT_TIMESTAMP,
    updated_at   timestamp    DEFAULT CURRENT_TIMESTAMP,
    deleted_at   timestamp    DEFAULT NULL
);

CREATE UNIQUE INDEX idx_sys_user_account ON sys_user (account);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sys_user_set_updated_at
BEFORE UPDATE ON sys_user
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
