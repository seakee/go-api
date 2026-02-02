CREATE TABLE auth_app
(
    id           bigserial PRIMARY KEY,
    app_id       varchar(30)  NOT NULL,
    app_name     varchar(50)  DEFAULT NULL,
    app_secret   varchar(256) NOT NULL,
    redirect_uri varchar(500) DEFAULT NULL,
    description  text,
    status       smallint     NOT NULL DEFAULT 0,
    created_at   timestamp    DEFAULT CURRENT_TIMESTAMP,
    updated_at   timestamp    DEFAULT CURRENT_TIMESTAMP,
    deleted_at   timestamp    DEFAULT NULL
);

CREATE INDEX idx_auth_app_app_id ON auth_app (app_id);
CREATE INDEX idx_auth_app_app_name ON auth_app (app_name);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_auth_app_set_updated_at
BEFORE UPDATE ON auth_app
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
