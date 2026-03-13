CREATE TABLE sys_user_passkey
(
    id                    bigserial PRIMARY KEY,
    user_id               bigint       NOT NULL,
    credential_id         varchar(512) NOT NULL,
    credential_public_key text         NOT NULL,
    sign_count            bigint       DEFAULT 0,
    aaguid                varchar(64)  DEFAULT NULL,
    transports_json       text         DEFAULT NULL,
    user_handle           varchar(255) DEFAULT NULL,
    display_name          varchar(100) DEFAULT NULL,
    last_used_at          timestamp    DEFAULT NULL,
    created_at            timestamp    DEFAULT CURRENT_TIMESTAMP,
    updated_at            timestamp    DEFAULT CURRENT_TIMESTAMP,
    deleted_at            timestamp    DEFAULT NULL
);

CREATE UNIQUE INDEX uk_sys_user_passkey_credential_id
    ON sys_user_passkey (credential_id);

CREATE INDEX idx_sys_user_passkey_user_id
    ON sys_user_passkey (user_id);

CREATE TRIGGER trg_sys_user_passkey_set_updated_at
BEFORE UPDATE ON sys_user_passkey
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
