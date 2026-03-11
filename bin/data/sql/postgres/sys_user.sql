CREATE TABLE sys_user
(
    id           bigserial PRIMARY KEY,
    email        varchar(200) DEFAULT NULL,
    phone        varchar(30)  DEFAULT NULL,
    password     varchar(255) DEFAULT NULL,
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

CREATE UNIQUE INDEX uk_sys_user_email ON sys_user (email);
CREATE UNIQUE INDEX uk_sys_user_phone ON sys_user (phone);

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
