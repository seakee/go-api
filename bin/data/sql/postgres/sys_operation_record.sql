CREATE TABLE sys_operation_record
(
    id            bigserial PRIMARY KEY,
    ip            varchar(50)  DEFAULT NULL,
    method        varchar(10)  DEFAULT NULL,
    path          varchar(500) DEFAULT NULL,
    status        integer      DEFAULT NULL,
    latency       double precision DEFAULT NULL,
    agent         varchar(512) DEFAULT NULL,
    error_message text,
    user_id       bigint       DEFAULT NULL,
    params        text,
    resp          text,
    trace_id      varchar(64)  DEFAULT NULL,
    created_at    timestamp    DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp    DEFAULT CURRENT_TIMESTAMP,
    deleted_at    timestamp    DEFAULT NULL
);

CREATE INDEX idx_sys_operation_record_created_at ON sys_operation_record (created_at DESC);
CREATE INDEX idx_sys_operation_record_user_id ON sys_operation_record (user_id);
CREATE INDEX idx_sys_operation_record_trace_id ON sys_operation_record (trace_id);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sys_operation_record_set_updated_at
BEFORE UPDATE ON sys_operation_record
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
