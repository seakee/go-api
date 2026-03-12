CREATE TABLE sys_user_identity
(
    id               bigserial PRIMARY KEY,
    user_id          bigint      NOT NULL,
    provider         varchar(50) NOT NULL,
    provider_tenant  varchar(200) NOT NULL,
    provider_subject varchar(200) NOT NULL,
    display_name     varchar(100) DEFAULT NULL,
    avatar_url       text,
    raw_profile_json text         DEFAULT NULL,
    bound_at         timestamp    DEFAULT CURRENT_TIMESTAMP,
    last_login_at    timestamp    DEFAULT NULL,
    created_at       timestamp    DEFAULT CURRENT_TIMESTAMP,
    updated_at       timestamp    DEFAULT CURRENT_TIMESTAMP,
    deleted_at       timestamp    DEFAULT NULL,
    CONSTRAINT fk_sys_user_identity_user_id FOREIGN KEY (user_id) REFERENCES sys_user (id)
);

CREATE UNIQUE INDEX uk_sys_user_identity_provider_subject
    ON sys_user_identity (provider, provider_tenant, provider_subject);

CREATE INDEX idx_sys_user_identity_user_id
    ON sys_user_identity (user_id);

CREATE TRIGGER trg_sys_user_identity_set_updated_at
BEFORE UPDATE ON sys_user_identity
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
